// TODO:
// - move this to use sftp package based on golang's ssh package
//
// This replicates functionalities of nitrous-io/sftp.js to simplify the
// porting process.
package sftp

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/kr/pty"
)

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
	c.Pty, err = pty.Start(c.sftpCmd())
	return
}

func (c *Conn) Ls(filePath string) {
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
	for scanner.Scan() {
		data := scanner.Text()
		strings.Replace(data, "\r", "", -1)
		res += data + "\n"
		matched, err := regexp.MatchString("(^|\n)sftp> $", data)
		if err == nil && matched {
			break
		}
	}
	return res, nil
}

func escape(filePath string) string {
	return "'" + strings.Replace(filePath, "'", "'\"'\"'", -1) + "'"
}
