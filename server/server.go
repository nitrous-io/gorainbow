package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/nitrous-io/gorainbow/multiplexer"
	"github.com/nitrous-io/gorainbow/sockjs"
	"github.com/nitrous-io/gorainbow/verifier"
)

type (
	config struct {
		Address     string `json:"address"`
		AESKey      string `json:"aes_key"`
		SecretToken string `json:"secret_token"`
		DatabaseUrl string `json:"database_url"`
	}

	Server struct {
		Config *config
		Router *mux.Router

		db       *sql.DB
		mulx     *multiplexer.Multiplexer
		verifier verifier.UserVerifier
		signer   verifier.SignatureVerifier
	}
)

const (
	heartbeat = 2 * time.Second
)

func (s *Server) CreateRouter() *mux.Router {
	r := mux.NewRouter()

	// setup http handler
	m := map[string]map[string]HttpHandler{
		"GET": {
			"/":         s.getHome,
			"/download": s.getDownload,
		},
		"OPTIONS": {
			"/upload": s.optionsUpload,
		},
		"POST": {
			"/upload": s.postUpload,
		},
	}
	for method, routes := range m {
		for route, handler := range routes {
			localRoute := route
			localHandler := handler
			localMethod := method
			r.HandleFunc(localRoute, makeHttpHandler(localHandler)).Methods(localMethod)
		}
	}

	// setup websocket handler
	wsRouter := r.PathPrefix("/ws").Subrouter()
	sockjs.New(wsRouter, s.wsHandler, heartbeat)

	return r
}

func New(configData []byte) (*Server, error) {
	config := &config{}
	if err := json.Unmarshal(configData, config); err != nil {
		log.Printf("gorainbow: invalid config data %v", string(configData))
		return nil, err
	}
	// get database connection
	db, err := sql.Open("postgres", config.DatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("Failed to open db connection %v", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to validate db connection %v", err)
	}
	// get signer
	signer := verifier.NewTokenVerifier(config.SecretToken)
	// get verifier
	v := verifier.NewAuthVerifier("rainbow", signer)
	// get multiplexer
	mulx := multiplexer.NewMultiplexer()

	server := &Server{
		Config:   config,
		db:       db,
		verifier: v,
		mulx:     mulx,
		signer:   signer,
	}
	server.Router = server.CreateRouter()
	return server, nil
}
