package multiplexer

import "testing"

func TestLockable(t *testing.T) {
	l := new(lockable)
	l.Lock()
	if !l.locked {
		t.Fatal("should be locked")
	}
	l.Unlock()
	if l.locked {
		t.Fatal("should be unlocked")
	}
}
