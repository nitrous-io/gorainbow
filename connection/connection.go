package connection

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/multiplexer"
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
	"github.com/nitrous-io/gorainbow/verifier"
)

type Connection struct {
	ws       sockjs.ReadWriteCloser
	verifier verifier.UserVerifier
	db       *sql.DB
	mulx     *multiplexer.Multiplexer
	aesKey   string

	userId int
	cmds   map[string]command.Command
}

func (c *Connection) authenticate(msg *protocol.Message) error {
	authMsg := msg.Args[0].(string)
	userId, err := c.verifier.Verify(authMsg, time.Now())
	if err != nil {
		return err
	}
	c.userId = userId
	c.write(protocol.CallbackMsg(msg.CallbackId))
	return nil
}

func (c *Connection) initCommand(msg *protocol.Message) error {
	// if this is a command socket, we just ignore it and return back the
	// callback. This should go away once we only multiplex with one socket
	if msg.Namespace == "command" {
		c.write(protocol.CallbackMsg(msg.CallbackId))
		return nil
	}
	cmd, err := command.InitCommand(msg.Namespace, &command.CommandConfig{
		DB: c.db, Mulx: c.mulx, WS: c.ws, UserId: c.userId, AESKey: c.aesKey,
	})
	if err != nil {
		return err
	}
	c.cmds[msg.Namespace] = cmd
	c.write(protocol.CallbackMsg(msg.CallbackId))
	return nil
}

func (c *Connection) read() (*protocol.Message, error) {
	data, err := c.ws.Read()
	if err != nil {
		return nil, err
	}
	return protocol.ParseClientMsg(data)
}

func (c *Connection) write(msg *protocol.Message) {
	c.ws.Write(protocol.EncodeServerMsg(msg))
}

func (c *Connection) Start() error {
	for {
		msg, err := c.read()
		if err != nil {
			return err
		}
		if msg.Namespace == "global" || msg.Event == "auth" {
			if err := c.authenticate(msg); err != nil {
				return err
			}
		} else if msg.Event == "init" {
			if err := c.initCommand(msg); err != nil {
				return err
			}
		} else {
			// special handle for command namespace.
			// Ideally we should just remove it since it does nothing useful
			// This will simply start both filefind and search command
			if msg.Namespace == "command" {
				if msg.Event == "start" {
					// start filefind
					cmd, ok := c.cmds["filefind"]
					if !ok {
						return fmt.Errorf("Missing command %v", msg)
					}
					if err := cmd.Handle(msg); err != nil {
						return err
					}
					// TODO: start search
					// notify that we started
					c.write(protocol.CallbackMsg(msg.CallbackId))
				}
			}
			// Accepted namespace
			// - ssh
			// - parts
			// - sftp
			// - filefind
			if msg.Namespace == "ssh" || msg.Namespace == "parts" ||
				msg.Namespace == "sftp" || msg.Namespace == "filefind" {
				cmd, ok := c.cmds[msg.Namespace]
				if !ok {
					return fmt.Errorf("Missing command %v", msg)
				}
				if err := cmd.Handle(msg); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func NewConnection(ws sockjs.ReadWriteCloser, verifier verifier.UserVerifier,
	db *sql.DB, mulx *multiplexer.Multiplexer, aesKey string) *Connection {
	c := &Connection{
		ws:       ws,
		verifier: verifier,
		db:       db,
		mulx:     mulx,
		cmds:     make(map[string]command.Command),
		aesKey:   aesKey,
	}
	return c
}
