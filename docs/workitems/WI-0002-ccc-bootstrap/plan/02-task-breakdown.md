# Task Breakdown: WI-0002-ccc-bootstrap

## Task 1: Project Scaffolding and Dependencies

**Summary:** Create directory structure, install Go dependencies, and build a minimal Cobra-based CLI entrypoint that compiles.

**Steps:**
1. Create directories: `cmd/ccc/`, `internal/config/`, `internal/copilot/`, `internal/sensitive/`, `internal/logging/`, `internal/tui/`
2. Run `go get` for all dependencies: bubbletea, lipgloss, huh, cobra, viper
3. Create `cmd/ccc/main.go` with Cobra root command (`ccc`), `--version` flag, `--log-level` flag
4. Verify `go build ./cmd/ccc/` compiles successfully

**Files Touched:**
- `cmd/ccc/main.go` (new)
- `go.mod` (updated)
- `go.sum` (new)

**Acceptance Criteria:**
- All directories exist
- `go build ./cmd/ccc/` produces a `ccc` binary without errors
- `./ccc --version` prints the version string
- `./ccc --help` prints usage with `--log-level` flag documented

**Test Requirements:**
- Build verification only (no unit tests yet)

**Risk Level:** Low

**Dependencies:** None

**References:** ADR-0002

---

## Task 2: Logging Package

**Summary:** Implement `internal/logging` with slog-based file logging per CC-0003.

**Steps:**
1. Create `internal/logging/logging.go` — `Init(level slog.Level, logPath string) error`
2. Opens/creates log file, sets `slog.SetDefault` with `TextHandler`
3. Create `internal/logging/logging_test.go` — tests init, write, level filtering
4. Wire `logging.Init()` call into `cmd/ccc/main.go` (PersistentPreRunE)

**Files Touched:**
- `internal/logging/logging.go` (new)
- `internal/logging/logging_test.go` (new)
- `cmd/ccc/main.go` (updated)

**Acceptance Criteria:**
- `logging.Init(slog.LevelWarn, "/tmp/test.log")` creates the log file
- `slog.Info("msg")` at warn level does NOT write to file
- `slog.Warn("msg")` at warn level DOES write to file
- Log entries are structured (key=value format)
- No output goes to stdout/stderr

**Test Requirements:**
- UT: Init creates file, writes at correct level, filters below-level
- UT: Init with invalid path returns wrapped error

**Risk Level:** Low

**Dependencies:** Task 1

**References:** CC-0003

---

## Task 3: Sensitive Data Package

**Summary:** Implement `internal/sensitive` with field classification, SHA-256 masking, and token pattern detection per CC-0005.

**Steps:**
1. Create `internal/sensitive/sensitive.go` — `IsSensitive(field string) bool`, `MaskValue(value any) string`, `LooksLikeToken(value string) bool`, `SensitiveFields` var
2. `IsSensitive` checks against known list: `copilot_tokens`, `logged_in_users`, `last_logged_in_user`, `staff`
3. `MaskValue` for strings: SHA-256 hash truncated to 12 hex chars + `...`; for objects/arrays: `[redacted — N items]`
4. `LooksLikeToken` checks prefixes: `gho_`, `ghp_`, `github_pat_`
5. Create `internal/sensitive/sensitive_test.go`

**Files Touched:**
- `internal/sensitive/sensitive.go` (new)
- `internal/sensitive/sensitive_test.go` (new)

**Acceptance Criteria:**
- `IsSensitive("copilot_tokens")` returns `true`
- `IsSensitive("model")` returns `false`
- `MaskValue("gho_abc123...")` returns a 12-char hex string + `...`
- `MaskValue(map[string]any{...})` returns `[redacted — N items]`
- `LooksLikeToken("gho_xyz")` returns `true`
- `LooksLikeToken("hello")` returns `false`
- Same input always produces same masked output (deterministic)

**Test Requirements:**
- UT: All four sensitive fields return true
- UT: Non-sensitive fields return false
- UT: MaskValue string produces correct hash prefix
- UT: MaskValue object/array produces summary
- UT: LooksLikeToken with each prefix pattern
- UT: LooksLikeToken negative cases

**Risk Level:** Low

**Dependencies:** Task 1

**References:** CC-0005

---

## Task 4: Copilot Detection Package

**Summary:** Implement `internal/copilot` with version detection and schema parsing from `copilot help config` per CC-0004.

