package filefind

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/nitrous-io/gorainbow/agent"
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
	client agent.Client
}

func NewFilefindCmd(config *command.Config) (command.Command, error) {
	return &FilefindCmd{
		ws:     config.Socket,
		client: config.Client,
	}, nil
}

func (c *FilefindCmd) Close() {
}

func (c *FilefindCmd) Handle(msg *protocol.Message) error {
	if msg.Event == "filefind" {
		term := msg.Args[0].(string)
		location := msg.Args[1].(string)
		id := term + searchKeyDelimiter + location + idDataDelimiter

		// we can callback immediately without waiting for result
		c.writeOne(protocol.CallbackMsg(msg.CallbackId))

		// start searching
		err := c.find(id, term, location)
		c.complete(id, err)
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
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	cmdout := utils.NewPrefixWriter(id,
		utils.NewBufferWriter(10, utils.NewOutputWriter("filefind", c.ws)))
	go io.Copy(cmdout, stdout)

	// build search command
	cmd := fmt.Sprintf("sh -c '%s -i -g %s %s | %s %s'",
		findCmd, utils.ShellEscape(regexTerm(term)), utils.GetPath(location),
		matcherCmd, utils.ShellEscape(term))
	log.Printf("filefind: starting cmd: %v", cmd)

	// start and wait for cmd to finish or timeout
	if err := session.Start(cmd); err != nil {
		return err
	}
	defer session.Close()

	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()

	select {
	case <-time.After(timeout):
		return fmt.Errorf("Timed out %v", cmd)
	case err := <-done:
		return err
	}
	return nil
}

func (c *FilefindCmd) complete(id string, err error) {
	if err != nil {
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

func (c *FilefindCmd) writeOne(msg *protocol.Message) {
	utils.WriteOne(c.ws, msg)
}

func regexTerm(term string) string {
	chars := strings.Split(term, "")
	return ".*?" + strings.Join(chars, ".*?") + ".*?"
}
