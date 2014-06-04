package command

import (
	"fmt"
	"os"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/terminal"
	"github.com/kr/pty"
	"github.com/nitrous-io/gorainbow/cipher"
	"github.com/nitrous-io/gorainbow/models"
	"github.com/nitrous-io/gorainbow/protocol"
)

const DefaultUsername = "action"

type TerminalCommand struct {
	cfg *CommandConfig

	session *ssh.Session
	master  *os.File
	slave   *os.File
}

func (c *TerminalCommand) end() {
	c.master.Close()
	c.slave.Close()
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
	c.master = master
	c.slave = slave
	c.session.Stdout = master
	c.session.Stdin = slave
	c.session.Stderr = slave

	// Set up terminal modes
	if _, err = terminal.MakeRaw(int(slave.Fd())); err != nil {
		return err
	}
	termWidth, termHeight, err := terminal.GetSize(int(slave.Fd()))
	if err != nil {
		return err
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		return err
	}
	// Start remote shell
	if err := session.Shell(); err != nil {
		return err
	}
	return nil
}

func (c *TerminalCommand) writeOne(msg *protocol.Message) {
	c.cfg.WS.Write(protocol.EncodeServerMsg(msg))
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
	}
	return nil
}

func newTerminal(cfg *CommandConfig) (Command, error) {
	c := &TerminalCommand{cfg: cfg}
	return c, nil
}
