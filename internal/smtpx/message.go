package smtpx

import (
	"bytes"
	"fmt"
	"net/mail"
	"strings"
)

func BuildMessage(from string, message Message) ([]byte, []string, error) {
	fromHeader, _, err := normalizeAddress(from)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid from address: %w", err)
	}

	toHeaders, toRecipients, err := normalizeAddresses(message.To)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid to address: %w", err)
	}
	ccHeaders, ccRecipients, err := normalizeAddresses(message.Cc)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid cc address: %w", err)
	}
	_, bccRecipients, err := normalizeAddresses(message.Bcc)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid bcc address: %w", err)
	}

	recipients := append(append([]string{}, toRecipients...), ccRecipients...)
	recipients = append(recipients, bccRecipients...)
	if len(recipients) == 0 {
		return nil, nil, fmt.Errorf("at least one recipient is required")
	}

	subject, err := sanitizeHeader(message.Subject)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid subject: %w", err)
	}

	var builder bytes.Buffer
	writeHeader(&builder, "From", fromHeader)
	if len(toHeaders) > 0 {
		writeHeader(&builder, "To", strings.Join(toHeaders, ", "))
	}
	if len(ccHeaders) > 0 {
		writeHeader(&builder, "Cc", strings.Join(ccHeaders, ", "))
	}
	writeHeader(&builder, "Subject", subject)
	writeHeader(&builder, "MIME-Version", "1.0")
	writeHeader(&builder, "Content-Type", `text/plain; charset="utf-8"`)
	writeHeader(&builder, "Content-Transfer-Encoding", "8bit")
	builder.WriteString("\r\n")
	builder.WriteString(normalizeBody(message.Text))
	return builder.Bytes(), recipients, nil
}

func writeHeader(buffer *bytes.Buffer, key string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	buffer.WriteString(key)
	buffer.WriteString(": ")
	buffer.WriteString(value)
	buffer.WriteString("\r\n")
}

func sanitizeHeader(value string) (string, error) {
	if strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf("header value must not contain newlines")
	}
	return strings.TrimSpace(value), nil
}

func normalizeBody(body string) string {
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.ReplaceAll(normalized, "\n", "\r\n")
	if !strings.HasSuffix(normalized, "\r\n") {
		normalized += "\r\n"
	}
	return normalized
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func normalizeAddresses(values []string) ([]string, []string, error) {
	headers := make([]string, 0, len(values))
	recipients := make([]string, 0, len(values))
	for _, value := range compactStrings(values) {
		header, recipient, err := normalizeAddress(value)
		if err != nil {
			return nil, nil, err
		}
		headers = append(headers, header)
		recipients = append(recipients, recipient)
	}
	return headers, recipients, nil
}

func normalizeAddress(value string) (string, string, error) {
	if strings.ContainsAny(value, "\r\n") {
		return "", "", fmt.Errorf("address must not contain newlines")
	}
	address, err := mail.ParseAddress(strings.TrimSpace(value))
	if err != nil {
		return "", "", err
	}
	return address.String(), address.Address, nil
}
