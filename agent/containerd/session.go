package containerd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

func newSession(id, username string) *session {
	s := &session{
		id:       id,
		username: username,
	}
	s.inR, s.inW = io.Pipe()
	s.outR, s.outW = io.Pipe()
	s.errR, s.errW = io.Pipe()
	return s
}

type session struct {
	id       string
	username string
	shellId  string

	inW io.WriteCloser
	inR io.Reader

	outW io.WriteCloser
	outR io.Reader

	errW io.WriteCloser
	errR io.Reader

	conn *httputil.ClientConn
}

func (s *session) Close() error {
	return nil
}

func getConn() (*httputil.ClientConn, error) {
	dial, err := net.Dial("unix", containerdSocket)
	if err != nil {
		return nil, fmt.Errorf("Can't connect to containerd daemon %v", containerdSocket)
	}
	return httputil.NewClientConn(dial, nil), nil
}

func (s *session) Output(cmd string) ([]byte, error) {
	params := url.Values{}
	params.Set("username", s.username)
	url := fmt.Sprintf("/containers/%s/command?%s", s.id, params.Encode())
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "plain/text")
	req.Host = containerdSocket

	clientconn, err := getConn()
	if err != nil {
		return nil, fmt.Errorf("Can't connect to containerd daemon %v", containerdSocket)
	}
	defer clientconn.Close()
	_, err = clientconn.Do(req)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *session) Run(cmd string) error {
	return nil
}

func (s *session) Start(cmd string) error {
	return nil
}

func (s *session) Wait() error {
	return nil
}

func (s *session) StdinPipe() (io.WriteCloser, error) {
	return s.inW, nil
}

func (s *session) StdoutPipe() (io.Reader, error) {
	return s.outR, nil
}

func (s *session) StderrPipe() (io.Reader, error) {
	return s.errR, nil
}

func (s *session) AllPipe() (stdin io.WriteCloser, stdout io.Reader, stderr io.Reader, err error) {
	if stdin, err = s.StdinPipe(); err != nil {
		return
	}
	if stdout, err = s.StdoutPipe(); err != nil {
		return
	}
	if stderr, err = s.StderrPipe(); err != nil {
		return
	}
	return
}

func (s *session) RemoteShell(term string, h, w int) error {
	params := url.Values{}
	params.Set("username", s.username)

	// get the shell id
	endpoint := fmt.Sprintf("/containers/%s/shell?%s", s.id, params.Encode())
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "plain/text")
	req.Host = containerdSocket

	clientconn, err := getConn()
	if err != nil {
		return fmt.Errorf("Can't connect to containerd daemon %v %v", containerdSocket, err)
	}
	resp, err := clientconn.Do(req)
	if err != nil {
		return err
	}
	id, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	s.shellId = string(id)
	clientconn.Close()

	log.Println("shellId", s.shellId)

	// attach to the shell and set stdin, stdout properly
	params = url.Values{}
	params.Set("username", s.username)
	params.Set("term", term)
	params.Set("width", strconv.Itoa(w))
	params.Set("height", strconv.Itoa(h))

	endpoint = fmt.Sprintf("/shell/%s/attach?%s", s.shellId, params.Encode())
	req, err = http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "plain/text")
	req.Host = containerdSocket

	clientconn, err = getConn()
	if err != nil {
		return fmt.Errorf("Can't attach containerd daemon %v", containerdSocket)
	}
	defer clientconn.Close()
	clientconn.Do(req)

	// hijack the http connection here. We have the stream now
	// set this stream as the session stdin/stdout/stderr
	rwc, _ := clientconn.Hijack()

	log.Println("hijack done")
	go io.Copy(rwc, s.inR)
	go io.Copy(s.outW, rwc)

	return nil
}

func (s *session) Write(p []byte) (int, error) {
	log.Println(string(p))
	return len(p), nil
}

func (s *session) Resize(w, h float64) error {
	if s.shellId == "" {
		return errors.New("missing shell id. Please call RemoteShell")
	}
	params := url.Values{}
	params.Set("username", s.username)
	params.Set("width", strconv.FormatFloat(w, 'f', 0, 64))
	params.Set("height", strconv.FormatFloat(h, 'f', 0, 64))

	log.Println("resize", params)
	endpoint := fmt.Sprintf("/shell/%s/resize?%s", s.shellId, params.Encode())
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}
	req.Host = containerdSocket

	clientconn, err := getConn()
	if err != nil {
		return fmt.Errorf("Can't connect to containerd daemon %v", containerdSocket)
	}
	defer clientconn.Close()
	_, err = clientconn.Do(req)
	return err
}
