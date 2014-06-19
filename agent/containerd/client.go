package containerd

import "github.com/nitrous-io/gorainbow/agent"

const (
	containerdSocket = "/var/run/containerd.sock"
)

func init() {
	agent.RegisterClient("containerd", NewClient)
}

type client struct {
	id       string
	username string
}

func (c client) Close() error {
	return nil
}

func (c client) NewSession() (agent.Session, error) {
	return newSession(c.id, c.username), nil
}

func (c client) NewFilesystem() (agent.Filesystem, error) {
	return &filesystem{id: c.id, username: c.username}, nil
}

// Given a `ClientConfig`, this will try to create an ssh connection
func NewClient(c *agent.ClientConfig) (agent.Client, error) {
	username := c.Username
	id := c.ContainerId

	return client{id: id, username: username}, nil
}
