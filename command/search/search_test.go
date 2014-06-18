package search

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nitrous-io/gorainbow/agent"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/command/utils"
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
			Namespace: "search",
			Event:     "commandComplete",
			Args:      []interface{}{fmt.Sprintf("1!!\n%s", status)},
		}),
	)
	return responses
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

func TestLimitError(t *testing.T) {
	_, ws, client, s := prepare()
	s.Stdout = bytes.NewBufferString(strings.Repeat("line\n", 6))
	gcmd, _ := NewSearchCmd(&command.Config{
		Socket: ws,
		Client: client,
	})
	cmd := gcmd.(*SearchCmd)
	cmd.maxLines = 5

	done := make(chan error, 1)
	go func() {
		done <- cmd.Handle(&protocol.Message{
			Event:      "search",
			CallbackId: 42,
			Args:       []interface{}{"1", "line", ".", "ignoreCase,literal"},
		})
	}()

	responses := makeResponses("1",
		&protocol.Message{
			Namespace: "search",
			Event:     "commandOutput",
			Args:      []interface{}{fmt.Sprintf("1!!\n%s", strings.Repeat("line\n", 4))},
		},
		&protocol.Message{
			Namespace: "search",
			Event:     "commandError",
			Args:      []interface{}{fmt.Sprintf("1!!\n%s", utils.ErrLimitLines)},
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

func TestKnownErrorHandling(t *testing.T) {
	var knownErrors = []struct {
		expected     error
		sampleStderr string
	}{
		{ErrBadPath, "ERR: Error stat()ing: workspace/nopath\n"},
		{ErrBadRegex, "ERR: pcre_compile failed at position 5. Error: missing )\n"},
		{fmt.Errorf(ErrfMaxMatch, "README.md"), "ERR: Too many matches in README.md. Skipping\n"},
	}

	for _, c := range knownErrors {
		cmd, ws, _, s := prepare()
		s.Stderr = bytes.NewBufferString(c.sampleStderr)

		go cmd.Handle(&protocol.Message{
			Event:      "search",
			CallbackId: 42,
			Args:       []interface{}{"1", "hello", ".", "ignoreCase,literal"},
		})

		responses := makeResponses("1",
			&protocol.Message{
				Namespace: "search",
				Event:     "commandError",
				Args:      []interface{}{fmt.Sprintf("1!!\n%s", c.expected)},
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

func TestIdentifyErrorOnUnknownStderr(t *testing.T) {
	err := identifyError("ERR: Something we haven't seen yet\n")
	if err != nil {
		t.Errorf("Expected nil, got %s", err)
	}
}
