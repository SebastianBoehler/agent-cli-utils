# Agent CLI Utils

[![Test](https://github.com/SebastianBoehler/agent-cli-utils/actions/workflows/test.yml/badge.svg)](https://github.com/SebastianBoehler/agent-cli-utils/actions/workflows/test.yml)
[![Build](https://github.com/SebastianBoehler/agent-cli-utils/actions/workflows/build.yml/badge.svg)](https://github.com/SebastianBoehler/agent-cli-utils/actions/workflows/build.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-yellow.svg)](./LICENSE)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg)](./CONTRIBUTING.md)
[![Awesome Ecosystem](https://img.shields.io/badge/awesome-go%20clis-blue)](./AWESOME.md)

Small Go CLIs for agent runtimes on constrained machines: Raspberry Pis, hobby ARM boards, cheap VPSes, and embedded Linux boxes.

The repo is intentionally simple:

- Go only
- tiny binaries
- bounded memory where it matters
- machine-readable output by default
- built for shell-based agents such as PicoClaw and other Go-first automation runners

## Tools

### `agentq`

Query or convert JSON/YAML from stdin or a file.

```bash
# read from stdin
cat config.yaml | agentq -q .server.port -format raw

# read from file
agentq -input state.json -q .jobs[0].id -format raw

# convert yaml to json
agentq -input config.yaml -format json
```

### `agentenv`

Validate required environment variables without leaking values unless explicitly requested.

```bash
agentenv OPENAI_API_KEY HOME
agentenv -file required-env.txt
agentenv -file required-env.txt -format text
```

### `agentfs`

Probe a directory tree and return compact metadata, optional previews, and optional hashes.

```bash
agentfs -root /opt/agent -max-depth 2
agentfs -root /var/log -preview-lines 3 -format text
agentfs -root ./workspace -hash -max-entries 200
```

### `agentrun`

Run a command with a timeout and capped output capture so agents do not wedge themselves on noisy processes.

```bash
agentrun -timeout 5s -- sh -c 'echo ok'
agentrun -timeout 2s -max-output 4096 -- journalctl -n 100
agentrun -format text -- sh -c 'uname -a'
```

### `agentedit`

Apply deterministic targeted file edits from a JSON or YAML spec. This is usually a better fit for agents than generic patch parsing because edits can be exact-match validated before writing.

```bash
cat <<'EOF' > edits.yaml
edits:
  - path: README.md
    action: replace
    old: "Small Go CLIs"
    new: "Small, deterministic Go CLIs"
EOF

agentedit -spec edits.yaml
agentedit -spec edits.yaml -dry-run -format text
```

## Install

Install individual tools:

```bash
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentq@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentenv@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentfs@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentrun@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentedit@latest
```

Build all tools locally:

```bash
mkdir -p bin
go build -o bin/agentq ./cmd/agentq
go build -o bin/agentenv ./cmd/agentenv
go build -o bin/agentfs ./cmd/agentfs
go build -o bin/agentrun ./cmd/agentrun
go build -o bin/agentedit ./cmd/agentedit
```

For smaller static Linux binaries on low-power boards:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/agentq ./cmd/agentq
```

## Typical Agent Workflows

Check that a device is ready:

```bash
agentenv OPENAI_API_KEY AGENT_HOME MODEL_NAME
agentfs -root "$AGENT_HOME" -max-depth 1
agentrun -timeout 3s -- sh -c 'uptime && df -h'
```

Inspect structured state written by another tool:

```bash
agentrun -- sh -c 'curl -s http://127.0.0.1:8080/status' \
  | agentq -q .workers[0].healthy -format raw
```

Summarize a lightweight workspace snapshot:

```bash
agentfs -root . -max-depth 2 -preview-lines 2 > snapshot.json
agentq -input snapshot.json -q .entries[0]
```

Apply a safe targeted change:

```bash
cat <<'EOF' | agentedit -format text
edits:
  - path: config.toml
    action: insert_after
    anchor: "[server]\n"
    new: "port = 8080\n"
EOF
```

## Design Goals

- low startup overhead
- predictable exit codes
- JSON/YAML outputs for agent parsing
- no Python or Node requirement
- suitable for cross-compiling to ARM and minimal Linux images
- exact-match file editing for agent-generated changes

## CI

GitHub Actions is configured for:

- tests on push and pull request
- formatting checks with `gofmt`
- cross-platform builds for Linux, macOS, and Windows
- build artifacts for all CLI tools

## Open Source Project

This repository is set up to be maintained as a proper open source project.

- [Contributing](./CONTRIBUTING.md)
- [Code of Conduct](./CODE_OF_CONDUCT.md)
- [Security Policy](./SECURITY.md)
- [Maintainers](./MAINTAINERS.md)
- [License](./LICENSE)
- [Awesome Go CLI Ecosystem](./AWESOME.md)

## Awesome Ecosystem

This repo now does two things:

- ships small Go CLIs built here for agent runtimes
- curates external Go CLI projects that are useful in the same environments

The curated list lives in [AWESOME.md](./AWESOME.md). It includes upstream Go tools for Google APIs and Google Workspace-adjacent workflows, structured data processing, shell automation, and low-overhead sync and backup.

Links are preferred over git submodules here. That is the normal awesome-list pattern and gives users the right upstream release history, docs, and ownership context without making this repository heavier.

## Maintainers Welcome

If you use these tools in agent runtimes, especially low-compute deployments or PicoClaw-like systems, maintainers and long-term contributors are welcome.

Good contributions include:

- new small CLI tools that solve real agent problems
- better ARM and low-memory support
- workflow examples from embedded and edge deployments
- tests, docs, packaging, and release automation

Start with an issue or a pull request, and read the contribution guide first.
