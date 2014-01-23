package main

import (
	"github.com/action-io/gorainbow"
	"github.com/igm/sockjs-go/sockjs"
	"log"
	"net/http"
)

func main() {
	port := "3001"
	config := sockjs.DefaultConfig
	config.DecodeFrames = true
	sockjs.Install("/ws", gorainbow.ServerHandler, config)
	err := http.ListenAndServe(":"+port, nil)
	log.Fatal(err)
}
