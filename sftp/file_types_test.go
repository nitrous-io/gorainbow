package sftp

import (
	"testing"
)

func TestIsKnownTextFile(t *testing.T) {
	expected := map[string]bool{
		"/tmp/test.rb": true,
		"/tmp/.env":    true,
		"/tmp/README":  true,
		"/tmp/iotop":   false,
	}

	for name, r := range expected {
		if IsKnownTextFile(name) != r {
			t.Log(name)
			t.Fail()
		}
	}
}

func TestGetFileType(t *testing.T) {
	f, _ := GetFileType("./fixtures/ls.txt")
	if f != "ASCII text" {
		t.Fatal(f)
	}
	_, err := GetFileType("./fixtures/wahever.txt")
	if err == nil {
		t.Fatal("should fail but didnt")
	}
}
