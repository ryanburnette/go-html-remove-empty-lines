# [html-remove-empty-lines](https://github.com/ryanburnette/go-html-remove-empty-lines)

`go-html-remove-empty-lines` is a command-line utility written in Go to remove empty lines from HTML files recursively in a directory. It is particularly useful for cleaning up generated HTML output.

## Usage

```bash
go-html-remove-empty-lines -file <file> -dir <directory> [-ext <list>] [-d]
```

-   `-file`: File path (required if not using -dir). Specifies a single file to process.
-   `-dir`: Directory path (required if not using -file). Specifies the directory to process HTML files recursively.
-   `-exts`: Extensions lists. Defaults to `.html`, but a comma-separated list can be provided. Only works when using `-dir`.
-   `-d`: Dry run mode (optional). If provided, it performs a dry run without modifying the files.

**Note:** The `-file` and `-dir` options are mutually exclusive.

**Note:** The utility removes empty lines from HTML files while preserving the content within `<pre></pre>` blocks.
