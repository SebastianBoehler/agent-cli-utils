package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReadsFromStdin(t *testing.T) {
	var stdout bytes.Buffer
	err := run([]string{"-name", "payload.json", "-format", "json"}, strings.NewReader(`{"ok":true}`), &stdout)
	if err != nil {
		t.Fatalf("run from stdin: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, `"format": "json"`) {
		t.Fatalf("expected json output, got %q", output)
	}
	if !strings.Contains(output, `"source": "payload.json"`) {
		t.Fatalf("expected logical source name, got %q", output)
	}
}

func TestRunReadsFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.html")
	if err := os.WriteFile(path, []byte("<html><body><h1>Title</h1><p>Hello</p></body></html>"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var stdout bytes.Buffer
	err := run([]string{"-input", path}, strings.NewReader("ignored"), &stdout)
	if err != nil {
		t.Fatalf("run from file: %v", err)
	}

	if !strings.Contains(stdout.String(), "# Title") {
		t.Fatalf("expected markdown heading, got %q", stdout.String())
	}
}

func TestRunRejectsUnknownOutputFormat(t *testing.T) {
	var stdout bytes.Buffer
	err := run([]string{"-name", "notes.txt", "-format", "raw"}, strings.NewReader("hello"), &stdout)
	if err == nil {
		t.Fatal("expected error for unsupported output format")
	}
	if !strings.Contains(err.Error(), `unsupported format "raw"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
