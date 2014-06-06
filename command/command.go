package command

import (
	"database/sql"
	"fmt"

	"github.com/nitrous-io/gorainbow/multiplexer"
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
)

type (
	Command interface {
		Handle(msg *protocol.Message) error // process data from sockjs connection
	}

	CommandConfig struct {
		DB     *sql.DB
		Mulx   *multiplexer.Multiplexer
		WS     sockjs.ReadWriteCloser
		UserId int
		AESKey string
	}
)

type NewCommand func(config *CommandConfig) (Command, error)

var commands map[string]NewCommand

func init() {
	commands = map[string]NewCommand{}
}

func InitCommand(namespace string, config *CommandConfig) (Command, error) {
	newCommand, ok := commands[namespace]
	if !ok {
		return nil, fmt.Errorf("Invalid command %v", namespace)
	}
	return newCommand(config)
}

func Register(namespace string, init NewCommand) {
	commands[namespace] = init
}
