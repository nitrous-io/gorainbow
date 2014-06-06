package utils

import (
	"bytes"
	"reflect"
	"testing"
)

func TestBufferWriter(t *testing.T) {
	output := bytes.NewBuffer(nil)
	bw := NewBufferWriter(5, output)

	for _, text := range []string{
		"wow\n hello",
		" world\n should",
		" finish",
	} {
		i, err := bw.Write([]byte(text))
		if i != len(text) || err != nil {
			t.Fatal("failed to write %v", text)
		}
	}

	final := output.Bytes()
	expected := `wow\n hello world\n should finish`
	if reflect.DeepEqual(final, []byte(expected)) {
		t.Fatal("unexpected result", string(final))
	}
}

func TestBufferWriterLimit(t *testing.T) {
	output := bytes.NewBuffer(nil)
	bw := NewBufferWriter(2, output)

	if _, err := bw.Write([]byte("wow\n hello")); err != nil {
		t.Fatal("Unexpected error", err)
	}
	if _, err := bw.Write([]byte(" world\n should")); err != ErrLimitLines {
		t.Fatal("Unexpected error", err)
	}

	final := output.Bytes()
	expected := `wow\n hello world\n should`
	if reflect.DeepEqual(final, []byte(expected)) {
		t.Fatal("unexpected result", string(final))
	}
}