**Steps:**
1. Create `internal/copilot/copilot.go` — `DetectVersion() (string, error)`, `DetectSchema() (*Schema, error)`
2. `DetectVersion()` runs `copilot version`, parses version string (e.g., "0.0.412")
3. `DetectSchema()` runs `copilot help config`, parses output into `[]SchemaField` structs
4. `SchemaField` struct: `Name`, `Type` (bool/string/enum/list), `Default`, `Options []string`, `Description`
5. Parser handles: boolean settings, string settings, enum settings (with option lists), list settings
6. Create `internal/copilot/copilot_test.go` with captured output samples
7. Create `internal/copilot/errors.go` — sentinel errors: `ErrCopilotNotInstalled`, `ErrVersionParseFailed`, `ErrSchemaParseFailed`

**Files Touched:**
- `internal/copilot/copilot.go` (new)
- `internal/copilot/errors.go` (new)
- `internal/copilot/copilot_test.go` (new)
- `internal/copilot/testdata/copilot-help-config.txt` (new — captured output sample)
- `internal/copilot/testdata/copilot-version.txt` (new — captured output sample)

**Acceptance Criteria:**
- `DetectVersion()` returns `"0.0.412"` (or current version) when copilot is installed
- `DetectVersion()` returns `ErrCopilotNotInstalled` when copilot is not found
- `DetectSchema()` returns 25 schema fields matching `copilot help config` output
- Boolean fields have `Type: "bool"` and correct defaults
- Enum fields (model, theme, banner) have `Type: "enum"` with correct option lists
- List fields (allowed_urls, denied_urls, etc.) have `Type: "list"`
- Schema parser handles the `model` field's 17 options correctly

**Test Requirements:**
- UT: Parse captured version output → correct version string
- UT: Parse captured schema output → correct field count and types
- UT: Each field type (bool, string, enum, list) parsed correctly
- UT: Model field has all 17 options
- UT: Error wrapping for not-installed case

**Risk Level:** Medium — parser depends on `copilot help config` output format

**Dependencies:** Task 1

**References:** CC-0004

---

## Task 5: Config Management Package

**Summary:** Implement `internal/config` with JSON config loading, saving, and round-trip preservation per CC-0004.

**Steps:**
1. Create `internal/config/config.go` — `Config` struct, `LoadConfig(path string) (*Config, error)`, `SaveConfig(path string, cfg *Config) error`, `DefaultPath() string`
2. `Config` struct has typed known fields (model, theme, banner, etc.) plus `Extra map[string]any` for unknown fields
3. `Config.Get(key string) any` and `Config.Set(key string, value any)` methods
4. `DefaultPath()` checks `XDG_CONFIG_HOME` then falls back to `~/.copilot/config.json`
5. `SaveConfig` writes JSON with 2-space indent, preserves unknown fields, preserves sensitive fields unchanged
6. Create `internal/config/errors.go` — sentinel errors: `ErrConfigNotFound`, `ErrConfigInvalid`
7. Create `internal/config/config_test.go`

**Files Touched:**
- `internal/config/config.go` (new)
- `internal/config/errors.go` (new)
- `internal/config/config_test.go` (new)
- `internal/config/testdata/valid-config.json` (new — test fixture)
- `internal/config/testdata/minimal-config.json` (new — test fixture)

**Acceptance Criteria:**
- `LoadConfig("path")` reads JSON into Config struct
- `LoadConfig("nonexistent")` returns `ErrConfigNotFound`
- `SaveConfig` writes back with 2-space indented JSON
- Round-trip: load → save → load produces identical config
- Unknown fields in JSON are preserved after save
- Sensitive field values are preserved exactly (not masked) on save
- `DefaultPath()` respects `XDG_CONFIG_HOME` environment variable
- `Config.Set("model", "gpt-5.2")` updates the model field
- `Config.Get("model")` returns the current model value

**Test Requirements:**
- UT: Load valid config → correct typed fields
- UT: Load missing file → ErrConfigNotFound
- UT: Load invalid JSON → ErrConfigInvalid
- UT: Round-trip preserves all fields including unknown ones
- UT: Set/Get work for known and unknown fields
- UT: DefaultPath with and without XDG_CONFIG_HOME
- UT: SaveConfig produces 2-space indented JSON

**Risk Level:** Low

**Dependencies:** Task 1, Task 3 (for sensitive field awareness)

**References:** CC-0004, CC-0005

---

## Task 6: TUI Application

**Summary:** Implement `internal/tui` with Bubbletea model, Huh form builder, and Lipgloss styling.

