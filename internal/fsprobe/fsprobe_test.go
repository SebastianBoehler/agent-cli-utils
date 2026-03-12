package fsprobe

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProbe(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config.txt"), []byte("line1\nline2\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "nested"), 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "nested", "state.json"), []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := Probe(root, Options{
		MaxDepth:     2,
		PreviewLines: 1,
		MaxEntries:   10,
	})
	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if len(result.Entries) != 4 {
		t.Fatalf("len(Entries) = %d, want 4", len(result.Entries))
	}
	if result.Files != 2 {
		t.Fatalf("Files = %d, want 2", result.Files)
	}
	if result.Directories != 2 {
		t.Fatalf("Directories = %d, want 2", result.Directories)
	}
}

func TestProbeHiddenSkip(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "config"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := Probe(root, Options{MaxDepth: 2, MaxEntries: 10})
	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	if len(result.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(result.Entries))
	}
}
