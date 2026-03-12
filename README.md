# Agent CLI Utils

Fast CLI utilities for AI agents, written in Go. Replaces common Python tools with single-binary, zero-dependency (except stdlib) tools.

## `agent-cli-utils` (this repo)

Query and transform JSON/YAML data from stdin.

```bash
# Get a nested field
cat config.yaml | agent-cli-utils -q .server.port

# Convert JSON to YAML
cat data.json | agent-cli-utils -f yaml

# Convert YAML to JSON
cat config.yaml | agent-cli-utils -f json

# Chain with other tools
curl -s api/data | agent-cli-utils -q .results[0].id
```

## Installation

```bash
go install github.com/SebastianBoehler/agent-cli-utils@latest
```

Or download binary from Releases.

## Why?

- No Python runtime needed on agent systems
- Fast startup, low memory
- Single static binary
- Works in restricted environments
