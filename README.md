# html-remove-empty-lines

Remove blank lines from HTML files in place. The tool keeps line endings on retained lines and preserves whitespace inside `<pre>`, `<textarea>`, scripts, styles, and multiline markup.

## Install

With Go 1.27 or later, run:

```sh
go install github.com/ryanburnette/go-html-remove-empty-lines@latest
```

Go installs the command as `go-html-remove-empty-lines`.

## Usage

```sh
go-html-remove-empty-lines -file index.html
go-html-remove-empty-lines -dir public -ext .html,.htm
go-html-remove-empty-lines -dir public -d
```

Use either `-file` or `-dir`. Directory mode searches recursively, skips symlinks, and matches `.html` by default. Use `-ext` to set a comma-separated list of extensions. Matching is case-insensitive.

Add `-d` to list files that would change. A failed file or directory returns a nonzero exit status.

## License

[MIT](LICENSE)
