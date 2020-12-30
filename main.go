package main

import (
	"context"
	"log"
	"net/http"

	socketio "github.com/googollee/go-socket.io"
	"github.com/yomorun/yomo/pkg/quic"
	"github.com/yomorun/yomo/pkg/rx"
)

const (
	socketioRoom   = "yomo-demo"
	socketioAddr   = "0.0.0.0:8000"
	sinkServerAddr = "0.0.0.0:4141"
)

func main() {
	socketioServer, err := newSocketIOServer()
	if err != nil {
		log.Printf("❌ Initialize the socket.io server failure with err: %v", err)
		return
	}

	// sink server which will receive the data from `yomo-sink`.
	go serveSinkServer(socketioServer, sinkServerAddr)

	// serve socket.io server.
	go socketioServer.Serve()
	defer socketioServer.Close()

	http.Handle("/socket.io/", socketioServer)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
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

// serveSinkServer serves the Sink server over QUIC.
func serveSinkServer(socketioServer *socketio.Server, addr string) {
	log.Print("Starting sink server...")
	handler := &quicServerHandler{
		socketioServer,
	}
	quicServer := quic.NewServer(handler)

	err := quicServer.ListenAndServe(context.Background(), addr)
	if err != nil {
		log.Printf("❌ Serve the sink server on %s failure with err: %v", addr, err)
	}
}

type quicServerHandler struct {
	socketioServer *socketio.Server
}

func (s *quicServerHandler) Listen() error {
	// you can add the customized codes which will be triggered when QUIC server is listening.
	return nil
}

func (s *quicServerHandler) Read(st quic.Stream) error {
	// receive the data from `yomo-flow` and use rx (ReactiveX) to process the stream.
	rxStream := rx.FromReader(st).
		Y3Decoder("0x10", float32(0)). // decode the data via Y3 Codec.
		StdOut()

	go func() {
		for customer := range rxStream.Observe() {
			if customer.Error() {
				log.Print(customer.E.Error())
			} else if customer.V != nil {
				// broadcast message to all connected user.
				s.socketioServer.BroadcastToRoom("", socketioRoom, "receive_sink", customer.V)
			}
			st.Write([]byte("Received data."))
		}
	}()

	return nil
}
