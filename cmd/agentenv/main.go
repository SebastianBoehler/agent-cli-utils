package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/SebastianBoehler/agent-cli-utils/internal/envcheck"
	"github.com/SebastianBoehler/agent-cli-utils/internal/output"
)

func main() {
	filePath := flag.String("file", "", "load variable names from file, or - for stdin")
	allowEmpty := flag.Bool("allow-empty", false, "treat empty values as valid")
	showValues := flag.Bool("show-values", false, "include raw values in output")
	format := flag.String("format", "json", "json, yaml, or text")
	strict := flag.Bool("strict", true, "exit non-zero if variables are missing")
	flag.Parse()

	names, err := collectNames(*filePath, flag.Args())
	if err != nil {
		fail(err)
	}
	if len(names) == 0 {
		fail(fmt.Errorf("provide variable names as args or via -file"))
	}

	report := envcheck.Check(names, *allowEmpty, *showValues)

	switch *format {
	case "json", "yaml":
		if err := output.Write(*format, report); err != nil {
			fail(err)
		}
	case "text":
		renderText(report)
	default:
		fail(fmt.Errorf("unsupported format %q", *format))
	}

	if *strict && !report.AllPresent {
		os.Exit(1)
	}
}

func collectNames(filePath string, args []string) ([]string, error) {
	names := make([]string, 0, len(args))
	names = append(names, args...)

	if filePath == "" {
		return names, nil
	}

	var (
		reader *os.File
		err    error
	)

	if filePath == "-" {
		reader = os.Stdin
	} else {
		reader, err = os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
	}

	loaded, err := envcheck.LoadNames(reader)
	if err != nil {
		return nil, err
	}

	return append(names, loaded...), nil
}

func renderText(report envcheck.Report) {
	status := "ok"
	if !report.AllPresent {
		status = "missing"
	}

	fmt.Printf("status: %s\n", status)
	if len(report.Missing) > 0 {
		fmt.Printf("missing: %s\n", strings.Join(report.Missing, ", "))
	}
	if len(report.Empty) > 0 {
		fmt.Printf("empty: %s\n", strings.Join(report.Empty, ", "))
	}
	for _, variable := range report.Variables {
		fmt.Printf("%-20s present=%t empty=%t length=%d\n", variable.Name, variable.Present, variable.Empty, variable.Length)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
