// create an in memory client and session that stores/transmit the data
// in-memory.
package testing

import (
	"bytes"
	"io"
	"io/ioutil"
	"time"
)

var (
	defaultWaitTime = 100 * time.Millisecond
)

type Session struct {
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

func (s *Session) Close() error {
	s.Closed = true
	return nil
}

func (s *Session) Output(cmd string) ([]byte, error) {
	s.Cmd = cmd
	return ioutil.ReadAll(s.Stdout)
}

func (s *Session) Run(cmd string) error {
	s.Cmd = cmd
	return s.RunErr
}

func (s *Session) Start(cmd string) error {
	s.Cmd = cmd
	return s.RunErr
}

func (s *Session) Wait() error {
	time.Sleep(defaultWaitTime)
	return s.RunErr
}

func (s *Session) StdinPipe() (io.WriteCloser, error) {
	in, out := io.Pipe()
	go func() {
		_, err := io.Copy(s.Stdin, in)
		in.CloseWithError(err)
	}()
	return out, nil
}

func (s *Session) StdoutPipe() (io.Reader, error) {
	in, out := io.Pipe()
	go func() {
		_, err := io.Copy(out, s.Stdout)
		out.CloseWithError(err)
	}()
	return in, nil
}

func (s *Session) StderrPipe() (io.Reader, error) {
	in, out := io.Pipe()
	go func() {
		_, err := io.Copy(out, s.Stderr)
		out.CloseWithError(err)
	}()
	return in, nil
}

func (s *Session) AllPipe() (io.WriteCloser, io.Reader, io.Reader, error) {
	stdin, _ := s.StdinPipe()
	stdout, _ := s.StdoutPipe()
	stderr, _ := s.StderrPipe()
	return stdin, stdout, stderr, nil

}

func (s *Session) RemoteShell(term string, h, w int) error {
	s.Term = term
	s.Width = float64(w)
	s.Height = float64(h)
	return nil
}

func (s *Session) Resize(w, h float64) error {
	s.Width = float64(w)
	s.Height = float64(h)
	return nil
}
