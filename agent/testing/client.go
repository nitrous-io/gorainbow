// create an in memory client and session that stores/transmit the data
// in-memory.
package testing

import (
	"bytes"

	"github.com/nitrous-io/gorainbow/agent"
)

func init() {
	agent.RegisterClient("testing", NewClient)
}

type Client struct {
	Filesystem *Filesystem
	Session    *Session
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) Wait() error {
	return nil
}

func (c *Client) NewFilesystem() (agent.Filesystem, error) {
	return c.Filesystem, nil
}

func (c *Client) NewSession() (agent.Session, error) {
	if c.Session != nil {
		return c.Session, nil
	}
	s := new(Session)
	s.Stdin = bytes.NewBuffer(nil)
	s.Stdout = bytes.NewBuffer(nil)
	s.Stderr = bytes.NewBuffer(nil)
	c.Session = s
	return s, nil
}

func NewClient(c *agent.ClientConfig) (agent.Client, error) {
	return new(Client), nil
}
