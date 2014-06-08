package filefind

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/command/utils"
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
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

	defaultUsername = "action"

	timeout = 1 * time.Minute
)

var (
	ErrPathNotFound = errors.New("NOPATH")
	ErrNOBINARY     = errors.New("NOBINARY")
)

func init() {
	command.Register("filefind", NewFilefindCmd)
}

type FilefindCmd struct {
	ws     sockjs.Writer
	client *ssh.Client
}

func NewFilefindCmd(ws sockjs.Writer, client *ssh.Client) (command.Command, error) {
	return &FilefindCmd{ws: ws, client: client}, nil
}

func (c *FilefindCmd) Handle(msg *protocol.Message) error {
	if msg.Event == "filefind" {
		term := msg.Args[0].(string)
		location := msg.Args[1].(string)
		id := term + searchKeyDelimiter + location + idDataDelimiter
		// we can callback immediately without waiting for result
		c.writeOne(protocol.CallbackMsg(msg.CallbackId))

		// start searching
		if err := c.find(id, term, location); err != nil {
			c.writeOne(&protocol.Message{
				Namespace: "filefind",
				Event:     "commandError",
				Args:      []interface{}{err},
			})
			c.writeOne(&protocol.Message{
				Namespace: "filefind",
				Event:     "commandComplete",
				Args:      []interface{}{id + "1"},
			})
		} else {
			c.writeOne(&protocol.Message{
				Namespace: "filefind",
				Event:     "commandComplete",
				Args:      []interface{}{id + "0"},
			})
		}
	}
	return nil
}

func (c *FilefindCmd) find(id, term, location string) error {
	// open a new session
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	// flush stdout/stderr to the socket directly
	session.Stdout = utils.NewPrefixWriter(id,
		utils.NewBufferWriter(10, utils.NewOutputWriter("filefind", c.ws)))

	cmd := fmt.Sprintf("sh -c '%s -i -g %s %s | %s %s'",
		findCmd, utils.ShellEscape(regexTerm(term)), utils.GetPath(location),
		matcherCmd, utils.ShellEscape(term))
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

func (c *FilefindCmd) writeOne(msg *protocol.Message) {
	utils.WriteOne(c.ws, msg)
}

func regexTerm(term string) string {
	chars := strings.Split(term, "")
	return ".*?" + strings.Join(chars, ".*?") + ".*?"
}
