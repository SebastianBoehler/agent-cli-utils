package datax

import (
	"strings"
	"testing"
)

func TestParseAndQueryJSON(t *testing.T) {
	payload := []byte(`{"items":[{"id":"abc123","enabled":true}]}`)

	value, format, err := ParseStructured(payload)
	if err != nil {
		t.Fatalf("ParseStructured() error = %v", err)
	}
	if format != "json" {
		t.Fatalf("format = %q, want json", format)
	}

	got, err := Query(value, ".items[0].id")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if got != "abc123" {
		t.Fatalf("Query() = %v, want abc123", got)
	}
}

func TestParseYAMLNormalizesMaps(t *testing.T) {
	payload := []byte("server:\n  port: 8080\n")

	value, format, err := ParseStructured(payload)
	if err != nil {
		t.Fatalf("ParseStructured() error = %v", err)
	}
	if format != "yaml" {
		t.Fatalf("format = %q, want yaml", format)
	}

	got, err := Query(value, ".server.port")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if got != 8080 {
		t.Fatalf("Query() = %v, want 8080", got)
	}
}

func TestRenderRaw(t *testing.T) {
	payload, err := Render("ready", "raw")
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if strings.TrimSpace(string(payload)) != "ready" {
		t.Fatalf("Render() = %q, want ready", string(payload))
	}
}
