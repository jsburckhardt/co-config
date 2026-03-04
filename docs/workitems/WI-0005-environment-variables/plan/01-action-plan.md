# Action Plan: Environment Variables Panel with Tab Navigation

## Feature
- **ID:** WI-0005-environment-variables
- **Research Brief:** docs/workitems/WI-0005-environment-variables/research/00-research.md

## ADRs Created
- **ADR-0004** — [TUI Multi-View Tab Navigation](../../architecture/ADR/ADR-0004-tui-multi-view-navigation.md): Defines the multi-view navigation pattern using left/right/tab keys to switch between the Config View and the new Env Vars View. Adds `StateEnvVars` to the state machine. Establishes view vs. panel terminology. Declares the env vars view as read-only.

## Core-Components Created
- None. No new core-components are required.

### Core-Component Scope Decisions
- **CC-0004 (Configuration Management):** `DetectEnvVars` follows the same "run CLI command, parse output" pattern established by `DetectSchema` and `DetectVersion`. It lives in the same `internal/copilot` package and follows the same conventions. No CC-0004 update is needed — the pattern is already established; `DetectEnvVars` is an implementation detail.
- **CC-0005 (Sensitive Data Handling):** The existing `sensitive.IsSensitive` function checks config field names (`copilot_tokens`, `logged_in_users`, etc.) — it does NOT cover env var names like `COPILOT_GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_TOKEN`. The `sensitive.LooksLikeToken` function catches values with `gho_`/`ghp_`/`github_pat_` prefixes but tokens may not always have those prefixes. **Action:** Add a `SensitiveEnvVars` list and `IsEnvVarSensitive(name string) bool` function to the `sensitive` package. The env vars panel will check both name-based sensitivity (`IsEnvVarSensitive`) and value-based sensitivity (`LooksLikeToken`) — belt-and-suspenders approach consistent with CC-0005 principles. This is a minor extension to the existing `sensitive` package, not a CC-level change.

## Implementation Tasks

### Task 1: Add `EnvVarInfo` struct and `DetectEnvVars`/`ParseEnvVars` to `internal/copilot`
**Files:** `internal/copilot/copilot.go`, `internal/copilot/errors.go`
- Add `EnvVarInfo` struct with `Names []string`, `Description string`, `Qualifier string`
- Add `ErrEnvVarsParseFailed` sentinel error
- Add `DetectEnvVars() ([]EnvVarInfo, error)` — runs `copilot help environment`, calls `ParseEnvVars`
- Add `ParseEnvVars(output string) ([]EnvVarInfo, error)` — regex-based parser for backtick-quoted names, optional parenthesised qualifier, multi-line descriptions
- Return `nil, nil` (not error) if command succeeds but produces no output (graceful degradation)
- **Tests:** `internal/copilot/copilot_test.go` — fixture-driven tests for `ParseEnvVars` with captured `copilot help environment` output; test multi-name entries, single-name entries, multi-line descriptions

### Task 2: Add `SensitiveEnvVars` and `IsEnvVarSensitive` to `internal/sensitive`
**Files:** `internal/sensitive/sensitive.go`, `internal/sensitive/sensitive_test.go`
- Add `SensitiveEnvVars` list: `COPILOT_GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_TOKEN`
- Add `IsEnvVarSensitive(name string) bool` — case-insensitive match against env var name
- **Tests:** Test that the three token env var names return `true`; test that other env var names return `false`

### Task 3: Add `StateEnvVars` to TUI state machine
**Files:** `internal/tui/state.go`
- Add `StateEnvVars State = iota` constant after `StateExiting`
- Update `String()` method to return `"EnvVars"` for the new state

### Task 4: Add `Left`/`Right` key bindings to `KeyMap`
**Files:** `internal/tui/keys.go`
- Add `Left key.Binding` and `Right key.Binding` fields to `KeyMap`
- Wire in `DefaultKeyMap()`: `Left` → `key.WithKeys("left", "h")`, `Right` → `key.WithKeys("right", "l")`
- Update `Tab` help text from `"switch"` to `"switch view"` for clarity

### Task 5: Create `EnvVarsPanel` component
**Files:** `internal/tui/env_panel.go` (NEW)
- `EnvVarsPanel` struct: holds `[]copilot.EnvVarInfo`, scroll offset, dimensions
- `NewEnvVarsPanel(envVars []copilot.EnvVarInfo) *EnvVarsPanel`
- `View() string` — renders scrollable list of env var entries:
  - Primary name bold/highlighted; aliases in muted style
  - Current value via `os.Getenv`: if sensitive → `🔒 set` (masked); if set → truncated value; if unset → `(not set)` muted
  - Qualifier (if any) in muted style
  - Description wrapped to panel width
