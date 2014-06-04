package command

import (
	"fmt"
	"io"
	"log"
	"os"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/terminal"
	"github.com/kr/pty"
	"github.com/nitrous-io/gorainbow/cipher"
	"github.com/nitrous-io/gorainbow/models"
	"github.com/nitrous-io/gorainbow/pkg/term"
	"github.com/nitrous-io/gorainbow/protocol"
)

const DefaultUsername = "action"

type resizeRequest struct {
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
}

type TerminalCommand struct {
	cfg *CommandConfig

	session *ssh.Session
	master  *os.File
	slave   *os.File
}

func (c *TerminalCommand) end() {
	c.master.Close()
	c.session.Close()
}

func (c *TerminalCommand) start(boxId int) (err error) {
	mulx := c.cfg.Mulx
	userId := c.cfg.UserId
	db := c.cfg.DB

	// fetch box and keypair
	box, keypair, err := models.FetchBoxAndKeyPair(db, boxId, userId)
	if err != nil {
		return err
	}
	privateKey, err := cipher.Decrypt(
		[]byte(c.cfg.AESKey),
		[]byte(keypair.EncryptedPrivateKeyIv),
		[]byte(keypair.EncryptedPrivateKey),
	)
	if err != nil {
		return err
	}

	// get ssh connection from the multiplexer
	sshClient := mulx.Get(userId)
	if sshClient == nil {
		sshClient, err = mulx.Add(userId, DefaultUsername,
			fmt.Sprintf("%s:%d", box.Host, box.Port),
			privateKey)
		if err != nil {
			return err
		}
	}

	// open a new session
	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	c.session = session

	// allocate a pty attach to the ssh session
	master, slave, err := pty.Open()
	if err != nil {
		return nil
	}
	c.slave = slave
	c.master = master
	// prepare session input, output
	c.session.Stdout = slave
	c.session.Stderr = slave
	c.session.Stdin = slave

	go func() {
		io.Copy(c, master)
	}()

	// Set up terminal modes
	if _, err = terminal.MakeRaw(int(slave.Fd())); err != nil {
		return err
	}
	termWidth, termHeight, err := terminal.GetSize(int(slave.Fd()))
	if err != nil {
		return err
	}
	modes := ssh.TerminalModes{}
	// Request pseudo terminal
	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		return err
	}
	// Start remote shell
	if err := session.Shell(); err != nil {
		return err
	}

	// wait til the session finishes
	go func() {
		if err := session.Wait(); err != nil {
			c.end()
		}
	}()

	return nil
}

// write output from ssh shell to sockjs protocol
func (c *TerminalCommand) Write(p []byte) (int, error) {
	c.writeOne(&protocol.Message{
		Namespace: "ssh",
		Event:     "p",
		Args:      []interface{}{string(p)},
	})
	return len(p), nil
}

func (c *TerminalCommand) writeOne(msg *protocol.Message) {
	c.cfg.WS.Write(protocol.EncodeServerMsg(msg))
}

func (c *TerminalCommand) resize(w, h float64) error {
	ws := &term.Winsize{Height: uint16(h), Width: uint16(w)}
	if err := term.SetWinsize(c.master.Fd(), ws); err != nil {
		return err
	}
	req := &resizeRequest{
		Columns: uint32(w),
		Rows:    uint32(h),
		Width:   uint32(w),
		Height:  uint32(h),
	}
	// resize remote terminal too
	if _, err := c.session.SendRequest("window-change", false,
		ssh.Marshal(req)); err != nil {
		return err
	}
	return nil
}

// How we do it
// - create a prt that represents our shell
// - makeRaw the pty
// - handle input/output with the pty
// - set pty as the session stdin/out/err
func (c *TerminalCommand) Handle(msg *protocol.Message) error {
	if msg.Event == "start" {
		boxId := int(msg.Args[0].(float64))
		if err := c.start(boxId); err != nil {
			return err
		}
		c.writeOne(protocol.CallbackMsg(msg.CallbackId))
	} else if msg.Event == "r" {
		w := msg.Args[0].(float64)
		h := msg.Args[1].(float64)
		c.resize(w, h)
	} else if msg.Event == "d" {
		_, err := c.master.Write([]byte(msg.Args[0].(string)))
		if err != nil {
			log.Println("write error", err)
		}
	}
	return nil
}

func newTerminal(cfg *CommandConfig) (Command, error) {
	c := &TerminalCommand{cfg: cfg}
	return c, nil
}
