# CORE-COMPONENT-0004: Configuration Management

## Status

Adopted

## Purpose

Define how `ccc` reads, validates, edits, and writes the Copilot CLI configuration file. This component ensures config integrity, schema awareness, and safe round-tripping of unknown fields.

## Scope

The `internal/config` package. Affects how the TUI presents config fields and how changes are persisted. Covers all three configuration scopes (user, project, project-local) and their path resolution.

## Definition

### Rules
- The canonical user config path is `~/.copilot/config.json`, or `$XDG_CONFIG_HOME/copilot/config.json` if `XDG_CONFIG_HOME` is set
- Three config scopes are supported, cascading from general to specific:
  1. **User** (global): `~/.copilot/config.json` — personal defaults
  2. **Project**: `<project-root>/.copilot/settings.json` — team-shared, committed to VCS
  3. **Project-local**: `<project-root>/.copilot/settings.local.json` — personal per-project overrides, gitignored
- Each scope is loaded and edited independently; no merged or effective view is presented
- When writing config, the active scope is the write target; values from other scopes are never modified
- Config is read as raw JSON and decoded into a typed struct for known fields; unknown fields are preserved via a `map[string]any` catch-all
- Config schema (available keys, types, defaults, descriptions) is auto-detected at startup by running `copilot help config` and parsing the output
- The installed copilot version is detected by running `copilot version` and parsing the output
- When writing config, only known editable fields are updated; sensitive, token-like, and unknown fields are preserved unchanged and displayed as read-only in the TUI
- Config validation occurs before writing — invalid values are rejected with user-friendly errors
- The TUI must track per-field dirty state (`Modified` flag) for in-memory changes that have not yet been persisted to disk, and surface this to the user via a "(not-saved)" indicator
- After a successful `SaveConfig`, the TUI must re-read the config file from disk via `LoadConfig` to verify round-trip integrity and reflect the actual persisted state
- The "✓ Saved" UI indicator must be cleared immediately when any new in-memory change is committed, so the banner never shows stale information

### Interfaces
- `Config` struct holds typed known fields plus a raw `map[string]any` for round-tripping
- `LoadConfig(path string) (*Config, error)` — reads and parses the config file
- `SaveConfig(path string, cfg *Config) error` — writes config back preserving unknown fields
- `DefaultPath() string` — returns the user-level config path (`~/.copilot/config.json` or XDG equivalent)
- `ProjectSettingsPath(projectDir string) string` — returns `<projectDir>/.copilot/settings.json`
- `ProjectLocalSettingsPath(projectDir string) string` — returns `<projectDir>/.copilot/settings.local.json`
- `Scope` type with values `ScopeUser`, `ScopeProject`, `ScopeProjectLocal`
- `DetectSchema() (*Schema, error)` — runs `copilot help config` and parses available settings
- `DetectVersion() (string, error)` — runs `copilot version` and extracts the version string

### Expectations
- If the config file doesn't exist, the tool shows an empty/default config and creates it on save
- If a project-scope or project-local config file does not exist, the tool shows an empty config and creates the file and `.copilot/` directory on first save
- The TUI header indicates which scope is currently active, alongside the file path
- If `copilot` is not installed, the tool shows an error screen with installation instructions
- JSON formatting is preserved (indented with 2 spaces) to match copilot CLI's own output
- No data loss — fields the tool doesn't understand are never dropped

## Rationale

Auto-detecting the schema from the copilot CLI itself ensures `ccc` stays compatible with future copilot versions without code changes. Round-tripping unknown fields prevents data loss when the config contains keys `ccc` doesn't know about yet.

## Usage Examples

```go
// Load user-level config (existing behavior)
cfg, err := config.LoadConfig(config.DefaultPath())
if err != nil {
    // handle error
}

// Load project-level config
projectDir, _ := os.Getwd()
projectCfg, err := config.LoadConfig(config.ProjectSettingsPath(projectDir))
if err != nil {
    if errors.Is(err, config.ErrConfigNotFound) {
        projectCfg = config.NewConfig() // empty config — file will be created on save
    }
}

// Detect schema
schema, err := config.DetectSchema()
if err != nil {
    // copilot not installed or other error
}

// Modify a field
cfg.Set("model", "claude-sonnet-4.5")

// Save back to the active scope's path
err = config.SaveConfig(config.DefaultPath(), cfg)
```

## Integration Guidelines

- The TUI layer reads from `Config` and `Schema` to build form fields
- The TUI layer calls `Config.Set()` to update values
- On form submission, call `SaveConfig()` to persist
- Always load schema before presenting the form so field metadata is available
- After `SaveConfig()` succeeds, call `LoadConfig()` and rebuild list entries from the reloaded config to verify the round-trip
- Mark fields as modified when `Config.Set()` is called from the TUI; clear all modified markers after a successful save-and-reload cycle
- If post-save `LoadConfig()` fails, show a non-fatal error in the UI header and keep the in-memory state

## Exceptions

- If `copilot help config` output format changes dramatically, a fallback hardcoded schema for a known version may be used temporarily

## Enforcement

- [x] Code review checklist
- [x] Test coverage requirements — round-trip tests (load → modify → save → load) must verify no data loss
- [ ] Automated checks (future: integration test against copilot binary)

## Related ADRs

- [ADR-0002-go-charm-tui-stack](../ADR/ADR-0002-go-charm-tui-stack.md)
- [ADR-0008-multi-scope-configuration](../ADR/ADR-0008-multi-scope-configuration.md)
