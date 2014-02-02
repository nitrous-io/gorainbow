package sftp

import (
	"io"
	. "launchpad.net/gocheck"
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
	conn := &Conn{Stdin: w, Stdout: r}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer func() {
			w := conn.Stdin.(*io.PipeWriter)
			w.Close()
		}()
		defer wg.Done()
		io.WriteString(conn.Stdin, "ls\r\n")
		io.WriteString(conn.Stdin, "foo\r\n")
		io.WriteString(conn.Stdin, "bar\r\nbaz")
		io.WriteString(conn.Stdin, "\r\nqux\r\nsftp> ")
	}()

	content := ""
	wg.Add(1)
	go func() {
		defer func() {
			conn.Stdout.Close()
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
sftp> 
`
	c.Assert(content, Equals, expected)
}
