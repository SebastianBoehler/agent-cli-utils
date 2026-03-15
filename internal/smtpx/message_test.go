package smtpx

import (
	"strings"
	"testing"
)

func TestBuildMessageOmitsBccHeader(t *testing.T) {
	payload, recipients, err := BuildMessage("sender@example.com", Message{
		To:      []string{"to@example.com"},
		Cc:      []string{"cc@example.com"},
		Bcc:     []string{"bcc@example.com"},
		Subject: "Subject",
		Text:    "Hello",
	})
	if err != nil {
		t.Fatalf("BuildMessage() error = %v", err)
	}

	if strings.Contains(string(payload), "\r\nBcc:") {
		t.Fatalf("payload unexpectedly contains Bcc header: %q", payload)
	}
	if got := strings.Join(recipients, ","); got != "to@example.com,cc@example.com,bcc@example.com" {
		t.Fatalf("recipients = %q", got)
	}
}

func TestBuildMessageRejectsHeaderInjection(t *testing.T) {
	_, _, err := BuildMessage("sender@example.com", Message{
		To:      []string{"to@example.com"},
		Subject: "oops\r\nX-Test: 1",
		Text:    "Hello",
	})
	if err == nil {
		t.Fatalf("BuildMessage() error = nil, want header validation error")
	}
}
