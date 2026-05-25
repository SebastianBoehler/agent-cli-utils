---
name: company-cli
description: Use the local `company` CLI for German Handelsregister, OffeneRegister, and OpenCorporates company lookup with structured output. Prefer it when an agent needs registry-backed company research before lead scoring, outreach, or CRM/database updates.
---

# Company CLI

Use `company` to search company registries and return normalized JSON/YAML/text output.

## Invoke the Tool

Prefer the installed binary when it exists:

```bash
company search "Acme GmbH" --city Berlin --source handelsregister --format json
```

If working inside this repository or the binary is not installed, run:

```bash
go run ./cmd/company -- search "Acme GmbH" --city Berlin --source handelsregister --format json
```

## Choose Flags

- Use `search` for registry lookup.
- Use `--source handelsregister` for the official German commercial register.
- Use `--source all` when broader discovery is useful and partial source failures are acceptable.
- Use `--city` to disambiguate local companies.
- Use `--exact` only when the legal name is known.
- Use `--limit 5` for agent research unless a larger list is explicitly needed.
- Use `--format json` or `--format yaml` for downstream parsing.
- Set `OPENCORPORATES_API_TOKEN` when using `opencorporates` or `all`.

## Examples

Official German register lookup:

```bash
company search "Kiez Zahnzentrum Berlin" --source handelsregister --city Berlin --format json
```

Multi-source research:

```bash
company search "Acme GmbH" --source all --limit 5 --format yaml
```

OpenCorporates quota check:

```bash
company quota --format text
```
