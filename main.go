package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	socketio "github.com/googollee/go-socket.io"
	y3 "github.com/yomorun/y3-codec-golang"
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
	socketioRoom = "yomo-demo"
	socketioAddr = "0.0.0.0:8000"
	zipperAddr   = "localhost:9000"
	serviceName  = "Noise"
)

var (
	socketioServer *socketio.Server
	err            error
)

func main() {
	socketioServer, err = newSocketIOServer()
	if err != nil {
		log.Printf("❌ Initialize the socket.io server failure with err: %v", err)
		return
	}

	// sink server which will receive the data from `yomo-zipper`.
	go connectToZipper(zipperAddr)

	// serve socket.io server.
	go socketioServer.Serve()
	defer socketioServer.Close()

	router := gin.New()
	router.Use(ginMiddleware())
	router.GET("/socket.io/*any", gin.WrapH(socketioServer))
	router.POST("/socket.io/*any", gin.WrapH(socketioServer))
	router.StaticFS("/public", http.Dir("./asset"))
	router.Run(socketioAddr)

	log.Print("✅ Serving socket.io on ", socketioAddr)
	err = http.ListenAndServe(socketioAddr, nil)
	if err != nil {
		log.Printf("❌ Serving the socket.io server on %s failure with err: %v", socketioAddr, err)
		return
	}
}

func newSocketIOServer() (*socketio.Server, error) {
	log.Print("Starting socket.io server...")
	server, err := socketio.NewServer(nil)
	if err != nil {
		return nil, err
	}

	// add all connected user to the room "yomo-demo".
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Print("connected:", s.ID())
		s.Join(socketioRoom)

		return nil
	})

	return server, nil
}

func ginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestOrigin := c.Request.Header.Get("Origin")

		c.Writer.Header().Set("Access-Control-Allow-Origin", requestOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Request.Header.Del("Origin")

		c.Next()
	}
}

// connectToZipper connects to `yomo-zipper` and receives the real-time data.
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

// Handler receives the real-time data and broadcast it to socket.io clients.
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

// broadcastData broadcasts the real-time data to socket.io clients.
func broadcastData(_ context.Context, data interface{}) (interface{}, error) {
	if socketioServer != nil && data != nil {
		socketioServer.BroadcastToRoom("", socketioRoom, "receive_sink", data)
	}

	return data, nil
}
