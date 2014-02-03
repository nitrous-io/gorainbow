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
	conn := &Conn{in: w, out: r}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer func() {
			w := conn.in.(*io.PipeWriter)
			w.Close()
		}()
		defer wg.Done()
		io.WriteString(conn.in, "ls\r\n")
		io.WriteString(conn.in, "foo\r\n")
		io.WriteString(conn.in, "bar\r\nbaz")
		io.WriteString(conn.in, "\r\nqux\r\nsftp> ")
	}()

	content := ""
	wg.Add(1)
	go func() {
		defer func() {
			r := conn.out.(*io.PipeReader)
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
	// replace cmdout with a dummy reader so we dont write into pty
	conn.cmdout = bytes.NewBuffer([]byte{})
	// setup the pipes so input/output are functional
	conn.setupPipes()

	files, _ := conn.Ls("/home/action")
	c.Assert(files[0], Equals, FileEntry{".", false, 0})
	c.Assert(files[1], Equals, FileEntry{"..", false, 0})
	c.Assert(files[2], Equals, FileEntry{".VERSION", false, 26})
	c.Assert(files[3], Equals, FileEntry{".bashrc", false, 736})
}
