package gorainbow

import (
	"github.com/igm/sockjs-go/sockjs"
	"log"
)

func ServerHandler(socket sockjs.Conn) {
	sockets := make(map[string]socketHandler)
	conn := &Connection{Conn: socket, Sockets: sockets}
	defer conn.Close()
	for {
		if err := conn.Process(); err != nil {
			log.Println("ERROR!!", err)
			return
		}
	}
}
