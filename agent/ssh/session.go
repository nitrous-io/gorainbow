package agent

import (
	"io"

	"golang.org/x/crypto/ssh"
)

type resizeRequest struct {
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
}

type session struct {
	*ssh.Session
}

func (s session) AllPipe() (stdin io.WriteCloser, stdout io.Reader, stderr io.Reader, err error) {
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

func (s session) Resize(w, h float64) error {
	req := &resizeRequest{
		Columns: uint32(w),
		Rows:    uint32(h),
		Width:   uint32(w),
		Height:  uint32(h),
	}
	_, err := s.Session.SendRequest("window-change", false, ssh.Marshal(req))
	return err
}

func (s session) RemoteShell(term string, h, w int) error {
	// Request pseudo terminal
	if err := s.Session.RequestPty(term, h, w, nil); err != nil {
		return err
	}
	// Start remote shell
	return s.Session.Shell()
}
