package filefind

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

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
	cmd, _ := NewFilefindCmd(&command.Config{
		Socket: ws,
		Client: client,
	})
	return cmd, ws.(*testsockjs.Writer), client.(*testagent.Client), s.(*testagent.Session)
}

// makeResponses returns the given messages as expected responses sandwiched
// between the initial global callback response and the final commandComplete.
// `status` argument is exit code of commandComplete message
func makeResponses(status string, messages ...*protocol.Message) [][]interface{} {
	responses := [][]interface{}{
		protocol.EncodeServerMsg(&protocol.Message{
			Namespace: "global",
			Event:     "callback",
			Args:      []interface{}{42},
		}),
	}
	for _, m := range messages {
		responses = append(responses, protocol.EncodeServerMsg(m))
	}
	responses = append(responses,
		protocol.EncodeServerMsg(&protocol.Message{
			Namespace: "filefind",
			Event:     "commandComplete",
			Args:      []interface{}{fmt.Sprintf("are\\\\.!!\n%s", status)},
		}),
	)
	return responses
}

func TestFindCmd(t *testing.T) {
	cmd, _, _, s := prepare()

	cmd.Handle(&protocol.Message{
		Event: "filefind",
		Args:  []interface{}{"are", "."},
	})

	expectedCmd := `sh -c '/usr/local/bin/ag -i -g .*?a.*?r.*?e.*? ./. | /usr/local/bin/matcher are'`
	if s.Cmd != expectedCmd {
		t.Fatal("Failed to send correct cmd", s.Cmd, expectedCmd)
	}
}

func TestFindResponse(t *testing.T) {
	cmd, ws, _, s := prepare()
	s.Stdout = bytes.NewBufferString(`./search/search.go
./dollar/search.go
`)

	go cmd.Handle(&protocol.Message{
		Event:      "filefind",
		CallbackId: 42,
		Args:       []interface{}{"are", "."},
	})

	responses := makeResponses("0",
		&protocol.Message{
			Namespace: "filefind",
			Event:     "commandOutput",
			Args: []interface{}{`are\\.!!
./search/search.go
./dollar/search.go
`},
		},
	)

	after := time.After(120 * time.Millisecond)
	for {
		select {
		case <-after:
			t.Fatal("timed out")
			return
		default:
			time.Sleep(50 * time.Millisecond)
			if len(ws.Stdin) < len(responses) {
				continue
			}
			for i, r := range ws.Stdin {
				if !reflect.DeepEqual(r, responses[i]) {
					t.Fatal("unexpected response", r, responses[i])
				}
			}
			return
		}
	}
}

func TestKnownErrorHandling(t *testing.T) {
	var knownErrors = []struct {
		expected     error
		sampleStderr string
	}{
		{ErrBadPath, "ERR: Error stat()ing: workspace/nopath\n"},
		{ErrNoBinary, "matcher: not found\n"},
	}

	for _, c := range knownErrors {
		cmd, ws, _, s := prepare()
		s.Stderr = bytes.NewBufferString(c.sampleStderr)

		go cmd.Handle(&protocol.Message{
			Event:      "filefind",
			CallbackId: 42,
			Args:       []interface{}{"are", "."},
		})

		responses := makeResponses("1",
			&protocol.Message{
				Namespace: "filefind",
				Event:     "commandError",
				Args:      []interface{}{fmt.Sprintf("are\\\\.!!\n%s", c.expected)},
			},
		)
		after := time.After(110 * time.Millisecond)
		for {
			select {
			case <-after:
				t.Fatal("timed out")
				return
			default:
				time.Sleep(50 * time.Millisecond)
				if len(ws.Stdin) < len(responses) {
					continue
				}
				for i, r := range ws.Stdin {
					if !reflect.DeepEqual(r, responses[i]) {
						t.Error("unexpected response", r, responses[i])
					}
				}
				return
			}
		}
	}
}
