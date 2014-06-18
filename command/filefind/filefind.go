package filefind

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
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
	// test if the path is not found
	badPathTest *regexp.Regexp

	// test if matcher binary not present
	// specifically for edge case of paid users not restarting boxes
	// and therefore not picking up base template changes
	noBinaryTest *regexp.Regexp

	ErrTimeout  = errors.New("timed out")
	ErrBadPath  = errors.New("NOPATH")
	ErrNoBinary = errors.New("NOBINARY")
)

func init() {
	badPathTest = regexp.MustCompile(`Error stat\(\)ing`)
	noBinaryTest = regexp.MustCompile(`matcher: not found`)
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

	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}
	var cmderr bytes.Buffer
	go io.Copy(bufio.NewWriter(&cmderr), stderr)

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
		return ErrTimeout
	case err := <-done:
		if knownErr := identifyError(cmderr.String()); knownErr != nil {
			return knownErr
		} else {
			return err
		}
	}
	return nil
}

func (c *FilefindCmd) complete(id string, err error) {
	if err != nil {
		c.writeOne(&protocol.Message{
			Namespace: "filefind",
			Event:     "commandError",
			Args:      []interface{}{id + err.Error()},
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

func identifyError(stderr string) error {
	switch {
	case badPathTest.MatchString(stderr):
		return ErrBadPath
	case noBinaryTest.MatchString(stderr):
		return ErrNoBinary
	default:
		return nil
	}
}

func (c *FilefindCmd) writeOne(msg *protocol.Message) {
	utils.WriteOne(c.ws, msg)
}

func regexTerm(term string) string {
	chars := strings.Split(term, "")
	return ".*?" + strings.Join(chars, ".*?") + ".*?"
}
