package agent

import (
	"io"
	"os"
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
	NewSession() (Session, error)
	NewFilesystem() (Filesystem, error)
}
