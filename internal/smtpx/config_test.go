package smtpx

import (
	"path/filepath"
	"os"
	"testing"
	"time"
)

func TestResolveGmailDefaults(t *testing.T) {
	config, err := Resolve(Config{Provider: "gmail", From: "student@gmail.com"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if config.Host != "smtp.gmail.com" {
		t.Fatalf("Host = %q, want smtp.gmail.com", config.Host)
	}
	if config.Port != 587 {
		t.Fatalf("Port = %d, want 587", config.Port)
	}
	if config.Security != "starttls" {
		t.Fatalf("Security = %q, want starttls", config.Security)
	}
	if config.Auth != "password" {
		t.Fatalf("Auth = %q, want password", config.Auth)
	}
	if config.Username != "student@gmail.com" {
		t.Fatalf("Username = %q, want student@gmail.com", config.Username)
	}
	if config.Timeout != 15*time.Second {
		t.Fatalf("Timeout = %s, want 15s", config.Timeout)
	}
}

func TestInspectUsesConfiguredEnvSecret(t *testing.T) {
	t.Setenv("SMTP_SECRET", "secret")

	profile := Inspect(Config{
		Provider:    "google-workspace",
		Host:        "smtp.gmail.com",
		Port:        587,
		Security:    "starttls",
		Auth:        "password",
		Username:    "student@example.edu",
		From:        "student@example.edu",
		Timeout:     10 * time.Second,
		PasswordEnv: "SMTP_SECRET",
	})

	if !profile.HasSecret {
		t.Fatalf("HasSecret = false, want true")
	}
	if profile.SecretSource != "env:SMTP_SECRET" {
		t.Fatalf("SecretSource = %q, want env:SMTP_SECRET", profile.SecretSource)
	}
}

func TestInspectMissingPasswordFileDoesNotClaimSecret(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.secret")
	profile := Inspect(Config{
		Provider:     "generic",
		Host:         "smtp.example.com",
		Port:         587,
		Security:     "starttls",
		Auth:         "password",
		Username:     "student@example.edu",
		From:         "student@example.edu",
		Timeout:      10 * time.Second,
		PasswordFile: path,
	})

	if profile.HasSecret {
		t.Fatalf("HasSecret = true, want false")
	}
	if profile.SecretSource != "file:"+path {
		t.Fatalf("SecretSource = %q, want %q", profile.SecretSource, "file:"+path)
	}
}

func TestInspectEmptyPasswordFileDoesNotClaimSecret(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.secret")
	if err := os.WriteFile(path, []byte(" \n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	profile := Inspect(Config{
		Provider:     "generic",
		Host:         "smtp.example.com",
		Port:         587,
		Security:     "starttls",
		Auth:         "password",
		Username:     "student@example.edu",
		From:         "student@example.edu",
		Timeout:      10 * time.Second,
		PasswordFile: path,
	})

	if profile.HasSecret {
		t.Fatalf("HasSecret = true, want false")
	}
	if profile.SecretSource != "file:"+path {
		t.Fatalf("SecretSource = %q, want %q", profile.SecretSource, "file:"+path)
	}
}
