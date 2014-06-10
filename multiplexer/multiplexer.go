package multiplexer

import "io"

type Multiplexer struct {
	// mapping of user_id -> ssh.Client
	clients map[int]*activeClient

	locker map[int]*lockable
}

// get the connection associate with the id. If there's an active connection,
// inrease the active count of that connection by 1
func (m *Multiplexer) Get(id int) io.Closer {
	m.lock(id)
	defer m.unlock(id)
	active, ok := m.clients[id]
	if ok {
		active.addOne()
		return active.client
	}
	return nil
}

func (m *Multiplexer) Add(id int, addr string, client io.Closer) {
	m.lock(id)
	defer m.unlock(id)
	active := &activeClient{
		addr:   addr,
		client: client,
	}
	active.addOne()
	m.clients[id] = active
}

// if after release, we dont maintain any active ssh connections for the box,
// then release it
func (m *Multiplexer) Release(id int) error {
	m.lock(id)
	defer m.unlock(id)
	active, ok := m.clients[id]
	// pass if we dont maitain the id
	if !ok || active == nil {
		return nil
	}
	if active.closeOne() {
		delete(m.clients, id)
	}
	return nil
}

func (m *Multiplexer) lock(id int) {
	var locker *lockable
	locker, ok := m.locker[id]
	if !ok {
		m.locker[id] = new(lockable)
		locker = m.locker[id]
	}
	locker.Lock()
}

func (m *Multiplexer) unlock(id int) {
	locker, ok := m.locker[id]
	if ok {
		locker.Unlock()
	}
}

func NewMultiplexer() *Multiplexer {
	return &Multiplexer{
		clients: make(map[int]*activeClient),
		locker:  make(map[int]*lockable),
	}
}
