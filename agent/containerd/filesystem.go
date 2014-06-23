package containerd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/nitrous-io/gorainbow/agent"
	"github.com/nitrous-io/gorainbow/pkg/httpfile"
)

type filesystem struct {
	id       string
	username string
	root     string
}

func newFilesystem(id, username, root string) agent.Filesystem {
	f := &filesystem{
		id:       id,
		username: username,
		root:     root,
	}
	return f
}

func (f *filesystem) Close() error {
	return nil
}

func (f *filesystem) createFile(method, path string) (agent.File, error) {
	filename := f.rootPath(path)

	params := url.Values{}
	params.Set("username", f.username)
	params.Set("path", filename)
	endpoint := fmt.Sprintf("/containers/%s/file?%s", f.id, params.Encode())
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return nil, err
	}

	conn, err := getConn()
	if err != nil {
		return nil, err
	}

	return httpfile.NewFileWriter(conn, req, filename)
}

func (f *filesystem) Create(path string) (agent.File, error) {
	return f.createFile("PUT", path)
}

func (f *filesystem) CreateNew(path string) (agent.File, error) {
	return f.createFile("POST", path)
}

func (f *filesystem) Open(path string) (agent.File, error) {
	filename := f.rootPath(path)

	params := url.Values{}
	params.Set("username", f.username)
	params.Set("path", filename)
	endpoint := fmt.Sprintf("/containers/%s/file?%s", f.id, params.Encode())
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	conn, err := getConn()
	if err != nil {
		return nil, err
	}

	return httpfile.NewFileReader(conn, req, filename)
}

func (f *filesystem) Stat(path string) (os.FileInfo, error) {
	filename := f.rootPath(path)

	params := url.Values{}
	params.Set("username", f.username)
	params.Set("path", filename)
	endpoint := fmt.Sprintf("/containers/%s/fileinfo?%s", f.id, params.Encode())
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	conn, err := getConn()
	if err != nil {
		return nil, err
	}

	fs, err := httpfile.NewFileStat(conn, req, filename)
	if err != nil {
		return nil, err
	}
	return fs.Stat()
}

func (f *filesystem) Mkdir(path string) error {
	filename := f.rootPath(path)

	params := url.Values{}
	params.Set("username", f.username)
	params.Set("path", filename)
	endpoint := fmt.Sprintf("/containers/%s/dir?%s", f.id, params.Encode())
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}

	conn, err := getConn()
	if err != nil {
		return err
	}

	resp, err := conn.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("POST: %v failed with resp status %v ", endpoint,
			resp.StatusCode)
	}
	return nil
}

func (f *filesystem) Rename(oldname, newname string) error {
	params := url.Values{}
	params.Set("username", f.username)
	params.Set("from", f.rootPath(oldname))
	params.Set("to", f.rootPath(newname))
	endpoint := fmt.Sprintf("/containers/%s/rename?%s", f.id, params.Encode())
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}

	conn, err := getConn()
	if err != nil {
		return err
	}

	resp, err := conn.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("POST: %v failed with resp status %v ", endpoint,
			resp.StatusCode)
	}
	return nil
}

func (f *filesystem) Remove(path string) error {
	params := url.Values{}
	params.Set("username", f.username)
	params.Set("path", f.rootPath(path))
	endpoint := fmt.Sprintf("/containers/%s/path?%s", f.id, params.Encode())
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	conn, err := getConn()
	if err != nil {
		return err
	}

	resp, err := conn.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("POST: %v failed with resp status %v ", endpoint,
			resp.StatusCode)
	}
	return nil
	return nil
}

func (f *filesystem) ReadDir(path string) ([]os.FileInfo, error) {
	params := url.Values{}
	params.Set("username", f.username)
	params.Set("path", f.rootPath(path))
	endpoint := fmt.Sprintf("/containers/%s/dir?%s", f.id, params.Encode())
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	conn, err := getConn()
	if err != nil {
		return nil, err
	}

	resp, err := conn.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET: %v failed with resp status %v ", endpoint,
			resp.StatusCode)
	}
	defer resp.Body.Close()

	var entries []*httpfile.FileInfo
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}
	var info []os.FileInfo
	for _, e := range entries {
		info = append(info, e)
	}
	return info, nil
}

// return the path that includes root
func (f *filesystem) rootPath(name string) string {
	return filepath.Clean(path.Join(f.root, name))
}
