package containerd

import (
	"fmt"
	"log"
	"net"
	"net/http/httputil"

	"github.com/nitrous-io/gorainbow/agent"
)

const (
	containerdSocket = "/var/run/containerd.sock"
)

func getConn() (*httputil.ClientConn, error) {
	dial, err := net.Dial("unix", containerdSocket)
	if err != nil {
		return nil, fmt.Errorf("Can't connect to containerd daemon %v", containerdSocket)
	}
	return httputil.NewClientConn(dial, nil), nil
}

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

func (c client) Wait() error {
	return nil
}

func (c client) NewSession() (agent.Session, error) {
	return newSession(c.id, c.username), nil
}

func (c client) NewFilesystem() (agent.Filesystem, error) {
	log.Println("hello new filesystem")
	return newFilesystem(c.id, c.username, "/home/action"), nil
}

// Given a `ClientConfig`, this will try to create an ssh connection
func NewClient(c *agent.ClientConfig) (agent.Client, error) {
	username := c.Username
	id := c.ContainerId
	return client{id: id, username: username}, nil
}
