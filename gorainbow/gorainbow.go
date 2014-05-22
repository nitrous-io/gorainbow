package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/nitrous-io/gorainbow"
)

var version string

func main() {
	flConfig := flag.String("config", "", "path to the config file")
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
	server, err := gorainbow.New(configData)
	if err != nil {
		log.Printf("gorainbow: failed to create server %v", err)
		os.Exit(1)
	}
	log.Println(server.Router)
	http.Handle("/", server.Router)
	log.Printf("gorainbow: starting server at %v", server.Config.Address)
	if err := http.ListenAndServe(server.Config.Address, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
