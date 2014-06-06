package command

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
		esc := shellEscape(tc.src)
		if tc.expected != esc {
			t.Fatalf("Expect %v but got %v", tc.expected, esc)
		}
	}
}

func TestQuoteEscape(t *testing.T) {
	testCases := []struct {
		src      string
		expected string
	}{
		{"3'o clock at harry's", `'3'"'"'o clock at harry'"'"'s'`},
	}

	for _, tc := range testCases {
		esc := quoteEscape(tc.src)
		if tc.expected != esc {
			t.Fatalf("Expect %v but got %v", tc.expected, esc)
		}
	}
}

func TestGetPath(t *testing.T) {
	testCases := []struct {
		src      string
		expected string
	}{
		{"", `./`},
		{"README.md", `./README.md`},
		{"/home/action", `/home/action`},
	}

	for _, tc := range testCases {
		esc := getPath(tc.src)
		if tc.expected != esc {
			t.Fatalf("Expect %v but got %v", tc.expected, esc)
		}
	}
}
