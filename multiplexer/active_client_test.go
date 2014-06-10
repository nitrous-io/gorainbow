package multiplexer

import "testing"

func TestActiveClient(t *testing.T) {
	client := new(closer)
	active := &activeClient{
		addr:   "localhost",
		client: client,
	}

	if active.addOne(); active.count != 1 {
		t.Fatal("incorrect count", active)
	}
	if active.addOne(); active.count != 2 {
		t.Fatal("incorrect count", active)
	}

	if closed := active.closeOne(); active.count != 1 || closed || client.closed {
		t.Fatal("incorrect close", active)
	}

	if closed := active.closeOne(); active.count != 0 || !closed || !client.closed {
		t.Fatal("incorrect close", active)
	}
}
