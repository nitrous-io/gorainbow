package agent

import (
	"os"

	"github.com/pkg/sftp"
)

type filesystem struct {
	*sftp.Client
	client Client
}

func (fs filesystem) Create(p string) (File, error) {
	return fs.Client.Create(p)
}

func (fs filesystem) CreateNew(p string) (File, error) {
	return fs.Client.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_EXCL)
}

func (fs filesystem) Open(p string) (File, error) {
	return fs.Client.Open(p)
}
