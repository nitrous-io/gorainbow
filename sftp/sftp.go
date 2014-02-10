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
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/kr/pty"
)

type (
	FileEntry struct {
		Name     string
		IsDir    bool
		FileSize int
	}

	response struct {
		output string
		err    error
	}

	Conn struct {
		Host   string
		Port   string
		User   string
		Key    string
		pty    io.ReadWriter // *os.File acts like ReadWriter
		cmdin  io.Writer
		ptyout io.Reader
	}
)

var lineRegexp = regexp.MustCompile(`^\s*([a-z\-])([rwx\-]+)\s+([\d]+)\s+([\w]+)\s+([\w]+)\s+([\d]+)\s+([\w\s\d]+)([\d]{2}\:?[\d]{2})\s+(.*)$`)

func (c *Conn) Connect() (err error) {
	c.pty, err = pty.Start(c.sftpCmd())
	c.setupPipes()
	// FIXME: should have timeout here
	// wait until the prompt is ready
	c.bufferDataUntilPrompt()
	return
}

func (c *Conn) Ls(filePath string) (files []FileEntry, err error) {
	data, err := c.runCommand("ls -la " + escape(filePath))
	if err != nil {
		return
	}
	lines := strings.Split(data, "\n")
	errors := []error{}
	for _, line := range lines[1 : len(lines)-2] {
		matched := lineRegexp.FindStringSubmatch(line)
		if len(matched) < 10 {
			errors = append(errors, fmt.Errorf("sftp: Ls failed to parse %s", line))
		} else {
			entry := FileEntry{}
			entry.Name = matched[9]
			if entry.Name != "." && entry.Name != ".." {
				entry.IsDir = matched[1] == "d"
				entry.FileSize, _ = strconv.Atoi(matched[6])
			}
			files = append(files, entry)
		}
	}
	err = fmt.Errorf("sftp: Ls failed %v", errors)
	return
}

func (c *Conn) Mkdir(filePath string) error {
	return c.doBlankResponse("mkdir", filePath)
}

func (c *Conn) Rmdir(filePath string) error {
	return c.doBlankResponse("rmdir", filePath)
}

func (c *Conn) runCommand(command string) (string, error) {
	res := make(chan response, 1)
	go func() {
		o, err := c.bufferDataUntilPrompt()
		res <- response{o, err}
	}()
	io.WriteString(c.cmdin, command+"\n")
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

func (c *Conn) setupPipes() {
	var in *io.PipeWriter
	var ptyin *io.PipeReader
	c.ptyout, in = io.Pipe()
	ptyin, c.cmdin = io.Pipe()
	go func() {
		io.Copy(in, c.pty)
	}()
	go func() {
		io.Copy(c.pty, ptyin)
	}()
}

func (c *Conn) bufferDataUntilPrompt() (string, error) {
	res := ""
	scanner := bufio.NewScanner(c.ptyout)
	scanner.Split(splitSftp)
	for scanner.Scan() {
		data := scanner.Text()
		data = strings.Replace(data, "\r", "", -1)
		res += data + "\n"
		matched, err := regexp.MatchString(`(^|\n)sftp>\s*$`, data)
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
