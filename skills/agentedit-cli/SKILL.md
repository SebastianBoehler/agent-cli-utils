---
name: agentedit-cli
description: Apply deterministic targeted file edits with the local `agentedit` CLI. Use when an agent needs exact-match replacements, anchored inserts, line-range updates, or whole-file writes from a JSON or YAML edit spec.
---

# AgentEdit CLI

Use `agentedit` for safe, targeted file changes. Prefer it over broad patch parsing when the caller can describe edits as explicit replacements, inserts, line-range updates, appends, or writes.

## Invoke the Tool

Prefer the installed binary when it exists:

```bash
agentedit -spec edits.yaml
```

If working inside this repository or the binary is not installed, run:

```bash
go run ./cmd/agentedit -- -spec edits.yaml
```

## Choose Actions

- Use `replace` for one exact old-to-new substitution.
- Use `replace` with `replace_all: true` only when every match should change.
- Use `insert_before` or `insert_after` with one exact `anchor`.
- Use `replace_lines` for deterministic line-range rewrites.
- Use `append` to add content at the end of a file.
- Use `write` for full-file replacement or creation.

## Apply Good Defaults

- Keep `-fail-on-noop=true` so missed matches fail loudly.
- Use `-dry-run` first when the edit target may have drifted.
- Prefer narrow anchors and exact old strings to avoid ambiguity.
- Set `create: true` only when the task really intends to create a file.

## Example Spec

```yaml
edits:
  - path: README.md
    action: replace
    old: "Small Go CLIs"
    new: "Small, deterministic Go CLIs"
  - path: config.toml
    action: insert_after
    anchor: "[server]\n"
    new: "port = 8080\n"
```

Run it:

```bash
go run ./cmd/agentedit -- -spec edits.yaml -dry-run -format text
```
