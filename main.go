package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct {
	conns map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]bool),
	}
}

// handleWSFeed sends data to users subscribed to the feed
// every 2 seconds a message is broadcasted
func (s *Server) handleWSFeed(ws *websocket.Conn) {
	fmt.Println("new incoming connection from client to feed:", ws.RemoteAddr())

	for {
		payload := fmt.Sprintf("feed data -> %d\n", time.Now().UnixNano())
		ws.Write([]byte(payload))
		time.Sleep(time.Second * 2)
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	fmt.Println("new incoming connection from client: ", ws.RemoteAddr())

	// mutex in prod
	s.conns[ws] = true
	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)

	for {
		n, err := ws.Read(buf)
		if err != nil {
			// conn on other side closed
			if err == io.EOF {
				break
			}

			fmt.Println("read error: ", err)
			continue
		}

		msg := buf[:n]
		s.broadcast(msg)
	}
}

// broadcast loops through the connections and sends the bytes
func (s *Server) broadcast(b []byte) {
	for ws := range s.conns {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				fmt.Println("write error: ", err)
			}
		}(ws)
	}
}

func main() {

	server := NewServer()
	http.Handle("/ws", websocket.Handler(server.handleWS))
	http.Handle("/feed", websocket.Handler(server.handleWSFeed))
	http.ListenAndServe(":3000", nil)
}
