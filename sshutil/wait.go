package sshutil

import (
	"log"
	"strings"
	"time"

	"code.google.com/p/go.crypto/ssh"
)

const (
	MaxAttempts = 5
)

func waitForSSH(cfg *ClientConfig, wait time.Duration) (client *ssh.Client, err error) {
	// generate the ssh config
	sshConfig := &ssh.ClientConfig{
		User: cfg.username,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(cfg.signer)},
	}

	// trying to connect to ssh here
	attempts := 0 // number of attemps to connect to ssh
	for {
		select {
		case <-time.After(wait):
		}
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
				continue
			}
			return nil, err
		}
		break
	}
	return client, nil
}
