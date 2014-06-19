package containerd

import (
	"os"

	"github.com/nitrous-io/gorainbow/agent"
)

type filesystem struct {
	id       string
	username string
}

func (f *filesystem) Close() error {
	return nil
}
func (f *filesystem) Create(path string) (agent.File, error) {
	return nil, nil
}
func (f *filesystem) CreateNew(path string) (agent.File, error) {
	return nil, nil
}
func (f *filesystem) Open(path string) (agent.File, error) {
	return nil, nil
}
func (f *filesystem) Mkdir(path string) error {
	return nil
}
func (f *filesystem) Rename(oldname, newname string) error {
	return nil
}
func (f *filesystem) Remove(path string) error {
	return nil
}
func (f *filesystem) ReadDir(p string) ([]os.FileInfo, error) {
	return nil, nil
}
