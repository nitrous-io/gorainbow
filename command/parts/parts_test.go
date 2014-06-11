package parts

import (
	"testing"

	at "github.com/nitrous-io/gorainbow/agent/testing"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/protocol"
	st "github.com/nitrous-io/gorainbow/sockjs/testing"
)

func TestListAll(t *testing.T) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	s, _ := client.NewSession()
	cmd, _ := NewPartsCmd(&command.Config{
		Socket: ws,
		Client: client,
	})

	cmd.Handle(&protocol.Message{
		Event: "listAll",
		Args:  []interface{}{42},
	})

	expectedCmd := `parts search --json && parts list --json`
	if s.(*at.MemSession).Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.(*at.MemSession).Cmd, expectedCmd)
	}
}

func TestStatus(t *testing.T) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	s, _ := client.NewSession()
	cmd, _ := NewPartsCmd(&command.Config{
		Socket: ws,
		Client: client,
	})

	cmd.Handle(&protocol.Message{
		Event: "status",
		Args:  []interface{}{42},
	})

	expectedCmd := `parts status --json`
	if s.(*at.MemSession).Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.(*at.MemSession).Cmd, expectedCmd)
	}
}

func TestInstall(t *testing.T) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	s, _ := client.NewSession()
	cmd, _ := NewPartsCmd(&command.Config{
		Socket: ws,
		Client: client,
	})

	cmd.Handle(&protocol.Message{
		Event: "install",
		Args:  []interface{}{"ack"},
	})

	expectedCmd := `parts install 'ack'`
	if s.(*at.MemSession).Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.(*at.MemSession).Cmd, expectedCmd)
	}
}

func TestUninstall(t *testing.T) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	s, _ := client.NewSession()
	cmd, _ := NewPartsCmd(&command.Config{
		Socket: ws,
		Client: client,
	})

	cmd.Handle(&protocol.Message{
		Event: "uninstall",
		Args:  []interface{}{"ack"},
	})

	expectedCmd := `parts uninstall 'ack'`
	if s.(*at.MemSession).Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.(*at.MemSession).Cmd, expectedCmd)
	}
}

func TestStart(t *testing.T) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	s, _ := client.NewSession()
	cmd, _ := NewPartsCmd(&command.Config{
		Socket: ws,
		Client: client,
	})

	cmd.Handle(&protocol.Message{
		Event: "partStart",
		Args:  []interface{}{"apache2"},
	})

	expectedCmd := `parts start 'apache2'`
	if s.(*at.MemSession).Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.(*at.MemSession).Cmd, expectedCmd)
	}
}

func TestStop(t *testing.T) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	s, _ := client.NewSession()
	cmd, _ := NewPartsCmd(&command.Config{
		Socket: ws,
		Client: client,
	})

	cmd.Handle(&protocol.Message{
		Event: "partStop",
		Args:  []interface{}{"apache2"},
	})

	expectedCmd := `parts stop 'apache2'`
	if s.(*at.MemSession).Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.(*at.MemSession).Cmd, expectedCmd)
	}
}
