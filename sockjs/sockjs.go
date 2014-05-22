package sockjs

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type httpFunc func(w http.ResponseWriter, r *http.Request)

// accept a mux router and inject necessary routes for sock.js implementation
func New(router *mux.Router, handler SocketHandler, heartbeatTimeout time.Duration) {
	router.HandleFunc("/info", getInfo).Methods("GET")
	router.HandleFunc("/info", optionsInfo).Methods("OPTIONS")
	router.HandleFunc("/{serverId:[0-9]+}/{sessionId}/websocket",
		makeSocketHandler(handler, heartbeatTimeout))
}
