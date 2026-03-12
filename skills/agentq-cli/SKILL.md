---
name: agentq-cli
description: Query and convert JSON or YAML with the local `agentq` CLI. Use when a task needs field extraction, format conversion, or machine-readable structured data filtering in this repository or from an installed `agentq` binary.
---

# AgentQ CLI

Use `agentq` to extract fields from JSON or YAML, or to convert between formats without pulling in Python or `jq`-style tooling.

## Invoke the Tool

Prefer the installed binary when it exists:

```bash
agentq -input state.json -q .jobs[0].id -format raw
```

If working inside this repository or the binary is not installed, run:

```bash
go run ./cmd/agentq -- -input state.json -q .jobs[0].id -format raw
```

## Choose Flags

- Use `-input` to read from a file. Omit it to read from stdin.
- Use `-q` with paths like `.server.port` or `.items[0].id`.
- Use `-format json` or `-format yaml` for structured output.
- Use `-format raw` when the caller only needs a scalar string or number.

## Apply Good Defaults

- Prefer `raw` for shell pipelines and agent follow-up steps.
- Prefer `json` when another tool will parse the result.
- Fail fast if the path is ambiguous or missing instead of guessing.

## Examples

Convert YAML to JSON:

```bash
go run ./cmd/agentq -- -input config.yaml -format json
```

Extract one value from stdin:

```bash
cat status.json | go run ./cmd/agentq -- -q .workers[0].healthy -format raw
```
