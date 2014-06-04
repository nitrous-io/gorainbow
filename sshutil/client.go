package sshutil

import (
	"time"

	"code.google.com/p/go.crypto/ssh"
)

var (
	MaxTimeWait = 5 * time.Second
)

type ClientConfig struct {
	username string
	address  string
	signer   ssh.Signer
}

// accept username, address and private key data
func NewClientConfig(username, address string, data []byte) (*ClientConfig, error) {
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, err
	}
	return &ClientConfig{username: username, address: address, signer: signer}, nil

}

// Given a `ClientConfig`, this will try to create an ssh connection
func NewClient(cfg *ClientConfig) (*ssh.Client, error) {
	return waitForSSH(cfg, MaxTimeWait)
}
