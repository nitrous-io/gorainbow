package command

import (
	"fmt"
	"github.com/nitrous-io/gorainbow/agent"
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
	"github.com/nitrous-io/gorainbow/verifier"
)

type (
	Command interface {
		Handle(msg *protocol.Message) error // process data from sockjs connection
		Close()
	}

	Config struct {
		Socket sockjs.Writer
		Client agent.Client
		Signer verifier.SignatureVerifier
		Env    map[string]string
	}
)

type NewCommand func(config *Config) (Command, error)

var commands map[string]NewCommand

func init() {
	commands = map[string]NewCommand{}
}

func InitCommand(namespace string, config *Config) (Command, error) {
	newCommand, ok := commands[namespace]
	if !ok {
		return nil, fmt.Errorf("Invalid command %v", namespace)
	}
	return newCommand(config)
}

func Register(namespace string, init NewCommand) {
	commands[namespace] = init
}
