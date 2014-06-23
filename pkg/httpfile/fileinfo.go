package httpfile

import (
	"os"
	"time"
)

type FileInfo struct {
	StatName    string      `json:"Name"`
	StatSize    int64       `json:"Size"`
	StatIsDir   bool        `json:"IsDir"`
	StatMode    os.FileMode `json:"Mode"`
	StatModTime time.Time   `json:"ModTime"`
}

func (s *FileInfo) Name() string       { return s.StatName }
func (s *FileInfo) Size() int64        { return s.StatSize }
func (s *FileInfo) IsDir() bool        { return s.StatIsDir }
func (s *FileInfo) Mode() os.FileMode  { return s.StatMode }
func (s *FileInfo) ModTime() time.Time { return s.StatModTime }
func (s *FileInfo) Sys() interface{}   { return new(interface{}) }
