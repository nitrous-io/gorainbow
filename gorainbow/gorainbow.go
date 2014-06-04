package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/nitrous-io/gorainbow/server"
)

var version string

func main() {
	flConfig := flag.String("c", "config.json", "path to the config file")
	flVersion := flag.Bool("v", false, "current version")
	flag.Parse()
	if *flVersion {
		fmt.Printf("Gorainbow: revision %s\n", version)
		os.Exit(0)
		return
	}
	if len(*flConfig) == 0 {
		log.Printf("gorainbow: missing config file")
		os.Exit(1)
	}
	configData, err := ioutil.ReadFile(*flConfig)
	if err != nil {
		log.Printf("gorainbow: failed to read config file")
		os.Exit(1)
	}
	// start the server
	srv, err := server.New(configData)
	if err != nil {
		log.Printf("gorainbow: failed to create server %v", err)
		os.Exit(1)
	}
	http.Handle("/", srv.Router)
	log.Printf("gorainbow: starting server at %v", srv.Config.Address)
	log.Fatal("ListenAndServe: ", http.ListenAndServe(srv.Config.Address, nil))
}
