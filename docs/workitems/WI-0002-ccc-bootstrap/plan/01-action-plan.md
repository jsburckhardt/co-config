# Action Plan: WI-0002-ccc-bootstrap

## Objective

Scaffold and implement the `ccc` (Copilot Config CLI) application — a Go TUI tool that reads `~/.copilot/config.json`, auto-detects config schema from the Copilot CLI, and presents an interactive terminal UI for viewing and editing settings. Produces a single `ccc` binary.

## Non-Goals

- MCP config editing (`mcp-config.json`) — out of scope for v1
- Config file locking or concurrent access handling
- Log rotation
- CI/CD pipeline setup
- Remote config sync or multi-machine support

## Chosen Approach

Full scaffold (Option A from research). All packages built in one workitem since architecture is already fully defined by ADR-0002 and CC-0002–0005.

## Impacted Areas

All new files — greenfield implementation:

| Area | Packages | Purpose |
|------|----------|---------|
| CLI entrypoint | `cmd/ccc/` | Cobra root command, version flag, log init |
| Config management | `internal/config/` | Load, save, round-trip JSON config |
| Copilot detection | `internal/copilot/` | Version detection, schema parsing from `copilot help config` |
| Sensitive data | `internal/sensitive/` | Field classification, SHA-256 masking, token detection |
| Logging | `internal/logging/` | slog file-based logger init |
| TUI | `internal/tui/` | Bubbletea model, Huh form builder, Lipgloss styling |

## Dependencies

External Go modules to install:

| Module | Purpose | Reference |
|--------|---------|-----------|
| `github.com/charmbracelet/bubbletea` | TUI framework | ADR-0002 |
| `github.com/charmbracelet/lipgloss` | Terminal styling | ADR-0002 |
| `github.com/charmbracelet/huh` | Form components | ADR-0002 |
| `github.com/spf13/cobra` | CLI framework | ADR-0002 |
| `github.com/spf13/viper` | Config binding | ADR-0002 |

## Implementation Phases

### Phase 1: Project Scaffolding

- Create directory structure (`cmd/ccc/`, `internal/*/`)
- Install Go dependencies (`go get`)
- Create `cmd/ccc/main.go` with Cobra root command (version, log-level flag)
- Create `internal/logging/logging.go` with slog file init (CC-0003)
- Verify: `go build ./cmd/ccc/` compiles

### Phase 2: Core Packages

- **`internal/copilot/`** — `DetectVersion()` runs `copilot version`, parses output; `DetectSchema()` runs `copilot help config`, parses all 25 settings into typed `SchemaField` structs (name, type, default, options, description)
- **`internal/config/`** — `LoadConfig(path)` reads JSON into typed struct + `map[string]any` catch-all; `SaveConfig(path, cfg)` writes back preserving unknown fields; `DefaultPath()` resolves `~/.copilot/config.json` with `XDG_CONFIG_HOME` override (CC-0004)
- **`internal/sensitive/`** — `IsSensitive(field)` checks known list + token pattern detection (`gho_`, `ghp_`, `github_pat_`); `MaskValue(value)` returns truncated SHA-256; `SensitiveFields` list (CC-0005)
- Verify: unit tests pass for all three packages

### Phase 3: TUI Application

- **`internal/tui/model.go`** — Bubbletea model holding config, schema, and form state
- **`internal/tui/form.go`** — Huh form builder: booleans → toggles, enums → selects, strings → text inputs, lists → text areas, sensitive → read-only styled display
- **`internal/tui/styles.go`** — Lipgloss styles (header, field labels, sensitive field highlight, save confirmation)
- **`internal/tui/keys.go`** — Key bindings (save, quit, navigate)
- Wire TUI into Cobra command: load config + schema → build form → run Bubbletea → save on confirm
- Verify: `ccc` launches, displays config fields, allows editing, saves changes

### Phase 4: Testing & Polish

- Round-trip test: load → modify → save → load → verify no data loss
- Schema parser test with captured `copilot help config` output sample
- Sensitive field detection test (known fields + token patterns)
- Masking test (SHA-256 truncation)
- Error path tests (missing config, copilot not installed)
- `go vet ./...` clean
- Verify: `go test ./...` all pass

## Architectural References

| Decision | Source | Key Rule |
|----------|--------|----------|
| Go + Charm TUI + Cobra + Viper | ADR-0002 | Use Bubbletea/Lipgloss/Huh for TUI, Cobra for CLI, Viper for config binding |
| Error handling | CC-0002 | `fmt.Errorf` + `%w`, sentinel errors per package, user-friendly TUI messages |
| Logging | CC-0003 | `log/slog` to `~/.copilot/ccc.log`, never stdout, default level warn |
| Config management | CC-0004 | Round-trip JSON, auto-detect schema, `XDG_CONFIG_HOME` support |
| Sensitive data | CC-0005 | Mask with truncated SHA-256, read-only in TUI, never log tokens |

## Rollout Strategy

Single deployment — build `ccc` binary and verify it works end-to-end.

## Observability

- File-based logging at `~/.copilot/ccc.log` (CC-0003)
- `--log-level` flag and `CCC_LOG_LEVEL` env var for debug output

## Security Considerations

- Sensitive fields (tokens, credentials) are never displayed in plain text (CC-0005)
- Sensitive fields are never logged at any level (CC-0003, CC-0005)
- Config file permissions are preserved on save
- No network access — all operations are local

## Testing Strategy

- **Unit tests** per package: config, copilot, sensitive, logging
- **Integration test**: schema parser against live `copilot help config` output
- **Round-trip test**: load → modify → save → reload → verify
- **Build test**: `go build ./cmd/ccc/` produces binary
- **Coverage target**: all public functions tested

## Acceptance Criteria

- [ ] `go build ./cmd/ccc/` produces a working `ccc` binary
- [ ] `ccc` launches and displays a TUI with all detected config fields
- [ ] Boolean settings render as toggles
- [ ] Enum settings (model, theme, banner) render as selects with correct options
- [ ] List settings (allowed_urls, trusted_folders) render as editable text areas
- [ ] Sensitive fields display masked values and are read-only
- [ ] Editing a field and saving persists the change to `config.json`
- [ ] Unknown config fields are preserved on save (no data loss)
- [ ] `copilot version` and `copilot help config` output is parsed correctly
- [ ] `go test ./...` passes with no failures
- [ ] `go vet ./...` reports no issues
- [ ] Logging goes to file, not stdout/stderr

## Open Questions

None — all architectural decisions are already made.
