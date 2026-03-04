# Task Breakdown: Environment Variables Panel with Tab Navigation

- **Workitem:** WI-0005-environment-variables
- **Action Plan:** [01-action-plan.md](01-action-plan.md)

---

## Task 1: Add `EnvVarInfo` struct, `ParseEnvVars`, and `DetectEnvVars` to `internal/copilot`

- **Status:** Not Started
- **Complexity:** Medium
- **Dependencies:** None
- **Related ADRs:** [ADR-0004](../../../architecture/ADR/ADR-0004-tui-multi-view-navigation.md)
- **Related Core-Components:** [CC-0002](../../../architecture/core-components/CORE-COMPONENT-0002-error-handling.md), [CC-0004](../../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md)

### Description

Add the data model and parsing logic for environment variable metadata sourced from `copilot help environment`. This follows the established `DetectSchema`/`ParseSchema` pattern in `internal/copilot/copilot.go`.

1. Define `EnvVarInfo` struct in `internal/copilot/copilot.go`:
   ```go
   type EnvVarInfo struct {
       Names       []string // primary name is Names[0]; aliases follow
       Description string
       Qualifier   string   // e.g. "in order of precedence", may be empty
   }
   ```

2. Add `ErrEnvVarsParseFailed` sentinel error to `internal/copilot/errors.go`.

3. Implement `ParseEnvVars(output string) ([]EnvVarInfo, error)`:
   - Parse backtick-quoted names (`` `NAME` ``) from each entry's leading line.
   - Extract optional parenthesised qualifier text (e.g., `(in order of precedence)`).
   - Collect multi-line descriptions (everything after the colon, continuing on indented lines until the next entry or end of output).
   - Return `nil, nil` for empty output (graceful degradation, not an error).
   - Return `ErrEnvVarsParseFailed` only if the output is non-empty but completely unparseable.

4. Implement `DetectEnvVars() ([]EnvVarInfo, error)`:
   - Check for `copilot` binary via `exec.LookPath`.
   - Run `copilot help environment`.
   - Call `ParseEnvVars` on the output.
   - Return `ErrCopilotNotInstalled` if binary not found.

5. Create test fixture file `internal/copilot/testdata/copilot-help-environment.txt` with the captured output from the research brief.

### Acceptance Criteria

- [ ] `EnvVarInfo` struct is exported with `Names []string`, `Description string`, `Qualifier string` fields.
- [ ] `ErrEnvVarsParseFailed` sentinel error exists in `internal/copilot/errors.go`.
- [ ] `ParseEnvVars` correctly parses all 11 env var entries from the fixture data.
- [ ] Multi-name entries (e.g., `COPILOT_EDITOR`, `VISUAL`, `EDITOR`) produce a single `EnvVarInfo` with 3 names.
- [ ] Qualifier text is extracted (e.g., `"in order of precedence"`).
- [ ] Multi-line descriptions are concatenated with spaces.
- [ ] `ParseEnvVars("")` returns `nil, nil` (not an error).
- [ ] `DetectEnvVars` runs `copilot help environment` (never `gh copilot`).
- [ ] Errors are wrapped with `fmt.Errorf` and `%w` per CC-0002.

### Test Coverage

- UT-COP-010: `ParseEnvVars` with full fixture data returns 11 entries.
- UT-COP-011: `ParseEnvVars` multi-name entry (`COPILOT_EDITOR`, `VISUAL`, `EDITOR`) returns 3 names.
- UT-COP-012: `ParseEnvVars` single-name entry (`COPILOT_MODEL`) returns 1 name.
- UT-COP-013: `ParseEnvVars` extracts qualifier `"in order of precedence"`.
- UT-COP-014: `ParseEnvVars` multi-line description is concatenated.
- UT-COP-015: `ParseEnvVars` with empty string returns `nil, nil`.
- UT-COP-016: `ParseEnvVars` with malformed output returns `ErrEnvVarsParseFailed`.

---

