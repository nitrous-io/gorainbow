package command

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/nitrous-io/gorainbow/cipher"
	"github.com/nitrous-io/gorainbow/models"
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/pkg/sftp"
)

const (
	currentDirectory = "/home/action"
	textContent      = "text"
	pngContent       = "image/png"
	jpgContent       = "image/jpg"
	gifContent       = "image/gif"
	iconContent      = "image/x-icon"
	sizeLimit        = 2 * 1024 * 1024 // 2 megabytes
)

var (
	ErrFileExist    = errors.New("aio:file_exists")
	ErrFileTooLarge = errors.New("aio:file_too_large")
	ErrNotTextFile  = errors.New("aio:not_text_file")
)

type FilesystemCommand struct {
	cfg    *CommandConfig
	client *sftp.Client
}

func (c *FilesystemCommand) Handle(msg *protocol.Message) error {
	if msg.Event == "start" {
		boxId := int(msg.Args[0].(float64))
		if err := c.start(boxId); err != nil {
			return err
		}
		writeOne(c.cfg.WS, protocol.CallbackMsg(msg.CallbackId))
	} else if msg.Event == "list" {
		dir := msg.Args[0].(string)
		files := [][]interface{}{}
		entries, err := c.list(dir)
		if err != nil {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, err, files))
		} else {
			for _, f := range entries {
				files = append(files, fileInfoArgs(f))
			}
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, nil, dir, files))
		}
	} else if msg.Event == "read" {
		fileName := msg.Args[0].(string)
		info, err := c.read(fileName)
		if err != nil {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, err))
		} else {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, nil, info...))
		}
	} else if msg.Event == "download" {
		// TODO: implement download
	} else if msg.Event == "write" {
		fileName := msg.Args[0].(string)
		content := msg.Args[1].(string)
		if err := c.write(fileName, content); err != nil {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, err))
		} else {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, nil, fileName))
		}
	} else if msg.Event == "writeNew" {
		fileName := msg.Args[0].(string)
		content := msg.Args[1].(string)
		if err := c.writeNew(fileName, content); err != nil {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, err))
		} else {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, nil, fileName))
		}
	} else if msg.Event == "makeDir" {
		dirName := msg.Args[0].(string)
		if err := c.mkdir(dirName); err != nil {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, err))
		} else {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, nil, dirName))
		}
	} else if msg.Event == "deleteDir" {
		// TODO: delete dir with ssh session and rm -rf
	} else if msg.Event == "deleteFile" {
		fileName := msg.Args[0].(string)
		if err := c.deleteFile(fileName); err != nil {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, err))
		} else {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, nil, fileName))
		}
	} else if msg.Event == "rename" {
		oldPath := msg.Args[0].(string)
		newPath := msg.Args[1].(string)
		if err := c.rename(oldPath, newPath); err != nil {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, err))
		} else {
			writeOne(c.cfg.WS, makeCallbackMsg(msg.CallbackId, nil, oldPath))
		}
	} else if msg.Event == "upload" {
		// TODO: implement upload
	}
	return nil
}

func (c *FilesystemCommand) rename(oldPath, newPath string) error {
	return c.client.Rename(getPath(oldPath), getPath(newPath))
}

func (c *FilesystemCommand) deleteFile(fileName string) error {
	return c.client.Remove(getPath(fileName))
}

func (c *FilesystemCommand) mkdir(dir string) error {
	return c.client.Mkdir(getPath(dir))
}

func (c *FilesystemCommand) write(path, content string) error {
	f, err := c.client.Create(getPath(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewBufferString(content))
	return err
}

func (c *FilesystemCommand) writeNew(path, content string) error {
	f, err := c.client.OpenFile(getPath(path), os.O_CREATE|os.O_EXCL)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewBufferString(content))
	return err
}

func (c *FilesystemCommand) read(path string) ([]interface{}, error) {
	log.Println("sftp: open path", path, getPath(path))
	// open the file
	f, err := c.client.Open(getPath(path))
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
	log.Println("content: ", path, string(data))
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

func (c *FilesystemCommand) list(dir string) ([]os.FileInfo, error) {
	return c.client.ReadDir(getPath(dir))
}

func (c *FilesystemCommand) start(boxId int) error {
	mulx := c.cfg.Mulx
	userId := c.cfg.UserId
	db := c.cfg.DB

	// fetch box and keypair
	box, keypair, err := models.FetchBoxAndKeyPair(db, boxId, userId)
	if err != nil {
		return err
	}
	privateKey, err := cipher.Decrypt(
		[]byte(c.cfg.AESKey),
		[]byte(keypair.EncryptedPrivateKeyIv),
		[]byte(keypair.EncryptedPrivateKey),
	)
	if err != nil {
		return err
	}

	// get ssh connection from the multiplexer
	sshClient := mulx.Get(userId)
	if sshClient == nil {
		sshClient, err = mulx.Add(userId, DefaultUsername,
			fmt.Sprintf("%s:%d", box.Host, box.Port),
			privateKey)
		if err != nil {
			return err
		}
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}
	c.client = sftpClient
	return nil
}

func newFilesystem(cfg *CommandConfig) (Command, error) {
	return &FilesystemCommand{cfg: cfg}, nil
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

func getPath(name string) string {
	if len(name) > 0 {
		name = path.Join(currentDirectory, escapeName(name))
	} else {
		name = currentDirectory
	}
	return name
}
