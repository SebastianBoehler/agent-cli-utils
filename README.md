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

### `agentdoctor`

Check host dependencies and missing tools before a workflow starts. This is useful for issues like “SMB client missing” when connecting Windows shares or other external systems.

```bash
agentdoctor -profile smb-client -format text
agentdoctor -profile smb-client -strict=false
agentdoctor -cmd git -cmd curl -format json
```

### `agentrunpod`

Submit, inspect, and cancel RunPod serverless jobs with structured output and env-based configuration.

```bash
agentrunpod submit -endpoint "$RUNPOD_ENDPOINT_ID" -payload '{"prompt":"hello"}'
agentrunpod status -endpoint "$RUNPOD_ENDPOINT_ID" -request "$JOB_ID"
agentrunpod submit -sync -endpoint "$RUNPOD_ENDPOINT_ID" -input payload.yaml
```

### `agentfal`

Submit, inspect, and cancel fal queue requests with structured output and env-based configuration.

```bash
agentfal submit -model "$FAL_MODEL" -payload '{"prompt":"hello"}'
agentfal status -model "$FAL_MODEL" -request "$REQUEST_ID" -logs
agentfal result -model "$FAL_MODEL" -request "$REQUEST_ID"
```

### `agentprint`

Wrap local CUPS tools for queue discovery, queue repair, and bounded print submission with machine-readable output by default.

```bash
agentprint list
agentprint discover -format tsv
agentprint ensure -printer office -match "Office Laser" -default
agentprint print -printer office -input ./label.pdf -duplex -media A4 -fit-to-page
agentprint print -printer office https://example.com/report.pdf -color-mode monochrome
```

Output defaults to JSON. `list` and `discover` also support `-format text` and `-format tsv` for quick terminal inspection.

### `agentsmtp`

Send or validate SMTP submission with machine-readable output, config-file loading, and provider defaults for Gmail or Google Workspace.

```bash
agentsmtp profile -provider google-workspace -from student@example.edu -password-env SMTP_PASSWORD -format text
agentsmtp test -provider gmail -from student@gmail.com -password-env GMAIL_APP_PASSWORD
printf 'hello from agentsmtp\n' | agentsmtp send -provider google-workspace -from student@example.edu -password-env SMTP_PASSWORD -to advisor@example.edu -subject "Status"
```

### `agenttv`

Discover AirPlay, DLNA, and DIAL endpoints on the local network, wake compatible devices, pair AirPlay where needed, and hand off media.

This focuses on realistic network control:

- probe per-device protocol reachability and auth requirements
- AirPlay URL playback handoff
- experimental Apple TV AirPlay pairing via `atvremote` and playback via an embedded `pyatv` helper
- DLNA / UPnP MediaRenderer playback and stop
- Wake-on-LAN for compatible TVs

It does not implement universal screen mirroring, which is device-specific and usually tied to OS-level stacks rather than an open TV API.

```bash
agenttv discover -format text
agenttv probe -device "Samsung 6 Series"
agenttv pair -protocol airplay -host 192.168.178.33 -pin 1234
agenttv play -device "Living Room TV" -url http://192.168.1.20:8080/stream.m3u8
agenttv play -host 192.168.1.50:7000 -protocol airplay -url https://example.com/video.mp4
agenttv stop -device "Living Room TV"
agenttv wake -mac AA:BB:CC:DD:EE:FF
```

`agenttv play` will use stored Apple AirPlay credentials when present for the target host. Apple TV pairing currently relies on a locally installed `atvremote` binary, while Apple playback uses an embedded `pyatv` helper that keeps the AirPlay session open longer and reports the raw `/play` response. Apple playback should still be treated as experimental.

### `agentsamsung`

Samsung-native TV control with token pairing, remote keys, and app launch commands.

Examples:

```bash
agentsamsung pair -host 192.168.178.86
agentsamsung remote -host 192.168.178.86 -key KEY_HOME
agentsamsung remote -host 192.168.178.86 -key KEY_SOURCE
agentsamsung launch -host 192.168.178.86 -app-id 111299001912
```

`agentsamsung` is the place for Samsung-specific auth and control flows. `agenttv` stays focused on generic AirPlay, DLNA, DIAL, and Wake-on-LAN behavior.

### `agentyt`

Discover DIAL YouTube receivers on the local network and send public DIAL-style launch requests using YouTube video IDs or URLs.

This stays on the public path:

- discover receivers exposing a YouTube DIAL app endpoint
- inspect current YouTube app state on a receiver
- launch a YouTube video using public-style `launch=dial` and `v=<video-id>` parameters

Examples:

```bash
agentyt discover -format text
agentyt status -device "Samsung 6 Series"
agentyt play -host 192.168.178.86 -video https://www.youtube.com/watch?v=wU90dfDDUiQ
agentyt play -device "Samsung 6 Series" -video wU90dfDDUiQ -start 43
```

