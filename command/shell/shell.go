package shell

import (
	"io"
	"log"
	"os"

	"github.com/kr/pty"
	"github.com/nitrous-io/gorainbow/agent"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/command/utils"
	"github.com/nitrous-io/gorainbow/pkg/term"
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	command.Register("ssh", NewShell)
}

const (
	termMode        = "xterm-256color"
	defaultUsername = "action"
)

type Shell struct {
	ws     sockjs.Writer
	client agent.Client

	session agent.Session
	master  *os.File
	slave   *os.File
}

func (c *Shell) Close() {
	log.Printf("shell: session ended")
	if c.slave != nil {
		c.slave.Close()
		c.slave = nil
	}
	if c.master != nil {
		c.master.Close()
		c.master = nil
	}
	if c.session != nil {
		c.session.Close()
		c.session = nil
	}
}

func NewShell(config *command.Config) (command.Command, error) {
	ws := config.Socket
	client := config.Client
	c := &Shell{ws: ws, client: client}
	// open a new session
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	c.session = session
	// allocate a pty attach to the ssh session
	master, slave, err := pty.Open()
	if err != nil {
		return nil, err
	}
	c.slave = slave
	c.master = master
	// prepare session input, output, error
	stdin, stdout, stderr, err := session.AllPipe()
	if err != nil {
		return nil, err
	}
	go io.Copy(slave, stdout)
	go io.Copy(slave, stderr)
	go io.Copy(stdin, slave)
	go io.Copy(c, master)

	// Set up terminal modes
	if _, err = terminal.MakeRaw(int(slave.Fd())); err != nil {
		return nil, err
	}
	termWidth, termHeight, err := terminal.GetSize(int(slave.Fd()))
	if err != nil {
		return nil, err
	}
	// Request pseudo terminal
	if err := session.RemoteShell(termMode, termHeight, termWidth); err != nil {
		return nil, err
	}
	// closes the cmmand when the session finishes
	go func() {
		if err := session.Wait(); err != nil {
			c.Close()
		}
	}()
	return c, nil
}

// How we do it
// - create a prt that represents our shell
// - makeRaw the pty
// - handle input/output with the pty
// - set pty as the session stdin/out/err
func (c *Shell) Handle(msg *protocol.Message) error {
	if msg.Event == "r" {
		w := msg.Args[0].(float64)
		h := msg.Args[1].(float64)
		c.resize(w, h)
	} else if msg.Event == "d" {
		_, err := c.master.Write([]byte(msg.Args[0].(string)))
		if err != nil {
			log.Printf("write error %v", err)
		}
	}
	return nil
}

// write output from ssh shell to sockjs protocol
// Shell fulfill `io.Writer` interface so it can be uses in `io.Copy` and co.
func (c *Shell) Write(p []byte) (int, error) {
	c.writeOne(&protocol.Message{
		Namespace: "ssh",
		Event:     "p",
		Args:      []interface{}{string(p)},
	})
	return len(p), nil
}

func (c *Shell) writeOne(msg *protocol.Message) {
	utils.WriteOne(c.ws, msg)
}

func (c *Shell) resize(w, h float64) error {
	// resize local terminal
	ws := &term.Winsize{Height: uint16(h), Width: uint16(w)}
	if err := term.SetWinsize(c.master.Fd(), ws); err != nil {
		return err
	}
	// resize remote terminal
	return c.session.Resize(w, h)
}
