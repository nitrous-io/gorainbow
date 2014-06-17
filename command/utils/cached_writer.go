package utils

import "bytes"

type CachedWriter struct {
	Store *bytes.Buffer
}

func (w *CachedWriter) Write(p []byte) (int, error) {
	return w.Store.Write(p)
}

func (w *CachedWriter) String() string {
	return w.Store.String()
}

func NewCachedWriter() *CachedWriter {
	return &CachedWriter{
		Store: bytes.NewBuffer(nil),
	}
}
