# Implementation Notes: WI-0005 Environment Variables Panel

## Task 3: Add StateEnvVars to TUI state machine

- **Status:** Complete
- **Files Changed:** `internal/tui/state.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 27
- **Tests Failed:** 0

### Changes Summary

Added the `StateEnvVars` state to the TUI state machine as specified in ADR-0004. This state represents the environment variables view being active (read-only).

1. **`internal/tui/state.go`** — Added `StateEnvVars` constant after `StateExiting` in the `iota` const block (value 4). Added `case StateEnvVars: return "EnvVars"` to the `String()` method.
2. **`internal/tui/tui_test.go`** — Added two new tests:
   - `TestStateEnvVarsString` (UT-TUI-026): Verifies `StateEnvVars.String()` returns `"EnvVars"`.
   - `TestStateEnvVarsDistinct` (UT-TUI-027): Verifies `StateEnvVars` has a distinct value from all other states (`StateBrowsing`, `StateEditing`, `StateSaving`, `StateExiting`).

### Test Results

All 27 TUI tests pass, including 25 existing tests (UT-TUI-001 through UT-TUI-025) and 2 new tests (UT-TUI-026, UT-TUI-027). `go build ./internal/tui/...` compiles without errors.

### Notes

- No existing state values or string representations were modified.
- The `StateEnvVars` iota value is 4, following `StateExiting` (3), preserving the existing state numbering.

---

## Task 4: Add Left/Right key bindings to KeyMap

- **Status:** Complete
- **Files Changed:** `internal/tui/keys.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 30 (all TUI tests)
- **Tests Failed:** 0

### Changes Summary

Extended `KeyMap` with `Left` and `Right` bindings for horizontal view navigation, as specified in ADR-0004.

1. **`internal/tui/keys.go`** — Added `Left` and `Right` fields to the `KeyMap` struct with corresponding bindings in `DefaultKeyMap()`:
   - `Left`: keys `"left"`, `"h"` with help text `"←/h" → "config"`
   - `Right`: keys `"right"`, `"l"` with help text `"→/l" → "env vars"`
   - Updated `Tab` help description from `"switch"` to `"switch view"` for clarity.

2. **`internal/tui/tui_test.go`** — Added `key` import and three new tests:
   - `TestDefaultKeyMapLeftBinding` (UT-TUI-028): Verifies Left binding matches `tea.KeyLeft` and `'h'` rune, and help desc is `"config"`.
   - `TestDefaultKeyMapRightBinding` (UT-TUI-029): Verifies Right binding matches `tea.KeyRight` and `'l'` rune, and help desc is `"env vars"`.
   - `TestDefaultKeyMapTabHelpText` (UT-TUI-030): Verifies Tab help desc is `"switch view"`.

### Test Results

All 30 TUI tests pass (25 existing + 2 from Task 3 + 3 new). Full suite (`go test ./...`) passes. `go vet ./...` clean.

### Notes

- Follows ADR-0004 key binding specification exactly.
- Left/Right fields are placed between Down and Enter in the struct to maintain logical grouping (vertical nav → horizontal nav → action keys).

---

## Task 6: Create EnvVarsPanel component

- **Status:** Complete
- **Files Changed:** `internal/tui/env_panel.go` (new), `internal/tui/tui_test.go`
- **Tests Passed:** 40
- **Tests Failed:** 0

### Changes Summary

Created `EnvVarsPanel` component in `internal/tui/env_panel.go`, a read-only scrollable list for environment variables. Modeled after the existing `ListPanel` pattern in `list_item.go`.

1. **`internal/tui/env_panel.go`** (new file) — Contains:
   - `EnvVarsPanel` struct with `envVars`, `cursor`, `offset`, `width`, `height` fields.
   - `NewEnvVarsPanel(envVars []copilot.EnvVarInfo) *EnvVarsPanel` constructor.
   - `SetSize(w, h int)`, `Up()`, `Down()`, `Cursor()` navigation methods.
   - `View()` renders entries with 4 lines each: primary name + value status, aliases/qualifier, description, blank separator.
   - `ensureVisible()` keeps cursor in viewport using `linesPerEntry=4` to calculate visible entries.
   - Value resolution: iterates `entry.Names`, calls `os.Getenv()` for each, uses first non-empty value.
   - Sensitivity: checks `sensitive.IsEnvVarSensitive(name)` for each name AND `sensitive.LooksLikeToken(value)` for the resolved value.
   - Display: sensitive+set → `🔒 set`, non-sensitive+set → value (truncated to 30 chars), unset → `(not set)`.
   - Empty/nil envVars → renders `"No environment variables detected"` placeholder.
   - Zero width/height → returns `""`.
   - Uses styles from Task 5: `envVarNameStyle`, `envVarAliasStyle`, `envVarValueSetStyle`, `envVarValueUnsetStyle`, `envVarSensitiveStyle`, `envVarQualifierStyle`, `envVarDescStyle`.

2. **`internal/tui/tui_test.go`** — Added 10 new tests:
   - `TestNewEnvVarsPanelCursorStart` (UT-TUI-031): Non-empty slice starts cursor at 0.
   - `TestNewEnvVarsPanelEmptyNoPanic` (UT-TUI-032): nil and empty slice don't panic.
   - `TestEnvVarsPanelViewPrimaryNames` (UT-TUI-033): Renders primary name for each entry.
   - `TestEnvVarsPanelViewAliasNames` (UT-TUI-034): Renders alias names for multi-name entries.
   - `TestEnvVarsPanelViewSensitiveSet` (UT-TUI-035): Shows 🔒 for sensitive set env var, hides raw token.
   - `TestEnvVarsPanelViewUnset` (UT-TUI-036): Shows "not set" for unset env var.
   - `TestEnvVarsPanelViewNonSensitiveSet` (UT-TUI-037): Shows actual value for non-sensitive set env var.
   - `TestEnvVarsPanelCursorNavigation` (UT-TUI-038): Down/Up advance/retreat cursor within bounds.
   - `TestEnvVarsPanelViewQualifier` (UT-TUI-039): Renders qualifier text.
   - `TestEnvVarsPanelEmptyRender` (UT-TUI-040): Empty panel renders placeholder without panic.

