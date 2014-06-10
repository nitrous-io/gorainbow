package command

import (
	"fmt"

	"code.google.com/p/go.crypto/ssh"
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
		Client *ssh.Client
		Signer verifier.SignatureVerifier
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
