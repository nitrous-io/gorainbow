package utils

import (
	"bytes"
	"reflect"
	"testing"
)

func TestBufferWriter(t *testing.T) {
	cases := []struct {
		lines          []string
		lineLimit      int
		expectedErrors []error
		expected       string
	}{
		{
			[]string{"one write with no newlines"},
			1, []error{nil},
			"one write with no newlines",
		},
		{
			[]string{"one \nwrite with\nnewlines within limit\n"},
			5, []error{nil, nil, nil},
			"one \nwrite with\nnewlines within limit\n",
		},
		{
			[]string{"wow\n", " hello world\n", " should finish"},
			5, []error{nil, nil, nil},
			"wow\n hello world\n should finish",
		},
		{
			[]string{"wow\n hello", " world\n should truncate\n"},
			2, []error{nil, ErrLimitLines},
			"wow\n hello world\n",
		},
		{
			[]string{"wow\n", " hello world\n should truncate"},
			2, []error{nil, ErrLimitLines},
			"wow\n hello world\n",
		},
		{
			[]string{"wow\n hello world\n"},
			2, []error{nil, nil},
			"wow\n hello world\n",
		},
	}

	for _, c := range cases {
		output := bytes.NewBuffer(nil)
		bw := NewBufferWriter(c.lineLimit, output)
		for i, line := range c.lines {
			_, err := bw.Write([]byte(line))
			if err != c.expectedErrors[i] {
				t.Errorf("Expected error: %v, got error: %v", c.expectedErrors[i], err)
			}
		}
		written := output.Bytes()
		if !reflect.DeepEqual(written, []byte(c.expected)) {
			t.Errorf("Unexpected result: %s; expected: %s", written, c.expected)
		}
	}
}
