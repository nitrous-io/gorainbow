package command

import "github.com/nitrous-io/gorainbow/protocol"

type SearchCommand struct {
}

func (c *SearchCommand) Handle(msg *protocol.Message) error {
	return nil
}

func newSearch(cfg *CommandConfig) (Command, error) {
	return &SearchCommand{}, nil
}
