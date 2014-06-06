package search

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"github.com/nitrous-io/gorainbow/cipher"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/command/utils"
	"github.com/nitrous-io/gorainbow/models"
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
)

const (
	findCmd = "/usr/local/bin/ag"
	// multiple searches can be in-flight simultaneously, so we need to tag data
	// sent over the socket with the search it corresponds to.  This is the
	// separator used -- id is guaranteed numeric
	idDataDelimiter = "!!\n"

	defaultUsername = "action"

	timeout = 1 * time.Minute
)

var (
	optionsToArgs = map[string]string{
		"ignoreCase": "--ignore-case",
		"literal":    "--literal",
	}

	multibyteTest *regexp.Regexp
)

func init() {
	multibyteTest = regexp.MustCompile(`[^\x00-\x7F]`)
	command.Register("search", NewSearch)
}

type SearchCmd struct {
	ws     sockjs.Writer
	cfg    *command.CommandConfig
	client *ssh.Client
}

func (c *SearchCmd) writeOne(msg *protocol.Message) {
	utils.WriteOne(c.ws, msg)
}

func (c *SearchCmd) Handle(msg *protocol.Message) error {
	// NOTE: we dont send the response here because this is started as part of
	// command:start msg, so we will send response from the caller.
	// FIXME: when we remove command pattern, we can reply from here
	if msg.Event == "start" {
		boxId := int(msg.Args[0].(float64))
		if err := c.start(boxId); err != nil {
			return err
		}
	} else if msg.Event == "search" {
		tabId := msg.Args[0].(string)
		term := msg.Args[1].(string)
		location := msg.Args[2].(string)
		options := msg.Args[3].(string)
		id := tabId + idDataDelimiter

		// we can callback immediately without waiting for result
		c.writeOne(protocol.CallbackMsg(msg.CallbackId))

		// start searching
		if err := c.search(id, term, location, options); err != nil {
			c.writeOne(&protocol.Message{
				Namespace: "search",
				Event:     "commandError",
				Args:      []interface{}{err},
			})
			c.writeOne(&protocol.Message{
				Namespace: "search",
				Event:     "commandComplete",
				Args:      []interface{}{id + "1"},
			})
		} else {
			c.writeOne(&protocol.Message{
				Namespace: "search",
				Event:     "commandComplete",
				Args:      []interface{}{id + "0"},
			})
		}
	}
	return nil
}

func (c *SearchCmd) search(id, term, location, options string) error {
	// open a new session
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	// flush stdout/stderr to the socket directly
	session.Stdout = utils.NewPrefixWriter(id,
		utils.NewBufferWriter(500, utils.NewOutputWriter("search", c.ws)))

	// build arguments
	args := []string{"--ackmate", "--group", "--all-text"}
	optArgs := optionToArguments(options, willIgnoreCase(term))
	args = append(args, optArgs...)
	args = append(args, term, location)
	for i, a := range args {
		args[i] = utils.ShellEscape(a)
	}

	cmd := fmt.Sprintf("%s %s", findCmd, strings.Join(args, " "))
	log.Printf("search: cmd: %v", cmd)
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

func (c *SearchCmd) start(boxId int) error {
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
		sshClient, err = mulx.Add(userId, defaultUsername,
			fmt.Sprintf("%s:%d", box.Host, box.Port),
			privateKey)
		if err != nil {
			return err
		}
	}

	c.client = sshClient
	return nil
}

func NewSearch(cfg *command.CommandConfig) (command.Command, error) {
	return &SearchCmd{cfg: cfg, ws: cfg.WS}, nil
}

func willIgnoreCase(term string) bool {
	return multibyteTest.MatchString(term)
}

func optionToArguments(options string, ignoreCase bool) []string {
	args := []string{}
	for _, opt := range strings.Split(options, ",") {
		arg, ok := optionsToArgs[opt]
		if !ok || (opt == "ignoreCase" && !ignoreCase) {
			continue
		}
		args = append(args, arg)
	}
	return args
}
