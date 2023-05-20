# [html-remove-empty-lines](https://github.com/ryanburnette/go-html-remove-empty-lines)

`go-html-remove-empty-lines` is a command-line utility written in Go to remove empty lines from HTML files recursively in a directory. It is particularly useful for cleaning up generated HTML output.

## Usage

```bash
go-html-remove-empty-lines -dir <directory> [-d]
```

-   `-dir`: Directory path (required). Specifies the directory to process HTML files recursively.
-   `-d`: Dry run mode (optional). If provided, it performs a dry run without modifying the files.

**Note:** The utility removes empty lines from HTML files while preserving the content within `<pre></pre>` blocks.