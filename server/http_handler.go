package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nitrous-io/gorainbow/models"
	"github.com/nitrous-io/gorainbow/verifier"
)

const (
	defaultUploadMemory     = 20 << 20 // 20 MB
	tmpDir                  = "/tmp"
	httpUnprocessableEntity = 422
	badRequest              = `{"error":"bad request"}`
	badAuth                 = "You need to sign in or sign up before continuing."
	badBoxIdPath            = "Box id and path must be provided."
	badBoxId                = "Invalid box id."
	badAccess               = "You don;t have access to the given box"
	homeResp                = "rainbow ok"
)

type intKey int

type jsonResponse struct {
	Tokens []string `json:"tokens,omitempty"`
	Error  string   `json:"error,omitempty"`
}

type uploadToken struct {
	Name string `json:"fileName"`
	Path string `json:"tmpFilePath"`
}

func (s *Server) getHome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(homeResp))
}

func (s *Server) getDownload(w http.ResponseWriter, r *http.Request) {
	s.addCorsHeader(w)
	userId := auth(s.verifier, w, r)
	if userId == 0 {
		return
	}
	boxIdParams := r.FormValue("id")
	localPath := r.FormValue("path")
	if len(boxIdParams) == 0 || len(localPath) == 0 {
		replyError(w, badBoxIdPath, http.StatusBadRequest)
		return
	}
	boxId, err := strconv.Atoi(boxIdParams)
	if err != nil {
		replyError(w, badBoxId, http.StatusBadRequest)
		return
	}
	accessible, err := models.IsBoxAccessible(s.db, boxId, userId)
	if err != nil || !accessible {
		replyError(w, badAccess, http.StatusForbidden)
		return
	}
	localPath = path.Join(s.Config.DownloadDir, localPath)
	lf, err := os.Open(localPath)
	if err != nil {
		replyBadRequest(w)
		return
	}
	defer func() {
		lf.Close()
		os.Remove(localPath)
		os.Remove(filepath.Dir(localPath))
	}()

	fileName := filepath.Base(localPath)
	w.Header().Set("Content-Disposition", contentDisposition(fileName))
	http.ServeContent(w, r, fileName, time.Now(), lf)
}

func (s *Server) optionsUpload(w http.ResponseWriter, r *http.Request) {
	s.addNoCacheHeader(w)
	s.addCorsHeader(w)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) postUpload(w http.ResponseWriter, r *http.Request) {
	s.addNoCacheHeader(w)
	s.addCorsHeader(w)
	userId := auth(s.verifier, w, r)
	if userId == 0 {
		return
	}
	if err := r.ParseMultipartForm(defaultUploadMemory); err != nil {
		replyBadRequest(w)
		return
	}

	// process files and generate tokens
	uploadTokens, err := uploadToTokens(r.MultipartForm.File)
	if err != nil {
		replyBadRequest(w)
		return
	}

	// for each token, we sign it with `Server.signer`
	tokens := []string{}
	for _, ut := range uploadTokens {
		data, err := json.Marshal(ut)
		if err != nil {
			replyBadRequest(w)
			return
		}
		tokens = append(tokens, s.signer.Generate(data))
	}

	if len(tokens) > 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(httpUnprocessableEntity)
	}

	json.NewEncoder(w).Encode(&jsonResponse{Tokens: tokens})
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

func auth(v verifier.UserVerifier, w http.ResponseWriter, r *http.Request) int {
	userId, err := v.Verify(r.FormValue("auth_token"), time.Now())
	if err != nil || userId == 0 {
		replyError(w, badAuth, http.StatusUnauthorized)
		return 0
	}
	return userId
}

func replyError(w http.ResponseWriter, e string, code int) {
	resp := &jsonResponse{Error: "e"}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func replyBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(badRequest))
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

func contentDisposition(filename string) string {
	res := "attachment"
	if filename != "" {
		uriName := url.QueryEscape(filename)
		res = fmt.Sprintf("attachment; filename='%s'; filename*=UTF-8''%s",
			uriName, uriName)
	}
	return res
}
