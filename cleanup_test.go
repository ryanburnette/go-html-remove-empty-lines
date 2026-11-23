package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRemoveEmptyLines(t *testing.T) {
	longLine := "<p>" + strings.Repeat("x", 128<<10) + "</p>\n"
	cases := []struct {
		name, input, want string
	}{
		{"empty", "", ""},
		{"all blank", "\n \t\r\n\f", ""},
		{"mixed endings", "\r\na\r\n \t\r\nb\r\rc\n\n", "a\r\nb\rc\n"},
		{"no final newline", "a\n\n b", "a\n b"},
		{"long line", "before\n\n" + longLine + "\nafter\n", "before\n" + longLine + "after\n"},
		{"nonbreaking space", "a\n\u00a0\n\nb\n", "a\n\u00a0\nb\n"},
		{"pre with attributes", "\r\n<PRE class='code' data-x='>'>\r\n\r\na\r\n \r\n</pRe>\r\n\r\n", "<PRE class='code' data-x='>'>\r\n\r\na\r\n \r\n</pRe>\r\n"},
		{"multiline tag", "<pre\n\n class='code'>\n\n</pre>\n\n", "<pre\n\n class='code'>\n\n</pre>\n"},
		{"same line tags", "<pre>a</pre><pre>\n\nb\n</pre>\n\n", "<pre>a</pre><pre>\n\nb\n</pre>\n"},
		{"unterminated pre", "\n<pre>\n\n \t", "<pre>\n\n \t"},
		{"comment", "<!-- <pre>\n\n-->\n\n<p>x</p>\n", "<!-- <pre>\n\n-->\n<p>x</p>\n"},
		{"textarea", "\n<textarea>\n\n<pre>\n\n</textarea>\n\n", "<textarea>\n\n<pre>\n\n</textarea>\n"},
		{"script", "<script>let x = `<pre>\n\n`;\n</script>\n\n<p>x</p>\n", "<script>let x = `<pre>\n\n`;\n</script>\n<p>x</p>\n"},
		{"style", "<style>\n\n</style>\n\n", "<style>\n\n</style>\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := removeEmptyLines([]byte(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("got %d bytes, want %d bytes", len(got), len(tc.want))
			}
			again, err := removeEmptyLines(got)
			if err != nil || !bytes.Equal(again, got) {
				t.Fatalf("second cleanup changed output or failed: %v", err)
			}
		})
	}
}
