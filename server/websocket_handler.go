package server

import (
	"log"

	"github.com/nitrous-io/gorainbow/connection"
	"github.com/nitrous-io/gorainbow/sockjs"
)

func (s *Server) wsHandler(c sockjs.ReadWriteCloser) {
	socket := connection.NewConnection(c, s.verifier, s.db, s.mulx,
		s.Config.AESKey)
	if err := socket.Start(); err != nil {
		log.Printf("Websocket connection: ended %v", err)
		socket.Close()
		c.Close()
	}
}
