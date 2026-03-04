# Test Plan: WI-0006 — User Experience Saving and Viewing Configuration Values

> **Note:** This file should be at `plan/03-test-plan.md` once the `plan/` directory is created.

- **Workitem:** WI-0006-user-experience-saving
- **Action Plan:** [01-action-plan.md](01-action-plan.md)
- **Task Breakdown:** [02-task-breakdown.md](02-task-breakdown.md)

## Overview

All tests are unit tests in `internal/tui/tui_test.go` using `go test`. No integration tests against external binaries are required — the config `LoadConfig`/`SaveConfig` round-trip uses temp files. All tests follow the existing pattern: construct `config.NewConfig()`, define a schema, build a `Model`, simulate key presses via `tea.KeyMsg`, and assert model state.

### Test Summary

| Test ID | Title | Type | Task | Priority |
|---------|-------|------|------|----------|
| UT-TUI-026 | Enter commits enum field and returns to Browsing | Unit | Task 2 | P0 |
| UT-TUI-027 | Enter commits string field and returns to Browsing | Unit | Task 2 | P0 |
| UT-TUI-028 | Enter commits bool field and returns to Browsing | Unit | Task 2 | P0 |
| UT-TUI-029 | Enter on list field stays in Editing | Unit | Task 2 | P0 |
| UT-TUI-030 | Modified flag default and after UpdateItemValue | Unit | Task 3 | P0 |
| UT-TUI-031 | renderItem appends (not-saved) when Modified | Unit | Task 4 | P0 |
| UT-TUI-032 | ClearAllModified resets all Modified flags | Unit | Task 3 | P0 |
| UT-TUI-033 | Saved flag cleared after commit | Unit | Task 1/6 | P0 |
| UT-TUI-034 | Help bar shows enter confirm for non-list editing | Unit | Task 9 | P1 |
| UT-TUI-035 | Help bar omits enter confirm for list editing | Unit | Task 9 | P1 |
| UT-TUI-008+ | Extended: Esc commit clears saved and marks Modified | Unit | Task 1/3/6 | P0 |
| UT-TUI-036 | CurrentFieldType accessor returns correct type | Unit | Task 8 | P1 |
| UT-TUI-037 | CurrentFieldType returns empty string for nil field | Unit | Task 8 | P1 |
| UT-TUI-038 | (not-saved) not shown when Modified is false | Unit | Task 4 | P1 |
| UT-TUI-039 | Narrow terminal with (not-saved) does not panic | Unit | Task 4 | P1 |
| UT-TUI-040 | Post-save reload preserves cursor by field name | Unit | Task 5 | P0 |
| UT-TUI-041 | Modified flags cleared after successful save | Unit | Task 7 | P0 |
| UT-TUI-042 | Save failure does not clear Modified flags | Unit | Task 7 | P1 |

---

## Test UT-TUI-026: Enter commits enum field and returns to Browsing

- **Type:** Unit
- **Task:** Task 2
- **Priority:** P0

### Setup

1. Create `config.NewConfig()` with `cfg.Set("model", "gpt-4")`
2. Define schema with one enum field: `{Name: "model", Type: "enum", Options: ["gpt-4", "gpt-3.5-turbo"]}`
3. Create model via `NewModel(cfg, schema, "0.0.412", "/tmp/config.json")`
4. Set `model.windowWidth = 100`, `model.windowHeight = 30`, call `model.updateSizes()`
5. Transition to editing: send `tea.KeyMsg{Type: tea.KeyEnter}` to enter editing mode

### Steps

1. Verify `model.state == StateEditing`
2. Send `tea.KeyMsg{Type: tea.KeyEnter}` (Enter key while editing enum field)
3. Inspect the returned model

### Expected Result

- `m.state == StateBrowsing` — returned to browsing
- `m.cfg.Get("model")` reflects the committed value
- `m.listPanel.SelectedItem().Value` matches the committed value

---

## Test UT-TUI-027: Enter commits string field and returns to Browsing

- **Type:** Unit
- **Task:** Task 2
- **Priority:** P0

### Setup

1. Create `config.NewConfig()` with `cfg.Set("theme", "dark")`
2. Define schema with one string field: `{Name: "theme", Type: "string", Default: ""}`
3. Create model, set window size, and enter editing mode on the "theme" field

### Steps

1. Verify `model.state == StateEditing`
2. Send `tea.KeyMsg{Type: tea.KeyEnter}`
3. Inspect the returned model

### Expected Result

- `m.state == StateBrowsing`
- The string value is committed to `m.cfg`

---

## Test UT-TUI-028: Enter commits bool field and returns to Browsing

- **Type:** Unit
- **Task:** Task 2
- **Priority:** P0

### Setup

1. Create `config.NewConfig()` with `cfg.Set("stream", true)`
2. Define schema with one bool field: `{Name: "stream", Type: "bool", Default: "true"}`
3. Create model, set window size, and enter editing mode

### Steps

1. Verify `model.state == StateEditing`
2. Optionally toggle the bool (send space key)
3. Send `tea.KeyMsg{Type: tea.KeyEnter}`
4. Inspect the returned model

