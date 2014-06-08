package command

import (
	"fmt"

	"code.google.com/p/go.crypto/ssh"
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
)

type (
	Command interface {
		Handle(msg *protocol.Message) error // process data from sockjs connection
		Close()
	}
)

type NewCommand func(ws sockjs.Writer, client *ssh.Client) (Command, error)

var commands map[string]NewCommand

func init() {
	commands = map[string]NewCommand{}
}

func InitCommand(namespace string, ws sockjs.Writer, client *ssh.Client) (Command, error) {
	newCommand, ok := commands[namespace]
	if !ok {
		return nil, fmt.Errorf("Invalid command %v", namespace)
	}
	return newCommand(ws, client)
}

func Register(namespace string, init NewCommand) {
	commands[namespace] = init
}
