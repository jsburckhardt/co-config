# Research Brief: ccc Bootstrap — Copilot Config CLI

## Title

ccc Bootstrap — Interactive TUI for GitHub Copilot CLI Configuration

## Idea Summary

Build `ccc`, a Go CLI/TUI application that reads `~/.copilot/config.json`, auto-detects the installed Copilot CLI version and available config keys via `copilot help config`, and presents them in a Bubbletea-based terminal UI for viewing and editing. Sensitive fields (tokens, credentials) are masked and read-only. Unknown config fields are preserved on save (no data loss).

## Scope Type

```
scope_type: workitem
```

## Related Workitem

WI-0002-ccc-bootstrap

## Existing Repo Context

- **Go module** already initialized: `github.com/jsburckhardt/co-config` (Go 1.25.0)
- **No Go source code exists yet** — the project is greenfield
- **No dependencies installed** — no Charm, Cobra, or Viper packages in go.mod
- **Architecture decisions already made:**
  - ADR-0002: Go + Charm TUI stack (Bubbletea, Lipgloss, Huh, Cobra, Viper)
  - CC-0002: Error handling (fmt.Errorf + %w, sentinel errors)
  - CC-0003: Logging (log/slog to file, never stdout)
  - CC-0004: Configuration management (read/write/schema detection)
  - CC-0005: Sensitive data handling (SHA-256 masking, read-only display)
- **Application docs** define the package layout:
  - `cmd/ccc/` — CLI entrypoint
  - `internal/config/` — Config file reading, writing, schema detection
  - `internal/copilot/` — Copilot CLI version and schema detection
  - `internal/sensitive/` — Sensitive data masking
  - `internal/logging/` — Structured file-based logging
  - `internal/tui/` — Bubbletea TUI application

## External References

- [Bubbletea](https://github.com/charmbracelet/bubbletea) — Elm-architecture TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [Huh](https://github.com/charmbracelet/huh) — Form components
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Viper](https://github.com/spf13/viper) — Configuration binding
- Copilot CLI v0.0.412 — target CLI version for schema detection

## Copilot CLI Config Schema (from `copilot help config`)

The Copilot CLI exposes 25 config settings. The config file is `~/.copilot/config.json`.

### Boolean Settings (default)

| Setting | Default | Description |
|---------|---------|-------------|
| `alt_screen` | false | Alternate screen buffer |
| `auto_update` | true | Auto-download CLI updates |
| `bash_env` | false | BASH_ENV support for bash |
| `beep` | true | Beep on attention required |
| `compact_paste` | true | Collapse large pastes |
| `custom_agents.default_local_only` | false | Local agents only |
| `experimental` | false | Experimental features |
| `include_coauthor` | true | Co-authored-by trailer |
| `parallel_tool_execution` | true | Parallel tool execution |
| `render_markdown` | true | Terminal markdown rendering |
| `screen_reader` | false | Screen reader optimizations |
| `stream` | true | Streaming mode |
| `streamer_mode` | false | Hide model/quota info |
| `undo_without_confirmation` | false | Skip undo confirm (experimental) |
| `update_terminal_title` | true | Update terminal title |

### String/Enum Settings

| Setting | Default | Options |
|---------|---------|---------|
| `banner` | "once" | always, never, once |
| `log_level` | "default" | default, all |
| `model` | — | 17 models (claude-sonnet-4.6, gpt-5.3-codex, etc.) |
| `theme` | "auto" | auto, dark, light |

### List Settings

| Setting | Description |
|---------|-------------|
| `allowed_urls` | URL/domain allowlist (supports wildcards) |
| `denied_urls` | URL/domain denylist (takes precedence) |
| `launch_messages` | Startup banner messages |
| `trusted_folders` | Granted folder paths |

### Undocumented Settings (found in config.json)

| Setting | Type | Notes |
|---------|------|-------|
| `reasoning_effort` | string | Found in live config, not in `copilot help config` |

### Sensitive Fields (read-only, masked)

| Setting | Type | Notes |
|---------|------|-------|
| `copilot_tokens` | object | OAuth tokens keyed by host:user |
| `logged_in_users` | array | Array of {host, login} objects |
| `last_logged_in_user` | object | {host, login} object |
| `staff` | bool | Internal staff flag |

### Config File Structure

- Format: JSON with 2-space indentation
- Path: `~/.copilot/config.json` (respects `XDG_CONFIG_HOME`)
- Round-trip safe: unknown fields must be preserved on save

## Options Considered

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| A. Full scaffold | Implement all packages (cmd, config, copilot, sensitive, logging, tui) in one workitem | Complete working application | Large scope, harder to review |
| B. Layered build | Core packages first (config, copilot, sensitive, logging), then TUI, then CLI wiring | Testable increments, easier review | More workitems needed |
| C. TUI-first | Build TUI with hardcoded data, then wire real config/copilot detection | Fast visual feedback | Throwaway code, testing harder |

## Recommendation

**Option A — Full scaffold.** The architecture is already well-defined by the existing ADRs and core-components. The package layout is documented. All decisions are made. This is an implementation workitem, not a design exercise. Building it all at once avoids partial states where the binary doesn't do anything useful.

## Risks & Unknowns

1. **Schema detection fragility** — `copilot help config` output format may change across versions. Mitigation: parser tests with captured output samples.
2. **Undocumented settings** — Some settings (e.g., `reasoning_effort`) appear in config but not in `copilot help config`. Mitigation: display them as editable text fields with no type constraints.
3. **Huh form limitations** — Huh may not support all desired form patterns (e.g., list editors for `allowed_urls`). Mitigation: start with simple text input for lists (comma-separated), iterate later.
4. **Config file locking** — No file locking on config.json. Risk of race condition if copilot CLI writes simultaneously. Mitigation: accepted risk for v1; warn user in docs.

## Required ADRs

No — ADR-0002 already covers the tech stack decision.

## Required Core-Components

No — CC-0002 through CC-0005 already cover all cross-cutting concerns.

## Verification Strategy

- **Unit tests:** Config load/save round-trip, schema parsing, sensitive field detection, masking
- **Integration test:** Run `copilot help config` parsing against live output
- **Manual verification:** Run `ccc` binary, verify TUI renders, edit a field, confirm save
- **Build verification:** `go build ./cmd/ccc/` produces a working binary

## Architect Handoff Notes

- All architectural decisions are already in place (ADR-0002, CC-0002–0005)
- No new ADRs or core-components are needed
- The Architect stage should produce an action plan covering:
  1. Project scaffolding (go mod, dependencies, directory structure)
  2. Internal packages (config, copilot, sensitive, logging)
  3. TUI implementation (bubbletea model, huh forms)
  4. CLI wiring (cobra root command)
  5. Tests for each package
- The package layout from `docs/application/README.md` should be followed exactly
