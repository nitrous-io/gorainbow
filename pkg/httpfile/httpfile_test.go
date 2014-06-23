package httpfile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
	"time"
)

func getConn(uri string) (*httputil.ClientConn, error) {
	dial, err := net.Dial("tcp", uri)
	if err != nil {
		return nil, fmt.Errorf("Can't connect to %v", uri)
	}
	return httputil.NewClientConn(dial, nil), nil
}

func TestReadFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	}))

	req, _ := http.NewRequest("GET", srv.URL, nil)
	conn, _ := getConn(srv.Listener.Addr().String())
	rf, err := NewFileReader(conn, req, "read-file-test")
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	result := bytes.NewBuffer(nil)
	io.Copy(result, rf)
	if content := result.String(); content != "Hello world" {
		t.Fatalf("Unexpected content: %v", content)
	}

	// test close
	rf.Close()
	if _, err := io.Copy(bytes.NewBuffer(nil), rf.reader); err == nil {
		t.Fatalf("should fail with closed body")
	}
}

func TestWriteFile(t *testing.T) {
	done := make(chan struct{}, 1)
	dest := bytes.NewBuffer(nil)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _ := w.(http.Hijacker).Hijack()
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.file.raw-stream\r\n\r\n")
		io.Copy(dest, conn)
		conn.Close()
		done <- struct{}{}
	}))

	req, _ := http.NewRequest("POST", srv.URL, nil)
	conn, _ := getConn(srv.Listener.Addr().String())
	rf, err := NewFileWriter(conn, req, "write-file-test")
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	_, err = io.Copy(rf, bytes.NewBufferString("Hello world"))
	rf.Close()
	<-done
	if content := dest.String(); content != "Hello world" {
		t.Fatalf("Unexpected content: %v", content)
	}
}

func TestReadStat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(&FileInfo{
			StatName:    "hello.txt",
			StatSize:    10,
			StatIsDir:   false,
			StatMode:    0755,
			StatModTime: time.Now(),
		})
	}))

	req, _ := http.NewRequest("GET", srv.URL, nil)
	conn, _ := getConn(srv.Listener.Addr().String())
	rf, err := NewFileStat(conn, req, "stat-file-test")
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	stat, err := rf.Stat()
	if err != nil || stat.Name() != "hello.txt" {
		t.Fatalf("Unexpected stat: %v, err: %v", stat, err)
	}
}
