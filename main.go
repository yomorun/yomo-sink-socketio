package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	y3 "github.com/yomorun/y3-codec-golang"
	"github.com/yomorun/yomo-sink-socketio-server-example/ws"
	"github.com/yomorun/yomo/pkg/client"
	"github.com/yomorun/yomo/pkg/rx"
)

// noiseDataKey represents the Tag of a Y3 encoded data packet.
const noiseDataKey = 0x10

type noiseData struct {
	Noise float32 `y3:"0x11" json:"noise"` // Noise value
	Time  int64   `y3:"0x12" json:"time"`  // Timestamp (ms)
	From  string  `y3:"0x13" json:"from"`  // Source IP
}

const (
	wsAddr      = "0.0.0.0:8003"
	zipperAddr  = "localhost:9000"
	serviceName = "Noise"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var hub *ws.Hub

func main() {
	// connect to `yomo-zipper`.
	go connectToZipper(zipperAddr)

	// WebSocket server
	hub = ws.NewHub()
	go hub.Run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, w, r)
	})
	log.Print("✅ Serving WebSocket on ", wsAddr)
	log.Fatal(http.ListenAndServe(wsAddr, nil))
}

// serveWS handles websocket requests from the peer.
func serveWS(hub *ws.Hub, w http.ResponseWriter, r *http.Request) {
	// CORS
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// always return true when checking origin
		return true
	}

	// represents a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &ws.Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256)}
	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
}

func broadcastData(_ context.Context, data interface{}) (interface{}, error) {
	if hub != nil && data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
		} else {
			hub.Broadcast <- b
		}
	}

	return data, nil
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "asset/index.html")
}

func connectToZipper(zipperAddr string) error {
	urls := strings.Split(zipperAddr, ":")
	if len(urls) != 2 {
		return fmt.Errorf(`❌ The format of url "%s" is incorrect, it should be "host:port", f.e. localhost:9000`, zipperAddr)
	}
	host := urls[0]
	port, _ := strconv.Atoi(urls[1])
	cli, err := client.NewServerless(serviceName).Connect(host, port)
	if err != nil {
		return err
	}
	defer cli.Close()

	cli.Pipe(Handler)
	return nil
}

func Handler(rxstream rx.RxStream) rx.RxStream {
	stream := rxstream.
		Subscribe(noiseDataKey).
		OnObserve(decode).
		Map(broadcastData).
		Encode(noiseDataKey)

	return stream
}

func decode(v []byte) (interface{}, error) {
	// decode the data via Y3 Codec.
	data := noiseData{}
	err := y3.ToObject(v, &data)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return data, nil
}
