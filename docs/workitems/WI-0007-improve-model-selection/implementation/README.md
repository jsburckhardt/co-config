# Implementation Notes: WI-0007 — Improve Model Selection UX

## Task 1: Add `StateModelPicker` to state machine

- **Status:** Complete
- **Files Changed:** `internal/tui/state.go`
- **Tests Passed:** 25
- **Tests Failed:** 0

### Changes Summary

Added `StateModelPicker` constant to the `State` iota block in `internal/tui/state.go`, positioned between `StateEditing` and `StateSaving`. Added the corresponding `case StateModelPicker: return "ModelPicker"` to the `String()` switch statement.

No existing constants or their iota ordering were changed. `StateSaving` and `StateExiting` shift their numeric values by +1 due to the new constant, but this is safe since they are only compared by name, never by raw integer value.

### Test Results

All 25 existing tests (UT-TUI-001 through UT-TUI-025) pass. No dedicated unit test is needed for the constant itself per the task breakdown — it will be verified through state machine integration tests in Task 6 (UT-TUI-026, UT-TUI-027).

### Notes

- The `StateModelPicker` constant is not yet referenced by any code; it will be wired in Task 4.
- Compilation verified with `go build ./internal/tui/...`.

## Task 3: Add `Filter` key binding

- **Status:** Complete
- **Files Changed:** `internal/tui/keys.go`
- **Tests Passed:** 25
- **Tests Failed:** 0

### Changes Summary

Added a `Filter key.Binding` field to the `KeyMap` struct and initialized it in `DefaultKeyMap()` with key `/` and help text `"/ filter"`. No existing bindings were modified.

### Test Results

All 25 existing tests pass. Task 3 has no dedicated unit test; it is covered indirectly by UT-TUI-033 (help bar rendering for `StateModelPicker`), which will be added in Task 6.

### Notes

- The `Filter` field is placed after `Tab` in the struct, consistent with the task breakdown.
- The binding uses `/` as the trigger key, matching common TUI conventions (e.g., vim, less).

## Task 2: Create `ModelPickerPanel` component

- **Status:** Complete
- **Files Changed:** `internal/tui/model_picker_panel.go` (created, ~90 lines)
- **Tests Passed:** 25 (all existing) + 2 manual verification tests for UT-TUI-028 and UT-TUI-029
- **Tests Failed:** 0

### Changes Summary

Created `internal/tui/model_picker_panel.go` containing:

1. **`modelItem` struct** — Implements `list.Item` interface with `Title()`, `Description()`, and `FilterValue()` methods. Description returns empty string since model options don't have descriptions.

2. **`ModelPickerPanel` struct** — Wraps `bubbles/list.Model` with a `selected` string fallback field.

3. **`NewModelPickerPanel(options []string, current string)`** constructor:
   - Converts options to `[]list.Item` as `modelItem` values
   - Creates a `list.NewDefaultDelegate()` with custom styling:
     - `SelectedTitle` foreground set to `primaryColor` (from `styles.go`)
     - `SelectedTitle` border foreground set to `primaryColor`
     - `SelectedDesc` hidden (height 0, no colors) since we have no descriptions
   - Creates `list.New(items, delegate, 0, 0)` with:
     - `SetShowStatusBar(false)` — no status bar
     - `SetFilteringEnabled(true)` — fuzzy filtering via `sahilm/fuzzy`
     - `SetShowHelp(false)` — avoids conflicting with app's help bar
   - Title set to `"Select Model"` styled with bold + primaryColor
   - Pre-selects current value by iterating options and calling `l.Select(i)`

4. **Methods:**
   - `SetSize(w, h int)` — delegates to `list.SetSize`
   - `Update(msg tea.Msg) tea.Cmd` — delegates to `list.Update`
   - `SelectedValue() string` — returns highlighted item's name or fallback `selected`
   - `View() string` — returns `list.View()`
   - `FilterState() list.FilterState` — returns current filter state (needed by Task 4 for Enter key handling)

### Test Results

All 25 existing tests (UT-TUI-001 through UT-TUI-025) pass without modification. Manual verification confirmed UT-TUI-028 (creation + pre-selection) and UT-TUI-029 (View renders non-empty with title) behavior. Formal tests will be added in Task 6.

### Notes

- No new dependencies added to `go.mod` — `bubbles/list` was already present (v0.21.1) and `sahilm/fuzzy` is an indirect dependency of bubbles.
- The `primaryColor` variable is referenced directly from `styles.go` (same package).
- The `FilterState()` method is included per the task breakdown — it will be used by Task 4 to distinguish between "filtering active" (Enter confirms filter text) vs "not filtering" (Enter confirms selection).
- File not committed per instructions.

## Task 4: Wire `StateModelPicker` into main model

- **Status:** Complete
- **Files Changed:** `internal/tui/model.go`
- **Tests Passed:** 25
- **Tests Failed:** 0

### Changes Summary

Integrated the `ModelPickerPanel` into the main `Model` in `internal/tui/model.go` with 7 sub-changes:

1. **4a. Added `modelPickerPanel *ModelPickerPanel` field** to the `Model` struct, positioned after `detailPanel`.

2. **4b. Added non-key message forwarding** for `StateModelPicker` in `Update()` — when `m.state == StateModelPicker && m.modelPickerPanel != nil`, non-key messages (e.g., blink timers) are forwarded to the picker's `Update()`.

3. **4c. Modified `StateBrowsing` / `enter` case** in `handleKeyPress` — before the existing `StateEditing` transition, checks if the field is an enum with `len(item.Field.Options) >= 5`. If so, creates a `NewModelPickerPanel`, sizes it via `updateSizes()`, assigns to `m.modelPickerPanel`, and transitions to `StateModelPicker`. The `isSensitiveItem()` guard still blocks entry for sensitive fields (unchanged).

