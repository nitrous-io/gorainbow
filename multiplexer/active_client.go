package multiplexer

import (
	"io"
	"log"
)

type activeClient struct {
	addr   string
	client io.Closer
	count  int
}

func (a *activeClient) addOne() {
	log.Printf("multiplexer: add %v", a.addr)
	a.count++
}

// return true if the client will be closed, false if it will remain open
func (a *activeClient) closeOne() bool {
	log.Printf("multiplexer: release %v", a.addr)
	a.count--
	if a.count <= 0 {
		if a.client != nil {
			log.Printf("multiplexer: close %v", a.addr)
			a.client.Close()
		}
		return true
	}
	return false
}
