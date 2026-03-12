---
name: agentdoctor-cli
description: Check host dependencies with the local `agentdoctor` CLI. Use when an agent workflow fails because a required tool such as an SMB client, SSH client, or HTTP fetch utility is missing or when a task needs a machine-readable preflight report.
---

# AgentDoctor CLI

Use `agentdoctor` to diagnose missing host capabilities before a workflow starts or immediately after a dependency-related failure.

## Invoke the Tool

Prefer the installed binary when it exists:

```bash
agentdoctor -profile smb-client -format text
```

If working inside this repository or the binary is not installed, run:

```bash
go run ./cmd/agentdoctor -- -profile smb-client -format text
```

## Choose Profiles

- Use `-profile smb-client` when a workflow needs SMB or CIFS access to a Windows host or share.
- Use `-profile ssh-client` when the task depends on SSH or SCP.
- Use `-profile web-fetch` when the task depends on `curl` or PowerShell HTTP support.
- Use repeated `-cmd` flags for custom one-off dependency checks.

## Apply Good Defaults

- Keep `-strict=true` for startup and health checks so missing tools fail early.
- Use `-format json` when another agent step will branch on the result.
- Use `-strict=false` when the task only needs advisory diagnostics.
- Prefer `agentdoctor` over a blind failing connection attempt when the problem may be “tool missing” rather than “credentials wrong.”

## Examples

Diagnose the SMB client case:

```bash
go run ./cmd/agentdoctor -- -profile smb-client -format text
```

Check a custom command set:

```bash
go run ./cmd/agentdoctor -- -cmd git -cmd curl -format json
```
