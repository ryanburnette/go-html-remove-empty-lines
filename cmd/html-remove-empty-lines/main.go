package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	os.Exit(run(os.Args[0], os.Args[1:], os.Stdout, os.Stderr))
}

func run(name string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet(filepath.Base(name), flag.ContinueOnError)
	flags.SetOutput(stderr)
	directory := flags.String("dir", "", "Process matching files recursively")
	file := flags.String("file", "", "Process a single file")
	var extensions string
	flags.StringVar(&extensions, "ext", ".html", "Comma-separated extensions for -dir")
	flags.StringVar(&extensions, "exts", ".html", "Alias for -ext")
	dryRun := flags.Bool("d", false, "List files that would change")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if (*directory == "") == (*file == "") || flags.NArg() != 0 {
		fmt.Fprintln(stderr, "Specify exactly one of -dir or -file, with no positional arguments.")
		flags.Usage()
		return 2
	}

	process := func(path string) error {
		changed, err := processFile(path, *dryRun)
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}
		if *dryRun {
			fmt.Fprintf(stdout, "Would update: %s\n", path)
		} else {
			fmt.Fprintf(stdout, "Updated: %s\n", path)
		}
		return nil
	}

	var err error
	if *file != "" {
		err = process(*file)
	} else {
		err = processDir(*directory, strings.Split(extensions, ","), process)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func processDir(dir string, extensions []string, process func(string) error) error {
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s: expected a directory", dir)
	}

	for i, ext := range extensions {
		ext = strings.TrimSpace(ext)
		if !strings.HasPrefix(ext, ".") || len(ext) == 1 {
			return fmt.Errorf("invalid extension %q: expected a value such as .html", ext)
		}
		extensions[i] = ext
	}

	var failures []error
	err = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			failures = append(failures, err)
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		ext := filepath.Ext(path)
		for _, candidate := range extensions {
			if !strings.EqualFold(ext, candidate) {
				continue
			}
			if err := process(path); err != nil {
				failures = append(failures, err)
			}
			break
		}
		return nil
	})
	if err != nil {
		failures = append(failures, err)
	}
	return errors.Join(failures...)
}
