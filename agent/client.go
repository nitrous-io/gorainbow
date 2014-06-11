package agent

import (
	"log"
	"strings"
	"time"

	"code.google.com/p/go.crypto/ssh"
)

const (
	MaxAttempts = 5
	MaxTimeWait = 1 * time.Second
)

type ClientConfig struct {
	username string
	address  string
	signer   ssh.Signer
	timeWait time.Duration
}

// accept username, address and private key data
func NewClientConfig(username, address string, data []byte) (*ClientConfig, error) {
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, err
	}
	return &ClientConfig{
		username: username,
		address:  address,
		signer:   signer,
		timeWait: MaxTimeWait,
	}, nil

}

// Given a `ClientConfig`, this will try to create an ssh connection
func NewClient(cfg *ClientConfig) (*ssh.Client, error) {
	return waitForSSH(cfg)
}

func waitForSSH(cfg *ClientConfig) (client *ssh.Client, err error) {
	// generate the ssh config
	sshConfig := &ssh.ClientConfig{
		User: cfg.username,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(cfg.signer)},
	}

	// trying to connect to ssh here
	attempts := 0 // number of attemps to connect to ssh
	for {
		client, err = ssh.Dial("tcp", cfg.address, sshConfig)
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
				case <-time.After(cfg.timeWait):
				}
				continue
			} else {
				return nil, err
			}
		} else {
			break
		}
	}
	return client, nil
}