## Task 2: Add `SensitiveEnvVars` list and `IsEnvVarSensitive` to `internal/sensitive`

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** None
- **Related ADRs:** [ADR-0004](../../../architecture/ADR/ADR-0004-tui-multi-view-navigation.md)
- **Related Core-Components:** [CC-0005](../../../architecture/core-components/CORE-COMPONENT-0005-sensitive-data-handling.md)

### Description

Extend the `internal/sensitive` package to handle environment variable name-based sensitivity detection. The existing `IsSensitive` function checks config field names; this task adds a parallel function for env var names.

1. Add `SensitiveEnvVars` package-level list: `COPILOT_GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_TOKEN`.
2. Add `IsEnvVarSensitive(name string) bool`:
   - Case-insensitive match against `SensitiveEnvVars` list.
   - Returns `true` if the env var name is in the list.
3. The env vars panel will use belt-and-suspenders: check both `IsEnvVarSensitive(name)` for the env var name AND `LooksLikeToken(value)` for the env var value.

### Acceptance Criteria

- [ ] `SensitiveEnvVars` exported variable contains exactly `["COPILOT_GITHUB_TOKEN", "GH_TOKEN", "GITHUB_TOKEN"]`.
- [ ] `IsEnvVarSensitive("COPILOT_GITHUB_TOKEN")` returns `true`.
- [ ] `IsEnvVarSensitive("GH_TOKEN")` returns `true`.
- [ ] `IsEnvVarSensitive("GITHUB_TOKEN")` returns `true`.
- [ ] `IsEnvVarSensitive("copilot_github_token")` returns `true` (case-insensitive).
- [ ] `IsEnvVarSensitive("COPILOT_MODEL")` returns `false`.
- [ ] `IsEnvVarSensitive("")` returns `false`.
- [ ] Existing `IsSensitive` and `LooksLikeToken` functions remain unchanged and all existing tests pass.

### Test Coverage

- UT-SEN-014: `IsEnvVarSensitive` returns `true` for `COPILOT_GITHUB_TOKEN`.
- UT-SEN-015: `IsEnvVarSensitive` returns `true` for `GH_TOKEN`.
- UT-SEN-016: `IsEnvVarSensitive` returns `true` for `GITHUB_TOKEN`.
- UT-SEN-017: `IsEnvVarSensitive` is case-insensitive (lowercase input returns `true`).
- UT-SEN-018: `IsEnvVarSensitive` returns `false` for non-sensitive env var names.
- UT-SEN-019: `IsEnvVarSensitive` returns `false` for empty string.

---

## Task 3: Add `StateEnvVars` to TUI state machine

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** None
- **Related ADRs:** [ADR-0004](../../../architecture/ADR/ADR-0004-tui-multi-view-navigation.md)
- **Related Core-Components:** None

### Description

Extend the state machine in `internal/tui/state.go` with the new `StateEnvVars` state as specified in ADR-0004.

1. Add `StateEnvVars State = iota` constant after `StateExiting` in the const block.
2. Update the `String()` method to return `"EnvVars"` for the new state.

### Acceptance Criteria

- [ ] `StateEnvVars` constant is defined and has the correct `iota` value (4, after `StateExiting`).
- [ ] `StateEnvVars.String()` returns `"EnvVars"`.
- [ ] Existing state constants and their `String()` values remain unchanged.

### Test Coverage

- UT-TUI-026: `StateEnvVars.String()` returns `"EnvVars"`.
- UT-TUI-027: `StateEnvVars` has a distinct value from all other states.

---

## Task 4: Add `Left`/`Right` key bindings to `KeyMap`

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** None
- **Related ADRs:** [ADR-0004](../../../architecture/ADR/ADR-0004-tui-multi-view-navigation.md)
- **Related Core-Components:** None

### Description

Extend the `KeyMap` struct in `internal/tui/keys.go` with Left and Right key bindings for horizontal navigation between views, as specified in ADR-0004.

