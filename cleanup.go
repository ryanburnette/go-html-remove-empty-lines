package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/net/html"
)

func processFile(path string, dryRun bool) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("%s: expected a regular file", path)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	cleaned, err := removeEmptyLines(content)
	if err != nil {
		return false, fmt.Errorf("process %s: %w", path, err)
	}
	if bytes.Equal(content, cleaned) {
		return false, nil
	}
	if !dryRun {
		if err := replaceFile(path, cleaned, info); err != nil {
			return false, fmt.Errorf("replace %s: %w", path, err)
		}
	}
	return true, nil
}

func replaceFile(path string, content []byte, original os.FileInfo) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".html-cleanup-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	_, writeErr := tmp.Write(content)
	modeErr := tmp.Chmod(original.Mode().Perm())
	if err := errors.Join(writeErr, modeErr, tmp.Close()); err != nil {
		return err
	}

	current, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !os.SameFile(original, current) || original.Size() != current.Size() ||
		!original.ModTime().Equal(current.ModTime()) || original.Mode() != current.Mode() {
		return errors.New("file changed during processing")
	}
	return os.Rename(tmp.Name(), path)
}

type byteSpan struct {
	start, end int
}

// Keep the original bytes rather than rendering tokens back to HTML.
func removeEmptyLines(content []byte) ([]byte, error) {
	protected, err := protectedSpans(content)
	if err != nil {
		return nil, err
	}
	cleaned := make([]byte, 0, len(content))
	for start := 0; start < len(content); {
		end := start + lineLength(content[start:])
		for len(protected) > 0 && protected[0].end <= start {
			protected = protected[1:]
		}
		line := content[start:end]
		if (len(protected) > 0 && protected[0].start < end) ||
			len(bytes.Trim(line, " \t\r\n\f")) != 0 {
			cleaned = append(cleaned, line...)
		}
		start = end
	}
	return cleaned, nil
}

func lineLength(content []byte) int {
	end := bytes.IndexAny(content, "\r\n")
	if end == -1 {
		return len(content)
	}
	if content[end] == '\r' && end+1 < len(content) && content[end+1] == '\n' {
		return end + 2
	}
	return end + 1
}

func protectedSpans(content []byte) ([]byteSpan, error) {
	tokenizer := html.NewTokenizer(bytes.NewReader(content))
	var spans []byteSpan
	preDepth := 0
	rawTag := ""
	for start := 0; ; {
		kind := tokenizer.Next()
		raw := tokenizer.Raw()
		end := start + len(raw)
		if kind == html.ErrorToken && !errors.Is(tokenizer.Err(), io.EOF) {
			return nil, tokenizer.Err()
		}

		// Preserve multiline markup and comments, as well as text in pre
		// elements and raw-text elements.
		keep := kind != html.TextToken || preDepth > 0 || rawTag != ""
		if keep && bytes.IndexAny(raw, "\r\n") >= 0 {
			spans = append(spans, byteSpan{start, end})
		}
		if kind == html.ErrorToken {
			return spans, nil
		}
		tagName, _ := tokenizer.TagName()
		tag := string(tagName)
		switch kind {
		case html.StartTagToken, html.SelfClosingTagToken:
			switch tag {
			case "pre", "listing":
				preDepth++
			case "textarea", "script", "style", "title", "xmp", "iframe", "noembed", "noframes", "noscript", "plaintext":
				rawTag = tag
			}
		case html.EndTagToken:
			switch {
			case tag == rawTag:
				rawTag = ""
			case tag == "pre" || tag == "listing":
				preDepth = max(0, preDepth-1)
			}
		}
		start = end
	}
}
