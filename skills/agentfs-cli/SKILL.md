---
name: agentfs-cli
description: Inspect filesystem state with the local `agentfs` CLI. Use when an agent needs a bounded directory snapshot, file previews, or optional hashes without traversing an uncontrolled tree.
---

# AgentFS CLI

Use `agentfs` to inspect a file tree with explicit limits so the result stays cheap enough for low-compute systems and downstream agent parsing.

## Invoke the Tool

Prefer the installed binary when it exists:

```bash
agentfs -root . -max-depth 2 -format json
```

If working inside this repository or the binary is not installed, run:

```bash
go run ./cmd/agentfs -- -root . -max-depth 2 -format json
```

## Choose Flags

- Use `-root` to set the inspection root.
- Use `-max-depth` to keep traversal bounded.
- Use `-max-entries` to avoid oversized snapshots.
- Use `-preview-lines` for cheap text previews.
- Use `-hash` only when content integrity matters.
- Use `-hidden` only when hidden files are relevant.
- Use `-format json` or `-format yaml` for automation, `text` for review.

## Apply Good Defaults

- Keep depth and entry limits conservative unless the task explicitly needs more.
- Prefer previews over full file reads when building context.
- Hash only a small set of files, since hashing scales with file size.

## Examples

Take a bounded repo snapshot:

```bash
go run ./cmd/agentfs -- -root . -max-depth 2 -preview-lines 2
```

Inspect logs in text form:

```bash
go run ./cmd/agentfs -- -root /var/log -max-depth 1 -preview-lines 3 -format text
```
