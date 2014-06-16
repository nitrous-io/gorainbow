package server

import (
	"log"

	"github.com/nitrous-io/gorainbow/connection"
	"github.com/nitrous-io/gorainbow/sockjs"
)

func (s *Server) wsHandler(c sockjs.ReadWriteCloser) {
	env := map[string]string{
		"downloadDir": s.Config.DownloadDir,
	}
	socket := connection.NewConnection(c, s.verifier, s.signer, s.db, s.mulx,
		s.Config.AESKey, env)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Websocket: PANIC %+v", r)
		}
		socket.Close()
		c.Close()
	}()
	if err := socket.Start(); err != nil {
		log.Printf("Websocket connection: ended %v", err)
	}
}