- `Up()` / `Down()` — scroll through entries
- `SetSize(w, h int)` — set content dimensions
- Uses `sensitive.IsEnvVarSensitive(name)` and `sensitive.LooksLikeToken(value)` for masking decisions

### Task 6: Add new styles for env vars panel
**Files:** `internal/tui/styles.go`
- Add `envVarNameStyle` (bold, primary colour)
- Add `envVarAliasStyle` (muted/dim)
- Add `envVarValueSetStyle` (green)
- Add `envVarValueUnsetStyle` (dim/muted)
- Add `envVarSensitiveStyle` (yellow with lock icon)

### Task 7: Wire view navigation in `Model`
**Files:** `internal/tui/model.go`
- Add `envVars []copilot.EnvVarInfo` and `envPanel *EnvVarsPanel` fields to `Model`
- Update `NewModel` signature to accept `envVars []copilot.EnvVarInfo`; construct `EnvVarsPanel`
- In `handleKeyPress`:
  - `StateBrowsing` + `"right"/"l"/"tab"` → `StateEnvVars`
  - `StateEnvVars` + `"left"/"h"/"tab"` → `StateBrowsing`
  - `StateEnvVars` + `"up"/"k"` → `envPanel.Up()`
  - `StateEnvVars` + `"down"/"j"` → `envPanel.Down()`
  - Suppress `ctrl+s` in `StateEnvVars` (read-only view, no saving)
- In `View()`: branch on `m.state == StateEnvVars` to render `envPanel.View()` instead of the two-panel config layout
- In `updateSizes()`: set `envPanel.SetSize()` for full-width layout
- Update `ShortHelp(state)` to return correct bindings per state per ADR-0004

### Task 8: Update `main.go` to call `DetectEnvVars`
**Files:** `cmd/ccc/main.go`
- Call `copilot.DetectEnvVars()` after `DetectSchema()` — log warning on error, default to empty slice
- Pass `envVars` to `tui.NewModel(cfg, schema, envVars, copilotVersion, configPath)`

### Task 9: Write tests
**Files:** `internal/tui/tui_test.go`, `internal/copilot/copilot_test.go`
- **ParseEnvVars tests:** fixture data covering all 11 entries, multi-name entries, empty output, malformed output
- **State transition tests:** `StateBrowsing` → right → `StateEnvVars`; `StateEnvVars` → left → `StateBrowsing`; tab toggles; editing blocks view switch
- **EnvVarsPanel rendering tests:** non-sensitive env var shows value; sensitive env var shows masked value; unset env var shows `(not set)`
- **Help bar tests:** correct bindings shown per state
- **Sensitive env var tests:** `IsEnvVarSensitive` for known token names

## Dependency Order

```
Task 1 (DetectEnvVars)          — no dependencies
Task 2 (SensitiveEnvVars)       — no dependencies
Task 3 (StateEnvVars)           — no dependencies
Task 4 (Key bindings)           — no dependencies
Task 6 (Styles)                 — no dependencies
  ↓
Task 5 (EnvVarsPanel)           — depends on Tasks 1, 2, 6
  ↓
Task 7 (Model wiring)           — depends on Tasks 3, 4, 5
  ↓
Task 8 (main.go)                — depends on Tasks 1, 7
  ↓
Task 9 (Tests)                  — depends on all above
```

## Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| `copilot help environment` output format changes | Pin `ParseEnvVars` tests against a captured fixture; fail gracefully (empty panel, log warning) |
| Token env var values logged accidentally | Apply `sensitive.MaskValue` before any `slog` call; never log raw `os.Getenv` results for sensitive vars |
| `NewModel` signature change breaks existing tests | Single call site in `main.go`; update tests to pass `nil`/empty env vars slice |
| Terminal too narrow for env var entry display | Word-wrap descriptions; truncate values with ellipsis |

## Verification

- All `go test ./...` pass
- Manual: run `ccc`, press `→` or `tab` — env vars panel appears
- Manual: press `←` or `tab` — config view returns
- Manual: confirm `COPILOT_GITHUB_TOKEN` shows `🔒 set` if set, `(not set)` if unset
- Manual: confirm help bar changes per view
- Manual: confirm `ctrl+s` does nothing in env vars view
