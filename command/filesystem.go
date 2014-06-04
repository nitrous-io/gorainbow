package command

import "github.com/nitrous-io/gorainbow/protocol"

type FilesystemCommand struct {
}

func (c *FilesystemCommand) Handle(msg *protocol.Message) error {
	return nil
}

func newFilesystem(cfg *CommandConfig) (Command, error) {
	return &FilesystemCommand{}, nil
}
