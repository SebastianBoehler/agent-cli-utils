package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/SebastianBoehler/agent-cli-utils/internal/mdconv"
	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
)

func main() {
	inputPath := flag.String("input", "", "read from file instead of stdin")
	name := flag.String("name", "", "logical filename used for format detection")
	format := flag.String("format", "markdown", "markdown, json, or yaml")
	inputType := flag.String("type", "", "force input type like txt, html, docx, xlsx, pptx, or zip")
	maxArchiveEntries := flag.Int("max-archive-entries", 64, "maximum files converted from a zip archive")
	archiveDepth := flag.Int("archive-depth", 2, "maximum nested archive depth")
	flag.Parse()

	data, inferredName, err := readInput(*inputPath)
	if err != nil {
		fail(err)
	}

	if *name == "" {
		*name = inferredName
	}

	result, err := mdconv.NewService().Convert(data, mdconv.Options{
		Name:              *name,
		Format:            *inputType,
		MaxArchiveEntries: *maxArchiveEntries,
		ArchiveDepth:      *archiveDepth,
	})
	if err != nil {
		fail(err)
	}

	switch *format {
	case "markdown":
		if _, err := fmt.Fprintln(os.Stdout, result.Markdown); err != nil {
			fail(err)
		}
	case "json", "yaml":
		if err := output.Write(*format, result); err != nil {
			fail(err)
		}
	default:
		fail(fmt.Errorf("unsupported format %q", *format))
	}
}

func readInput(path string) ([]byte, string, error) {
	if path == "" {
		data, err := io.ReadAll(os.Stdin)
		return data, "", err
	}

	data, err := os.ReadFile(path)
	return data, path, err
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
