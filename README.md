# ccc — Copilot Config CLI

[![CI](https://github.com/jsburckhardt/co-config/actions/workflows/ci.yml/badge.svg)](https://github.com/jsburckhardt/co-config/actions/workflows/ci.yml)
[![govulncheck](https://github.com/jsburckhardt/co-config/actions/workflows/govulncheck.yml/badge.svg)](https://github.com/jsburckhardt/co-config/actions/workflows/govulncheck.yml)

A TUI tool to interactively configure and view GitHub Copilot CLI settings.

`ccc` reads your `~/.copilot/config.json`, auto-detects the installed Copilot CLI version and available config keys, and presents them in a beautiful terminal UI for editing. Sensitive fields (tokens, credentials) are masked and read-only.

## Installation

### Quick install (curl)

```bash
curl -sSfL https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh | sh
```

> When installing without root privileges, the binary is placed in `~/.local/bin` and the installer automatically adds it to your `PATH` via your shell profile (`~/.bashrc`, `~/.zshrc`, etc.). To opt out of automatic PATH configuration:
>
> ```bash
> NO_PATH_UPDATE=1 curl -sSfL https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh | sh
> ```

### Version-pinned install

```bash
curl -sSfL https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh | sh -s -- --version v1.0.0
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.ps1 | iex
```

Version-pinned:

```powershell
$env:CCC_VERSION = "v1.0.0"; irm https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.ps1 | iex
```

### Go install

```bash
go install github.com/jsburckhardt/co-config/cmd/ccc@latest
```

### Build from source

```bash
git clone https://github.com/jsburckhardt/co-config.git
cd co-config
go build -o ccc ./cmd/ccc
```

### Run

```bash
ccc
```

## Verify Release Artifacts

### SHA256 checksum verification

Download the `checksums.txt` file from the GitHub Release, then verify:

```bash
sha256sum --check checksums.txt
```

### cosign signature verification

```bash
cosign verify-blob \
  --certificate-identity-regexp='github.com/jsburckhardt/co-config' \
  --certificate-oidc-issuer='https://token.actions.githubusercontent.com' \
  checksums.txt \
  --bundle checksums.txt.bundle
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
