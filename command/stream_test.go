package command

import (
	"bytes"
	"reflect"
	"testing"
)

func TestScannerOutput(t *testing.T) {
	output := bytes.NewBuffer(nil)

	sc := &scannerOutput{
		outBuffer: bytes.NewBuffer(nil),
		output:    output,
	}

	texts := []string{
		"wow\n hello",
		" world\n should",
		" finish",
	}
	for _, text := range texts {
		i, err := sc.Write([]byte(text))
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

func TestScannerOutputWithLimit(t *testing.T) {
	output := bytes.NewBuffer(nil)

	sc := &scannerOutput{
		lineLimit: 2,
		outBuffer: bytes.NewBuffer(nil),
		output:    output,
	}

	_, err := sc.Write([]byte("wow\n hello"))
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	_, err = sc.Write([]byte(" world\n should"))
	if err != ErrLimitLines {
		t.Fatal("Unexpected error", err)
	}

	final := output.Bytes()
	expected := `wow\n hello world\n should`
	if reflect.DeepEqual(final, []byte(expected)) {
		t.Fatal("unexpected result", string(final))
	}
}

func TestCachedStream(t *testing.T) {
	c := &cachedStream{limit: 5, content: bytes.NewBuffer(nil)}

	c.Write([]byte("hello"))
	c.Write([]byte("ignore"))
	c.Write([]byte("ignore"))

	final := c.content.String()
	if final != "hello" {
		t.Fatal("Unexpected content", final)
	}
}
