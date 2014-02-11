package sftp

import (
	"bytes"
	"io"
	. "launchpad.net/gocheck"
	"os"
	"sync"
)

type SftpSuite struct{}

var _ = Suite(&SftpSuite{})

func (s *SftpSuite) TestEscape(c *C) {
	str := "3'o clock at harry's"
	expected := "'3'\"'\"'o clock at harry'\"'\"'s'"
	c.Assert(escape(str), Equals, expected)
}

func (s *SftpSuite) TestBufferDataUntilPrompt(c *C) {
	r, w := io.Pipe()
	conn := &Conn{ptyout: r}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer func() {
			w.Close()
		}()
		defer wg.Done()
		io.WriteString(w, "ls\r\n")
		io.WriteString(w, "foo\r\n")
		io.WriteString(w, "bar\r\nbaz")
		io.WriteString(w, "\r\nqux\r\nsftp> ")
	}()

	content := ""
	wg.Add(1)
	go func() {
		defer func() {
			r.Close()
		}()
		defer wg.Done()
		content, _ = conn.bufferDataUntilPrompt()
	}()
	wg.Wait()
	expected := `ls
foo
bar
baz
qux` + `
sftp>` + " \n"
	c.Assert(content, Equals, expected)
}

func (s *SftpSuite) TestLs(c *C) {
	conn := &Conn{Port: "22"}
	// replace pty with a fixture
	conn.pty, _ = os.Open("./fixtures/ls.txt")
	// setup the pipes so input/output are functional
	conn.setupPipes()

	files, _ := conn.Ls("/home/action")
	c.Assert(files[0], Equals, FileEntry{".", false, 0})
	c.Assert(files[1], Equals, FileEntry{"..", false, 0})
	c.Assert(files[2], Equals, FileEntry{".VERSION", false, 26})
	c.Assert(files[3], Equals, FileEntry{".bashrc", false, 736})
}

func (s *SftpSuite) TestMkdir(c *C) {
	conn := &Conn{Port: "22"}
	conn.pty = bytes.NewBuffer([]byte("sftp> \n"))
	conn.setupPipes()
	conn.cmdin = bytes.NewBuffer([]byte{})

	conn.Mkdir("/home/action/test")
	b := conn.cmdin.(*bytes.Buffer)
	c.Assert(b.String(), Equals, "mkdir '/home/action/test'\n")
}

func (s *SftpSuite) TestRmdir(c *C) {
	conn := &Conn{Port: "22"}
	conn.pty = bytes.NewBuffer([]byte("sftp> \n"))
	conn.setupPipes()
	conn.cmdin = bytes.NewBuffer([]byte{})

	conn.Rmdir("/home/action/test")
	b := conn.cmdin.(*bytes.Buffer)
	c.Assert(b.String(), Equals, "rmdir '/home/action/test'\n")
}

func (s *SftpSuite) TestRename(c *C) {
	conn := &Conn{Port: "22"}
	conn.pty = bytes.NewBuffer([]byte("sftp> \n"))
	conn.setupPipes()
	conn.cmdin = bytes.NewBuffer([]byte{})

	conn.Rename("/home/action/test", "/home/action/newtest")
	b := conn.cmdin.(*bytes.Buffer)
	c.Assert(b.String(), Equals, "rename '/home/action/test' '/home/action/newtest'\n")
}

func (s *SftpSuite) TestRm(c *C) {
	conn := &Conn{Port: "22"}
	conn.pty = bytes.NewBuffer([]byte("sftp> \n"))
	conn.setupPipes()
	conn.cmdin = bytes.NewBuffer([]byte{})

	conn.Rm("/home/action/test")
	b := conn.cmdin.(*bytes.Buffer)
	c.Assert(b.String(), Equals, "rm '/home/action/test'\n")
}
