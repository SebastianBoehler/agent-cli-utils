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
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fail(err)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	flags := flag.NewFlagSet("agentmd", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	inputPath := flags.String("input", "", "read from file instead of stdin")
	name := flags.String("name", "", "logical filename used for format detection")
	format := flags.String("format", "markdown", "markdown, json, or yaml")
	inputType := flags.String("type", "", "force input type like txt, html, docx, xlsx, pptx, or zip")
	maxArchiveEntries := flags.Int("max-archive-entries", 64, "maximum files converted from a zip archive")
	archiveDepth := flags.Int("archive-depth", 2, "maximum nested archive depth")
	if err := flags.Parse(args); err != nil {
		return err
	}

	data, inferredName, err := readInput(stdin, *inputPath)
	if err != nil {
		return err
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
		return err
	}

	switch *format {
	case "markdown":
		_, err = fmt.Fprintln(stdout, result.Markdown)
		return err
	case "json":
		return writeStructured(stdout, *format, result)
	case "yaml":
		return writeStructured(stdout, *format, result)
	default:
		return fmt.Errorf("unsupported format %q", *format)
	}
}

func readInput(stdin io.Reader, path string) ([]byte, string, error) {
	if path == "" {
		data, err := io.ReadAll(stdin)
		return data, "", err
	}

	data, err := os.ReadFile(path)
	return data, path, err
}

func writeStructured(stdout io.Writer, format string, value any) error {
	return output.WriteTo(stdout, format, value)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