`agentyt` does not implement YouTube's private pairing or queue-control flows. It only uses the public DIAL-style launch surface that some TVs expose.

### `agentmd`

Convert common files to Markdown using a native Go binary with no Python or Node runtime.

Day-1 support focuses on low-friction and common formats:

- plain text and Markdown
- HTML
- CSV
- JSON, YAML, and XML
- ZIP archives with recursive conversion of supported entries
- basic OOXML extraction for DOCX, XLSX, and PPTX

```bash
agentmd -input ./notes.html
agentmd -input ./report.docx
agentmd -input ./sheet.xlsx
cat ./payload.json | agentmd -name payload.json
agentmd -input ./bundle.zip -format json
```

### `company`

Search company registries with normalized output and best-effort multi-source fallback.

```bash
company search "Acme GmbH"
company search "Acme GmbH" --city Berlin --exact --format text
company search "Acme" --source opencorporates --limit 5
company quota
```

`company search` defaults to `source=all` and currently queries:

- `handelsregister` for Germany's official registry search
- `opencorporates` for enrichment and stable IDs
- `offeneregister` as a best-effort Datasette source when available

If one backend fails, the command still returns the others and includes per-source errors in the structured output. `company quota` calls OpenCorporates account status and expects `OPENCORPORATES_API_TOKEN` or `--opencorporates-api-token`.
## Install

Install individual tools:

```bash
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentq@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentenv@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentfs@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentrun@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentedit@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentdoctor@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentrunpod@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentfal@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentprint@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentmd@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/company@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agentsmtp@latest
go install github.com/SebastianBoehler/agent-cli-utils/cmd/agenttv@latest
```

Build all tools locally:

```bash
mkdir -p bin
go build -o bin/agentq ./cmd/agentq
go build -o bin/agentenv ./cmd/agentenv
go build -o bin/agentfs ./cmd/agentfs
go build -o bin/agentrun ./cmd/agentrun
go build -o bin/agentedit ./cmd/agentedit
go build -o bin/agentdoctor ./cmd/agentdoctor
go build -o bin/agentrunpod ./cmd/agentrunpod
go build -o bin/agentfal ./cmd/agentfal
go build -o bin/agentprint ./cmd/agentprint
go build -o bin/agentmd ./cmd/agentmd
go build -o bin/company ./cmd/company
go build -o bin/agentsmtp ./cmd/agentsmtp
go build -o bin/agenttv ./cmd/agenttv
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

Diagnose a missing SMB client before connecting a Windows server:

```bash
agentdoctor -profile smb-client -format text
```

Submit a prompt to a remote inference queue:

```bash
agentfal submit -model "$FAL_MODEL" -payload '{"prompt":"an orange robot"}'
```

Discover and repair a printer queue before printing:

```bash
agentprint discover -format json
agentprint ensure -printer office -match "Office Laser" -default
agentprint print -printer office -input ./document.pdf -copies 2 -duplex -fit-to-page
```

Convert mixed agent artifacts into Markdown for indexing:

```bash
agentmd -input ./incident.docx
agentmd -input ./metrics.xlsx
agentmd -input ./archive.zip -format json
```

Downstream PicoClaw migration:

```bash
# old helper surface
printer print --printer office ./document.pdf

# new shared CLI
agentprint print -printer office -input ./document.pdf
```

If an existing board image or shell wrapper still expects `printer`, keep a thin compatibility shim that shells out to `agentprint` during the transition.

Kick off a RunPod serverless job and poll it later:

```bash
agentrunpod submit -endpoint "$RUNPOD_ENDPOINT_ID" -payload '{"prompt":"hello"}'
agentrunpod result -endpoint "$RUNPOD_ENDPOINT_ID" -request "$JOB_ID"
```

## Design Goals

- low startup overhead
- predictable exit codes
- JSON/YAML outputs for agent parsing
- no Python or Node requirement
- suitable for cross-compiling to ARM and minimal Linux images
- exact-match file editing for agent-generated changes
- preflight dependency checks for workflows like SMB, SSH, and HTTP access
- bounded CUPS printer discovery, queue repair, and print submission
- native Markdown conversion for common agent-facing formats

## CI

GitHub Actions is configured for:

- tests on push and pull request
- formatting checks with `gofmt`
- cross-platform builds for Linux, macOS, and Windows
- build artifacts for all CLI tools

## Codex Skills

For easier agent integration, this repository also includes Codex skills under [`skills/`](./skills/).

Available skills:

- `$agentq-cli`
- `$agentenv-cli`
- `$agentfs-cli`
- `$agentrun-cli`
- `$agentedit-cli`
- `$agentdoctor-cli`
- `$agentrunpod-cli`
- `$agentfal-cli`
- `$agentprint-cli`

Each skill tells another Codex instance when to use the CLI, how to invoke it from this repo or from `PATH`, and which flags are the right default for agent workflows.

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
