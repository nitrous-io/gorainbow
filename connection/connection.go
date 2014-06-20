package connection

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/nitrous-io/gorainbow/agent"
	_ "github.com/nitrous-io/gorainbow/agent/ssh"

	"github.com/nitrous-io/gorainbow/command"
	_ "github.com/nitrous-io/gorainbow/command/filefind"
	_ "github.com/nitrous-io/gorainbow/command/fs"
	_ "github.com/nitrous-io/gorainbow/command/parts"
	_ "github.com/nitrous-io/gorainbow/command/search"
	_ "github.com/nitrous-io/gorainbow/command/shell"

	"github.com/nitrous-io/gorainbow/cipher"
	"github.com/nitrous-io/gorainbow/models"
	"github.com/nitrous-io/gorainbow/multiplexer"
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
	"github.com/nitrous-io/gorainbow/verifier"
)

const (
	defaultUsername = "action"
	maxTimeWait     = 1 * time.Second
)

type Connection struct {
	ws       sockjs.ReadWriteCloser
	verifier verifier.UserVerifier
	signer   verifier.SignatureVerifier
	db       *sql.DB
	mulx     *multiplexer.Multiplexer
	aesKey   string
	env      map[string]string // shared env variable with the commands

	userId int
	boxId  int
	cmds   map[string]command.Command
}

func NewConnection(ws sockjs.ReadWriteCloser, verifier verifier.UserVerifier,
	signer verifier.SignatureVerifier, db *sql.DB, mulx *multiplexer.Multiplexer,
	aesKey string, env map[string]string) *Connection {
	c := &Connection{
		ws:       ws,
		verifier: verifier,
		db:       db,
		mulx:     mulx,
		cmds:     make(map[string]command.Command),
		aesKey:   aesKey,
		signer:   signer,
		env:      env,
	}
	return c
}

// start the loop to handle messages from the socket. This is only for handling
// incoming messages, the commands can write messages themselves.
func (c *Connection) Start() (err error) {
	for {
		var msg *protocol.Message
		if msg, err = c.read(); err != nil {
			return err
		}
		switch {
		case msg.Namespace == "global" || msg.Event == "auth":
			// first try to authenticate the user
			if err := c.authenticate(msg); err != nil {
				return err
			}
		case msg.Event == "init":
			// if this is init msg, ignore it. we will initialize our `Command`
			// when we get `start` msg instead
			c.write(protocol.CallbackMsg(msg.CallbackId))
		case msg.Event == "start" && msg.Namespace == "command":
			boxId := int(msg.Args[0].(float64))
			for _, namespace := range []string{"filefind", "search"} {
				if err := c.startCommand(namespace, boxId); err != nil {
					return err
				}
			}
			c.write(protocol.CallbackMsg(msg.CallbackId))
		case msg.Event == "start":
			boxId := int(msg.Args[0].(float64))
			if err := c.startCommand(msg.Namespace, boxId); err != nil {
				return err
			}
			c.write(protocol.CallbackMsg(msg.CallbackId))
		default:
			cmd, ok := c.cmds[msg.Namespace]
			if !ok {
				return fmt.Errorf("Missing command %v", msg)
			}
			if err := cmd.Handle(msg); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Connection) Close() {
	for _, cmd := range c.cmds {
		cmd.Close()
		c.mulx.Release(c.boxId)
	}
}

// initialize the commands, this handles the establishments of ssh connections
func (c *Connection) startCommand(namespace string, boxId int) error {
	c.boxId = boxId
	// fetch box and keypair
	box, keypair, err := models.FetchBoxAndKeyPair(c.db, boxId, c.userId)
	if err != nil {
		return err
	}
	privateKey, err := cipher.Decrypt(
		[]byte(c.aesKey),
		[]byte(keypair.EncryptedPrivateKeyIv),
		[]byte(keypair.EncryptedPrivateKey),
	)
	if err != nil {
		return err
	}
	// get ssh connection from the multiplexer
	var client agent.Client
	addr := fmt.Sprintf("%s:%d", box.Host, box.Port)
	muxClient, err := c.mulx.Add(boxId, addr, func() (multiplexer.WaitCloser, error) {
		return newClient(defaultUsername, addr, privateKey)
	})
	if err != nil {
		return err
	} else {
		client = muxClient.(agent.Client)
	}
	cmd, err := command.InitCommand(namespace, &command.Config{
		Socket: c.ws,
		Client: client,
		Signer: c.signer,
		Env:    c.env,
	})
	if err != nil {
		return err
	}
	c.cmds[namespace] = cmd
	return nil
}

// authenticate the current message and returne an error if the current user is
// not verified
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

// read the current message from websocket
func (c *Connection) read() (*protocol.Message, error) {
	data, err := c.ws.Read()
	if err != nil {
		return nil, err
	}
	return protocol.ParseClientMsg(data)
}

// write a message to websocket
func (c *Connection) write(msg *protocol.Message) {
	c.ws.Write(protocol.EncodeServerMsg(msg))
}

// create a new client to communicate with the box
func newClient(username string, addr string, priv []byte) (agent.Client, error) {
	log.Printf("connection: new ssh connection %s", addr)
	return agent.GetClient("ssh", &agent.ClientConfig{
		Username:   username,
		Address:    addr,
		PrivateKey: priv,
		Wait:       maxTimeWait,
	})
}
