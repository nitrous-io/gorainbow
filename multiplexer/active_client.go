package multiplexer

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
)

type WaitCloser interface {
	io.Closer
	Wait() error
}

type ActiveClient struct {
	id    string
	addr  string
	count int

	Client WaitCloser
}

func newActiveClient(addr string, client WaitCloser) *ActiveClient {
	a := &ActiveClient{
		id:     randomString(),
		addr:   addr,
		Client: client,
	}
	a.addOne()
	return a
}

func (a *ActiveClient) addOne() {
	log.Printf("multiplexer: add %v", a.addr)
	a.count++
}

// return true if the client will be closed, false if it will remain open
func (a *ActiveClient) closeOne() bool {
	log.Printf("multiplexer: release %v", a.addr)
	a.count--
	if a.count <= 0 {
		if a.Client != nil {
			log.Printf("multiplexer: close %v", a.addr)
			a.Client.Close()
		}
		return true
	}
	return false
}

func randomString() string {
	id := make([]byte, 32)

	if _, err := io.ReadFull(rand.Reader, id); err != nil {
		panic(err) // This shouldn't happen
	}
	return hex.EncodeToString(id)
}
