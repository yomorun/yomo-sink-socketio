package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	y3 "github.com/yomorun/y3-codec-golang"
	"github.com/yomorun/yomo-sink-socketio-server-example/ws"
	"github.com/yomorun/yomo/pkg/quic"
)

type noiseData struct {
	Noise float32 `y3:"0x11" json:"noise"` // Noise value
	Time  int64   `y3:"0x12" json:"time"`  // Timestamp (ms)
	From  string  `y3:"0x13" json:"from"`  // Source IP
}

const (
	wsAddr         = "0.0.0.0:8003"
	sinkServerAddr = "0.0.0.0:4141"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var hub *ws.Hub

func main() {
	// sink server which will receive the data from `yomo-zipper`.
	go serveSinkServer(sinkServerAddr)

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

func broadcastData(data interface{}) {
	if hub == nil || data == nil {
		return
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
	} else {
		hub.Broadcast <- b
	}
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

// serveSinkServer serves the Sink server over QUIC.
func serveSinkServer(addr string) {
	log.Print("Starting sink server...")
	handler := &quicServerHandler{}
	quicServer := quic.NewServer(handler)

	err := quicServer.ListenAndServe(context.Background(), addr)
	if err != nil {
		log.Printf("❌ Serve the sink server on %s failure with err: %v", addr, err)
	}
}

type quicServerHandler struct {
}

func (s *quicServerHandler) Listen() error {
	// you can add the customized codes which will be triggered when QUIC server is listening.
	return nil
}

func (s *quicServerHandler) Read(st quic.Stream) error {
	// decode the data via Y3 Codec.
	ch := y3.
		FromStream(st).
		Subscribe(0x10).
		OnObserve(onObserve)

	go func() {
		for item := range ch {
			// broadcast message to all connected user.
			broadcastData(item)
		}
	}()

	return nil
}

func onObserve(v []byte) (interface{}, error) {
	// decode the data via Y3 Codec.
	data := noiseData{}
	err := y3.ToObject(v, &data)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return data, nil
}
