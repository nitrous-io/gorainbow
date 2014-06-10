package sshutil

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
	"testing"
	"time"

	"code.google.com/p/go.crypto/ssh"
)

func getPubKey(priv *rsa.PrivateKey) ([]byte, error) {
	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return []byte{}, err
	}
	return ssh.MarshalAuthorizedKey(pub), nil
}

func getPem(priv *rsa.PrivateKey) []byte {
	// ASN.1 DER encoded form
	priv_der := x509.MarshalPKCS1PrivateKey(priv)
	priv_blk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   priv_der,
	}
	return pem.EncodeToMemory(&priv_blk)
}

func generateSSHKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 512)
}

type pubKeyCallback func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error)

func dummyPubKeyCallback(priv *rsa.PrivateKey) pubKeyCallback {
	pub, _ := getPubKey(priv)
	return func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if bytes.Equal(pub, ssh.MarshalAuthorizedKey(key)) {
			return nil, nil
		}
		return nil, errors.New("failed to authenticate with public key")
	}
}

func TestWaitForSSH(t *testing.T) {
	priv, _ := generateSSHKey()
	// start the server
	serverConfig := &ssh.ServerConfig{
		PublicKeyCallback: dummyPubKeyCallback(priv),
	}
	signer, _ := ssh.ParsePrivateKey(getPem(priv))
	serverConfig.AddHostKey(signer)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_, _, reqs, err := ssh.NewServerConn(conn, serverConfig)
			if err != nil {
				continue
			}
			go ssh.DiscardRequests(reqs)
		}
	}()

	// creating an ssh connection to the server
	config, _ := NewClientConfig("root", listener.Addr().String(), getPem(priv))
	config.timeWait = 10 * time.Millisecond
	client, err := NewClient(config)
	if err != nil || client == nil {
		t.Fatal("failed to wait for ssh connection", err)
	}
}
