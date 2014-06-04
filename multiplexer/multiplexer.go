package multiplexer

import (
	"code.google.com/p/go.crypto/ssh"
	"github.com/nitrous-io/gorainbow/sshutil"
)

type Multiplexer struct {
	// mapping of user_id -> ssh.Client
	clients map[int]*ssh.Client

	locker map[int]*lockable
}

func (m *Multiplexer) Get(userId int) *ssh.Client {
	m.lock(userId)
	defer m.unlock(userId)
	c, ok := m.clients[userId]
	if !ok {
		return nil
	}
	return c
}

func (m *Multiplexer) Add(userId int, username, addr string, priv []byte) (*ssh.Client, error) {
	m.lock(userId)
	defer m.unlock(userId)
	cfg, err := sshutil.NewClientConfig(username, addr, priv)
	if err != nil {
		return nil, err
	}
	client, err := sshutil.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	m.clients[userId] = client
	return client, nil
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
		clients: make(map[int]*ssh.Client),
		locker:  make(map[int]*lockable),
	}
}
