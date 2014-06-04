package command

import "github.com/nitrous-io/gorainbow/protocol"

type PartsCommand struct {
}

func (c *PartsCommand) Handle(msg *protocol.Message) error {
	return nil
}

func newParts(cfg *CommandConfig) (Command, error) {
	return &PartsCommand{}, nil
}
