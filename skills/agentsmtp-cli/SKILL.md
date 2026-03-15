---
name: agentsmtp-cli
description: Send or validate SMTP mail flows with the `agentsmtp` CLI. Use when an agent needs a small SMTP submission tool with provider presets for Gmail or Google Workspace, structured output, and configurable auth through password or app-password.
---

# AgentSMTP CLI

Use `agentsmtp` when a workflow needs a thin SMTP client instead of a full mail app. The tool resolves provider defaults, loads config from file or environment, and keeps output machine-readable.

## Invoke the Tool

Prefer the installed binary when it exists:

```bash
agentsmtp profile -provider gmail -from student@gmail.com
```

If working inside this repository or the binary is not installed, run:

```bash
go run ./cmd/agentsmtp -- profile -provider gmail -from student@gmail.com
```

## Choose Commands

- Use `profile` to inspect the resolved SMTP host, port, auth mode, and whether a secret source is configured.
- Use `test` to open an SMTP session, authenticate, and run `NOOP` without sending mail.
- Use `send` to submit a plain-text message with repeated `-to`, `-cc`, and `-bcc` flags.

## Apply Good Defaults

- Prefer `-provider gmail` or `-provider google-workspace` for Google-hosted mail because the tool fills `smtp.gmail.com:587` and `STARTTLS`.
- Keep secrets in `-password-env` or `-password-file` rather than inline flags when possible.
- Pipe message text over stdin for agent workflows instead of shell-escaping large bodies.
- Use `-format json` for downstream parsing and `-format text` for quick debugging.

## Examples

Inspect a Google Workspace profile:

```bash
go run ./cmd/agentsmtp -- profile -provider google-workspace -from student@example.edu -password-env SMTP_PASSWORD -format text
```

Test Gmail SMTP with an app password:

```bash
go run ./cmd/agentsmtp -- test -provider gmail -from student@gmail.com -password-env GMAIL_APP_PASSWORD -format text
```

Send a plain-text message from stdin:

```bash
printf 'Hello from agentsmtp\n' | go run ./cmd/agentsmtp -- send -provider google-workspace -from student@example.edu -password-env SMTP_PASSWORD -to advisor@example.edu -subject "Status"
```
