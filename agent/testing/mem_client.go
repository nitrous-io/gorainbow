// create an in memory client and session that stores/transmit the data
// in-memory.
package testing

import (
	"bytes"

	"github.com/nitrous-io/gorainbow/agent"
)

type MemClient struct {
	Filesystem *MemFilesystem
	Session    *MemSession
}

func (c *MemClient) Close() error {
	return nil
}

func (c *MemClient) NewFilesystem() (agent.Filesystem, error) {
	return c.Filesystem, nil
}

func (c *MemClient) NewSession() (agent.Session, error) {
	if c.Session != nil {
		return c.Session, nil
	}
	s := new(MemSession)
	s.Stdin = bytes.NewBuffer(nil)
	s.Stdout = bytes.NewBuffer(nil)
	s.Stderr = bytes.NewBuffer(nil)
	c.Session = s
	return s, nil
}

func NewMemClient() (agent.Client, error) {
	return new(MemClient), nil
}
