---
name: agentrun-cli
description: Execute shell commands with the local `agentrun` CLI. Use when an agent needs timeout control, bounded stdout and stderr capture, or structured command results instead of direct raw process execution.
---

# AgentRun CLI

Use `agentrun` when direct command execution would be too noisy or risky because the process can hang, flood output, or return unstructured results.

## Invoke the Tool

Prefer the installed binary when it exists:

```bash
agentrun -timeout 5s -- sh -c 'echo ok'
```

If working inside this repository or the binary is not installed, run:

```bash
go run ./cmd/agentrun -- -timeout 5s -- sh -c 'echo ok'
```

## Choose Flags

- Use `-timeout` to cap command duration.
- Use `-max-output` to cap captured stdout and stderr bytes.
- Use `-dir` when the command must run in a specific directory.
- Use `-stdin` only if the child process must read caller input.
- Use `-format json` or `-format yaml` for programmatic consumers.
- Use `-format text` for quick inspection.

## Apply Good Defaults

- Keep short timeouts for probes and readiness checks.
- Keep small output caps for logs and diagnostic commands.
- Prefer `agentrun` over raw `sh -c` when the result needs to be parsed later.

## Examples

Run a health check:

```bash
go run ./cmd/agentrun -- -timeout 3s -format json -- sh -c 'uptime && df -h'
```

Capture a bounded log sample:

```bash
go run ./cmd/agentrun -- -timeout 2s -max-output 4096 -- journalctl -n 100
```