1. Add `Left key.Binding` and `Right key.Binding` fields to `KeyMap`.
2. In `DefaultKeyMap()`:
   - `Left` → `key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "config"))`.
   - `Right` → `key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "env vars"))`.
3. Update `Tab` help text from `"switch"` to `"switch view"` for clarity.

### Acceptance Criteria

- [ ] `KeyMap` struct has `Left` and `Right` fields of type `key.Binding`.
- [ ] `DefaultKeyMap().Left` responds to `"left"` and `"h"` keys.
- [ ] `DefaultKeyMap().Right` responds to `"right"` and `"l"` keys.
- [ ] `DefaultKeyMap().Left.Help().Desc` is `"config"`.
- [ ] `DefaultKeyMap().Right.Help().Desc` is `"env vars"`.
- [ ] `DefaultKeyMap().Tab.Help().Desc` is `"switch view"`.

### Test Coverage

- UT-TUI-028: `DefaultKeyMap().Left` has keys `["left", "h"]`.
- UT-TUI-029: `DefaultKeyMap().Right` has keys `["right", "l"]`.
- UT-TUI-030: `DefaultKeyMap().Tab` help text is `"switch view"`.

---

## Task 5: Add new styles for env vars panel

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** None
- **Related ADRs:** [ADR-0003](../../../architecture/ADR/ADR-0003-two-panel-tui-layout.md), [ADR-0004](../../../architecture/ADR/ADR-0004-tui-multi-view-navigation.md)
- **Related Core-Components:** None

### Description

Add Lipgloss style definitions in `internal/tui/styles.go` for the environment variables panel rendering.

1. Add `envVarNameStyle` — bold, primary colour (for the primary env var name).
2. Add `envVarAliasStyle` — muted/dim (for alias names).
3. Add `envVarValueSetStyle` — green/success colour (for set env var values).
4. Add `envVarValueUnsetStyle` — dim/muted (for `(not set)` display).
5. Add `envVarSensitiveStyle` — yellow or warning colour with lock icon support (for `🔒 set` display).
6. Add `envVarQualifierStyle` — muted, italic (for qualifier text like "in order of precedence").
7. Add `envVarDescStyle` — standard text colour (for descriptions, matching existing `detailDescStyle`).

### Acceptance Criteria

- [ ] All seven style variables are defined and exported in `internal/tui/styles.go`.
- [ ] `envVarNameStyle` has `Bold(true)` and uses `primaryColor`.
- [ ] `envVarAliasStyle` uses `mutedColor`.
- [ ] `envVarValueSetStyle` uses `successColor`.
- [ ] `envVarValueUnsetStyle` uses `mutedColor`.
- [ ] `envVarSensitiveStyle` uses a warning/alert colour.
- [ ] No existing styles are modified.
- [ ] Code compiles without errors.

### Test Coverage

- No unit tests needed for pure style definitions; visual correctness is verified during Task 6 (EnvVarsPanel) tests and manual verification.

---

## Task 6: Create `EnvVarsPanel` component

- **Status:** Not Started
- **Complexity:** High
- **Dependencies:** Task 1 (EnvVarInfo struct), Task 2 (IsEnvVarSensitive), Task 5 (styles)
- **Related ADRs:** [ADR-0004](../../../architecture/ADR/ADR-0004-tui-multi-view-navigation.md)
- **Related Core-Components:** [CC-0005](../../../architecture/core-components/CORE-COMPONENT-0005-sensitive-data-handling.md)

### Description

Create the new `EnvVarsPanel` component in `internal/tui/env_panel.go`. This is the read-only, scrollable panel that displays environment variable entries in the env vars view.

1. Define `EnvVarsPanel` struct:
   - `envVars []copilot.EnvVarInfo` — the env var entries to display.
   - `cursor int` — currently highlighted entry index.
   - `offset int` — scroll offset for viewport.
   - `width int`, `height int` — panel content dimensions.

2. Implement `NewEnvVarsPanel(envVars []copilot.EnvVarInfo) *EnvVarsPanel`.

