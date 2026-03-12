package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/SebastianBoehler/agent-cli-utils/internal/datax"
)

func main() {
	inputPath := flag.String("input", "", "read from file instead of stdin")
	query := flag.String("q", "", "dot path like .items[0].id")
	format := flag.String("format", "json", "json, yaml, or raw")
	flag.Parse()

	input, err := readInput(*inputPath)
	if err != nil {
		fail(err)
	}

	value, _, err := datax.ParseStructured(input)
	if err != nil {
		fail(err)
	}

	if *query != "" {
		value, err = datax.Query(value, *query)
		if err != nil {
			fail(err)
		}
	}

	payload, err := datax.Render(value, *format)
	if err != nil {
		fail(err)
	}

	if _, err := os.Stdout.Write(payload); err != nil {
		fail(err)
	}
}

func readInput(path string) ([]byte, error) {
	if path == "" {
		return io.ReadAll(os.Stdin)
	}

	return os.ReadFile(path)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
