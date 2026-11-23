package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "page.html")
	writeFixture(t, path, "first\r\n\r\nlast\r\n")
	original, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	checkRun(t, 0, "Would update: "+path+"\n", "-file", path, "-d")
	checkFile(t, path, "first\r\n\r\nlast\r\n")
	checkRun(t, 0, "Updated: "+path+"\n", "-file", path)
	checkFile(t, path, "first\r\nlast\r\n")
	cleaned, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if cleaned.Mode().Perm() != original.Mode().Perm() {
		t.Fatalf("permissions changed from %v to %v", original.Mode(), cleaned.Mode())
	}
	checkRun(t, 0, "", "-file", path)
	checkRun(t, 0, "", "-file", path, "-d")
	unchanged, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(cleaned, unchanged) || !cleaned.ModTime().Equal(unchanged.ModTime()) {
		t.Fatal("clean file was rewritten")
	}
}

func TestRunDirectory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	html := filepath.Join(dir, "a.HTML")
	htm := filepath.Join(sub, "b.htm")
	other := filepath.Join(dir, "c.txt")
	for _, path := range []string{html, htm, other} {
		writeFixture(t, path, "\n<p>hi</p>\n")
	}
	target := filepath.Join(t.TempDir(), "outside.html")
	writeFixture(t, target, "\n<p>outside</p>\n")
	link := filepath.Join(dir, "link.html")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	checkRun(t, 0, "Would update: "+html+"\n", "-dir", dir, "-d")
	checkRun(t, 0, "Updated: "+html+"\nUpdated: "+htm+"\n", "-dir", dir, "-exts", ".html, .htm")
	checkFile(t, html, "<p>hi</p>\n")
	checkFile(t, htm, "<p>hi</p>\n")
	checkFile(t, other, "\n<p>hi</p>\n")
	checkFile(t, target, "\n<p>outside</p>\n")
	checkRun(t, 1, "", "-file", link)
}

func TestRunArguments(t *testing.T) {
	dir := t.TempDir()
	checkRun(t, 2, "")
	checkRun(t, 2, "", "-file", "x", "-dir", "y")
	checkRun(t, 2, "", "-dir", dir, "extra")
	checkRun(t, 0, "", "-h")
	checkRun(t, 1, "", "-file", filepath.Join(dir, "missing.html"))
	checkRun(t, 1, "", "-dir", dir, "-ext", "")
	path := filepath.Join(dir, "page.html")
	writeFixture(t, path, "page")
	checkRun(t, 1, "", "-dir", path)
	checkRun(t, 1, "", "-file", dir)
}

func TestProcessDirContinuesAfterErrors(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.html", "b.html", "c.html"} {
		writeFixture(t, filepath.Join(dir, name), "")
	}
	first := errors.New("first failure")
	second := errors.New("second failure")
	var visited []string
	err := processDir(dir, []string{".html"}, func(path string) error {
		name := filepath.Base(path)
		visited = append(visited, name)
		switch name {
		case "a.html":
			return first
		case "b.html":
			return second
		default:
			return nil
		}
	})
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("expected both errors, got %v", err)
	}
	if got := strings.Join(visited, ","); got != "a.html,b.html,c.html" {
		t.Fatalf("visited %s", got)
	}
}

func TestReplaceFileRejectsChangedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "page.html")
	writeFixture(t, path, "original")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, path, "concurrent edit")
	if err := replaceFile(path, []byte("replacement"), info); err == nil {
		t.Fatal("expected replacement to fail")
	}
	checkFile(t, path, "concurrent edit")
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary file left behind: %v, %v", entries, err)
	}
}

func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0640); err != nil {
		t.Fatal(err)
	}
}

func checkFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s contains %q, want %q", path, got, want)
	}
}

func checkRun(t *testing.T, wantCode int, wantOut string, args ...string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if got := run("html-remove-empty-lines", args, &stdout, &stderr); got != wantCode {
		t.Fatalf("%v exited %d, want %d; stderr: %s", args, got, wantCode, &stderr)
	}
	if got := stdout.String(); got != wantOut {
		t.Fatalf("%v wrote %q to stdout, want %q", args, got, wantOut)
	}
	if wantCode != 0 && stderr.Len() == 0 {
		t.Fatalf("%v failed without a diagnostic", args)
	}
}
