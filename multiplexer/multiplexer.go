package multiplexer

import "io"

type ClientFunc func() (io.Closer, error)

type Multiplexer struct {
	// mapping of user_id -> ssh.Client
	clients map[int]*activeClient

	locker map[int]*lockable
}

// get the connection associate with the id. If there's an active connection,
// inrease the active count of that connection by 1
func (m *Multiplexer) get(id int) io.Closer {
	active, ok := m.clients[id]
	if ok {
		active.addOne()
		return active.client
	}
	return nil
}

// add or retrive the current client associated with the identifier
func (m *Multiplexer) Add(id int, addr string, getClient ClientFunc) (c io.Closer, err error) {
	m.lock(id)
	defer m.unlock(id)
	c = m.get(id)
	if c == nil {
		c, err = getClient()
		if err != nil {
			return nil, err
		}
		active := &activeClient{
			addr:   addr,
			client: c,
		}
		active.addOne()
		m.clients[id] = active
	}
	return c, nil
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
