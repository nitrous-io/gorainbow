package multiplexer

import (
	"sync"
	"testing"
	"time"
)

var clientFunc = func() (WaitCloser, error) {
	return new(closer), nil
}

var clientWaitFunc = func(done chan struct{}) ClientFunc {
	return func() (WaitCloser, error) {
		return &closer{done: done}, nil
	}
}

func TestMultiplexerConcurrentAdd(t *testing.T) {
	w := new(sync.WaitGroup)
	executed := 0
	atomicFunc := func() (WaitCloser, error) {
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
		c != m.clients[1] {
		t.Fatal("failed to add new client")
	}

	m.Add(2, "localhost", clientFunc)
	if len(m.clients) != 2 || len(m.locker) != 2 {
		t.Fatal("failed to add new client")
	}
}

func TestMultiplexerRelease(t *testing.T) {
	m := NewMultiplexer()
	c, _ := m.Add(1, "localhost", clientFunc)

	// release the only reference to that client
	m.Release(1, c)
	if _, ok := m.clients[1]; ok {
		t.Fatalf("should remove reference to that client")
	}

	// release a irrelevant client
	if err := m.Release(1, c); err == nil {
		t.Fatalf("should return error because of stale client")
	}

	// release one of the reference
	m.Add(2, "localhost", clientFunc)
	c, _ = m.Add(2, "localhost", clientFunc)
	m.Release(2, c)
	if c, ok := m.clients[2]; !ok || c.count != 1 {
		t.Fatal("should not closed the client", ok, c)
	}
}

func TestMultiplexerMonitorClient(t *testing.T) {
	m := NewMultiplexer()

	done := make(chan struct{})
	m.Add(1, "localhost", clientWaitFunc(done))
	if len(m.clients) != 1 || len(m.locker) != 1 {
		t.Fatal("failed to add new client")
	}
	if c := m.get(1); c == nil {
		t.Fatal("client should be added")
	}

	done <- struct{}{}
	select {
	case <-time.After(10 * time.Millisecond):
		t.Fatal("timed out before closed")
		return
	case _, ok := <-done:
		if !ok {
			if c := m.get(1); c != nil {
				t.Fatal("client should be released when its done")
			}
		} else {
			t.Fatal("done should be closed")
		}
	}
}
