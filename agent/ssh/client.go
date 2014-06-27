package agent

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nitrous-io/gorainbow/agent"
	"github.com/pkg/sftp"

	"code.google.com/p/go.crypto/ssh"
)

const (
	MaxAttempts = 3
)

func init() {
	agent.RegisterClient("ssh", NewClient)
}

// represent the ssh connection status. When the connection succeeds, then we
// have the client and err should be nil
type conn struct {
	client *ssh.Client
	err    error
}

type client struct {
	*ssh.Client
}

func (c client) NewSession() (agent.Session, error) {
	s, err := c.Client.NewSession()
	return session{s}, err
}

func (c client) NewFilesystem() (agent.Filesystem, error) {
	sftpClient, err := sftp.NewClient(c.Client)
	if err != nil {
		return nil, err
	}
	return filesystem{sftpClient, c}, nil
}

// Given a `ClientConfig`, this will try to create an ssh connection
func NewClient(c *agent.ClientConfig) (agent.Client, error) {
	var err error
	username := c.Username
	address := c.Address
	data := c.PrivateKey
	wait := c.Wait
	totalTimeWait := MaxAttempts * wait

	// generate a signer from the private key
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, err
	}

	// generate the ssh config
	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}

	ch := make(chan conn)

	go func() {
		// trying to connect to ssh here
		attempts := 0 // number of attemps to connect to ssh
		for {
			sshClient, err := ssh.Dial("tcp", address, sshConfig)
			if err == nil {
				ch <- conn{sshClient, nil}
				return
			}
			log.Printf("Waiting for ssh conn to be ready, got %v", err)
			// rejects failed authenticate and connection refused error
			if strings.Contains(err.Error(), "authenticate") ||
				strings.Contains(err.Error(), "connection refused") {
				log.Printf("Failed to authenticate with remote ssh: %v", err)
				attempts += 1
			}
			if attempts < MaxAttempts {
				time.Sleep(wait)
				continue
			} else {
				ch <- conn{nil, err}
			}
		}
	}()

	select {
	case c := <-ch:
		return client{c.client}, c.err
	case <-time.After(totalTimeWait):
		return nil, fmt.Errorf("ssh: cannot connect to %v", c.Address)
	}
}
