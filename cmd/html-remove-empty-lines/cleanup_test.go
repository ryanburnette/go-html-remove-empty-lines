package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRemoveEmptyLines(t *testing.T) {
	longLine := "<p>" + strings.Repeat("x", 128<<10) + "</p>\n"
	tests := []struct {
		name  string
		input string
		want  string
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := removeEmptyLines([]byte(tt.input))
			if err != nil {
				t.Fatalf("removeEmptyLines failed: %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("removeEmptyLines = %.120q (%d bytes), want %.120q (%d bytes)", got, len(got), tt.want, len(tt.want))
			}
			again, err := removeEmptyLines(got)
			if err != nil {
				t.Fatalf("second cleanup failed: %v", err)
			}
			if !bytes.Equal(again, got) {
				t.Errorf("second cleanup = %.120q, want %.120q", again, got)
			}
		})
	}
}
