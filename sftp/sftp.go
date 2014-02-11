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
	"io/ioutil"
	"os"
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

	FileData struct {
		File     *os.File
		FileType string
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

func (e FileEntry) ToArr() []interface{} {
	return []interface{}{e.Name, e.IsDir, e.FileSize}
}

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

// return *FileData which contains the open file.
// Caller of this function will have to be responsible for closing and removing
// the temp file once it finishes with the data.
//
// In case of error, the tmp file will be removed by the function. You dont
// have to do anything
func (c *Conn) Get(filePath string) (*FileData, error) {
	var tmpFile *os.File
	var fileType string
	tmpFile, err := ioutil.TempFile("", "sftp")
	if err != nil {
		return nil, err
	}
	tmpName := tmpFile.Name()
	data, err := c.runCommand(fmt.Sprintf("get %s %s", escape(filePath),
		escape(tmpName)))
	if err != nil || validateGetData(data) != nil {
		goto failed
	}
	if IsKnownTextFile(tmpName) {
		return &FileData{File: tmpFile, FileType: "text"}, nil
	}
	if fileType, err = GetFileType(tmpName); err != nil {
		goto failed
	}
	return &FileData{File: tmpFile, FileType: fileType}, nil
failed:
	tmpFile.Close()
	os.Remove(tmpName)
	return nil, fmt.Errorf("sftp: Failed to get %s", filePath)
}

func (c *Conn) Rename(oldPath, newPath string) error {
	return c.doBlankResponse("rename", oldPath, newPath)
}

func (c *Conn) Put(localPath, remotePath string) (err error) {
	stat, err := os.Stat(localPath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("sftp: Put non file %s", localPath)
	}
	data, err := c.runCommand(fmt.Sprintf("put %s %s", escape(localPath),
		escape(remotePath)))
	if err != nil {
		return fmt.Errorf("sftp: Put local %s to remote %s failed, err: %v",
			localPath, remotePath, err)
	}
	lines := strings.Split(data, "\n")
	if len(lines) < 2 {
		return fmt.Errorf("sftp: Put local %s to remote %s failed",
			localPath, remotePath)
	}
	matched, err := regexp.MatchString(`^Uploading\s`, lines[1])
	if err != nil {
		return fmt.Errorf("sftp: Put local %s to remote %s failed, err: %v",
			localPath, remotePath, err)
	}
	if !matched {
		return fmt.Errorf("sftp: failed to put %s to %s", localPath, remotePath)
	}
	return nil
}

func (c *Conn) PutData(remotePath string, contentData []byte) error {
	var tmpFile *os.File
	tmpFile, err := ioutil.TempFile("", "sftp")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		os.Remove(tmpName)
	}()
	_, err = tmpFile.Write(contentData)
	if err != nil {
		return err
	}
	err = c.Put(tmpName, remotePath)
	if err != nil {
		return err
	}
	return nil
}

func (c *Conn) Rm(pathName string) error {
	data, err := c.runCommand(fmt.Sprintf("rm %s", escape(pathName)))
	if err != nil {
		return fmt.Errorf("sftp: Remove remote %s failed, err: %v",
			pathName, err)
	}
	lines := strings.Split(data, "\n")
	if len(lines) < 2 {
		return fmt.Errorf("sftp: Remove remote %s failed", pathName)
	}
	matched, err := regexp.MatchString(`^Removing\s`, lines[1])
	if err != nil || !matched {
		return fmt.Errorf("sftp: Remove remote %s failed", pathName)
	}
	return nil
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

func validateGetData(data string) error {
	lines := strings.Split(data, "\n")
	if len(lines) != 3 {
		return fmt.Errorf("sftp: Get %v\n",
			strings.Join(lines[1:len(lines)-1], "\n"))
	}
	return nil
}
