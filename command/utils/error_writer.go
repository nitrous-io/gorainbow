package utils

import (
	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
)

type ErrorWriter struct {
	namespace string
	ws        sockjs.Writer
}

func (c *ErrorWriter) Write(p []byte) (int, error) {
	WriteOne(c.ws, &protocol.Message{
		Namespace: c.namespace,
		Event:     "commandError",
		Args:      []interface{}{string(p)},
	})
	return len(p), nil
}

func NewErrorWriter(namespace string, ws sockjs.Writer) *ErrorWriter {
	return &ErrorWriter{namespace, ws}
}
