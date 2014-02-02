// TODO:
// - move this to use sftp package based on golang's ssh package
//
// This replicates functionalities of nitrous-io/sftp.js to simplify the
// porting process.
package sftp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/kr/pty"
)

type response struct {
	output string
	err    error
}

type Conn struct {
	Host   string
	Port   string
	User   string
	Key    string
	Pty    *os.File
	Stdin  io.Writer
	Stdout io.ReadCloser
}

func (c *Conn) Connect() (err error) {
	c.Stdout, c.Stdin = io.Pipe()
	c.Pty, err = pty.Start(c.sftpCmd())
	go func() {
		io.Copy(c.Stdin, c.Pty)
	}()
	// wait until the prompt is ready
	c.bufferDataUntilPrompt()
	return
}

func (c *Conn) Ls(filePath string) {
	data, _ := c.runCommand("ls -la " + escape(filePath))
	fmt.Println(data)
}

func (c *Conn) Mkdir(filePath string) {
}

func (c *Conn) runCommand(command string) (string, error) {
	res := make(chan response, 1)
	go func() {
		o, err := c.bufferDataUntilPrompt()
		res <- response{o, err}
	}()
	io.WriteString(c.Pty, command+"\n")
	r := <-res
	return r.output, r.err
}

func (c *Conn) doBlankResponse(args ...string) error {
	command := args[0]
	paths := []string{}
	for _, path := range args[1:len(args)] {
		paths = append(paths, escape(path))
	}
	output, _ := c.runCommand(fmt.Sprintf("%s %s", command,
		strings.Join(paths, " ")))
	lines := strings.Split(output, "\n")
	if len(lines) == 2 {
		return nil
	} else {
		return fmt.Errorf("sftp: %v", strings.Join(lines, "\n"))
	}
}

func (c *Conn) sftpCmd() *exec.Cmd {
	return exec.Command("sftp",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "StrictHostKeyChecking=no",
		"-o", "PubkeyAuthentication=yes",
		"-o", "PasswordAuthentication=no",
		"-o", "LogLevel=FATAL",
		"-i", "/home/action/.ssh/nopass_id_rsa",
		"-P", c.Port,
		"-q",
		"action@192.168.254.254")
}

func (c *Conn) bufferDataUntilPrompt() (string, error) {
	res := ""
	scanner := bufio.NewScanner(c.Stdout)
	scanner.Split(splitSftp)
	for scanner.Scan() {
		data := scanner.Text()
		strings.Replace(data, "\r", "", -1)
		res += data + "\n"
		matched, err := regexp.MatchString("(^|\n)sftp> $", data)
		if err == nil && matched {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return res, nil
}

func escape(filePath string) string {
	return "'" + strings.Replace(filePath, "'", "'\"'\"'", -1) + "'"
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func splitSftp(data []byte, atEOF bool) (advance int, token []byte, err error) {
	matched, _ := regexp.Match("(^|\n)sftp> $", data)
	if matched {
		return len(data), dropCR(data), nil
	}
	return bufio.ScanLines(data, atEOF)
}
