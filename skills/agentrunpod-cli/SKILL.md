---
name: agentrunpod-cli
description: Use the local `agentrunpod` CLI to submit, inspect, and cancel RunPod serverless jobs with structured output. Prefer it when an agent needs an env-configurable wrapper around the RunPod HTTP API instead of ad hoc curl commands.
---

# AgentRunPod CLI

Use `agentrunpod` when a workflow needs a small deterministic wrapper around RunPod endpoint execution.

## Invoke the Tool

Prefer the installed binary when it exists:

```bash
agentrunpod submit -endpoint "$RUNPOD_ENDPOINT_ID" -input payload.json
```

If working inside this repository or the binary is not installed, run:

```bash
go run ./cmd/agentrunpod -- submit -endpoint "$RUNPOD_ENDPOINT_ID" -input payload.json
```

## Choose Flags

- Use `submit` for new jobs.
- Use `-sync` to hit RunPod's synchronous `runsync` path.
- Use `-input` or `-payload` to pass JSON or YAML input.
- Use `-raw-body` only when the payload already contains the full RunPod request body.
- Use `status`, `result`, or `cancel` with `-request`.
- Set `RUNPOD_API_KEY`, `RUNPOD_ENDPOINT_ID`, and optionally `RUNPOD_BASE_URL` through the environment when possible.
- Prefer `-format json` or `-format yaml` for downstream agent parsing.

## Examples

Submit a job:

```bash
go run ./cmd/agentrunpod -- submit -payload '{"prompt":"hello"}'
```

Fetch job status:

```bash
go run ./cmd/agentrunpod -- status -request "$JOB_ID"
```
