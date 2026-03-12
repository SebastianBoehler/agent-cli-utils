package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/SebastianBoehler/agent-cli-utils/internal/fsprobe"
	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
)

func main() {
	root := flag.String("root", ".", "root path to inspect")
	maxDepth := flag.Int("max-depth", 2, "maximum directory depth")
	includeHidden := flag.Bool("hidden", false, "include hidden files and directories")
	hash := flag.Bool("hash", false, "compute SHA256 for regular files")
	previewLines := flag.Int("preview-lines", 0, "capture the first N lines for text files")
	maxEntries := flag.Int("max-entries", 1000, "stop after N entries")
	format := flag.String("format", "json", "json, yaml, or text")
	flag.Parse()

	result, err := fsprobe.Probe(*root, fsprobe.Options{
		MaxDepth:      *maxDepth,
		IncludeHidden: *includeHidden,
		Hash:          *hash,
		PreviewLines:  *previewLines,
		MaxEntries:    *maxEntries,
	})
	if err != nil {
		fail(err)
	}

	switch *format {
	case "json", "yaml":
		if err := output.Write(*format, result); err != nil {
			fail(err)
		}
	case "text":
		fmt.Print(fsprobe.RenderText(result))
	default:
		fail(fmt.Errorf("unsupported format %q", *format))
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
