package search

import (
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
