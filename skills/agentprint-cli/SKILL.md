---
name: agentprint-cli
description: Manage local CUPS printers with the `agentprint` CLI. Use when an agent needs to discover printers, create or repair a queue, or submit a bounded print job from a file or HTTP(S) URL with machine-readable output.
---

# AgentPrint CLI

Use `agentprint` when the host already has CUPS tooling and the workflow needs structured printer discovery, queue setup, or non-interactive printing.

## Invoke the Tool

Prefer the installed binary when it exists:

```bash
agentprint list
```

If working inside this repository or the binary is not installed, run:

```bash
go run ./cmd/agentprint -- list
```

## Choose Commands

- Use `list` to inspect configured local queues and identify the default printer.
- Use `discover` to enumerate reachable IPP, IPPS, and DNS-SD printers through local CUPS tooling.
- Use `ensure` to create or repair a queue from `-uri` or a `-match` against discovered printers.
- Use `print` to submit a local file or HTTP(S) URL with CUPS options such as copies, duplex, media, scaling, color mode, and raw `-option name=value` flags.

## Apply Good Defaults

- Keep the default `-format json` when another step will parse the result.
- Use `-format tsv` for quick terminal inspection of `list` and `discover`.
- Prefer `ensure -match ... -default` on boards where queues need to self-heal before a print job.
- Reject or fix invalid combinations before running, especially `-scale-percent` with `-fit-to-page` or `-fill-page`.

## Examples

Discover reachable printers:

```bash
go run ./cmd/agentprint -- discover -format tsv
```

Repair a queue and set it as default:

```bash
go run ./cmd/agentprint -- ensure -printer office -match "Office Laser" -default
```

Print a remote PDF in monochrome:

```bash
go run ./cmd/agentprint -- print -printer office -input https://example.com/report.pdf -fit-to-page -color-mode monochrome
```
