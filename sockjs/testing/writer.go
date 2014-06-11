package testing

import "github.com/nitrous-io/gorainbow/sockjs"

type MemWriter struct {
	Stdin [][]interface{}
}

func (m *MemWriter) Write(a []interface{}) {
	m.Stdin = append(m.Stdin, a)
}

func NewMemWriter() sockjs.Writer {
	stdin := [][]interface{}{}
	return &MemWriter{Stdin: stdin}
}
