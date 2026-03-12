---
name: agentenv-cli
description: Validate required environment variables with the local `agentenv` CLI. Use when an agent workflow needs a startup check, deployment validation, or a machine-readable report of missing or empty variables.
---

# AgentEnv CLI

Use `agentenv` to check environment prerequisites before running a model, sync task, or automation step.

## Invoke the Tool

Prefer the installed binary when it exists:

```bash
agentenv OPENAI_API_KEY HOME
```

If working inside this repository or the binary is not installed, run:

```bash
go run ./cmd/agentenv -- OPENAI_API_KEY HOME
```

## Choose Flags

- Pass variable names as positional arguments for simple checks.
- Use `-file required-env.txt` to load names from a file.
- Use `-allow-empty` only when blank values are acceptable.
- Use `-show-values` only when exposing raw values is safe.
- Use `-format json` or `-format yaml` for automation.
- Use `-format text` for human review.

## Apply Good Defaults

- Keep `-strict=true` for CI, startup checks, and agent preflight.
- Avoid `-show-values` unless the environment is trusted and the output will not be logged.
- Prefer JSON output when a caller will branch on missing keys.

## Examples

Validate a known set:

```bash
go run ./cmd/agentenv -- OPENAI_API_KEY AGENT_HOME MODEL_NAME
```

Load from file:

```bash
go run ./cmd/agentenv -- -file required-env.txt -format json
```
