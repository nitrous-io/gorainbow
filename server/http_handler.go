package server

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	defaultUploadMemory     = 20 << 20 // 20 MB
	tmpDir                  = "/tmp"
	httpUnprocessableEntity = 422
)

type jsonResponse struct {
	Tokens []string `json:"tokens,omitempty"`
	Error  string   `json:"error,omitempty"`
}

type uploadToken struct {
	Name string `json:"fileName"`
	Path string `json:"tmpFilePath"`
}

type HttpHandler func(w http.ResponseWriter, r *http.Request) error

func (s *Server) getHome(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) getDownload(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Server) optionsUpload(w http.ResponseWriter, r *http.Request) error {
	s.addNoCacheHeader(w)
	s.addCorsHeader(w)
	w.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) postUpload(w http.ResponseWriter, r *http.Request) error {
	s.addNoCacheHeader(w)
	s.addCorsHeader(w)
	userId, err := s.verifier.Verify(r.FormValue("auth_token"), time.Now())
	if err != nil || userId == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		resp := &jsonResponse{
			Error: "You need to sign in or sign up before continuing.",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			return err
		}
		return nil
	}
	if err := r.ParseMultipartForm(defaultUploadMemory); err != nil {
		return err
	}

	// process files and generate tokens
	uploadTokens, err := uploadToTokens(r.MultipartForm.File)
	if err != nil {
		return err
	}

	// for each token, we sign it with `Server.signer`
	tokens := []string{}
	for _, ut := range uploadTokens {
		data, err := json.Marshal(ut)
		if err != nil {
			return err
		}
		tokens = append(tokens, s.signer.Generate(data))
	}

	if len(tokens) > 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(httpUnprocessableEntity)
	}

	resp := &jsonResponse{Tokens: tokens}
	return json.NewEncoder(w).Encode(resp)
}

func (s *Server) addNoCacheHeader(w http.ResponseWriter) {
	header := map[string]string{
		"Pragma":        "no-cache",
		"Cache-Control": "no-store, no-cache, must-revalidate",
	}
	for key, val := range header {
		w.Header().Set(key, val)
	}
}

func (s *Server) addCorsHeader(w http.ResponseWriter) {
	header := map[string]string{
		"Access-Control-Allow-Origin":      s.Config.CorsOrigin,
		"Access-Control-Allow-Methods":     "*",
		"Access-Control-Allow-Headers":     "*",
		"Access-Control-Allow-Credentials": "true",
	}
	for key, val := range header {
		w.Header().Set(key, val)
	}
}

func makeHttpHandler(handlerFunc HttpHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handlerFunc(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

func uploadToTokens(files map[string][]*multipart.FileHeader) ([]*uploadToken, error) {
	var tokens []*uploadToken
	var err error
	for fileName, fhs := range files {
		if len(fhs) > 0 {
			var tf *os.File
			var uf multipart.File
			if tf, err = ioutil.TempFile(tmpDir, fileName); err != nil {
				return nil, err
			}
			if uf, err = fhs[0].Open(); err != nil {
				return nil, err
			}
			io.Copy(tf, uf)
			tokens = append(tokens, &uploadToken{Name: fileName, Path: tf.Name()})
		}
	}
	return tokens, nil
}
