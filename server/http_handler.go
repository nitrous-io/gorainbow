package server

import "net/http"

type HttpHandler func(w http.ResponseWriter, r *http.Request) error

func (s *Server) getHome(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) getDownload(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) optionsUpload(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) postUpload(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func makeHttpHandler(handlerFunc HttpHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handlerFunc(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		w.Write([]byte("hello world"))
	}
}