**Steps:**
1. Create `internal/tui/styles.go` — Lipgloss styles for header, field labels, sensitive highlights, status bar
2. Create `internal/tui/form.go` — `BuildForm(cfg *config.Config, schema []copilot.SchemaField, sensitiveChecker func(string) bool) *huh.Form` — maps schema fields to Huh form components:
   - `bool` → `huh.NewConfirm()`
   - `enum` → `huh.NewSelect()` with options
   - `string` → `huh.NewInput()`
   - `list` → `huh.NewText()` (one item per line)
   - Sensitive fields → read-only styled display (not editable)
3. Create `internal/tui/model.go` — Bubbletea model: `Init()`, `Update(msg)`, `View()` — wraps Huh form, adds header/footer, handles save confirmation
4. Create `internal/tui/keys.go` — keybindings: Ctrl+S save, Ctrl+C/q quit, Tab/Shift+Tab navigate
5. Group form fields: General settings, Model & AI, URLs & Permissions, Display, Sensitive (read-only)

**Files Touched:**
- `internal/tui/styles.go` (new)
- `internal/tui/form.go` (new)
- `internal/tui/model.go` (new)
- `internal/tui/keys.go` (new)

**Acceptance Criteria:**
- TUI launches in terminal with styled header showing copilot version
- Boolean settings render as toggles with current values
- Enum settings render as selects with correct option lists
- String settings render as text inputs with current values
- List settings render as multi-line text areas
- Sensitive fields display masked values and cannot be edited
- Settings are grouped into logical sections
- Save confirmation prompts before writing
- Quit works via Ctrl+C or q (when not in an input field)

**Test Requirements:**
- UT: BuildForm produces correct component count for a given schema
- UT: Sensitive fields are excluded from editable form fields
- UT: Form field grouping is correct

**Risk Level:** Medium — Huh API may have limitations for some field types

**Dependencies:** Task 4, Task 5, Task 3

**References:** ADR-0002, CC-0005

---

## Task 7: CLI Wiring

**Summary:** Wire all packages together in `cmd/ccc/main.go` — load config, detect schema, build form, run TUI, save on confirm.

**Steps:**
1. Update `cmd/ccc/main.go` root command RunE:
   - Initialize logging
   - Call `copilot.DetectVersion()` — show error screen if not installed
   - Call `copilot.DetectSchema()` — fallback gracefully if parsing fails
   - Call `config.LoadConfig(config.DefaultPath())` — create default if missing
   - Build TUI form from config + schema
   - Run Bubbletea program
   - On save: call `config.SaveConfig()`
2. Add `--log-level` flag (maps to slog level)
3. Add `--version` flag (prints ccc version)

**Files Touched:**
- `cmd/ccc/main.go` (updated)

**Acceptance Criteria:**
- `ccc` runs end-to-end: detects copilot, loads config, shows TUI, saves on edit
- `ccc --version` prints version
- `ccc --log-level debug` enables debug logging
- When copilot is not installed, shows user-friendly error message
- When config doesn't exist, creates default config on save

**Test Requirements:**
- Build verification: `go build ./cmd/ccc/` succeeds
- Manual E2E: run binary, verify full flow

**Risk Level:** Low

**Dependencies:** Task 2, Task 4, Task 5, Task 6

**References:** ADR-0002, CC-0002, CC-0003

---

## Task 8: Integration Tests and Polish

**Summary:** Add round-trip integration tests, verify `go vet`, and ensure all tests pass.

**Steps:**
1. Add integration test in `internal/config/`: load real config → modify → save to temp → reload → verify
2. Add integration test in `internal/copilot/`: parse live `copilot help config` output → verify field count
3. Run `go vet ./...` and fix any issues
4. Run `go test ./...` and ensure all pass
5. Verify `go build ./cmd/ccc/` still produces clean binary

**Files Touched:**
- `internal/config/config_integration_test.go` (new)
- `internal/copilot/copilot_integration_test.go` (new)
- Various files (fixes from vet/test issues)

**Acceptance Criteria:**
- `go test ./...` passes with no failures
- `go vet ./...` reports no issues
- Round-trip integration test confirms no data loss
- Schema integration test confirms correct field count against live copilot

**Test Requirements:**
- Integration: config round-trip with real file
- Integration: schema parse against live copilot output

**Risk Level:** Low

**Dependencies:** Task 7

**References:** CC-0004

---

## Dependency Graph

```
Task 1 (Scaffold)
├── Task 2 (Logging)
├── Task 3 (Sensitive)
├── Task 4 (Copilot Detection)
├── Task 5 (Config) ← depends on Task 3
│
Task 6 (TUI) ← depends on Tasks 3, 4, 5
Task 7 (CLI Wiring) ← depends on Tasks 2, 4, 5, 6
Task 8 (Integration) ← depends on Task 7
```
