package multiplexer

import (
	"log"

	"code.google.com/p/go.crypto/ssh"
	"github.com/nitrous-io/gorainbow/sshutil"
)

type activeClient struct {
	client *ssh.Client
	count  int
}

func (a *activeClient) addOne() {
	log.Printf("multiplexer: add %v", a.client.RemoteAddr())
	a.count++
}

// return true if the client will be closed, false if it will remain open
func (a *activeClient) closeOne() bool {
	log.Printf("multiplexer: release %v", a.client.RemoteAddr())
	a.count--
	if a.count <= 0 {
		if a.client != nil {
			log.Printf("multiplexer: close %v", a.client.RemoteAddr())
			a.client.Close()
		}
		return true
	}
	return false
}

type Multiplexer struct {
	// mapping of user_id -> ssh.Client
	clients map[int]*activeClient

	locker map[int]*lockable
}

// get the connection associate with the id. If there's an active connection,
// inrease the active count of that connection by 1
func (m *Multiplexer) Get(id int) *ssh.Client {
	m.lock(id)
	defer m.unlock(id)
	active, ok := m.clients[id]
	if ok && active.count > 0 {
		active.addOne()
		return active.client
	} else {
		delete(m.clients, id)
		return nil
	}
}

func (m *Multiplexer) Add(id int, username, addr string, priv []byte) (*ssh.Client, error) {
	m.lock(id)
	defer m.unlock(id)
	cfg, err := sshutil.NewClientConfig(username, addr, priv)
	if err != nil {
		return nil, err
	}
	client, err := sshutil.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	active := &activeClient{client, 1}
	m.clients[id] = active
	return active.client, nil
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
