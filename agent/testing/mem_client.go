// create an in memory client and session that stores/transmit the data
// in-memory.
package testing

import (
	"bytes"
	"io"
	"io/ioutil"
	"time"

	"github.com/nitrous-io/gorainbow/agent"
)

var (
	defaultWaitTime = 100 * time.Millisecond
)

type MemClient struct {
	Session *MemSession
}

func (c *MemClient) Close() error {
	return nil
}

func (c *MemClient) NewFilesystem() (agent.Filesystem, error) {
	return nil, nil
}

func (c *MemClient) NewSession() (agent.Session, error) {
	if c.Session != nil {
		return c.Session, nil
	}
	s := new(MemSession)
	s.Stdin = bytes.NewBuffer(nil)
	s.Stdout = bytes.NewBuffer(nil)
	s.Stderr = bytes.NewBuffer(nil)
	c.Session = s
	return s, nil
}

func NewMemClient() (agent.Client, error) {
	return new(MemClient), nil
}

type MemSession struct {
	Stdin         *bytes.Buffer
	Stdout        *bytes.Buffer
	Stderr        *bytes.Buffer
	Closed        bool
	Width, Height float64
	Term          string
	RunErr        error
	Cmd           string
	WaitTime      time.Duration
}

func (s *MemSession) Close() error {
	s.Closed = true
	return nil
}

func (s *MemSession) Output(cmd string) ([]byte, error) {
	return ioutil.ReadAll(s.Stdout)
}

func (s *MemSession) Run(cmd string) error {
	s.Cmd = cmd
	return s.RunErr
}

func (s *MemSession) Start(cmd string) error {
	s.Cmd = cmd
	return s.RunErr
}

func (s *MemSession) Wait() error {
	time.Sleep(defaultWaitTime)
	return s.RunErr
}

func (s *MemSession) StdinPipe() (io.WriteCloser, error) {
	in, out := io.Pipe()
	go io.Copy(s.Stdin, in)
	return out, nil
}

func (s *MemSession) StdoutPipe() (io.Reader, error) {
	in, out := io.Pipe()
	go io.Copy(out, s.Stdout)
	return in, nil
}

func (s *MemSession) StderrPipe() (io.Reader, error) {
	in, out := io.Pipe()
	go io.Copy(out, s.Stderr)
	return in, nil
}

func (s *MemSession) AllPipe() (io.WriteCloser, io.Reader, io.Reader, error) {
	stdin, _ := s.StdinPipe()
	stdout, _ := s.StdoutPipe()
	stderr, _ := s.StderrPipe()
	return stdin, stdout, stderr, nil

}

func (s *MemSession) RemoteShell(term string, h, w int) error {
	s.Term = term
	s.Width = float64(w)
	s.Height = float64(h)
	return nil
}

func (s *MemSession) Resize(w, h float64) error {
	s.Width = float64(w)
	s.Height = float64(h)
	return nil
}
