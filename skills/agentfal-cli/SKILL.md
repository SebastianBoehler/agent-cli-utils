---
name: agentfal-cli
description: Use the local `agentfal` CLI to submit, inspect, and cancel fal queue requests with structured output. Prefer it when an agent needs a thin wrapper around fal's queue API and env-based auth.
---

# AgentFal CLI

Use `agentfal` when a workflow needs a small deterministic wrapper around fal queue requests.

## Invoke the Tool

Prefer the installed binary when it exists:

```bash
agentfal submit -model "$FAL_MODEL" -input payload.json
```

If working inside this repository or the binary is not installed, run:

```bash
go run ./cmd/agentfal -- submit -model "$FAL_MODEL" -input payload.json
```

## Choose Flags

- Use `submit` for new queue requests.
- Use `-input` or `-payload` to pass JSON or YAML arguments.
- Use `status`, `result`, or `cancel` with `-request`.
- Use `-logs` on `status` when the queue endpoint supports log streaming in status responses.
- Set `FAL_KEY` or `FAL_API_KEY`, `FAL_MODEL`, and optionally `FAL_BASE_URL` through the environment when possible.
- Prefer `-format json` or `-format yaml` for downstream agent parsing.

## Examples

Submit a request:

```bash
go run ./cmd/agentfal -- submit -payload '{"prompt":"hello"}'
```

Fetch request result:

```bash
go run ./cmd/agentfal -- result -request "$REQUEST_ID"
```
