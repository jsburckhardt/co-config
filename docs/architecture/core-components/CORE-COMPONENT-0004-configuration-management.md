# CORE-COMPONENT-0004: Configuration Management

## Status

Adopted

## Purpose

Define how `ccc` reads, validates, edits, and writes the Copilot CLI configuration file. This component ensures config integrity, schema awareness, and safe round-tripping of unknown fields.

## Scope

The `internal/config` package. Affects how the TUI presents config fields and how changes are persisted.

## Definition

### Rules
- The canonical config path is `~/.copilot/config.json`, overridden by `XDG_CONFIG_HOME` environment variable
- Config is read as raw JSON and decoded into a typed struct for known fields; unknown fields are preserved via a `map[string]any` catch-all
- Config schema (available keys, types, defaults, descriptions) is auto-detected at startup by running `copilot help config` and parsing the output
- The installed copilot version is detected by running `copilot version` and parsing the output
- When writing config, only known editable fields are updated; sensitive and unknown fields are preserved unchanged
- Config validation occurs before writing — invalid values are rejected with user-friendly errors

### Interfaces
- `Config` struct holds typed known fields plus a raw `map[string]any` for round-tripping
- `LoadConfig(path string) (*Config, error)` — reads and parses the config file
- `SaveConfig(path string, cfg *Config) error` — writes config back preserving unknown fields
- `DetectSchema() (*Schema, error)` — runs `copilot help config` and parses available settings
- `DetectVersion() (string, error)` — runs `copilot version` and extracts the version string

### Expectations
- If the config file doesn't exist, the tool shows an empty/default config and creates it on save
- If `copilot` is not installed, the tool shows an error screen with installation instructions
- JSON formatting is preserved (indented with 2 spaces) to match copilot CLI's own output
- No data loss — fields the tool doesn't understand are never dropped

## Rationale

Auto-detecting the schema from the copilot CLI itself ensures `ccc` stays compatible with future copilot versions without code changes. Round-tripping unknown fields prevents data loss when the config contains keys `ccc` doesn't know about yet.

## Usage Examples

```go
// Load existing config
cfg, err := config.LoadConfig(config.DefaultPath())
if err != nil {
    // handle error
}

// Detect schema
schema, err := config.DetectSchema()
if err != nil {
    // copilot not installed or other error
}

// Modify a field
cfg.Set("model", "claude-sonnet-4.5")

// Save back
err = config.SaveConfig(config.DefaultPath(), cfg)
```

## Integration Guidelines

- The TUI layer reads from `Config` and `Schema` to build form fields
- The TUI layer calls `Config.Set()` to update values
- On form submission, call `SaveConfig()` to persist
- Always load schema before presenting the form so field metadata is available

## Exceptions

- If `copilot help config` output format changes dramatically, a fallback hardcoded schema for a known version may be used temporarily

## Enforcement

- [x] Code review checklist
- [x] Test coverage requirements — round-trip tests (load → modify → save → load) must verify no data loss
- [ ] Automated checks (future: integration test against copilot binary)

## Related ADRs

- [ADR-0002-go-charm-tui-stack](../ADR/ADR-0002-go-charm-tui-stack.md)
