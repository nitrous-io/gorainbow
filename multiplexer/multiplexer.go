package multiplexer

import (
	"fmt"
	"log"
)

type ClientFunc func() (WaitCloser, error)

type Multiplexer struct {
	// mapping of user_id -> ssh.Client
	clients map[int]*ActiveClient

	locker map[int]*lockable
}

func (m *Multiplexer) monitor(id int, c *ActiveClient) {
	err := c.Client.Wait()
	log.Printf("multiplexer: client connection is closed %v", err)
	m.lock(id)
	c = nil
	delete(m.clients, id)
	m.unlock(id)
}

// get the connection associate with the id. If there's an active connection,
// inrease the active count of that connection by 1
func (m *Multiplexer) get(id int) *ActiveClient {
	active, ok := m.clients[id]
	if ok {
		active.addOne()
		return active
	}
	return nil
}

// add or retrive the current client associated with the identifier
func (m *Multiplexer) Add(id int, addr string, getClient ClientFunc) (*ActiveClient, error) {
	m.lock(id)
	defer m.unlock(id)
	active := m.get(id)
	if active == nil {
		c, err := getClient()
		if err != nil {
			return nil, err
		}
		active = newActiveClient(addr, c)
		m.clients[id] = active

		go m.monitor(id, active)
	}
	return active, nil
}

// if after release, we dont maintain any active ssh connections for the box,
// then release it
func (m *Multiplexer) Release(id int, a *ActiveClient) error {
	m.lock(id)
	defer m.unlock(id)
	active, ok := m.clients[id]
	if ok && a.id == active.id {
		if a.closeOne() {
			delete(m.clients, id)
		}
		return nil
	} else {
		return fmt.Errorf("multiplexer: release stale client: %v %v", id, a)
	}
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
		clients: make(map[int]*ActiveClient),
		locker:  make(map[int]*lockable),
	}
}
