package multiplexer

import "sync"

type lockable struct {
	lock   sync.Mutex
	locked bool
}

func (l *lockable) Lock() {
	l.locked = true
	l.lock.Lock()
}

func (l *lockable) Unlock() {
	l.lock.Unlock()
	l.locked = false
}
