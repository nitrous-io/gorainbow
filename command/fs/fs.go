package fs

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"code.google.com/p/go.crypto/ssh"
	"github.com/nitrous-io/gorainbow/command"
	"github.com/nitrous-io/gorainbow/command/utils"
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
	"github.com/nitrous-io/gorainbow/verifier"
	"github.com/pkg/sftp"
)

const (
	textContent     = "text"
	pngContent      = "image/png"
	jpgContent      = "image/jpg"
	gifContent      = "image/gif"
	iconContent     = "image/x-icon"
	sizeLimit       = 2 * 1024 * 1024 // 2 megabytes
	defaultUsername = "action"
)

var (
	ErrFileExist    = errors.New("aio:file_exists")
	ErrFileTooLarge = errors.New("aio:file_too_large")
	ErrNotTextFile  = errors.New("aio:not_text_file")
)

func init() {
	command.Register("sftp", NewFS)
}

type uploadToken struct {
	Name string `json:"fileName"`
	Path string `json:"tmpFilePath"`
}

type FsCmd struct {
	ws        sockjs.Writer
	client    *sftp.Client
	sshClient *ssh.Client
	signer    verifier.SignatureVerifier
}

func NewFS(config *command.Config) (command.Command, error) {
	if config.Signer == nil {
		return nil, errors.New("Signer is missing")
	}
	sftpClient, err := sftp.NewClient(config.Client)
	if err != nil {
		return nil, err
	}
	return &FsCmd{
		ws:        config.Socket,
		client:    sftpClient,
		sshClient: config.Client,
		signer:    config.Signer,
	}, nil
}

func (c *FsCmd) Close() {
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
}

func (c *FsCmd) Handle(msg *protocol.Message) error {
	if msg.Event == "list" {
		dir := msg.Args[0].(string)
		files := [][]interface{}{}
		entries, err := c.list(dir)
		if err != nil {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, err, files))
		} else {
			for _, f := range entries {
				files = append(files, fileInfoArgs(f))
			}
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, nil, dir, files))
		}
	} else if msg.Event == "read" {
		fileName := msg.Args[0].(string)
		info, err := c.read(fileName)
		if err != nil {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, err))
		} else {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, nil, info...))
		}
	} else if msg.Event == "download" {
		// TODO: implement download
	} else if msg.Event == "write" {
		fileName := msg.Args[0].(string)
		content := msg.Args[1].(string)
		if err := c.write(fileName, content); err != nil {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, err))
		} else {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, nil, fileName))
		}
	} else if msg.Event == "writeNew" {
		fileName := msg.Args[0].(string)
		content := msg.Args[1].(string)
		if err := c.writeNew(fileName, content); err != nil {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, err))
		} else {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, nil, fileName))
		}
	} else if msg.Event == "makeDir" {
		dirName := msg.Args[0].(string)
		if err := c.mkdir(dirName); err != nil {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, err))
		} else {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, nil, dirName))
		}
	} else if msg.Event == "deleteDir" {
		dirName := msg.Args[0].(string)
		if err := c.rmdir(dirName); err != nil {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, err))
		} else {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, nil, dirName))
		}
	} else if msg.Event == "deleteFile" {
		fileName := msg.Args[0].(string)
		if err := c.deleteFile(fileName); err != nil {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, err))
		} else {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, nil, fileName))
		}
	} else if msg.Event == "rename" {
		oldPath := msg.Args[0].(string)
		newPath := msg.Args[1].(string)
		if err := c.rename(oldPath, newPath); err != nil {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, err))
		} else {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, nil, oldPath))
		}
	} else if msg.Event == "upload" {
		remotePath := msg.Args[0].(string)
		token := msg.Args[1].(string)
		if err := c.upload(remotePath, token); err != nil {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, err))
		} else {
			c.writeOne(utils.MakeCallbackMsg(msg.CallbackId, nil, remotePath))
		}
	}
	return nil
}

func (c *FsCmd) upload(remotePath, token string) error {
	data, err := c.signer.Verify(token)
	if err != nil {
		return err
	}
	t := new(uploadToken)
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	return c.put(remotePath, t.Path)
}

func (c *FsCmd) put(remotePath, localPath string) error {
	var rf *sftp.File
	var lf *os.File
	var err error
	if lf, err = os.Open(localPath); err != nil {
		return err
	}
	defer func() {
		lf.Close()
		os.Remove(localPath)
	}()
	if rf, err = c.client.Create(remotePath); err != nil {
		return err
	}
	_, err = io.Copy(rf, lf)
	return err
}

func (c *FsCmd) rmdir(dirName string) error {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	cmd := fmt.Sprintf("rm -rf %s", utils.EscapeArg(dirName))
	return session.Run(cmd)
}

func (c *FsCmd) rename(oldPath, newPath string) error {
	return c.client.Rename(utils.GetPath(oldPath), utils.GetPath(newPath))
}

func (c *FsCmd) deleteFile(fileName string) error {
	return c.client.Remove(utils.GetPath(fileName))
}

func (c *FsCmd) mkdir(dir string) error {
	return c.client.Mkdir(utils.GetPath(dir))
}

func (c *FsCmd) write(path, content string) error {
	f, err := c.client.Create(utils.GetPath(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewBufferString(content))
	return err
}

func (c *FsCmd) writeNew(path, content string) error {
	f, err := c.client.OpenFile(utils.GetPath(path), os.O_CREATE|os.O_EXCL)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewBufferString(content))
	return err
}

func (c *FsCmd) read(path string) ([]interface{}, error) {
	// open the file
	f, err := c.client.Open(utils.GetPath(path))
	if err != nil {
		return nil, err
	}
	// get stat of the file
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	// check if the size is too large
	if stat.Size() > sizeLimit {
		return nil, ErrFileTooLarge
	}
	// get content of the file
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	// trying to guess content type
	var contentType string
	if stat.Size() == 0 {
		contentType = textContent
	} else {
		contentType, err = getContent(data)
		if err != nil {
			return nil, err
		}
	}
	// from the content type, encode data properly
	var content string
	switch contentType {
	case textContent:
		content = string(data)
	case pngContent, gifContent, jpgContent, iconContent:
		content = base64.StdEncoding.EncodeToString(data)
	}
	return []interface{}{path, contentType, content}, nil
}

func (c *FsCmd) list(dir string) ([]os.FileInfo, error) {
	return c.client.ReadDir(utils.GetPath(dir))
}

func (c *FsCmd) writeOne(msg *protocol.Message) {
	utils.WriteOne(c.ws, msg)
}

func fileInfoArgs(info os.FileInfo) []interface{} {
	return []interface{}{
		info.Name(),
		info.IsDir(),
		info.Size(),
	}
}

func getContent(data []byte) (string, error) {
	var contentType string
	detectedContent := http.DetectContentType(data)
	// if the file is a text file or empty
	if strings.Contains(detectedContent, "text") {
		contentType = textContent
	} else if detectedContent == "image/png" {
		contentType = pngContent
	} else if detectedContent == "image/gif" {
		contentType = gifContent
	} else if detectedContent == "image/jpeg" {
		contentType = jpgContent
	} else if strings.Contains(detectedContent, "icon") {
		contentType = iconContent
	}
	if len(contentType) == 0 {
		return "", ErrNotTextFile
	}
	return contentType, nil
}