4. **4d. Added `StateModelPicker` case** in `handleKeyPress` switch:
   - `Enter`: If `FilterState() != list.Filtering`, confirms selection (reads `SelectedValue()`, writes via `cfg.Set()`, updates `ListPanel` and `DetailPanel`, clears picker, returns to `StateBrowsing`). If filtering is active, forwards Enter to the list.
   - `Esc`: Always confirms current selection and returns to `StateBrowsing`.
   - All other keys: Forwarded to `m.modelPickerPanel.Update(msg)`.

5. **4e. Added `list` import** — `"github.com/charmbracelet/bubbles/list"` for `list.Filtering` constant.

6. **4f. Updated `updateSizes()`** — after existing list/detail sizing, propagates size to `modelPickerPanel` when non-nil with floor guards for width and height.

7. **4g. Updated `View()`** — when `state == StateModelPicker && modelPickerPanel != nil`, replaces the two-panel layout with a single full-width picker panel rendered using `focusedPanelStyle`.

### Test Results

All 25 existing tests (UT-TUI-001 through UT-TUI-025) pass without modification. Build succeeds with `go build ./internal/tui/...`. No existing logic for `StateBrowsing` or `StateEditing` was changed beyond the enter key branching in `StateBrowsing`.

### Notes

- Global keys (`ctrl+c`, `ctrl+s`) continue to work from `StateModelPicker` since they are handled before the `switch m.state` block.
- The `DetailPanel` code (`detail_panel.go`) and `ListPanel` code (`list_item.go`) are unchanged.
- Small enums (<5 options) and non-enum fields continue to use the existing `StateEditing` flow.
- The picker nil-guard in the `StateModelPicker` case provides a safe fallback to `StateBrowsing` if the picker is unexpectedly nil.

## Task 5: Update help bar for `StateModelPicker`

- **Status:** Complete
- **Files Changed:** `internal/tui/model.go`
- **Tests Passed:** 25
- **Tests Failed:** 0

### Changes Summary

Added a `case StateModelPicker:` to the `ShortHelp()` method in `internal/tui/model.go`, returning `[]key.Binding{k.Filter, k.Enter, k.Escape, k.Save, k.Quit}`. This displays the help bar `/ filter  •  enter edit  •  esc done  •  ctrl+s save  •  ctrl+c quit` when the model picker is active.

The new case was inserted between the existing `StateEditing` case and the `default` case. No existing cases were modified.

### Test Results

All 25 existing tests (UT-TUI-001 through UT-TUI-025) pass without modification. The help bar rendering will be further verified by UT-TUI-032 and UT-TUI-033 in Task 6.

### Notes

- The `Filter` binding (added in Task 3) is now displayed in the help bar only during `StateModelPicker`.
- The bindings order matches the user's interaction flow: filter → confirm → back → save → quit.

## Task 6: Add unit tests UT-TUI-026 through UT-TUI-033

- **Status:** Complete
- **Files Changed:** `internal/tui/tui_test.go`
- **Tests Passed:** 33
- **Tests Failed:** 0

### Changes Summary

Added 8 new unit tests to `internal/tui/tui_test.go` (UT-TUI-026 through UT-TUI-033) and added `"fmt"` to the imports. No existing tests (UT-TUI-001 through UT-TUI-025) were modified.

| Test ID | Test Function | Description |
|---------|--------------|-------------|
| UT-TUI-026 | `TestEnterOnLargeEnumTransitionsToModelPicker` | Enter on enum ≥5 options → `StateModelPicker` |
| UT-TUI-027 | `TestEnterOnSmallEnumTransitionsToEditing` | Enter on enum <5 options → `StateEditing` (not picker) |
| UT-TUI-028 | `TestModelPickerPanelCreation` | `NewModelPickerPanel` creates panel; `SelectedValue()` returns current |
| UT-TUI-029 | `TestModelPickerPanelView` | `ModelPickerPanel.View()` renders non-empty string with title |
| UT-TUI-030 | `TestEscFromModelPickerConfirmsAndReturns` | Esc writes value to config, returns to `StateBrowsing` |
| UT-TUI-031 | `TestEnterFromModelPickerConfirmsAndReturns` | Enter (not filtering) confirms, returns to `StateBrowsing` |
| UT-TUI-032 | `TestViewInModelPickerAtVariousSizes` | `View()` in `StateModelPicker` at 80×24, 120×40, 100×30 renders without panic |
| UT-TUI-033 | `TestCtrlSSaveFromModelPicker` | `ctrl+s` save works from `StateModelPicker` state |

### Test Results

All 33 tests pass with `go test ./internal/tui/... -v`:
- 25 existing tests (UT-TUI-001 through UT-TUI-025): all PASS
- 8 new tests (UT-TUI-026 through UT-TUI-033): all PASS
- UT-TUI-007 (`TestBrowsingToEditingTransition`) continues to pass — it uses a 2-option enum schema, below the 5-option threshold
- UT-TUI-032 uses table-driven subtests (`80x24`, `120x40`, `100x30`)

### Notes

- Added `"fmt"` import for `fmt.Sprintf` in `TestViewInModelPickerAtVariousSizes` subtest naming.
- UT-TUI-033 uses a non-existent config path (`/tmp/nonexistent/config.json`) to verify the save handler is reached; however, `config.SaveConfig` succeeded anyway (it creates the directory), so `m.saved == true` was set rather than `m.err != nil`. The test assertion `!m.saved && m.err == nil` correctly covers both cases.
- All tests follow the existing patterns: `NewModel` + set window dimensions + `updateSizes()` + send key messages via `Update()` + assert state.
