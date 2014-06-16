package testing

import "github.com/nitrous-io/gorainbow/sockjs"

type Writer struct {
	Stdin [][]interface{}
}

func (m *Writer) Write(a []interface{}) {
	m.Stdin = append(m.Stdin, a)
}

func NewWriter() sockjs.Writer {
	return &Writer{Stdin: [][]interface{}{}}
}
