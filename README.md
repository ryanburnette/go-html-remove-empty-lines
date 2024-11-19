# [html-remove-empty-lines](https://github.com/ryanburnette/go-html-remove-empty-lines)

`html-remove-empty-lines` is a Go-based CLI tool for removing empty lines from HTML files.

## Usage

```bash
html-remove-empty-lines -file <file> -dir <directory> [-exts <extensions>] [-d]
```

### Options

-   `-file <file>`: Specify the path to a single HTML file to process. Required if `-dir` is not used.
-   `-dir <directory>`: Specify the path to a directory to process all HTML files recursively. Required if `-file` is not used.
-   `-exts <extensions>`: Specify a comma-separated list of file extensions to process (e.g., `.html,.htm`). Defaults to `.html`. Only applicable when using `-dir`.
-   `-d`: Enable dry run mode. Displays changes without modifying the files.

### Notes

-   The `-file` and `-dir` options are mutually exclusive; you must use one or the other.
-   The utility removes empty lines from HTML files while preserving content within `<pre></pre>` blocks.