### Expected Result

- `m.state == StateBrowsing`
- `m.cfg.Get("stream")` reflects the (possibly toggled) bool value

---

## Test UT-TUI-029: Enter on list field stays in Editing

- **Type:** Unit
- **Task:** Task 2
- **Priority:** P0

### Setup

1. Create `config.NewConfig()` with `cfg.Set("allowed_urls", []any{"https://example.com"})`
2. Define schema with one list field: `{Name: "allowed_urls", Type: "list"}`
3. Create model, set window size, and enter editing mode

### Steps

1. Verify `model.state == StateEditing`
2. Verify `model.detailPanel.CurrentFieldType() == "list"`
3. Send `tea.KeyMsg{Type: tea.KeyEnter}`
4. Inspect the returned model

### Expected Result

- `m.state == StateEditing` — Enter was forwarded to textarea (inserts newline), does NOT exit editing
- The model is still in editing mode with the detail panel active

---

## Test UT-TUI-030: Modified flag default and after UpdateItemValue

- **Type:** Unit
- **Task:** Task 3
- **Priority:** P0

### Setup

1. Create list entries via `buildEntries()` with a sample schema
2. Create `NewListPanel(entries)`

### Steps

1. Iterate all entries and check `item.Modified` is `false`
2. Call `lp.UpdateItemValue("field_name", "new_value")`
3. Find the entry with `Field.Name == "field_name"` and check `item.Modified`

### Expected Result

- All entries initially have `Modified == false`
- After `UpdateItemValue`, the matching entry has `Modified == true`
- Non-matching entries still have `Modified == false`

---

## Test UT-TUI-031: renderItem appends (not-saved) when Modified is true

- **Type:** Unit
- **Task:** Task 4
- **Priority:** P0

### Setup

1. Create a `ListPanel` with entries
2. Set panel width (e.g., `lp.SetSize(60, 20)`)
3. Call `lp.UpdateItemValue("field_name", "new_value")` to set `Modified = true`

### Steps

1. Call `lp.View()` to render the list
2. Inspect the rendered output for the modified field

### Expected Result

- The rendered line for the modified field contains the substring `(not-saved)`
- The rendered line for unmodified fields does NOT contain `(not-saved)`

---

## Test UT-TUI-032: ClearAllModified resets all Modified flags

- **Type:** Unit
- **Task:** Task 3
- **Priority:** P0

### Setup

1. Create a `ListPanel` with multiple entries
2. Call `UpdateItemValue` on two or more entries to set `Modified = true`

### Steps

1. Verify at least two entries have `Modified == true`
2. Call `lp.ClearAllModified()`
3. Iterate all entries and check `Modified`

### Expected Result

- After `ClearAllModified()`, all entries have `Modified == false`

---

## Test UT-TUI-033: Saved flag cleared after commit

- **Type:** Unit
- **Task:** Task 1, Task 6
- **Priority:** P0

### Setup

1. Create model with a non-sensitive field
2. Set `model.saved = true` (simulating a prior Ctrl+S)
3. Enter editing mode

### Steps

1. **Sub-test A (Esc commit):** Send `tea.KeyMsg{Type: tea.KeyEsc}` → verify `m.saved == false`
2. **Sub-test B (Enter commit on non-list):** Set `model.saved = true` again, enter editing mode, send `tea.KeyMsg{Type: tea.KeyEnter}` → verify `m.saved == false`

### Expected Result

- In both sub-tests, `m.saved == false` after the commit
- `m.err == nil` after the commit
- `m.state == StateBrowsing` after the commit

---

## Test UT-TUI-034: Help bar shows enter confirm for non-list editing

- **Type:** Unit
- **Task:** Task 9
- **Priority:** P1

### Setup

1. Create `DefaultKeyMap()`
2. Optionally: create a model with a string/enum/bool field in editing state

### Steps

1. Call `keys.ShortHelp(StateEditing, "string")` (or equivalent field-type-aware call)
2. Inspect the returned key bindings

### Expected Result

- The returned bindings include a binding with help key `"enter"` and help desc containing `"confirm"`
- The returned bindings also include the `Escape` binding

---

## Test UT-TUI-035: Help bar omits enter confirm for list editing

- **Type:** Unit
- **Task:** Task 9
- **Priority:** P1

### Setup

1. Create `DefaultKeyMap()`

### Steps

1. Call `keys.ShortHelp(StateEditing, "list")` (or equivalent field-type-aware call)
2. Inspect the returned key bindings

### Expected Result

- The returned bindings do NOT include a binding with help key `"enter"` and help desc `"confirm"`
- The returned bindings include the `Escape` binding with `"done"` description
- The returned bindings include `Save` and `Quit`

---

## Test UT-TUI-008+ (Extended): Esc commit clears saved and marks Modified

- **Type:** Unit (extension of existing)
- **Task:** Task 1, Task 3, Task 6
- **Priority:** P0

### Setup

Same as existing UT-TUI-008:
1. Create model with enum field, set `model.state = StateEditing`
2. Additionally set `model.saved = true` before the Esc

### Steps

