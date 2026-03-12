package doctor

import (
	"errors"
	"testing"
)

func TestEvaluateSMBClientLinux(t *testing.T) {
	resolver := func(name string) (string, error) {
		switch name {
		case "smbclient":
			return "", errors.New("missing")
		case "mount.cifs":
			return "/usr/sbin/mount.cifs", nil
		default:
			return "", errors.New("missing")
		}
	}

	report, err := Evaluate("smb-client", nil, "linux", resolver)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if report.AllRequiredPresent {
		t.Fatalf("AllRequiredPresent = true, want false")
	}
	if len(report.Checks) != 2 {
		t.Fatalf("len(Checks) = %d, want 2", len(report.Checks))
	}
	if report.Checks[0].Name != "smbclient" || report.Checks[0].Present {
		t.Fatalf("first check = %#v, want missing smbclient", report.Checks[0])
	}
}

func TestEvaluateWithCustomCommand(t *testing.T) {
	resolver := func(name string) (string, error) {
		if name == "git" {
			return "/usr/bin/git", nil
		}
		return "", errors.New("missing")
	}

	report, err := Evaluate("", []string{"git"}, "linux", resolver)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if !report.AllRequiredPresent {
		t.Fatalf("AllRequiredPresent = false, want true")
	}
	if len(report.Checks) != 1 || report.Checks[0].Path != "/usr/bin/git" {
		t.Fatalf("Checks = %#v", report.Checks)
	}
}

func TestUnknownProfile(t *testing.T) {
	_, err := Evaluate("nope", nil, "linux", func(string) (string, error) {
		return "", nil
	})
	if err == nil {
		t.Fatal("Evaluate() error = nil, want error")
	}
}
