package search

import (
	"testing"

	at "github.com/nitrous-io/gorainbow/agent/testing"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/protocol"
	st "github.com/nitrous-io/gorainbow/sockjs/testing"
)

func TestSearch(t *testing.T) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	s, _ := client.NewSession()
	cmd, _ := NewSearchCmd(&command.Config{
		Socket: ws,
		Client: client,
	})

	cmd.Handle(&protocol.Message{
		Event:      "search",
		CallbackId: 42,
		Args:       []interface{}{"1", "hello", ".", "ignoreCase,literal"},
	})

	expectedCmd := `/usr/local/bin/ag --ackmate --group --all-text --ignore-case --literal hello .`
	if s.(*at.MemSession).Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.(*at.MemSession).Cmd, expectedCmd)
	}
}
