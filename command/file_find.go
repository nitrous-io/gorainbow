package command

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"github.com/nitrous-io/gorainbow/cipher"
	"github.com/nitrous-io/gorainbow/models"
	"github.com/nitrous-io/gorainbow/protocol"
)

const (
	findCmd     = "/usr/local/bin/ag"
	matcherCmd  = "/usr/local/bin/matcher"
	resultLimit = 10

	// multiple searches can be in-flight simultaneously, so we need to tag data
	// sent over the socket with the search it corresponds to.  This is the
	// separator used -- id is guaranteed numeric
	idDataDelimiter = "!!\n"

	// separator between file pattern and path so that (pattern1, path1) and
	// (pattern2, path2) do not overlap
	searchKeyDelimiter = "\\\\"
)

var (
	ErrPathNotFound = errors.New("NOPATH")
	ErrNOBINARY     = errors.New("NOBINARY")
)

type FileFindCommand struct {
	cfg       *CommandConfig
	client    *ssh.Client
	cmdOutput *scannerOutput
	cmdError  *cachedStream
}

func (c *FileFindCommand) Handle(msg *protocol.Message) error {
	// NOTE: we dont send the response here because this is started as part of
	// command:start msg, so we will send response from the caller.
	// FIXME: when we remove command pattern, we can reply from here
	if msg.Event == "start" {
		boxId := int(msg.Args[0].(float64))
		if err := c.start(boxId); err != nil {
			return err
		}
	} else if msg.Event == "filefind" {
		term := msg.Args[0].(string)
		location := msg.Args[1].(string)
		if err := c.find(term, location); err != nil {
			err = fmt.Errorf("Error: %v, stderr: %s", err,
				c.cmdError.content.String())
			log.Println("filefind: error", err)
			writeOne(c.cfg.WS, &protocol.Message{
				Namespace: "filefind",
				Event:     "commandError",
				Args:      []interface{}{err},
			})
		}
		writeOne(c.cfg.WS, &protocol.Message{
			Namespace: "filefind",
			Event:     "commandComplete",
			Args:      []interface{}{},
		})
	}
	return nil
}

func (c *FileFindCommand) find(term, location string) error {
	absLocation := getPath(location)
	// open a new session
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	// flush stdout/stderr to the socket directly
	session.Stdout = &prefixWriter{
		prefix: term + searchKeyDelimiter + location + idDataDelimiter,
		output: c.cmdOutput,
	}
	session.Stderr = c.cmdError

	cmd := fmt.Sprintf("sh -c '%s -i -g %s %s | %s %s'",
		findCmd, shellEscape(regexTerm(term)), absLocation,
		matcherCmd, shellEscape(term))
	log.Println("Find cmd: %v", cmd)
	// start and wait for cmd to finish or timeout
	if err := session.Start(cmd); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()
	select {
	case <-time.After(timeout):
		if err := session.Close(); err != nil {
			return fmt.Errorf("Session closed with err: %v.Timed out %v", err, cmd)
		}
		return fmt.Errorf("Timed out %v", cmd)
	case err := <-done:
		return err
	}
	return nil
}

func (c *FileFindCommand) start(boxId int) error {
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

	c.client = sshClient
	return nil
}

func newFileFind(cfg *CommandConfig) (Command, error) {
	scanner := &scannerOutput{
		lineLimit: 10,
		outBuffer: bytes.NewBuffer(nil),
		output:    &commandOutput{ws: cfg.WS, namespace: "filefind"},
	}
	cachedErr := &cachedStream{
		limit:   500,
		content: bytes.NewBuffer(nil),
	}
	return &FileFindCommand{
		cfg:       cfg,
		cmdOutput: scanner,
		cmdError:  cachedErr,
	}, nil
}
