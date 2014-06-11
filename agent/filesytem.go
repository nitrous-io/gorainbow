package agent

import (
	"fmt"
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
	return fs.Client.OpenFile(p, os.O_CREATE|os.O_EXCL)
}

func (fs filesystem) Open(p string) (File, error) {
	return fs.Client.Open(p)
}

func (fs filesystem) RemoveDir(p string) error {
	s, err := fs.client.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()
	return s.Run(fmt.Sprintf("rm -rf %s", p))
}
