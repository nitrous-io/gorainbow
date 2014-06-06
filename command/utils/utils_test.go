package utils

import "testing"

func TestShellEscape(t *testing.T) {
	testCases := []struct {
		src      string
		expected string
	}{
		{"new\nline", "new\\\\nline"},
		{"evil`rm -rf`", "evil\\`rm\\ -rf\\`"},
		{"evil$(rm -rf)", "evil\\$\\(rm\\ -rf\\)"},
		{"regex search\\w\\w", "regex\\ search\\\\w\\\\w"},
		{"test; rm -rf", "test\\;\\ rm\\ -rf"},
		{"*.p?", "*.p?"},
	}

	for _, tc := range testCases {
		esc := ShellEscape(tc.src)
		if tc.expected != esc {
			t.Fatalf("Expect %v but got %v", tc.expected, esc)
		}
	}
}
