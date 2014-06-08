package parts

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/command/utils"
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
)

const (
	listAll         = "parts search --json && parts list --json"
	status          = "parts status --json"
	timeout         = 1 * time.Minute
	defaultUsername = "action"
)

func init() {
	command.Register("parts", NewPartsCmd)
}

type Part struct {
	Name             string   `json:"name"`
	Description      string   `json:"description,omitempty"`
	Category         string   `json:"category,omitempty"`
	Versions         []string `json:"versions,omitempty"`
	InstalledVersion string   `json:"installed_version,omitempty"`
	Installed        bool     `json:"installed,omitempty"`
	Status           string   `json:"status,omitempty"`
}

type PartsCmd struct {
	ws        sockjs.Writer
	client    *ssh.Client
	cmdOutput *utils.OutputWriter
	cmdError  *utils.ErrorWriter
}

func NewPartsCmd(ws sockjs.Writer, client *ssh.Client) (command.Command, error) {
	return &PartsCmd{
		ws:        ws,
		client:    client,
		cmdOutput: utils.NewOutputWriter("parts", ws),
		cmdError:  utils.NewErrorWriter("parts", ws),
	}, nil
}

func (c *PartsCmd) Handle(msg *protocol.Message) error {
	if msg.Event == "listAll" {
		parts, err := c.listAll()
		if err != nil {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, err))
		} else {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, nil, parts))
		}
	} else if msg.Event == "status" {
		parts, err := c.status()
		if err != nil {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, err))
		} else {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, nil, parts))
		}
	} else if msg.Event == "install" {
		partName := msg.Args[0].(string)
		c.streamAction("install", partName)
	} else if msg.Event == "uninstall" {
		partName := msg.Args[0].(string)
		c.streamAction("uninstall", partName)
	} else if msg.Event == "partStart" {
		partName := msg.Args[0].(string)
		c.streamAction("start", partName)
	} else if msg.Event == "partStop" {
		partName := msg.Args[0].(string)
		c.streamAction("start", partName)
	}
	return nil
}

func (c *PartsCmd) streamAction(command, partName string) {
	code := 0
	if err := c.stream(command, partName); err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			code = exitErr.ExitStatus()
		} else {
			code = 1
		}
	}
	c.writeOne(&protocol.Message{
		Namespace: "parts",
		Event:     "commandComplete",
		Args:      []interface{}{code},
	})
}

func (c *PartsCmd) stream(command, partName string) error {
	// open a new session
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("parts %s %s", command, utils.EscapeArg(partName))
	// flush stdout/stderr to the socket directly
	session.Stdout = c.cmdOutput
	session.Stderr = c.cmdError

	// start and wait for cmd to finish or timeout
	if err := session.Start(cmd); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()
	select {
	case <-time.After(timeout):
		if err := session.Close(); err != nil {
			return fmt.Errorf("Session closed with err: %v.Timed out %v", err, cmd)
		}
		return fmt.Errorf("Timed out %v", cmd)
	case err := <-done:
		return err
	}
}

func (c *PartsCmd) status() ([]*Part, error) {
	// open a new session
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	output, err := session.Output(status)
	if err != nil {
		return nil, err
	}
	parts := []*Part{}
	if err := json.Unmarshal(output, &parts); err != nil {
		return nil, fmt.Errorf("Invalid status data %v", err)
	}
	return parts, nil
}

// list current installed parts
func (c *PartsCmd) listAll() ([]*Part, error) {
	// open a new session
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	output, err := toString(session.Output(listAll))
	if err != nil {
		return nil, err
	}
	parts := strings.Split(output, "\n")
	if len(parts) < 2 {
		return nil, fmt.Errorf("Invalid parts data %v", output)
	}

	allParts := []*Part{}
	if err := json.Unmarshal([]byte(parts[0]), &allParts); err != nil {
		return nil, fmt.Errorf("Invalid search data %v", err)
	}

	installedParts := []*Part{}
	if err := json.Unmarshal([]byte(parts[1]), &installedParts); err != nil {
		return nil, fmt.Errorf("Invalid installed data %v", err)
	}

	for _, installed := range installedParts {
		for _, part := range allParts {
			if installed.Name == part.Name {
				part.Installed = true
				part.InstalledVersion = installed.Versions[0]
			}
		}
	}

	return allParts, nil
}

func (c *PartsCmd) writeOne(msg *protocol.Message) {
	utils.WriteOne(c.ws, msg)
}

func toString(d []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return string(d), nil
}
