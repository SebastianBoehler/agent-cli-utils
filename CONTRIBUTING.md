# Contributing

Thanks for contributing to Agent CLI Utils.

## What Fits This Repo

This repository is for small Go CLI tools that help autonomous or semi-autonomous agents operate reliably on constrained systems, including boards and edge devices.

Good fits:

- tools that are easy to script from shell-based agents
- bounded-output helpers
- machine-readable diagnostics
- low-memory and low-dependency utilities
- commands that cross-compile cleanly

Bad fits:

- large daemon-style services
- tools that require Python or Node to function
- features that only make sense in heavyweight cloud environments
- commands with unstable or human-only output formats

## External Tool Curation

This repo also maintains [AWESOME.md](./AWESOME.md), a curated list of external Go CLIs that are useful for agents.

For curated external entries:

- link to the upstream repository instead of adding a submodule
- keep the description to one or two lines
- explain the agent or low-compute relevance
- prefer projects with clear documentation and active maintenance

## Development

Prerequisites:

- Go 1.21 or newer recommended

Run locally:

```bash
make check
make build
```

Before opening a pull request:

```bash
gofmt -w .
make check
```

## Design Expectations

- keep binaries small and startup fast
- prefer the standard library
- keep output parseable and stable
- use explicit exit codes
- document behavior changes in the README

## Pull Requests

Please keep pull requests focused. A good PR includes:

- the user or agent workflow it improves
- examples of input and output
- tests for non-trivial behavior
- notes about memory, timeout, or portability tradeoffs when relevant

## Becoming a Maintainer

Maintainers are welcome, especially from the agent tooling and edge-compute communities.

The usual path is:

1. Make several solid contributions.
2. Help review issues and pull requests.
3. Show good judgment around scope, stability, and backward compatibility.
4. Start a conversation about maintainer responsibilities.