3. Implement `View() string`:
   - For each visible env var entry, render:
     - **Primary name** in bold/highlighted style (`envVarNameStyle`).
     - **Alias names** in muted style (`envVarAliasStyle`), separated by spaces.
     - **Current value** via `os.Getenv(name)` for each name in order:
       - If any name is in `sensitive.IsEnvVarSensitive` OR value matches `sensitive.LooksLikeToken` → show `🔒 set` in `envVarSensitiveStyle`.
       - If set and not sensitive → show truncated value in `envVarValueSetStyle`.
       - If not set (none of the names have a value) → show `(not set)` in `envVarValueUnsetStyle`.
     - **Qualifier** (if non-empty) in muted/italic style.
     - **Description** wrapped to panel width.
   - Highlight the entry at `cursor` position.
   - Respect `offset` for scrolling.

4. Implement `Up()` / `Down()` — move cursor through entries, adjusting scroll offset.

5. Implement `SetSize(w, h int)` — set content dimensions, adjust scroll.

### Acceptance Criteria

- [ ] `EnvVarsPanel` struct is defined in `internal/tui/env_panel.go`.
- [ ] `NewEnvVarsPanel` creates a panel with cursor at index 0.
- [ ] `View()` renders all env var entries when they fit in the panel height.
- [ ] Primary name is rendered with `envVarNameStyle`.
- [ ] Alias names are rendered with `envVarAliasStyle`.
- [ ] A sensitive env var with a set value shows `🔒 set`.
- [ ] A non-sensitive set env var shows the truncated value.
- [ ] An unset env var shows `(not set)`.
- [ ] Qualifier text is rendered when present.
- [ ] Description text is included in the rendering.
- [ ] `Up()` and `Down()` move cursor and adjust scroll offset.
- [ ] `SetSize` updates dimensions and adjusts scroll.
- [ ] Panel handles empty `envVars` slice without panicking.
- [ ] Uses `sensitive.IsEnvVarSensitive(name)` and `sensitive.LooksLikeToken(value)` (belt-and-suspenders).

### Test Coverage

- UT-TUI-031: `NewEnvVarsPanel` with non-empty slice starts cursor at 0.
- UT-TUI-032: `NewEnvVarsPanel` with empty slice does not panic.
- UT-TUI-033: `View()` renders primary name for each entry.
- UT-TUI-034: `View()` renders alias names for multi-name entries.
- UT-TUI-035: `View()` shows `🔒 set` for sensitive env var that is set (using `t.Setenv`).
- UT-TUI-036: `View()` shows `(not set)` for an unset env var.
- UT-TUI-037: `View()` shows truncated value for a non-sensitive set env var (using `t.Setenv`).
- UT-TUI-038: `Down()` advances cursor; `Up()` retreats cursor.
- UT-TUI-039: `View()` renders qualifier text when present.
- UT-TUI-040: Empty panel renders without panic.

---

## Task 7: Wire view navigation in `Model`

- **Status:** Not Started
- **Complexity:** High
- **Dependencies:** Task 1 (EnvVarInfo), Task 3 (StateEnvVars), Task 4 (Left/Right keys), Task 6 (EnvVarsPanel)
- **Related ADRs:** [ADR-0004](../../../architecture/ADR/ADR-0004-tui-multi-view-navigation.md), [ADR-0003](../../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0003](../../../architecture/core-components/CORE-COMPONENT-0003-logging.md), [CC-0005](../../../architecture/core-components/CORE-COMPONENT-0005-sensitive-data-handling.md)

### Description

Update `internal/tui/model.go` to integrate the EnvVarsPanel and implement view navigation between Config View and Env Vars View per ADR-0004.

1. **Model fields**: Add `envVars []copilot.EnvVarInfo` and `envPanel *EnvVarsPanel` to `Model`.

2. **NewModel signature**: Change to `NewModel(cfg *config.Config, schema []copilot.SchemaField, envVars []copilot.EnvVarInfo, version, configPath string)`. Construct `EnvVarsPanel` from `envVars`.

