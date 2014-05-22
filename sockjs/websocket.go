package sockjs

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type SocketHandler func(msg []byte) []interface{}

type Socket struct {
	conn      *connection
	sessionId string
	serverId  string
	handler   SocketHandler
	heartbeat time.Duration
}

func (s *Socket) serveWs(w http.ResponseWriter, r *http.Request) {
	// check origin header
	origin := r.Header.Get("Origin")
	if origin != "http://"+r.Host && origin != "https://"+r.Host {
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}
	// get sessionId and serverId from URL
	vars := mux.Vars(r)
	s.serverId = vars["serverId"]
	s.sessionId = vars["sessionId"]
	log.Println("Accepted connection %v:%v", s.serverId, s.sessionId)

	done := make(chan struct{}, 1)
	// check if we can ugprade the connection to websocket
	s.conn = newConnection(w, r, done, s.heartbeat)
	if s.conn == nil {
		return
	}

	// send and receive message
	go s.handleMessage()

	// wait til the connection is closed
	<-done
	log.Println("websocket %v:%v closed", s.serverId, s.sessionId)
}

// - read msg from client
// - pass the message to handler func
// - write the response back to the client
//
// if there's any error, we close everything
func (s *Socket) handleMessage() {
	for {
		// if we cant read the message, just break the routine
		p, err := s.conn.Read()
		if err != nil {
			return
		}
		resp := s.handler(p)
		s.conn.Write(resp)
	}
}

func makeSocketHandler(handler SocketHandler, heartbeat time.Duration) httpFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		socket := &Socket{
			handler:   handler,
			heartbeat: heartbeat,
		}
		socket.serveWs(w, r)
	}
}
