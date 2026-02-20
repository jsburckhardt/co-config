# Application Documentation — ccc (Copilot Config CLI)

`ccc` is a Go CLI/TUI application that provides an interactive terminal interface for configuring the GitHub Copilot CLI.

## Tech Stack

- **Language:** Go
- **TUI Framework:** Bubbletea + Lipgloss + Huh (Charm stack)
- **Test Runner:** go test
- **Binary:** `ccc`

## Architecture

- `cmd/ccc/` — CLI entrypoint
- `internal/config/` — Config file reading, writing, schema detection
- `internal/copilot/` — Copilot CLI version and schema detection
- `internal/sensitive/` — Sensitive data masking
- `internal/logging/` — Structured file-based logging
- `internal/tui/` — Bubbletea TUI application

## Config Path

The tool reads/writes `~/.copilot/config.json` (respects `XDG_CONFIG_HOME` override).

## References

- [ADR-0002: Go with Charm TUI Stack](../architecture/ADR/ADR-0002-go-charm-tui-stack.md)
- [CC-0004: Configuration Management](../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md)
