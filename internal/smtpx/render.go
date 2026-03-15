package smtpx

import (
	"fmt"
	"strings"
)

func RenderProfileText(profile Profile) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "operation: %s\n", profile.Operation)
	fmt.Fprintf(&builder, "provider: %s\n", profile.Provider)
	fmt.Fprintf(&builder, "host: %s\n", profile.Host)
	fmt.Fprintf(&builder, "port: %d\n", profile.Port)
	fmt.Fprintf(&builder, "security: %s\n", profile.Security)
	fmt.Fprintf(&builder, "auth: %s\n", profile.Auth)
	if profile.Username != "" {
		fmt.Fprintf(&builder, "username: %s\n", profile.Username)
	}
	if profile.From != "" {
		fmt.Fprintf(&builder, "from: %s\n", profile.From)
	}
	fmt.Fprintf(&builder, "timeout: %s\n", profile.Timeout)
	fmt.Fprintf(&builder, "has_secret: %t\n", profile.HasSecret)
	if profile.SecretSource != "" {
		fmt.Fprintf(&builder, "secret_source: %s\n", profile.SecretSource)
	}
	return builder.String()
}

func RenderResultText(result Result) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "operation: %s\n", result.Operation)
	fmt.Fprintf(&builder, "status: %s\n", result.Status)
	fmt.Fprintf(&builder, "provider: %s\n", result.Provider)
	fmt.Fprintf(&builder, "host: %s\n", result.Host)
	fmt.Fprintf(&builder, "port: %d\n", result.Port)
	fmt.Fprintf(&builder, "security: %s\n", result.Security)
	fmt.Fprintf(&builder, "auth: %s\n", result.Auth)
	if result.Username != "" {
		fmt.Fprintf(&builder, "username: %s\n", result.Username)
	}
	if result.From != "" {
		fmt.Fprintf(&builder, "from: %s\n", result.From)
	}
	if len(result.To) > 0 {
		fmt.Fprintf(&builder, "to: %s\n", strings.Join(result.To, ", "))
	}
	if len(result.Cc) > 0 {
		fmt.Fprintf(&builder, "cc: %s\n", strings.Join(result.Cc, ", "))
	}
	if len(result.Bcc) > 0 {
		fmt.Fprintf(&builder, "bcc: %s\n", strings.Join(result.Bcc, ", "))
	}
	if result.Subject != "" {
		fmt.Fprintf(&builder, "subject: %s\n", result.Subject)
	}
	if result.MessageBytes > 0 {
		fmt.Fprintf(&builder, "message_bytes: %d\n", result.MessageBytes)
	}
	return builder.String()
}
