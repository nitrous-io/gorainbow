package utils

import (
	"bytes"
	"io"
)

// Prefix all output with `prefix` string
type PrefixWriter struct {
	prefix string
	output io.Writer
}

func (w *PrefixWriter) Write(p []byte) (int, error) {
	data := bytes.NewBufferString(w.prefix)
	if _, err := data.Write(p); err != nil {
		return 0, err
	}
	_, err := w.output.Write(data.Bytes())
	return len(p), err
}

func NewPrefixWriter(prefix string, output io.Writer) *PrefixWriter {
	return &PrefixWriter{prefix, output}
}
