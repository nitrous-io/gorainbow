package agent

import (
	"log"
	"strings"
	"time"

	"github.com/pkg/sftp"

	"code.google.com/p/go.crypto/ssh"
)

const (
	MaxAttempts = 5
	MaxTimeWait = 1 * time.Second
)

type client struct {
	*ssh.Client
}

func (c client) NewSession() (Session, error) {
	s, err := c.Client.NewSession()
	return session{s}, err
}

func (c client) NewFilesystem() (Filesystem, error) {
	sftpClient, err := sftp.NewClient(c.Client)
	if err != nil {
		return nil, err
	}
	return filesystem{sftpClient, c}, nil
}

// Given a `ClientConfig`, this will try to create an ssh connection
func NewClient(username, address string, data []byte, wait time.Duration) (Client, error) {
	var err error
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

	// trying to connect to ssh here
	attempts := 0 // number of attemps to connect to ssh
	var sshClient *ssh.Client
	for {
		sshClient, err = ssh.Dial("tcp", address, sshConfig)
		if err != nil {
			log.Printf("Waiting for ssh conn to be ready, got %v", err)
			// Only count this as an attempt if we were able to attempt
			// to authenticate. Note this is very brittle since it depends
			// on the string of the error... but I don't see any other way.
			if strings.Contains(err.Error(), "authenticate") {
				log.Printf("Failed to authenticate with remote ssh: %v", err)
				attempts += 1
			}
			if attempts < MaxAttempts {
				select {
				case <-time.After(wait):
				}
				continue
			} else {
				return nil, err
			}
		} else {
			break
		}
	}
	return client{sshClient}, nil
}
