package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
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
	checkRun(t, 0, "Would update: "+html+"\n", "-dir", dir, "-d")
	checkRun(t, 0, "Updated: "+html+"\nUpdated: "+htm+"\n", "-dir", dir, "-exts", ".html, .htm")
	checkFile(t, html, "<p>hi</p>\n")
	checkFile(t, htm, "<p>hi</p>\n")
	checkFile(t, other, "\n<p>hi</p>\n")
}

func TestRunSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(t.TempDir(), "outside.html")
	writeFixture(t, target, "\n<p>outside</p>\n")
	link := filepath.Join(dir, "link.html")
	if err := os.Symlink(target, link); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("creating symlinks is unavailable: %v", err)
		}
		t.Fatal(err)
	}

	checkRun(t, 0, "", "-dir", dir)
	checkFile(t, target, "\n<p>outside</p>\n")
	checkRun(t, 1, "", "-file", link)
}

func TestRunArguments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "page.html")
	writeFixture(t, path, "page")

	tests := []struct {
		name string
		args []string
		want int
	}{
		{name: "missing mode", want: 2},
		{name: "both modes", args: []string{"-file", "x", "-dir", "y"}, want: 2},
		{name: "positional argument", args: []string{"-dir", dir, "extra"}, want: 2},
		{name: "unknown flag", args: []string{"-unknown"}, want: 2},
		{name: "help", args: []string{"-h"}, want: 0},
		{name: "missing file", args: []string{"-file", filepath.Join(dir, "missing.html")}, want: 1},
		{name: "empty extension", args: []string{"-dir", dir, "-ext", ""}, want: 1},
		{name: "file as directory", args: []string{"-dir", path}, want: 1},
		{name: "directory as file", args: []string{"-file", dir}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkRun(t, tt.want, "", tt.args...)
		})
	}
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
	want := []string{"a.html", "b.html", "c.html"}
	if !slices.Equal(visited, want) {
		t.Errorf("visited %q, want %q", visited, want)
	}
}

func TestRunDirectoryReturnsFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires Unix directory permissions")
	}
	dir := t.TempDir()
	blocked := filepath.Join(dir, "a-read-only")
	if err := os.Mkdir(blocked, 0700); err != nil {
		t.Fatal(err)
	}
	first := filepath.Join(blocked, "page.html")
	last := filepath.Join(dir, "z.html")
	writeFixture(t, first, "\nfirst\n")
	writeFixture(t, last, "\nlast\n")
	if err := os.Chmod(blocked, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(blocked, 0700); err != nil {
			t.Error(err)
		}
	})
	probe, err := os.CreateTemp(blocked, "probe-*")
	if err == nil {
		probe.Close()
		t.Skip("process can write to a read-only directory")
	}
	if !errors.Is(err, os.ErrPermission) {
		t.Fatal(err)
	}

	checkRun(t, 1, "Updated: "+last+"\n", "-dir", dir)
	checkFile(t, first, "\nfirst\n")
	checkFile(t, last, "last\n")
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
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("temporary file left behind: %v", entries)
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
		t.Errorf("%s contains %q, want %q", path, got, want)
	}
}

func checkRun(t *testing.T, wantCode int, wantOut string, args ...string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if got := run("html-remove-empty-lines", args, &stdout, &stderr); got != wantCode {
		t.Errorf("run(%q) exited %d, want %d; stderr: %q", args, got, wantCode, stderr.String())
	}
	if got := stdout.String(); got != wantOut {
		t.Errorf("run(%q) wrote %q to stdout, want %q", args, got, wantOut)
	}
	if wantCode != 0 && stderr.Len() == 0 {
		t.Errorf("run(%q) failed without a diagnostic", args)
	}
}
