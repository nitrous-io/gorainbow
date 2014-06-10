package multiplexer

import (
	"io"
	"sync"
	"testing"
)

var clientFunc = func() (io.Closer, error) {
	return new(closer), nil
}

func TestMultiplexerConcurrentAdd(t *testing.T) {
	w := new(sync.WaitGroup)
	executed := 0
	atomicFunc := func() (io.Closer, error) {
		executed++
		return clientFunc()
	}

	m := NewMultiplexer()

	w.Add(1)
	go func() {
		defer w.Done()
		m.Add(1, "localhost", atomicFunc)
	}()
	w.Add(1)
	go func() {
		defer w.Done()
		m.Add(1, "localhost", atomicFunc)
	}()
	w.Wait()
	if executed != 1 {
		t.Fatal("should be atomic but not", executed)
	}
}

func TestMultiplexerAdd(t *testing.T) {
	m := NewMultiplexer()

	m.Add(1, "localhost", clientFunc)
	if len(m.clients) != 1 || len(m.locker) != 1 {
		t.Fatal("failed to add new client")
	}

	c, err := m.Add(1, "localhost", clientFunc)
	if len(m.clients) != 1 || len(m.locker) != 1 || err != nil ||
		c != m.clients[1].client {
		t.Fatal("failed to add new client")
	}

	m.Add(2, "localhost", clientFunc)
	if len(m.clients) != 2 || len(m.locker) != 2 {
		t.Fatal("failed to add new client")
	}
}

func TestMultiplexerRelease(t *testing.T) {
	m := NewMultiplexer()
	m.Add(1, "localhost", clientFunc)

	// release the only reference to that client
	if err := m.Release(1); err != nil {
		t.Fatal("should pass but failed", err)
	}
	_, ok := m.clients[1]
	if ok {
		t.Fatalf("should remove reference to that client")
	}

	// release one of the reference
	m.Add(2, "localhost", clientFunc)
	m.Add(2, "localhost", clientFunc)
	if err := m.Release(2); err != nil {
		t.Fatal("should pass but failed", err)
	}
	c, ok := m.clients[2]
	if !ok || c.count != 1 {
		t.Fatal("should not closed the client", ok, c)
	}
}
