package fs

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"

	at "github.com/nitrous-io/gorainbow/agent/testing"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/command/utils"
	"github.com/nitrous-io/gorainbow/protocol"
	st "github.com/nitrous-io/gorainbow/sockjs/testing"
	"github.com/nitrous-io/gorainbow/verifier"
)

func prepare(dir string) (command.Command, *st.MemWriter) {
	ws := st.NewMemWriter()
	client, _ := at.NewMemClient()
	client.(*at.MemClient).Filesystem = &at.MemFilesystem{
		Root: dir,
	}
	cmd, _ := NewFS(&command.Config{
		Socket: ws,
		Client: client,
		Signer: verifier.NewTokenVerifier("secret"),
		Env:    map[string]string{"downloadDir": dir},
	})

	return cmd, ws.(*st.MemWriter)
}

func TestList(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("/tmp", "list")
	os.MkdirAll(tmpDir+"/"+"test", 0755)
	defer os.RemoveAll(tmpDir)
	cmd, ws := prepare(tmpDir)

	cmd.Handle(&protocol.Message{
		Event:      "list",
		CallbackId: 42,
		Args:       []interface{}{"."},
	})

	stdin := ws.Stdin
	expected := protocol.EncodeServerMsg(utils.MakeCallbackMsg(
		42, nil, ".", []interface{}{
			[]interface{}{"test", true, 4096},
		},
	))

	res, _ := json.Marshal(stdin[0])
	expt, _ := json.Marshal(expected)
	if !bytes.Equal(res, expt) {
		t.Fatal("Unexpected response", string(res), string(expt))
	}
}

func TestRead(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("/tmp", "read")
	ioutil.WriteFile(path.Join(tmpDir, "hello.txt"), []byte("Hello World"), 0755)
	defer os.RemoveAll(tmpDir)
	cmd, ws := prepare(tmpDir)

	cmd.Handle(&protocol.Message{
		Event:      "read",
		CallbackId: 42,
		Args:       []interface{}{"hello.txt"},
	})

	stdin := ws.Stdin
	expected := protocol.EncodeServerMsg(utils.MakeCallbackMsg(
		42, nil, "hello.txt", "text", "Hello World",
	))

	res, _ := json.Marshal(stdin[0])
	expt, _ := json.Marshal(expected)
	if !bytes.Equal(res, expt) {
		t.Fatal("Unexpected response", string(res), string(expt))
	}
}

func TestWrite(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("/tmp", "write")
	ioutil.WriteFile(path.Join(tmpDir, "hello.txt"), []byte("Hello World"), 0755)
	defer os.RemoveAll(tmpDir)
	cmd, _ := prepare(tmpDir)

	cmd.Handle(&protocol.Message{
		Event:      "write",
		CallbackId: 42,
		Args:       []interface{}{"hello.txt", "Hello rainbow"},
	})

	expected, _ := ioutil.ReadFile(path.Join(tmpDir, "hello.txt"))
	if string(expected) != "Hello rainbow" {
		t.Fatal("Unexpected response", string(expected))
	}
}

func TestWriteNew(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("/tmp", "writeNew")
	defer os.RemoveAll(tmpDir)
	cmd, _ := prepare(tmpDir)

	cmd.Handle(&protocol.Message{
		Event:      "writeNew",
		CallbackId: 42,
		Args:       []interface{}{"hello.txt", "Hello rainbow"},
	})

	expected, _ := ioutil.ReadFile(path.Join(tmpDir, "hello.txt"))
	if string(expected) != "Hello rainbow" {
		t.Fatal("Unexpected response", string(expected))
	}
}

func TestMakeDir(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("/tmp", "makedir")
	defer os.RemoveAll(tmpDir)
	cmd, _ := prepare(tmpDir)

	cmd.Handle(&protocol.Message{
		Event:      "makeDir",
		CallbackId: 42,
		Args:       []interface{}{"hello"},
	})

	stat, _ := os.Stat(path.Join(tmpDir, "hello"))
	if !stat.IsDir() || stat.Name() != "hello" {
		t.Fatal("Unexpected response", stat)
	}
}

func TestDeleteDir(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("/tmp", "deletedir")
	defer os.RemoveAll(tmpDir)
	cmd, _ := prepare(tmpDir)

	// preload session to be checked later
	client := cmd.(*FsCmd).sshClient.(*at.MemClient)
	s, _ := client.NewSession()

	cmd.Handle(&protocol.Message{
		Event:      "deleteDir",
		CallbackId: 42,
		Args:       []interface{}{"tobedeleted"},
	})

	if s.(*at.MemSession).Cmd != `rm -rf 'tobedeleted'` {
		t.Fatal("Unexpected command", s.(*at.MemSession).Cmd)
	}
}

func TestDeleteFile(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("/tmp", "deletefile")
	ioutil.WriteFile(path.Join(tmpDir, "hello.txt"), []byte("Hello World"), 0755)
	defer os.RemoveAll(tmpDir)
	cmd, _ := prepare(tmpDir)

	cmd.Handle(&protocol.Message{
		Event:      "deleteFile",
		CallbackId: 42,
		Args:       []interface{}{"hello.txt"},
	})

	stat, err := os.Stat(path.Join(tmpDir, "hello.txt"))
	if !os.IsNotExist(err) {
		t.Fatal("Unexpected response", err, stat)
	}
}

func TestRename(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("/tmp", "rename")
	ioutil.WriteFile(path.Join(tmpDir, "hello.txt"), []byte("Hello World"), 0755)
	defer os.RemoveAll(tmpDir)
	cmd, _ := prepare(tmpDir)

	cmd.Handle(&protocol.Message{
		Event:      "rename",
		CallbackId: 42,
		Args:       []interface{}{"hello.txt", "world.txt"},
	})

	stat, err := os.Stat(path.Join(tmpDir, "hello.txt"))
	if !os.IsNotExist(err) {
		t.Fatal("Unexpected response", err, stat)
	}

	stat, err = os.Stat(path.Join(tmpDir, "world.txt"))
	if err != nil {
		t.Fatal("Unexpected response", err, stat)
	}
}

func TestUnzipCommand(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("/tmp", "unzip")
	defer os.RemoveAll(tmpDir)
	cmd, _ := prepare(tmpDir)

	client := cmd.(*FsCmd).sshClient.(*at.MemClient)
	s, _ := client.NewSession()

	cmd.Handle(&protocol.Message{
		Event:      "unzip",
		CallbackId: 42,
		Args:       []interface{}{"archive.zip"},
	})

	if cmd := s.(*at.MemSession).Cmd; cmd != "unzip -nq 'archive.zip' -x '__MACOSX/*'" {
		t.Fatal("Unexpected command", cmd)
	}
}