3. **handleKeyPress** — update key dispatch:
   - `StateBrowsing` + `"right"` / `"l"` / `"tab"` → transition to `StateEnvVars`.
   - `StateEnvVars` + `"left"` / `"h"` / `"tab"` → transition to `StateBrowsing`.
   - `StateEnvVars` + `"up"` / `"k"` → `envPanel.Up()`.
   - `StateEnvVars` + `"down"` / `"j"` → `envPanel.Down()`.
   - `StateEditing` — no change (left/right forwarded to input widget as before).
   - Suppress `ctrl+s` in `StateEnvVars` (read-only view, no saving).

4. **View()**: Branch on `m.state == StateEnvVars` to render `envPanel.View()` instead of the two-panel config layout. The header and outer frame remain constant across views per ADR-0004.

5. **updateSizes()**: Call `envPanel.SetSize()` using full-width layout for the env vars view.

6. **ShortHelp(state)**: Update to return correct bindings per state per ADR-0004:
   - `StateBrowsing`: `Up, Down, Enter, Right/Tab (env vars), Save, Quit`.
   - `StateEditing`: `Escape, Save, Quit`.
   - `StateEnvVars`: `Up, Down, Left/Tab (config), Quit`.

### Acceptance Criteria

- [ ] `NewModel` accepts `envVars []copilot.EnvVarInfo` parameter.
- [ ] `NewModel` constructs an `EnvVarsPanel` from the provided env vars.
- [ ] Pressing `right`, `l`, or `tab` in `StateBrowsing` transitions to `StateEnvVars`.
- [ ] Pressing `left`, `h`, or `tab` in `StateEnvVars` transitions to `StateBrowsing`.
- [ ] `up`/`k` and `down`/`j` in `StateEnvVars` scroll the env panel.
- [ ] `ctrl+s` in `StateEnvVars` does NOT save (read-only view).
- [ ] `left`/`right` keys in `StateEditing` are forwarded to the input widget (no view switch).
- [ ] `View()` renders the env vars panel content when in `StateEnvVars`.
- [ ] `View()` renders the two-panel config layout when in `StateBrowsing` or `StateEditing`.
- [ ] Header and outer frame are rendered identically in both views.
- [ ] `envPanel.SetSize()` is called in `updateSizes()`.
- [ ] `ShortHelp(StateBrowsing)` includes Right/Tab binding for env vars.
- [ ] `ShortHelp(StateEnvVars)` includes Left/Tab binding for config, omits Enter and Save.
- [ ] `ShortHelp(StateEditing)` remains unchanged (Escape, Save, Quit).
- [ ] Passing `nil` env vars to `NewModel` does not panic.

### Test Coverage

- UT-TUI-041: `StateBrowsing` + `right` key → `StateEnvVars`.
- UT-TUI-042: `StateBrowsing` + `l` key → `StateEnvVars`.
- UT-TUI-043: `StateBrowsing` + `tab` key → `StateEnvVars`.
- UT-TUI-044: `StateEnvVars` + `left` key → `StateBrowsing`.
- UT-TUI-045: `StateEnvVars` + `h` key → `StateBrowsing`.
- UT-TUI-046: `StateEnvVars` + `tab` key → `StateBrowsing`.
- UT-TUI-047: `StateEditing` + `right` key does NOT transition to `StateEnvVars` (stays in editing).
- UT-TUI-048: `StateEditing` + `left` key does NOT transition (stays in editing).
- UT-TUI-049: `ctrl+s` in `StateEnvVars` does not trigger save.
- UT-TUI-050: `View()` in `StateEnvVars` renders env panel content.
- UT-TUI-051: `ShortHelp(StateBrowsing)` includes right/tab binding.
- UT-TUI-052: `ShortHelp(StateEnvVars)` includes left/tab binding and omits enter/save.
- UT-TUI-053: `ShortHelp(StateEditing)` remains unchanged.
- UT-TUI-054: `NewModel` with nil envVars does not panic.

