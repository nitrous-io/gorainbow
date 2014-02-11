package gorainbow

import (
	"encoding/base64"
	"fmt"
	"github.com/action-io/gorainbow/sftp"
	"io/ioutil"
	"os"
	"strings"
)

const (
	SIZE_LIMIT = 2 * 1024 * 1024 // 2 megabytes
)

type SftpSocket struct {
	Conn *Connection
	Sftp *sftp.Conn
}

func (s *SftpSocket) OnMessage(msg Message) {
	debugln(msg)
	switch msg.Event {
	case "start":
		s.onStart(msg)
	case "list":
		s.onList(msg)
	case "read":
		s.onRead(msg)
	case "download":
		s.onDownload(msg)
	case "write":
		s.onWrite(msg)
	case "writeNew":
		s.onWriteNew(msg)
	case "makeDir":
		s.onMakeDir(msg)
	case "deleteDir":
		s.onDeleteDir(msg)
	case "deleteFile":
		s.onDeleteFile(msg)
	case "rename":
		s.onRename(msg)
	case "upload":
		s.onUpload(msg)
	}
}

func (s *SftpSocket) onStart(msg Message) {
	s.Sftp = &sftp.Conn{Port: "22"} // FIXME: change port
	if err := s.Sftp.Connect(); err != nil {
		// deal with it
	} else {
		startMsg := Message{Namespace: "global", Event: "callback",
			CallbackId: msg.CallbackId}
		debugln(startMsg)
		go s.Conn.Write(startMsg.Encode())
	}
}

func (s *SftpSocket) onList(msg Message) {
	filePath, _ := msg.Args[0].(string)
	files, _ := s.Sftp.Ls(filePath)
	list := [][]interface{}{}
	for _, f := range files {
		list = append(list, f.ToArr())
	}
	listMsg := Message{Namespace: "global", Event: "callback",
		CallbackId: msg.CallbackId, Args: []interface{}{nil, filePath, list}}
	go s.Conn.Write(listMsg.Encode())
}

func (s *SftpSocket) onRead(msg Message) {
	var err error
	var file sftp.FileEntry
	filePath, _ := msg.Args[0].(string)
	files, err := s.Sftp.Ls(filePath)
	if err != nil {
		// deal with it
		debugln("get", err, files)
		return
	}
	readMsg := Message{Namespace: "global", Event: "callback",
		CallbackId: msg.CallbackId}
	file = files[0]
	debugln("file", file)
	if file.FileSize > SIZE_LIMIT {
		// deal with it
		readMsg.Args = []interface{}{"aio:file_too_large", filePath}
		go s.Conn.Write(readMsg.Encode())
		return
	}
	data, err := s.Sftp.Get(filePath)
	debugln("data", data)
	if err != nil {
		// deal with it
		debugln("get", err, data)
		return
	}
	defer func() {
		data.File.Close()
		os.Remove(data.File.Name())
	}()
	content, _ := ioutil.ReadFile(data.File.Name())
	if strings.Contains(data.FileType, `text`) {
		readMsg.Args = []interface{}{nil, "text", filePath, string(content)}
	} else if strings.Contains(data.FileType, `PNG`) {
		readMsg.Args = []interface{}{nil, "image/png", filePath, tobase64(content)}
	} else if strings.Contains(data.FileType, `JPEG`) {
		readMsg.Args = []interface{}{nil, "image/jpg", filePath, tobase64(content)}
	} else if strings.Contains(data.FileType, `GIF`) {
		readMsg.Args = []interface{}{nil, "image/gif", filePath, tobase64(content)}
	} else if strings.Contains(data.FileType, `icon`) {
		readMsg.Args = []interface{}{nil, "image/x-icon", filePath, tobase64(content)}
	} else {
		readMsg.Args = []interface{}{"aio:not_text_file", filePath}
	}
	go s.Conn.Write(readMsg.Encode())
}

func (s *SftpSocket) onDownload(msg Message) {
}

func (s *SftpSocket) onWrite(msg Message) {
	remotePath, _ := msg.Args[0].(string)
	content, _ := msg.Args[1].(string)
	err := s.Sftp.PutData(remotePath, []byte(content))
	writeMsg := Message{Namespace: "global", Event: "callback",
		CallbackId: msg.CallbackId,
		Args:       []interface{}{fmt.Sprintf("%v", err), remotePath}}
	go s.Conn.Write(writeMsg.Encode())
}

func (s *SftpSocket) onWriteNew(msg Message) {
	remotePath, _ := msg.Args[0].(string)
	content, _ := msg.Args[1].(string)
	resp := Message{Namespace: "global", Event: "callback",
		CallbackId: msg.CallbackId}
	_, err := s.Sftp.Ls(remotePath)
	if err == sftp.ErrFileNotFound {
		s.Sftp.PutData(remotePath, []byte(content))
		resp.Args = []interface{}{nil, remotePath}
	} else {
		if err == nil {
			resp.Args = []interface{}{"aio:file_exists", remotePath}
		} else {
			resp.Args = []interface{}{fmt.Sprint(err), remotePath}
		}
	}
	go s.Conn.Write(resp.Encode())
}

func (s *SftpSocket) onMakeDir(msg Message) {
	remotePath, _ := msg.Args[0].(string)
	resp := Message{Namespace: "global", Event: "callback",
		CallbackId: msg.CallbackId}
	_, err := s.Sftp.Ls(remotePath)
	if err == sftp.ErrFileNotFound {
		s.Sftp.Mkdir(remotePath)
		resp.Args = []interface{}{nil, remotePath}
	} else {
		if err == nil {
			resp.Args = []interface{}{"aio:dir_exists", remotePath}
		} else {
			resp.Args = []interface{}{fmt.Sprint(err), remotePath}
		}
	}
	go s.Conn.Write(resp.Encode())
}

// FIXME: ssh and exec rm -rf here
func (s *SftpSocket) onDeleteDir(msg Message) {
}

func (s *SftpSocket) onDeleteFile(msg Message) {
	remotePath, _ := msg.Args[0].(string)
	err := s.Sftp.Rm(remotePath)
	resp := Message{Namespace: "global", Event: "callback",
		CallbackId: msg.CallbackId, Args: []interface{}{err}}
	go s.Conn.Write(resp.Encode())
}

func (s *SftpSocket) onRename(msg Message) {
	oldPath, _ := msg.Args[0].(string)
	newPath, _ := msg.Args[1].(string)
	resp := Message{Namespace: "global", Event: "callback",
		CallbackId: msg.CallbackId}
	_, err := s.Sftp.Ls(newPath)
	if err == nil {
		resp.Args = []interface{}{"aio:name_taken"}
	} else {
		err = s.Sftp.Rename(oldPath, newPath)
		resp.Args = []interface{}{err, oldPath}
	}
	go s.Conn.Write(resp.Encode())
}

func (s *SftpSocket) onUpload(msg Message) {
}

func tobase64(str []byte) string {
	return base64.StdEncoding.EncodeToString(str)
}
