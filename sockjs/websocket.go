package sockjs

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type SocketHandler func(c ReadWriteCloser)

type Socket struct {
	conn      *Connection
	sessionId string
	serverId  string
	handler   SocketHandler
	heartbeat time.Duration
}

func (s *Socket) serveWs(w http.ResponseWriter, r *http.Request) {
	// get sessionId and serverId from URL
	vars := mux.Vars(r)
	s.serverId = vars["serverId"]
	s.sessionId = vars["sessionId"]

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
}

// - pass the socket to client so it can handle its own stuffs
func (s *Socket) handleMessage() {
	s.handler(s.conn)
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
