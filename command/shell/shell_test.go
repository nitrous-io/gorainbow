package shell

import (
	"reflect"
	"testing"
	"time"

	"code.google.com/p/go.crypto/ssh/terminal"

	at "github.com/nitrous-io/gorainbow/agent/testing"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/protocol"
	st "github.com/nitrous-io/gorainbow/sockjs/testing"
)

func TestResize(t *testing.T) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	cmd, _ := NewShell(&command.Config{
		Socket: ws,
		Client: client,
	})

	cmd.Handle(&protocol.Message{
		Event: "r",
		Args:  []interface{}{float64(100), float64(50)},
	})

	s := cmd.(*Shell).session.(*at.MemSession)
	if s.Height != 50 || s.Width != 100 {
		t.Fatal("failed to resize", s)
	}

	termWidth, termHeight, err := terminal.GetSize(int(cmd.(*Shell).slave.Fd()))
	if err != nil || termHeight != 50 || termWidth != 100 {
		t.Fatal("failed to resize", termHeight, termWidth, err)
	}
}

func TestInput(t *testing.T) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	cmd, _ := NewShell(&command.Config{
		Socket: ws,
		Client: client,
	})

	cmd.Handle(&protocol.Message{
		Event: "d",
		Args:  []interface{}{"whoami\n"},
	})
	defer cmd.Close()

	s := cmd.(*Shell).session.(*at.MemSession)
	after := time.After(100 * time.Millisecond)
	for {
		select {
		case <-after:
			t.Fatal("timed out")
			return
		default:
			s.Stdin.UnreadByte()
			if s.Stdin.Len() <= 0 {
				continue
			}
			if s.Stdin.String() != "whoami\n" {
				t.Fatal("unexpected input")
			}
			return
		}
	}
}

func TestOutput(t *testing.T) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	cmd, _ := NewShell(&command.Config{
		Socket: ws,
		Client: client,
	})

	sh := cmd.(*Shell)
	defer sh.Close()

	sh.slave.Write([]byte("ls"))
	socket := ws.(*st.MemWriter)

	after := time.After(100 * time.Millisecond)
	expectedMsg := protocol.EncodeServerMsg(&protocol.Message{
		Namespace: "ssh",
		Event:     "p",
		Args:      []interface{}{"ls"},
	})
	for {
		select {
		case <-after:
			t.Fatal("timed out")
			return
		default:
			if len(socket.Stdin) <= 0 {
				continue
			}
			if !reflect.DeepEqual(socket.Stdin[0], expectedMsg) {
				t.Fatal("unexpected response", socket.Stdin[0])
			}
			return
		}
	}
}
