package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/SebastianBoehler/agent-cli-utils/internal/fileedit"
	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
)

func main() {
	specPath := flag.String("spec", "", "JSON/YAML edit spec file, or read stdin when omitted")
	format := flag.String("format", "json", "json, yaml, or text")
	dryRun := flag.Bool("dry-run", false, "show what would change without writing files")
	failOnNoop := flag.Bool("fail-on-noop", true, "fail when a targeted edit finds no match")
	flag.Parse()

	payload, err := readSpec(*specPath)
	if err != nil {
		fail(err)
	}

	request, err := fileedit.LoadRequest(payload)
	if err != nil {
		fail(err)
	}

	report, err := fileedit.Apply(request, fileedit.Options{
		DryRun:     *dryRun,
		FailOnNoop: *failOnNoop,
	})
	if err != nil {
		fail(err)
	}

	switch *format {
	case "json", "yaml":
		if err := output.Write(*format, report); err != nil {
			fail(err)
		}
	case "text":
		fmt.Print(fileedit.RenderText(report))
	default:
		fail(fmt.Errorf("unsupported format %q", *format))
	}
}

func readSpec(path string) ([]byte, error) {
	if path == "" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
