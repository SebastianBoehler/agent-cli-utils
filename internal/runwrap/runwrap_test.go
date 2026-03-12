package runwrap

import (
	"strings"
	"testing"
	"time"
)

func TestRunSuccess(t *testing.T) {
	result, err := Run([]string{"sh", "-c", "printf 'hello'"}, Options{
		Timeout:        2 * time.Second,
		MaxOutputBytes: 1024,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello" {
		t.Fatalf("Stdout = %q, want hello", result.Stdout)
	}
}

func TestRunTruncatesOutput(t *testing.T) {
	result, err := Run([]string{"sh", "-c", "printf '1234567890'"}, Options{
		Timeout:        2 * time.Second,
		MaxOutputBytes: 4,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.StdoutTruncated {
		t.Fatalf("StdoutTruncated = false, want true")
	}
	if result.Stdout != "1234" {
		t.Fatalf("Stdout = %q, want 1234", result.Stdout)
	}
}

func TestRunTimeout(t *testing.T) {
	result, err := Run([]string{"sh", "-c", "sleep 1"}, Options{
		Timeout:        50 * time.Millisecond,
		MaxOutputBytes: 1024,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.TimedOut {
		t.Fatalf("TimedOut = false, want true")
	}
	if !strings.Contains(result.Error, "timed out") {
		t.Fatalf("Error = %q, want timeout marker", result.Error)
	}
}
