package search

import (
	"fmt"
	"testing"

	"github.com/nitrous-io/gorainbow/agent"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/protocol"

	testagent "github.com/nitrous-io/gorainbow/agent/testing"
	testsockjs "github.com/nitrous-io/gorainbow/sockjs/testing"
)

func prepare() (command.Command, *testsockjs.Writer, *testagent.Client, *testagent.Session) {
	ws := testsockjs.NewWriter()
	client, _ := testagent.NewClient(new(agent.ClientConfig))
	s, _ := client.NewSession()
	cmd, _ := NewSearchCmd(&command.Config{
		Socket: ws,
		Client: client,
	})
	return cmd, ws.(*testsockjs.Writer), client.(*testagent.Client), s.(*testagent.Session)
}

func TestSearch(t *testing.T) {
	cmd, _, _, s := prepare()
	cmd.Handle(&protocol.Message{
		Event:      "search",
		CallbackId: 42,
		Args:       []interface{}{"1", "hello", ".", "ignoreCase,literal"},
	})
	expectedCmd := `/usr/local/bin/ag --ackmate --group --all-text --ignore-case --literal hello .`
	if s.Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.Cmd, expectedCmd)
	}
}

func TestIdentifyError(t *testing.T) {
	err := identifyError("ERR: Error stat()ing: workspace/nopath\n")
	expected := ErrBadPath
	if err != expected {
		t.Errorf("Expected %s, got %s", expected, err)
	}

	err = identifyError("ERR: pcre_compile failed at position 5. Error: missing )\n")
	expected = ErrBadRegex
	if err != expected {
		t.Errorf("Expected %s, got %s", expected, err)
	}

	err = identifyError("ERR: Too many matches in README.md. Skipping\n")
	expected = fmt.Errorf(ErrfMaxMatch, "README.md")
	if err.Error() != expected.Error() {
		t.Errorf("Expected %s, got %s", expected.Error(), err.Error())
	}

	err = identifyError("ERR: Something we haven't seen yet\n")
	if err != nil {
		t.Errorf("Expected nil, got %s", err)
	}
}
