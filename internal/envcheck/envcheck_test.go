package envcheck

import (
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	t.Setenv("TEST_PRESENT", "value")
	t.Setenv("TEST_EMPTY", "")

	report := Check([]string{"TEST_PRESENT", "TEST_EMPTY", "TEST_MISSING"}, false, false)
	if report.AllPresent {
		t.Fatalf("AllPresent = true, want false")
	}
	if len(report.Missing) != 1 || report.Missing[0] != "TEST_MISSING" {
		t.Fatalf("Missing = %v", report.Missing)
	}
	if len(report.Empty) != 1 || report.Empty[0] != "TEST_EMPTY" {
		t.Fatalf("Empty = %v", report.Empty)
	}
}

func TestLoadNames(t *testing.T) {
	reader := strings.NewReader("# comment\nOPENAI_API_KEY\n\nAGENT_NAME\n")

	names, err := LoadNames(reader)
	if err != nil {
		t.Fatalf("LoadNames() error = %v", err)
	}

	if len(names) != 2 || names[0] != "OPENAI_API_KEY" || names[1] != "AGENT_NAME" {
		t.Fatalf("LoadNames() = %v", names)
	}
}
