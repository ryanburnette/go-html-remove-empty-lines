package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	directory := flag.String("dir", "", "Directory path")
	file := flag.String("file", "", "Single file path")
	extensions := flag.String("ext", ".html", "File extensions")
	dryRun := flag.Bool("d", false, "Dry run")
	flag.Parse()

	if *directory == "" && *file == "" {
		fmt.Println("Please provide either -dir or -file flag")
		flag.Usage()
		os.Exit(1)
	}

	if *directory != "" {
		extSlice := strings.Split(*extensions, ",")
		err := processDir(*directory, *dryRun, extSlice)
		if err != nil {
			fmt.Printf("Error processing directory: %v\n", err)
			os.Exit(1)
		}
	}

	if *file != "" {
		err := processFile(*file, *dryRun)
		if err != nil {
			fmt.Printf("Error processing file: %v\n", err)
			os.Exit(1)
		}
	}
}

func processFile(file string, dryRun bool) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	newContent := removeEmptyLines(string(content))

	if !dryRun {
		err = ioutil.WriteFile(file, []byte(newContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", file, err)
		}
		fmt.Printf("Updated file: %s\n", file)
	} else {
		fmt.Printf("Dry run: Updated file: %s\n", file)
	}

	return nil
}

func findFilesByExt(directory string, extensions []string) ([]string, error) {
	var files []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && hasExtension(info.Name(), extensions) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func hasExtension(filename string, extensions []string) bool {
	ext := filepath.Ext(filename)
	for _, e := range extensions {
		if strings.EqualFold(ext, e) {
			return true
		}
	}
	return false
}

func processDir(dir string, dryRun bool, extensions []string) error {
	files, err := findFilesByExt(dir, extensions)
	if err != nil {
		return err
	}

	for _, file := range files {
		err := processFile(file, dryRun)
		if err != nil {
			fmt.Printf("Error processing file: %s: %v\n", file, err)
		}
	}

	return nil
}

func removeEmptyLines(content string) string {
	var lines []string
	insidePreTag := false
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if insidePreTag {
			lines = append(lines, line)
		} else {
			if strings.TrimSpace(line) != "" {
				lines = append(lines, line)
			}
		}
		if strings.Contains(line, "<pre>") {
			insidePreTag = true
		}
		if strings.Contains(line, "</pre>") {
			insidePreTag = false
		}
	}
	return strings.Join(lines, "\n")
}

func isInsidePreTag(content []byte) bool {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "<pre>") {
			return true
		}
		if strings.Contains(line, "</pre>") {
			return false
		}
	}
	return false
}
