package sockjs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// time allowed to write a message to an user
	writeWait = 5 * time.Second

	maxQueue = 50 // queue up to 50 messages per connection
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// allow all origins here
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// interfaces that encapsulates the bahavior of a sockjs socket
type (
	Reader interface {
		Read() ([]byte, error)
	}

	Writer interface {
		Write(a []interface{})
	}

	Closer interface {
		Close()
	}

	ReadWriteCloser interface {
		Reader
		Writer
		Closer
	}
)

type Connection struct {
	ws        *websocket.Conn // underlying websocket connection
	closing   chan struct{}   // notify that the socket should be closing soon
	done      chan struct{}   // notify the caller of socket's termination
	send      chan []byte     // buffered channel of outbound messages
	heartbeat time.Duration   // how often will heartbeat msg be sent
	closed    bool
}

func (c *Connection) Read() ([]byte, error) {
	_, p, err := c.ws.ReadMessage()
	if err != nil {
		log.Println("[ERROR] Failed to read msg", string(p), err)
		c.Close()
	}
	return p, err
}

// Array of json-encoded messages. For example: a["message"]
func (c *Connection) Write(a []interface{}) {
	msg, err := json.Marshal(a)
	if err != nil {
		c.Close()
	}
	log.Printf("[MSG] message: %v", string(msg))
	c.send <- []byte(fmt.Sprintf("a[%s]", strconv.Quote(string(msg))))
}

func (c *Connection) Close() {
	if !c.closed {
		c.closed = true
		// signal that the connection is closing soon
		close(c.closing)
	}
}

// Open frame. Every time a new session is established, the server must
// immediately send the open frame
func (c *Connection) start() {
	c.send <- []byte("o")
}

// Heartbeat frame. Most loadbalancers have arbitrary timeouts on connections.
// In order to keep connections from breaking, the server must send a heartbeat
// frame every now and then. The typical delay is 25 seconds and should be
// configurable.
func (c *Connection) sendHeartbeat() {
	c.send <- []byte("h")
}

// Close frame. This frame is send to the browser every time the client asks
// for data on closed connection. This may happen multiple times.
// Close frame contains a code and a string explaining a reason of closure,
// like: c[3000,"Go away!"].]
func (c *Connection) end() {
	log.Println("end")
	c.write([]byte(`c[3000,"Go away!"]`))
	c.ws.Close()
	close(c.done)
}

func (c *Connection) write(msg []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(websocket.TextMessage, msg)
}

func (c *Connection) deliver() {
	ticker := time.NewTicker(c.heartbeat)
	defer ticker.Stop()
	c.start()
	defer c.end()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.Close()
			}
			if err := c.write(msg); err != nil {
				c.Close()
			}
		case <-ticker.C:
			c.sendHeartbeat()
		case <-c.closing:
			return
		}
	}

}

func newConnection(w http.ResponseWriter, r *http.Request, done chan struct{},
	heartbeat time.Duration) *Connection {
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
	c := &Connection{
		ws:        ws,
		done:      done,
		send:      make(chan []byte, maxQueue),
		closing:   make(chan struct{}, 1),
		heartbeat: heartbeat,
	}
	// start messages delivery
	go c.deliver()
	return c
}
