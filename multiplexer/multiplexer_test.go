package multiplexer

import "testing"

func TestMultiplexerAdd(t *testing.T) {
	m := NewMultiplexer()

	client := new(closer)

	m.Add(1, "localhost", client)
	if len(m.clients) != 1 || len(m.locker) != 1 {
		t.Fatal("failed to add new client")
	}

	m.Add(1, "localhost", client)
	if len(m.clients) != 1 || len(m.locker) != 1 {
		t.Fatal("failed to add new client")
	}

	m.Add(2, "localhost", client)
	if len(m.clients) != 2 || len(m.locker) != 2 {
		t.Fatal("failed to add new client")
	}
}

func TestMultiplexerGet(t *testing.T) {
	m := NewMultiplexer()
	m.Add(1, "localhost", new(closer))

	// get a newly added client
	c := m.Get(1)
	active := m.clients[1]
	if c != active.client {
		t.Fatal("failed to get client")
	}
	if active.count != 2 {
		t.Fatal("failed to update reference to client", active)
	}

	// get non-existing client
	c = m.Get(2)
	if c != nil {
		t.Fatalf("should get nil but got %v", c)
	}
}

func TestMultiplexerRelease(t *testing.T) {
	m := NewMultiplexer()
	m.Add(1, "localhost", new(closer))

	// release the only reference to that client
	if err := m.Release(1); err != nil {
		t.Fatal("should pass but failed", err)
	}
	_, ok := m.clients[1]
	if ok {
		t.Fatalf("should remove reference to that client")
	}

	// release one of the reference
	m.Add(2, "localhost", new(closer))
	m.Get(2)
	if err := m.Release(2); err != nil {
		t.Fatal("should pass but failed", err)
	}
	c, ok := m.clients[2]
	if !ok || c.count != 1 {
		t.Fatal("should not closed the client", ok, c)
	}
}
