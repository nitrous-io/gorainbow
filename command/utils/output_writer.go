package utils

import (
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
)

type OutputWriter struct {
	namespace string
	ws        sockjs.Writer
}

func (c *OutputWriter) Write(p []byte) (int, error) {
	WriteOne(c.ws, &protocol.Message{
		Namespace: c.namespace,
		Event:     "commandOutput",
		Args:      []interface{}{string(p)},
	})
	return len(p), nil
}

func NewOutputWriter(namespace string, ws sockjs.Writer) *OutputWriter {
	return &OutputWriter{namespace, ws}
}
