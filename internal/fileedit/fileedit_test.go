package fileedit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceSingleMatch(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.txt")
	if err := os.WriteFile(path, []byte("hello world\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	report, err := Apply(Request{
		Edits: []Edit{{
			Path:   path,
			Action: "replace",
			Old:    "world",
			New:    "agent",
		}},
	}, Options{FailOnNoop: true})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if report.Changed != 1 {
		t.Fatalf("Changed = %d, want 1", report.Changed)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "hello agent\n" {
		t.Fatalf("content = %q, want %q", string(got), "hello agent\n")
	}
}

func TestReplaceRejectsAmbiguousMatch(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.txt")
	if err := os.WriteFile(path, []byte("x x\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Apply(Request{
		Edits: []Edit{{
			Path:   path,
			Action: "replace",
			Old:    "x",
			New:    "y",
		}},
	}, Options{FailOnNoop: true})
	if err == nil || !strings.Contains(err.Error(), "exactly 1 match") {
		t.Fatalf("Apply() error = %v, want ambiguity failure", err)
	}
}

func TestInsertAfter(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.txt")
	if err := os.WriteFile(path, []byte("a\nb\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Apply(Request{
		Edits: []Edit{{
			Path:   path,
			Action: "insert_after",
			Anchor: "a\n",
			New:    "x\n",
		}},
	}, Options{FailOnNoop: true})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "a\nx\nb\n" {
		t.Fatalf("content = %q, want %q", string(got), "a\nx\nb\n")
	}
}

func TestReplaceLines(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.txt")
	if err := os.WriteFile(path, []byte("1\n2\n3\n4\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Apply(Request{
		Edits: []Edit{{
			Path:      path,
			Action:    "replace_lines",
			StartLine: 2,
			EndLine:   3,
			New:       "x\ny\n",
		}},
	}, Options{FailOnNoop: true})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "1\nx\ny\n4\n" {
		t.Fatalf("content = %q, want %q", string(got), "1\nx\ny\n4\n")
	}
}

func TestDryRunDoesNotWrite(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.txt")
	if err := os.WriteFile(path, []byte("before\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	report, err := Apply(Request{
		Edits: []Edit{{
			Path:   path,
			Action: "write",
			New:    "after\n",
		}},
	}, Options{DryRun: true, FailOnNoop: true})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !report.DryRun {
		t.Fatalf("DryRun = false, want true")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "before\n" {
		t.Fatalf("content = %q, want %q", string(got), "before\n")
	}
}

func TestWriteCreatesFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "nested", "new.txt")

	report, err := Apply(Request{
		Edits: []Edit{{
			Path:   path,
			Action: "write",
			New:    "created\n",
			Create: true,
		}},
	}, Options{FailOnNoop: true})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if report.Results[0].Created != true {
		t.Fatalf("Created = false, want true")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "created\n" {
		t.Fatalf("content = %q, want %q", string(got), "created\n")
	}
}
