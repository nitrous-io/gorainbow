package sockjs

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

const MaxAge = 365 * 24 * 60 * 60

type info struct {
	Websocket    bool     `json:"websocket"`
	CookieNeeded bool     `json:"cookie_needed"`
	Origins      []string `json:"origins"`
	Entropy      int32    `json:"entropy"`
}

func infoHeader(w http.ResponseWriter, r *http.Request, next httpFunc) {
	origin := r.Header.Get("origin")
	if origin == "" || origin == "null" {
		origin = "*"
	}
	header := map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Cache-Control":                    "no-store, no-cache, must-revalidate, max-age=0",
		"Access-Control-Allow-Origin":      origin,
	}
	if allowHeaders := r.Header.Get("Access-Control-Request-Headers"); allowHeaders != "" && allowHeaders != "null" {
		header["Access-Control-Allow-Headers"] = allowHeaders
	}
	for k, v := range header {
		w.Header().Set(k, v)
	}

	next(w, r)
}

func getInfo(w http.ResponseWriter, r *http.Request) {
	infoHeader(w, r, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode(info{
			Websocket:    true,
			CookieNeeded: false,
			Origins:      []string{"*:*"},
			Entropy:      rand.Int31(),
		})
	})
}

func optionsInfo(w http.ResponseWriter, r *http.Request) {
	infoHeader(w, r, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET")
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", MaxAge))
		w.WriteHeader(http.StatusNoContent)
	})
}
