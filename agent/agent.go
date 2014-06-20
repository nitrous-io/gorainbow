package agent

import (
	"fmt"
	"io"
	"os"
	"time"
)

type File interface {
	io.ReadWriteCloser
	Stat() (os.FileInfo, error)
}

type Filesystem interface {
	io.Closer
	Create(path string) (File, error)
	CreateNew(path string) (File, error)
	Open(path string) (File, error)
	Mkdir(path string) error
	Rename(oldname, newname string) error
	Remove(path string) error
	ReadDir(p string) ([]os.FileInfo, error)
}

type Session interface {
	io.Closer
	Output(cmd string) ([]byte, error)
	Run(cmd string) error
	Start(cmd string) error
	Wait() error
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)
	AllPipe() (stdin io.WriteCloser, stdout io.Reader, stderr io.Reader, err error)
	RemoteShell(term string, h, w int) error
	Resize(w, h float64) error
}

type Client interface {
	io.Closer
	Wait() error // wait until the client is closed
	NewSession() (Session, error)
	NewFilesystem() (Filesystem, error)
}

type ClientConfig struct {
	Username   string
	Address    string
	PrivateKey []byte
	Wait       time.Duration
}

type NewClient func(c *ClientConfig) (Client, error)

var Agents map[string]NewClient

func init() {
	Agents = make(map[string]NewClient)
}

func RegisterClient(name string, f NewClient) {
	Agents[name] = f
}

func GetClient(name string, c *ClientConfig) (Client, error) {
	newClient, ok := Agents[name]
	if !ok {
		return nil, fmt.Errorf("agent: invalid name %v", name)
	}
	return newClient(c)
}
