package parts

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
	cmd, _ := NewPartsCmd(&command.Config{
		Socket: ws,
		Client: client,
	})
	return cmd, ws.(*testsockjs.Writer), client.(*testagent.Client), s.(*testagent.Session)
}

func TestListAll(t *testing.T) {
	cmd, _, _, s := prepare()
	cmd.Handle(&protocol.Message{
		Event: "listAll",
		Args:  []interface{}{42},
	})
	expectedCmd := `parts search --json && parts list --json`
	if s.Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.Cmd, expectedCmd)
	}
}

func TestStatus(t *testing.T) {
	cmd, _, _, s := prepare()
	cmd.Handle(&protocol.Message{
		Event: "status",
		Args:  []interface{}{42},
	})
	expectedCmd := `parts status --json`
	if s.Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.Cmd, expectedCmd)
	}
}

func TestInstall(t *testing.T) {
	cmd, _, _, s := prepare()
	cmd.Handle(&protocol.Message{
		Event: "install",
		Args:  []interface{}{"ack"},
	})
	expectedCmd := `parts install 'ack'`
	if s.Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.Cmd, expectedCmd)
	}
}

func TestUninstall(t *testing.T) {
	cmd, _, _, s := prepare()
	cmd.Handle(&protocol.Message{
		Event: "uninstall",
		Args:  []interface{}{"ack"},
	})
	expectedCmd := `parts uninstall 'ack'`
	if s.Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.Cmd, expectedCmd)
	}
}

func TestStart(t *testing.T) {
	cmd, _, _, s := prepare()
	cmd.Handle(&protocol.Message{
		Event: "partStart",
		Args:  []interface{}{"apache2"},
	})
	expectedCmd := `parts start 'apache2'`
	if s.Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.Cmd, expectedCmd)
	}
}

func TestStop(t *testing.T) {
	cmd, _, _, s := prepare()
	cmd.Handle(&protocol.Message{
		Event: "partStop",
		Args:  []interface{}{"apache2"},
	})
	expectedCmd := `parts stop 'apache2'`
	if s.Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.Cmd, expectedCmd)
	}
}
