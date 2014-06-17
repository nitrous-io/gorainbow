package search

import (
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

	// test if the regex contains non-ascii character
	multibyteTest *regexp.Regexp

	// test if the path is not found
	badPathTest *regexp.Regexp

	// test if the provided regex is bad
	badRegexTest *regexp.Regexp

	// test if there are too many matches
	maxMatchesTest *regexp.Regexp

	ErrTimeout   = errors.New("timed out")
	ErrBadPath   = errors.New("NOPATH")
	ErrBadRegex  = errors.New("BADREGEX")
	ErrfMaxMatch = "MAXMATCH:%s"
)

func init() {
	multibyteTest = regexp.MustCompile(`[^\x00-\x7F]`)
	badPathTest = regexp.MustCompile(`Error stat\(\)ing`)
	badRegexTest = regexp.MustCompile(`pcre_compile failed`)
	maxMatchesTest = regexp.MustCompile(`Too many matches in (.*?)\. Skipping`)
	command.Register("search", NewSearchCmd)
}

type SearchCmd struct {
	ws     sockjs.Writer
	client agent.Client
}

func NewSearchCmd(config *command.Config) (command.Command, error) {
	return &SearchCmd{
		ws:     config.Socket,
		client: config.Client,
	}, nil
}

func (c *SearchCmd) Close() {
}

func (c *SearchCmd) Handle(msg *protocol.Message) error {
	if msg.Event == "search" {
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
				Args:      []interface{}{id + err.Error()},
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
	// flush stdout to the socket directly
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	cmdout := utils.NewPrefixWriter(id,
		utils.NewBufferWriter(500, utils.NewOutputWriter("search", c.ws)))
	go io.Copy(cmdout, stdout)

	// handle stderr by caching it so we can look at it later to analyze the
	// failures
	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}
	cmderr := utils.NewCachedWriter()
	go io.Copy(cmderr, stderr)

	// build arguments
	args := []string{"--ackmate", "--group", "--all-text"}
	optArgs := optionToArguments(options, willIgnoreCase(term))
	args = append(args, optArgs...)
	args = append(args, term, location)
	for i, a := range args {
		args[i] = utils.ShellEscape(a)
	}

	// build search command
	cmd := fmt.Sprintf("%s %s", findCmd, strings.Join(args, " "))
	log.Printf("search: cmd: %v", cmd)

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

func (c *SearchCmd) writeOne(msg *protocol.Message) {
	utils.WriteOne(c.ws, msg)
}

func identifyError(stderr string) error {
	switch {
	case badPathTest.MatchString(stderr):
		return ErrBadPath
	case badRegexTest.MatchString(stderr):
		return ErrBadRegex
	case maxMatchesTest.MatchString(stderr):
		matches := maxMatchesTest.FindStringSubmatch(stderr)
		return fmt.Errorf(ErrfMaxMatch, matches[1])
	default:
		return nil
	}
}

func willIgnoreCase(term string) bool {
	return !multibyteTest.MatchString(term)
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
