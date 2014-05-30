package protocol

import (
	"reflect"
	"testing"
)

func TestParseClientMsg(t *testing.T) {
	testCases := []struct {
		data     string
		expected *Message
	}{
		{
			`["global", "auth", ["auth-token"], 86]`,
			&Message{
				Namespace:  "global",
				Event:      "auth",
				CallbackId: 86,
				Args:       []interface{}{"auth-token"},
			},
		},
		{
			`["sftp", "init", null, 42]`,
			&Message{
				Namespace:  "sftp",
				Event:      "init",
				Args:       []interface{}{},
				CallbackId: 42,
			},
		},
		{
			`["sftp", "rename", ["foo", "bar"], 77]`,
			&Message{
				Namespace:  "sftp",
				Event:      "rename",
				CallbackId: 77,
				Args:       []interface{}{"foo", "bar"},
			},
		},
	}

	for _, test := range testCases {
		msg, err := ParseClientMsg([]byte(test.data))
		if err != nil {
			t.Fatal("Unexpected error", err)
		}
		if !reflect.DeepEqual(msg, test.expected) {
			t.Fatalf("Wanted %+v but got %+v", test.expected, msg)
		}
	}
}

func TestEncodeServerMsg(t *testing.T) {
	testCases := []struct {
		msg      *Message
		expected []interface{}
	}{
		{
			&Message{
				Namespace:  "global",
				Event:      "callback",
				CallbackId: 42,
				Args:       []interface{}{"foo", "bar"},
			},
			[]interface{}{
				"global",
				"callback",
				[]interface{}{42, []interface{}{"foo", "bar"}},
			},
		},
	}

	for _, test := range testCases {
		data := EncodeServerMsg(test.msg)
		if !reflect.DeepEqual(data, test.expected) {
			t.Fatalf("Wanted %+v but got %+v", test.expected, data)
		}
	}
}
