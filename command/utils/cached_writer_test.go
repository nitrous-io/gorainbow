package utils

import "testing"

func TestCachedWriter(t *testing.T) {
	w := NewCachedWriter()
	w.Write([]byte("hello"))
	w.Write([]byte("world"))

	if res := w.String(); res != "helloworld" {
		t.Fatalf("Got %v", res)
	}
}
