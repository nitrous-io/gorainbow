package gorainbow

import (
	"github.com/kr/pty"
	"io"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

type SSHSocket struct {
	Conn *Connection
	Pty  *os.File
}

func (s *SSHSocket) OnMessage(msg Message) {
	switch msg.Event {
	case "start":
		s.Pty, _ = pty.Start(s.sshCmd())
		startMsg := Message{Namespace: "global", Event: "callback",
			CallbackId: msg.CallbackId}
		go s.Conn.Write(startMsg.Encode())
		go func() {
			io.Copy(s, s.Pty)
		}()
	case "d":
		b, _ := msg.Args[0].(string)
		go func() {
			s.Pty.WriteString(b)
		}()
	case "r":
		s.Resize(msg)
		startMsg := Message{Namespace: "global", Event: "callback",
			CallbackId: msg.CallbackId}
		go s.Conn.Write(startMsg.Encode())
	case "kill":
	}
}

func (s *SSHSocket) Write(data []byte) (int, error) {
	msg := Message{Namespace: "ssh", Event: "p", NoCallback: true,
		Args: []interface{}{string(data)}}
	res := msg.Encode()
	_, err := s.Conn.Write(res)
	return len(data), err
}

func (s *SSHSocket) Read(data []byte) (int, error) {
	return s.Pty.Write(data)
}

func (s *SSHSocket) Resize(msg Message) error {
	w, _ := msg.Args[0].(float64)
	h, _ := msg.Args[1].(float64)
	return SetWinsize(s.Pty.Fd(), &Winsize{Height: uint16(h), Width: uint16(w)})
}

func (s *SSHSocket) sshCmd() *exec.Cmd {
	return exec.Command("ssh",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "StrictHostKeyChecking=no",
		"-o", "PubkeyAuthentication=yes",
		"-o", "PasswordAuthentication=no",
		"-o", "LogLevel=FATAL",
		"-i", "/home/action/.ssh/nopass_id_rsa",
		"action@192.168.254.254")
}

type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16
	y      uint16
}

func GetWinsize(fd uintptr) (*Winsize, error) {
	ws := &Winsize{}
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))
	// Skipp errno = 0
	if err == 0 {
		return ws, nil
	}
	return ws, err
}

func SetWinsize(fd uintptr, ws *Winsize) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
	// Skipp errno = 0
	if err == 0 {
		return nil
	}
	return err
}