### Test Results

All 40 TUI tests pass (30 existing + 10 new). Full suite (`go test ./...`) passes. `go vet ./...` clean.

### Notes

- Follows the same scrollable panel pattern as `ListPanel` but is simpler (no group headers, no editing).
- Each entry occupies exactly 4 lines (name+value, aliases/qualifier, description, separator) for consistent scrolling.
- Uses `t.Setenv()` in tests for clean env var setup/teardown.
- The panel is read-only by design per ADR-0004; no editing or save functionality.

---

## Task 7: Wire view navigation in Model

- **Status:** Complete
- **Files Changed:** `internal/tui/model.go`, `internal/tui/tui_test.go`, `cmd/ccc/main.go`
- **Tests Passed:** 54
- **Tests Failed:** 0

### Changes Summary

Wired horizontal view navigation between Config View and Env Vars View in the Model, implementing the full state transition and rendering logic per ADR-0004.

1. **`internal/tui/model.go`**:
   - Added `envVars []copilot.EnvVarInfo` and `envPanel *EnvVarsPanel` fields to `Model` struct.
   - Changed `NewModel` signature from 4 args to 5 args: `func NewModel(cfg, schema, envVars, version, configPath)`. Constructs `EnvVarsPanel` in the body and stores both `envVars` and `envPanel` on the model.
   - **`handleKeyPress`**: Moved `ctrl+s` from global handler into per-state handling — available in `StateBrowsing` and `StateEditing`, but NOT in `StateEnvVars` (read-only view). Added horizontal navigation in `StateBrowsing` (`right`/`l`/`tab` → `StateEnvVars`). Added new `StateEnvVars` case (`left`/`h`/`tab` → `StateBrowsing`, `up`/`k`/`down`/`j` → panel cursor movement). In `StateEditing`, non-`esc`/non-`ctrl+s` keys are forwarded to detail panel via `default` case.
   - **`updateSizes`**: Added env panel sizing using full `innerWidth - 4` width and `panelHeight - 2` height, with floor guards.
   - **`View()`**: Branches on `m.state` for panel rendering. In `StateEnvVars`, renders env panel full-width with `focusedPanelStyle`. In `StateBrowsing`/`StateEditing`, renders existing two-panel layout. Header and outer frame remain constant.
   - **`ShortHelp`**: `StateBrowsing` now includes `Right`, `Tab` bindings. `StateEnvVars` includes `Up`, `Down`, `Left`, `Tab`, `Quit` (no `Enter`, `Save`). `StateEditing` unchanged.

2. **`cmd/ccc/main.go`**: Updated `tui.NewModel` call to pass `nil` for envVars (env var detection will be wired in Task 8).

3. **`internal/tui/tui_test.go`**: Updated all 13 existing `NewModel` calls to use new 5-argument signature (passing `nil` for envVars). Added 14 new tests:
   - `TestBrowsingRightToEnvVars` (UT-TUI-041): right key in StateBrowsing → StateEnvVars.
   - `TestBrowsingLToEnvVars` (UT-TUI-042): 'l' key in StateBrowsing → StateEnvVars.
   - `TestBrowsingTabToEnvVars` (UT-TUI-043): tab key in StateBrowsing → StateEnvVars.
   - `TestEnvVarsLeftToBrowsing` (UT-TUI-044): left key in StateEnvVars → StateBrowsing.
   - `TestEnvVarsHToBrowsing` (UT-TUI-045): 'h' key in StateEnvVars → StateBrowsing.
   - `TestEnvVarsTabToBrowsing` (UT-TUI-046): tab key in StateEnvVars → StateBrowsing.
   - `TestEditingRightNoTransition` (UT-TUI-047): right key in StateEditing stays in StateEditing.
   - `TestEditingLeftNoTransition` (UT-TUI-048): left key in StateEditing stays in StateEditing.
   - `TestEnvVarsCtrlSNoSave` (UT-TUI-049): ctrl+s in StateEnvVars does not trigger save.
   - `TestViewInEnvVarsState` (UT-TUI-050): View renders env panel content in StateEnvVars.
   - `TestShortHelpBrowsingIncludesRight` (UT-TUI-051): ShortHelp includes 'env vars' binding.
   - `TestShortHelpEnvVarsBindings` (UT-TUI-052): ShortHelp includes 'config', omits 'edit'/'save'.
   - `TestShortHelpEditingUnchanged` (UT-TUI-053): ShortHelp for editing unchanged.
   - `TestNewModelNilEnvVars` (UT-TUI-054): NewModel with nil envVars does not panic.

### Test Results

All 54 TUI tests pass (40 existing + 14 new). Full suite (`go test ./...`) passes. `go build ./...` compiles cleanly. `go vet ./...` reports no issues.

### Notes

- The `ctrl+s` handler was refactored from a global key to per-state handling. This is the cleanest way to prevent save in the read-only env vars view without adding a conditional check.
- In `StateEditing`, the `default` case in the inner switch forwards unrecognized keys (including left/right arrow keys) to the detail panel, preserving text cursor movement in input widgets.
- `cmd/ccc/main.go` passes `nil` for envVars — actual env var detection (`copilot.RunHelpEnvironment` + `ParseEnvVars`) will be wired in a subsequent task.
