# ccc — Copilot Config CLI

A TUI tool to interactively configure and view GitHub Copilot CLI settings.

`ccc` reads your `~/.copilot/config.json`, auto-detects the installed Copilot CLI version and available config keys, and presents them in a beautiful terminal UI for editing. Sensitive fields (tokens, credentials) are masked and read-only.

## Quick Start

```bash
# Install
go install github.com/jsburckhardt/co-config@latest

# Run
ccc
```

## Features

- 🎨 Beautiful TUI built with the Charm stack (Bubbletea + Lipgloss + Huh)
- 🔍 Auto-detects Copilot CLI version and available config schema
- 🔒 Masks sensitive fields (tokens, credentials) — read-only display
- 💾 Preserves unknown config fields on save — no data loss
- ⚡ Single static Go binary — no runtime dependencies

## Documentation

- [`CONTRIBUTING.md`](CONTRIBUTING.md) — pipeline workflow, how to start workitems, and where artifacts belong
- [`AGENTS.md`](AGENTS.md) — agent definitions, guardrails, and pipeline specification
- [`docs/`](docs/) — architecture decisions, core-components, and workitem artifacts
