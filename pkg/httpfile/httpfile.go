package httpfile

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
)

// data structure that exposes a remote file, and compatible with `agent.File`
// interface
type file struct {
	filename string    // remote path
	reader   io.Reader // read content of the file
	writer   io.Writer // write content to the file
	closer   io.Closer // close the http connection

	stat *FileInfo
}

func (f *file) Read(p []byte) (int, error) {
	return f.reader.Read(p)
}

func (f *file) Write(p []byte) (int, error) {
	return f.writer.Write(p)
}

func (f *file) Close() error {
	if f.closer != nil {
		return f.closer.Close()
	}
	return nil
}

func (f *file) Stat() (os.FileInfo, error) {
	return f.stat, nil
}

func NewFileStat(
	client *httputil.ClientConn,
	req *http.Request,
	filename string,
) (*file, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if !success(resp) {
		return nil, fmt.Errorf("request failed %v", resp.Status)
	}
	st := new(FileInfo)
	if err := json.NewDecoder(resp.Body).Decode(st); err != nil {
		return nil, fmt.Errorf("Invalid stat %v", err)
	}
	return &file{
		filename: filename,
		stat:     st,
	}, nil
}

func NewFileReader(
	client *httputil.ClientConn,
	req *http.Request,
	filename string,
) (*file, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if !success(resp) {
		return nil, fmt.Errorf("request failed %v", resp.Status)
	}
	return &file{
		filename: filename,
		reader:   resp.Body,
		closer:   resp.Body,
	}, nil
}

func NewFileWriter(
	client *httputil.ClientConn,
	req *http.Request,
	filename string,
) (*file, error) {
	client.Do(req)
	rwc, _ := client.Hijack()
	return &file{
		filename: filename,
		writer:   rwc,
		closer:   rwc,
	}, nil
}

func success(r *http.Response) bool {
	return r.StatusCode < 400
}
