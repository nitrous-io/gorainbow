package filefind

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	at "github.com/nitrous-io/gorainbow/agent/testing"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/protocol"
	st "github.com/nitrous-io/gorainbow/sockjs/testing"
)

func TestFindCmd(t *testing.T) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	cmd, _ := NewFilefindCmd(&command.Config{
		Socket: ws,
		Client: client,
	})

	cmd.Handle(&protocol.Message{
		Event: "filefind",
		Args:  []interface{}{"are", "."},
	})

	s := client.(*at.MemClient).Session
	expectedCmd := `sh -c '/usr/local/bin/ag -i -g .*?a.*?r.*?e.*? ./. | /usr/local/bin/matcher are'`
	if s.Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.Cmd, expectedCmd)
	}
}

func TestFindResponse(t *testing.T) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	s, _ := client.NewSession()
	s.(*at.MemSession).Stdout = bytes.NewBufferString(`./search/search.go
./dollar/search.go
`)
	cmd, _ := NewFilefindCmd(&command.Config{
		Socket: ws,
		Client: client,
	})

	go cmd.Handle(&protocol.Message{
		Event:      "filefind",
		CallbackId: 42,
		Args:       []interface{}{"are", "."},
	})

	responses := [][]interface{}{
		protocol.EncodeServerMsg(&protocol.Message{
			Namespace: "global",
			Event:     "callback",
			Args:      []interface{}{42},
		}),
		protocol.EncodeServerMsg(&protocol.Message{
			Namespace: "filefind",
			Event:     "commandOutput",
			Args: []interface{}{`are\\.!!
./search/search.go
./dollar/search.go
`},
		}),
		protocol.EncodeServerMsg(&protocol.Message{
			Namespace: "filefind",
			Event:     "commandComplete",
			Args: []interface{}{`are\\.!!
0`},
		}),
	}

	socket := ws.(*st.MemWriter)
	after := time.After(120 * time.Millisecond)
	for {
		select {
		case <-after:
			t.Fatal("timed out")
			return
		default:
			time.Sleep(50 * time.Millisecond)
			if len(socket.Stdin) < 3 {
				continue
			}
			for i, r := range socket.Stdin {
				if !reflect.DeepEqual(r, responses[i]) {
					t.Fatal("unexpected response", r, responses[i])
				}
			}
			return
		}
	}
}
