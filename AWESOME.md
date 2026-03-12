# Awesome Go CLIs for Agents

Curated Go command-line tools that pair well with low-compute agent runtimes, including PicoClaw-style workers, edge devices, and shell-first automation setups.

This list links to upstream projects instead of vendoring them as git submodules. That is the more common pattern for awesome-style repositories:

- lower maintenance burden
- users get the canonical upstream docs and releases
- no unnecessary repository weight
- no confusion about ownership or version drift

## In This Repo

- [`agentq`](./cmd/agentq) - Lightweight structured data query and conversion for JSON and YAML.
- [`agentenv`](./cmd/agentenv) - Environment validation for agent startup checks.
- [`agentfs`](./cmd/agentfs) - Bounded filesystem inspection for cheap machine-readable snapshots.
- [`agentrun`](./cmd/agentrun) - Timeout- and output-bounded command execution.

## Google Workspace and Google APIs

- [`google/oauth2l`](https://github.com/google/oauth2l) - Go CLI for Google OAuth 2.0 flows and token retrieval. Useful when an agent needs Google API access without embedding auth logic.
- [`mbrt/gmailctl`](https://github.com/mbrt/gmailctl) - Declarative Gmail filter management in Go. Good for automating inbox routing and deterministic email handling.
- [`tanaikech/goodls`](https://github.com/tanaikech/goodls) - Go CLI for downloading shared Google Drive files and folders, including Google Docs export flows.
- [`rclone/rclone`](https://github.com/rclone/rclone) - Broad cloud storage CLI written in Go with Google Drive support. Often the most practical answer for agent-side file sync.

## Structured Data and Shell Glue

- [`mikefarah/yq`](https://github.com/mikefarah/yq) - Portable structured data processor for YAML, JSON, XML, CSV, TOML, HCL, and more.
- [`charmbracelet/gum`](https://github.com/charmbracelet/gum) - Small terminal UX helpers for shell scripts. Useful when human-in-the-loop agent workflows need prompts or selections.
- [`go-task/task`](https://github.com/go-task/task) - Go-based task runner and Make alternative. Good for stable local orchestration on systems with minimal tooling.

## Sync, Backups, and State Management

- [`restic/restic`](https://github.com/restic/restic) - Fast encrypted backups in Go. Useful for protecting agent state, prompts, or device snapshots.
- [`rclone/rclone`](https://github.com/rclone/rclone) - Also belongs here for moving agent outputs, logs, and model artifacts across constrained systems.

## Meta Resources

- [`mantcz/awesome-go-cli`](https://github.com/mantcz/awesome-go-cli) - Broader curated list of Go CLI tools and resources. Good place to discover adjacent tooling.

## Selection Rules

Tools added here should generally meet most of these:

- written primarily in Go
- useful from shell scripts or agent runtimes
- relevant to low-compute, edge, or automation use cases
- actively maintained or clearly stable
- documented enough that users can adopt them without guesswork

## Contributing

If you want to add a tool:

1. Prefer upstream links over submodules.
2. Add one short sentence explaining why the tool matters for agents.
3. Avoid dumping generic developer tools that are not meaningfully relevant to automation or constrained environments.
