package command

import "github.com/nitrous-io/gorainbow/protocol"

type FileFindCommand struct {
}

func (c *FileFindCommand) Handle(msg *protocol.Message) error {
	return nil
}

func newFileFind(cfg *CommandConfig) (Command, error) {
	return &FileFindCommand{}, nil
}
