package sockjs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type connection struct {
	ws      *websocket.Conn
	closeCh chan struct{}
}

// Open frame. Every time a new session is established, the server must
// immediately send the open frame
func (c *connection) Start() {
	err := c.ws.WriteMessage(websocket.TextMessage, []byte("o"))
	if err != nil {
		close(c.closeCh)
	}
}

func (c *connection) Read() ([]byte, error) {
	_, p, err := c.ws.ReadMessage()
	return p, err
}

// Array of json-encoded messages. For example: a["message"]
func (c *connection) Write(a []interface{}) error {
	msg, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("Invalid response %v", a)
	}
	return c.ws.WriteMessage(websocket.TextMessage, []byte("a"+string(msg)))
}

// Heartbeat frame. Most loadbalancers have arbitrary timeouts on connections.
// In order to keep connections from breaking, the server must send a heartbeat
// frame every now and then. The typical delay is 25 seconds and should be
// configurable.
func (c *connection) Heartbeat() {
	err := c.ws.WriteMessage(websocket.TextMessage, []byte("h"))
	if err != nil {
		close(c.closeCh)
	}
}

func newConnection(w http.ResponseWriter, r *http.Request, closeCh chan struct{}) *connection {
	// check if we can ugprade the connection to websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			http.Error(w, `Can "Upgrade" only to "WebSocket".`, http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return nil
	}
	return &connection{ws, closeCh}
}

// Close frame. This frame is send to the browser every time the client asks
// for data on closed connection. This may happen multiple times.
// Close frame contains a code and a string explaining a reason of closure,
// like: c[3000,"Go away!"].]
func (c *connection) Close() {
	c.ws.WriteMessage(websocket.TextMessage, []byte(`c[3000,"Go away!"]`))
	c.ws.Close()
}

type SocketHandler func(msg []byte) []interface{}

type Socket struct {
	conn             *connection
	sessionId        string
	serverId         string
	handler          SocketHandler
	heartbeatTimeout time.Duration
	closeCh          chan struct{}
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

	s.closeCh = make(chan struct{}, 1)
	// check if we can ugprade the connection to websocket
	s.conn = newConnection(w, r, s.closeCh)
	if s.conn == nil {
		return
	}
	defer s.Close()

	// heartbeat
	go s.heartbeat()

	// send and receive message
	go s.handleMessage()

	// wait til the connection is closed
	<-s.closeCh
	log.Println("websocket %v:%v closed", s.serverId, s.sessionId)
}

// - read msg from client
// - pass the message to handler func
// - write the response back to the client
//
// if there's any error, we close everything
func (s *Socket) handleMessage() {
	s.conn.Start()
	for {
		var p []byte
		if p, err := s.conn.Read(); err != nil {
			log.Println("[ERROR] Failed to read msg", string(p), err)
			close(s.closeCh)
			return
		}
		resp := s.handler(p)
		if err := s.conn.Write(resp); err != nil {
			log.Println("[ERROR] Failed to write msg", resp, err)
			close(s.closeCh)
			return
		}
	}
}

// send heartbeat message every `heartbeatTimeout` duration. We stop sending
// heartbeat when the connection is closed
func (s *Socket) heartbeat() {
	ticker := time.NewTicker(s.heartbeatTimeout)
	defer ticker.Stop()
	for {
		select {
		case <-s.closeCh:
			return
		case <-ticker.C:
			s.conn.Heartbeat()
		}
	}
}

func (s *Socket) Close() {
	s.conn.Close()
}

func makeSocketHandler(handler SocketHandler, heartbeatTimeout time.Duration) httpFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		socket := &Socket{
			handler:          handler,
			heartbeatTimeout: heartbeatTimeout,
		}
		socket.serveWs(w, r)
	}
}
