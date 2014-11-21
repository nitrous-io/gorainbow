package shell

import (
	"reflect"
	"testing"
	"time"

	"github.com/nitrous-io/gorainbow/agent"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/protocol"
	"golang.org/x/crypto/ssh/terminal"

	testagent "github.com/nitrous-io/gorainbow/agent/testing"
	testsockjs "github.com/nitrous-io/gorainbow/sockjs/testing"
)

func prepare() (command.Command, *testsockjs.Writer, *testagent.Client, *testagent.Session) {
	ws := testsockjs.NewWriter()
	client, _ := testagent.NewClient(new(agent.ClientConfig))
	s, _ := client.NewSession()
	cmd, _ := NewShell(&command.Config{
		Socket: ws,
		Client: client,
	})
	return cmd, ws.(*testsockjs.Writer), client.(*testagent.Client), s.(*testagent.Session)
}

func TestResize(t *testing.T) {
	cmd, _, _, s := prepare()
	cmd.Handle(&protocol.Message{
		Event: "r",
		Args:  []interface{}{float64(100), float64(50)},
	})
	if s.Height != 50 || s.Width != 100 {
		t.Fatal("failed to resize", s)
	}
	termWidth, termHeight, err := terminal.GetSize(int(cmd.(*Shell).slave.Fd()))
	if err != nil || termHeight != 50 || termWidth != 100 {
		t.Fatal("failed to resize", termHeight, termWidth, err)
	}
}

func TestInput(t *testing.T) {
	cmd, _, _, s := prepare()
	cmd.Handle(&protocol.Message{
		Event: "d",
		Args:  []interface{}{"whoami\n"},
	})
	defer cmd.Close()

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
	cmd, ws, _, _ := prepare()

	sh := cmd.(*Shell)
	defer sh.Close()
	sh.slave.Write([]byte("ls"))

	expectedMsg := protocol.EncodeServerMsg(&protocol.Message{
		Namespace: "ssh",
		Event:     "p",
		Args:      []interface{}{"ls"},
	})

	after := time.After(100 * time.Millisecond)
	for {
		select {
		case <-after:
			t.Fatal("timed out")
			return
		default:
			if len(ws.Stdin) <= 0 {
				continue
			}
			if !reflect.DeepEqual(ws.Stdin[0], expectedMsg) {
				t.Fatal("unexpected response", ws.Stdin[0])
			}
			return
		}
	}
}
