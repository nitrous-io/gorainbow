package testing

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/nitrous-io/gorainbow/agent"
)

type Filesystem struct {
	Root   string
	Closed bool
}

func (fs *Filesystem) Close() error {
	fs.Closed = true
	os.RemoveAll(fs.Root)
	return nil
}

func (fs *Filesystem) Create(p string) (agent.File, error) {
	return os.Create(path.Join(fs.Root, p))
}

func (fs *Filesystem) CreateNew(p string) (agent.File, error) {
	return os.OpenFile(path.Join(fs.Root, p), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0755)
}

func (fs *Filesystem) Open(p string) (agent.File, error) {
	return os.Open(path.Join(fs.Root, p))
}

func (fs *Filesystem) Mkdir(p string) error {
	return os.Mkdir(path.Join(fs.Root, p), 0755)
}

func (fs *Filesystem) Rename(oldname, newname string) error {
	return os.Rename(path.Join(fs.Root, oldname), path.Join(fs.Root, newname))
}

func (fs *Filesystem) Remove(p string) error {
	return os.Remove(path.Join(fs.Root, p))
}

func (fs *Filesystem) ReadDir(p string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path.Join(fs.Root, p))
}