1. Send `tea.KeyMsg{Type: tea.KeyEsc}`
2. Check `m.state`, `m.saved`, and the selected item's `Modified` flag

### Expected Result

- `m.state == StateBrowsing` (existing assertion)
- `m.saved == false` (new assertion — stale banner cleared)
- The committed field's `Modified` flag is `true` in the list panel entry (new assertion — dirty tracking)

---

## Test UT-TUI-036: CurrentFieldType accessor returns correct type

- **Type:** Unit
- **Task:** Task 8
- **Priority:** P1

### Setup

1. Create a `DetailPanel` via `NewDetailPanel()`

### Steps

1. Call `dp.SetField(field, value)` for each type: `"string"`, `"bool"`, `"enum"`, `"list"`
2. After each `SetField`, call `dp.CurrentFieldType()`

### Expected Result

- Returns `"string"` after setting a string field
- Returns `"bool"` after setting a bool field
- Returns `"enum"` after setting an enum field
- Returns `"list"` after setting a list field

---

## Test UT-TUI-037: CurrentFieldType returns empty string for nil field

- **Type:** Unit
- **Task:** Task 8
- **Priority:** P1

### Setup

1. Create a `DetailPanel` via `NewDetailPanel()` — do NOT call `SetField()`

### Steps

1. Call `dp.CurrentFieldType()`

### Expected Result

- Returns `""` (empty string)

---

## Test UT-TUI-038: (not-saved) not shown when Modified is false

- **Type:** Unit
- **Task:** Task 4
- **Priority:** P1

### Setup

1. Create a `ListPanel` with entries (do NOT call `UpdateItemValue`)
2. Set panel size

### Steps

1. Call `lp.View()` to render the list

### Expected Result

- No line in the rendered output contains the substring `(not-saved)`

---

## Test UT-TUI-039: Narrow terminal with (not-saved) does not panic

- **Type:** Unit
- **Task:** Task 4
- **Priority:** P1

### Setup

1. Create a `ListPanel` with entries
2. Set panel width to narrow value: `lp.SetSize(30, 10)`
3. Call `UpdateItemValue` to set `Modified = true` on one entry

### Steps

1. Call `lp.View()` — must not panic

### Expected Result

- Function returns a non-empty string without panicking
- The `(not-saved)` suffix may be truncated at narrow widths but the rendering is safe

---

## Test UT-TUI-040: Post-save reload preserves cursor by field name

- **Type:** Unit
- **Task:** Task 5
- **Priority:** P0

### Setup

1. Create a model with multiple fields
2. Navigate cursor to a field that is NOT the first entry (e.g., second or third field)
3. Record the selected field name
4. Write config to a temp file so `SaveConfig` + `LoadConfig` round-trip works

### Steps

1. Send `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}` with `Alt: true` — or directly call the ctrl+s handler
2. After save+reload, check the selected item's field name

### Expected Result

- `m.listPanel.SelectedItem().Field.Name` matches the field name from before the save
- `m.saved == true`
- `m.err == nil`

---

## Test UT-TUI-041: Modified flags cleared after successful save

- **Type:** Unit
- **Task:** Task 7
- **Priority:** P0

### Setup

1. Create a model with a field
2. Enter editing, commit a change (so the field is marked `Modified`)
3. Write config to a temp file

### Steps

1. Trigger `ctrl+s` save
2. Iterate all list entries and check `Modified`

### Expected Result

- All entries have `Modified == false` after the save
- `m.saved == true`
- No `(not-saved)` text appears in `lp.View()`

---

## Test UT-TUI-042: Save failure does not clear Modified flags

- **Type:** Unit
- **Task:** Task 7
- **Priority:** P1

### Setup

1. Create a model with `configPath` pointing to an invalid/unwritable path (e.g., `/nonexistent/path/config.json`)
2. Mark a field as modified by committing a change

### Steps

1. Trigger `ctrl+s` save (which will fail because the path is unwritable)
2. Check the field's `Modified` flag

### Expected Result

- `m.err != nil` (save failed)
- `m.saved == false` (save did not succeed)
- The modified field still has `Modified == true`

---

## Regression Tests

All 25 existing tests (UT-TUI-001 through UT-TUI-025) must continue to pass unchanged. The following are particularly relevant for regression:

| Test ID | Risk | Why |
|---------|------|-----|
| UT-TUI-007 | Enter in Browsing → Editing transition | Must not be affected by Enter-commit logic (that only applies in StateEditing) |
| UT-TUI-008 | Esc in Editing → Browsing transition | Core commit path is being refactored into helper; must still work |
| UT-TUI-012 | formatValueCompact correctness | Adding `(not-saved)` suffix must not break the base formatting function |
| UT-TUI-018 | View at 80x24 | `(not-saved)` suffix must not cause overflow at standard terminal size |

## Execution

```bash
# Run all TUI tests
go test ./internal/tui/... -v

# Run only new tests (by name pattern)
go test ./internal/tui/... -v -run "TestEnter|TestModified|TestClearAll|TestSaved|TestHelp|TestCurrentFieldType|TestNotSaved|TestNarrow|TestPostSave"

# Run with race detector
go test ./internal/tui/... -race -v
```