---

## Task 8: Update `main.go` to call `DetectEnvVars`

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** Task 1 (DetectEnvVars), Task 7 (NewModel signature change)
- **Related ADRs:** [ADR-0004](../../../architecture/ADR/ADR-0004-tui-multi-view-navigation.md)
- **Related Core-Components:** [CC-0002](../../../architecture/core-components/CORE-COMPONENT-0002-error-handling.md), [CC-0003](../../../architecture/core-components/CORE-COMPONENT-0003-logging.md)

### Description

Update `cmd/ccc/main.go` to call `copilot.DetectEnvVars()` at startup and pass the results to the TUI model.

1. After the `DetectSchema()` call, add `copilot.DetectEnvVars()` call.
2. On error, log a warning with `slog.Warn` and default to an empty `[]copilot.EnvVarInfo{}` slice (graceful degradation — env vars panel will simply be empty).
3. On success, log the count with `slog.Info`.
4. Update the `tui.NewModel(...)` call to pass the `envVars` slice as the new parameter.

### Acceptance Criteria

- [ ] `copilot.DetectEnvVars()` is called after `DetectSchema()` in `main.go`.
- [ ] On `DetectEnvVars` error, a warning is logged and an empty slice is used (app does not crash).
- [ ] On success, the count of detected env vars is logged at info level.
- [ ] `tui.NewModel` is called with the env vars slice in the correct position.
- [ ] The `copilot` binary is invoked as `copilot` (never `gh copilot`).
- [ ] The application compiles and runs successfully.

### Test Coverage

- No new unit tests for `main.go` (it is an integration entry point). Verified via:
  - Compilation check (`go build ./...`).
  - Existing tests updated to use new `NewModel` signature.
  - Manual verification per action plan's Verification section.

---

## Task 9: Update existing tests for `NewModel` signature change

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** Task 7 (NewModel signature change)
- **Related ADRs:** [ADR-0004](../../../architecture/ADR/ADR-0004-tui-multi-view-navigation.md)
- **Related Core-Components:** None

### Description

All existing tests in `internal/tui/tui_test.go` call `NewModel(cfg, schema, version, configPath)` with 4 arguments. After Task 7, the signature becomes `NewModel(cfg, schema, envVars, version, configPath)` with 5 arguments. Update every existing test call site to pass `nil` (or an empty slice) as the `envVars` parameter.

### Acceptance Criteria

- [ ] All existing test functions in `internal/tui/tui_test.go` compile with the new 5-argument `NewModel` signature.
- [ ] All existing tests (UT-TUI-001 through UT-TUI-025) continue to pass.
- [ ] No test logic is changed — only the `NewModel` call signature is updated.

### Test Coverage

- Run `go test ./internal/tui/...` — all 25 existing tests pass unchanged.

---

## Dependency Graph

```
Task 1 (DetectEnvVars/ParseEnvVars)  ─── no deps
Task 2 (SensitiveEnvVars)            ─── no deps
Task 3 (StateEnvVars)                ─── no deps
Task 4 (Left/Right keys)            ─── no deps
Task 5 (Styles)                      ─── no deps
    │
    ├─── Task 6 (EnvVarsPanel)       ─── depends on Task 1, Task 2, Task 5
    │       │
    │       └─── Task 7 (Model wiring) ─── depends on Task 3, Task 4, Task 6
    │               │
    │               ├─── Task 8 (main.go) ─── depends on Task 1, Task 7
    │               │
    │               └─── Task 9 (Update existing tests) ─── depends on Task 7
    │
```

## Implementation Order

1. Tasks 1, 2, 3, 4, 5 — can be implemented in parallel (no dependencies).
2. Task 6 — after Tasks 1, 2, 5.
3. Task 7 — after Tasks 3, 4, 6.
4. Tasks 8, 9 — after Task 7 (can be done in parallel).
