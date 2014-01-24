package gorainbow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/igm/sockjs-go/sockjs"
	"io"
	"log"
)

type (
	Connection struct {
		Conn    sockjs.Conn
		Sockets map[string]socketHandler
	}

	Message struct {
		Namespace  string
		Event      string
		CallbackId int
		Args       []interface{}
		NoCallback bool
	}
)

func (m Message) Buffer() *bytes.Buffer {
	b, _ := m.Args[0].([]byte)
	return bytes.NewBuffer(b)
}

// encode the message into format suitable to transmit as a sockjs message
func (m Message) Encode() []byte {
	e := []interface{}{}
	e = append(e, m.Namespace)
	e = append(e, m.Event)
	if m.NoCallback {
		e = append(e, m.Args)
	} else {
		e = append(e, []interface{}{m.CallbackId, m.Args})
	}
	b, _ := json.Marshal(e)
	return b
}

func (m Message) EncodeBuffer() *bytes.Buffer {
	return bytes.NewBuffer(m.Encode())
}

func (c *Connection) Process() error {
	data, err := c.Conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("Cant read msg %v", err)
	}
	if len(data) == 0 {
		return nil
	}
	msg, _ := parseMsg(data)
	// handle authentication
	if msg.Namespace == "global" && msg.Event == "auth" {
		res := Message{Namespace: "global", Event: "callback",
			CallbackId: msg.CallbackId, Args: msg.Args}
		go c.Write(res.Encode())
	} else if msg.Event == "init" {
		// handle init of services
		socket := NewSocket(msg.Namespace, c)
		if socket != nil {
			c.Sockets[msg.Namespace] = socket
			res := Message{Namespace: "global", Event: "callback",
				CallbackId: msg.CallbackId, Args: msg.Args}
			go c.Write(res.Encode())
		} else {
			return fmt.Errorf("Invalid socket type %v", msg.Namespace)
		}
	} else {
		socket, ok := c.Sockets[msg.Namespace]
		if !ok {
			return fmt.Errorf("Invalid socket type %v", msg.Namespace)
		}
		socket.OnMessage(msg)
	}
	return nil
}

func (c *Connection) Write(b []byte) (int, error) {
	return c.Conn.WriteMessage(b)
}

func (c *Connection) ReadFrom(src io.Reader) (int64, error) {
	b := bytes.NewBuffer([]byte{})
	i, err := b.ReadFrom(src)
	c.Conn.WriteMessage(b.Bytes())
	return i, err
}

func (c *Connection) Close() error {
	log.Println("close conn")
	return c.Conn.Close()
}

// FIXME: check if parsing msg actually success
func parseMsg(data []byte) (Message, error) {
	msg := []interface{}{}
	_ = json.Unmarshal(data, &msg)
	namespace := msg[0].(string)
	event := msg[1].(string)
	args := msg[2].([]interface{})
	callback := 0
	if len(msg) > 3 {
		callback = int(msg[3].(float64))
	}
	return Message{Namespace: namespace, Event: event, CallbackId: callback,
		Args: args}, nil
}
