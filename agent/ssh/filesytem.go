package agent

import (
	"os"

	"github.com/nitrous-io/gorainbow/agent"
	"github.com/pkg/sftp"
)

type filesystem struct {
	*sftp.Client
	client client
}

func (fs filesystem) Create(p string) (agent.File, error) {
	return fs.Client.Create(p)
}

func (fs filesystem) CreateNew(p string) (agent.File, error) {
	return fs.Client.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_EXCL)
}

func (fs filesystem) Open(p string) (agent.File, error) {
	return fs.Client.Open(p)
}

func (fs filesystem) Stat(p string) (os.FileInfo, error) {
	return fs.Client.Lstat(p)
}
