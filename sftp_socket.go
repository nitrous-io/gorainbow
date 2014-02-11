package gorainbow

import (
	"github.com/action-io/gorainbow/sftp"
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
}

func (s *SftpSocket) onDownload(msg Message) {
}

func (s *SftpSocket) onWrite(msg Message) {
}

func (s *SftpSocket) onWriteNew(msg Message) {
}

func (s *SftpSocket) onMakeDir(msg Message) {
}

func (s *SftpSocket) onDeleteDir(msg Message) {
}

func (s *SftpSocket) onDeleteFile(msg Message) {
}

func (s *SftpSocket) onRename(msg Message) {
}

func (s *SftpSocket) onUpload(msg Message) {
}
