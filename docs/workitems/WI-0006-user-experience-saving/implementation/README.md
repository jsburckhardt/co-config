# Implementation Notes: WI-0006 — User Experience Saving

## Task 8: Expose CurrentFieldType accessor on DetailPanel

- **Status:** Complete
- **Files Changed:** `internal/tui/detail_panel.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 32
- **Tests Failed:** 0

### Changes Summary

Added `CurrentFieldType() string` method to `DetailPanel` in `detail_panel.go`. The method returns `d.field.Type` when a field is loaded, or `""` when `d.field` is nil. Placed after `SetSize()` method and before `View()`.

### Test Results

- UT-TUI-036 (CurrentFieldType accessor returns correct type): PASS — tests all four types (string, bool, enum, list)
- UT-TUI-037 (CurrentFieldType returns empty string for nil field): PASS

### Notes

This is a prerequisite for Task 2 (Enter intercept) and Task 9 (context-aware help bar).

---

## Task 3: Add dirty tracking to ConfigItem and ListPanel

- **Status:** Complete
- **Files Changed:** `internal/tui/list_item.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 32
- **Tests Failed:** 0

### Changes Summary

1. Added `Modified bool` field to `ConfigItem` struct
2. Modified `UpdateItemValue()` to set `Modified = true` on the matching entry
3. Added `ClearAllModified()` method to `ListPanel` that resets all `Modified` flags

### Test Results

- UT-TUI-030 (Modified flag default and after UpdateItemValue): PASS
- UT-TUI-032 (ClearAllModified resets all Modified flags): PASS
- All 25 existing tests continue to pass (no regression)

### Notes

`Modified` defaults to `false` (Go zero value) so existing code that creates `ConfigItem` structs without setting `Modified` behaves correctly.

---

## Task 4: Render "(not-saved)" indicator in list items

- **Status:** Complete
- **Files Changed:** `internal/tui/list_item.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 32
- **Tests Failed:** 0

### Changes Summary

Modified `renderItem()` in `list_item.go` to:
1. Reduce `valWidth` by 12 when `item.Modified` is true to make room for the " (not-saved)" suffix
2. Append " (not-saved)" to the value string when `item.Modified` is true
3. The suffix is only added for non-sensitive fields (sensitive fields show 🔒 regardless)
4. The `valWidth` floor of 3 prevents negative widths on narrow terminals

### Test Results

- UT-TUI-031 (renderItem appends (not-saved) when Modified): PASS
- UT-TUI-038 ((not-saved) not shown when Modified is false): PASS
- UT-TUI-039 (Narrow terminal with (not-saved) does not panic): PASS
- UT-TUI-018 (View at 80x24 regression): PASS

### Notes

At very narrow widths (< 30), the "(not-saved)" suffix may be truncated by the existing line-truncation logic (`line[:l.width-3] + "…"`), which is safe and expected behavior.

---

## Task 1 + Task 6: Extract commitAndReturnToBrowsing helper + clear saved flag

- **Status:** Complete
- **Files Changed:** `internal/tui/model.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 40
- **Tests Failed:** 0

### Changes Summary

Extracted the inline Esc commit logic from `handleKeyPress` into a new `commitAndReturnToBrowsing()` method on `*Model`. The method:
1. Calls `m.detailPanel.StopEditing()` to get the new value
2. Gets the selected item and updates config + list panel
3. Sets `m.saved = false` (clears stale "✓ Saved" banner — Task 6)
4. Sets `m.err = nil` (clears stale errors)
5. Sets `m.state = StateBrowsing`

### Test Results

- UT-TUI-008 (existing Esc transition): PASS (regression)
- UT-TUI-033 (Saved flag cleared after commit — both Esc and Enter sub-tests): PASS

### Notes

Task 6 is logically integrated into Task 1. The `m.saved = false` line ensures the "✓ Saved" banner disappears when a new edit is committed.

---

## Task 2: Intercept Enter in StateEditing for non-list fields

- **Status:** Complete
- **Files Changed:** `internal/tui/model.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 40
- **Tests Failed:** 0

### Changes Summary

Updated the `StateEditing` case in `handleKeyPress` to handle both Esc and Enter (for non-list fields) as commit triggers:
```go
if k == "esc" || (k == "enter" && m.detailPanel.CurrentFieldType() != "list") {
    m.commitAndReturnToBrowsing()
    return m, nil
}
```

- For `string`, `bool`, `enum` fields: Enter commits and returns to browsing
- For `list` fields: Enter is forwarded to textarea (inserts newline)
- Esc always commits and returns to browsing (all field types)

### Test Results

- UT-TUI-026 (Enter commits enum field): PASS
- UT-TUI-027 (Enter commits string field): PASS
- UT-TUI-028 (Enter commits bool field): PASS
- UT-TUI-029 (Enter on list field stays in Editing): PASS

### Notes

Uses `m.detailPanel.CurrentFieldType()` (from Task 8) to determine field type.

---

## Task 5 + Task 7: Post-save reload from disk + clear modified flags

- **Status:** Complete
- **Files Changed:** `internal/tui/model.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 40
- **Tests Failed:** 0

### Changes Summary

Updated the `ctrl+s` handler in `handleKeyPress` to:
1. After successful `SaveConfig`, call `config.LoadConfig` to reload from disk (Task 5)
2. On successful reload: save cursor position by field name, replace `m.cfg`, rebuild list entries via `buildEntries()`, restore cursor position, sync detail panel
3. On reload failure: set `m.err` with descriptive message, keep in-memory state
4. After successful save (regardless of reload result): call `m.listPanel.ClearAllModified()` (Task 7)
5. On save failure: do NOT clear modified flags

Added three helper methods:
- `listPanelWidth()` — returns content width for list panel
- `listPanelHeight()` — returns content height for list panel
- `selectFieldByName(name string)` — restores cursor to a field by name

Added `"fmt"` to imports for `fmt.Errorf`.

### Test Results

- UT-TUI-040 (Post-save reload preserves cursor by field name): PASS
- UT-TUI-041 (Modified flags cleared after successful save): PASS
- UT-TUI-042 (Save failure does not clear Modified flags): PASS

### Notes

The reload-from-disk pattern follows CC-0004 Integration Guidelines: "After `SaveConfig()` succeeds, call `LoadConfig()` and rebuild list entries from the reloaded config to verify the round-trip."

---

## Task 9: Context-aware help bar

- **Status:** Complete
- **Files Changed:** `internal/tui/keys.go`, `internal/tui/model.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 40
- **Tests Failed:** 0

### Changes Summary

1. Added `Confirm key.Binding` to `KeyMap` struct in `keys.go` with help text `"enter" / "confirm"`
2. Updated `ShortHelp` signature to `ShortHelp(state State, fieldType string)` in `model.go`
3. In `StateEditing` for non-list fields: returns `[Confirm, Escape, Save, Quit]`
4. In `StateEditing` for list fields: returns `[Escape, Save, Quit]` (no Enter confirm — Enter inserts newlines)
5. Updated call site in `View()` to pass `m.detailPanel.CurrentFieldType()`

### Test Results

- UT-TUI-034 (Help bar shows enter confirm for non-list editing): PASS
- UT-TUI-035 (Help bar omits enter confirm for list editing): PASS

### Notes

The `Confirm` binding uses the same key (`"enter"`) as the `Enter` binding but with different help text (`"confirm"` vs `"edit"`), so the help bar shows the appropriate context.

---

## Task 10: Add remaining unit tests

- **Status:** Complete
- **Files Changed:** `internal/tui/tui_test.go`
- **Tests Passed:** 40
- **Tests Failed:** 0

### Changes Summary

Added 10 new test functions covering all test plan entries not already implemented:
- UT-TUI-026: `TestEnterCommitsEnumField`
- UT-TUI-027: `TestEnterCommitsStringField`
- UT-TUI-028: `TestEnterCommitsBoolField`
- UT-TUI-029: `TestEnterOnListFieldStaysEditing`
- UT-TUI-033: `TestSavedFlagClearedAfterCommit` (with Esc and Enter sub-tests)
- UT-TUI-034: `TestHelpBarEnterConfirmNonList`
- UT-TUI-035: `TestHelpBarNoEnterConfirmList`
- UT-TUI-040: `TestPostSaveReloadPreservesCursor`
- UT-TUI-041: `TestModifiedFlagsClearedAfterSave`
- UT-TUI-042: `TestSaveFailureKeepsModifiedFlags`

### Test Results

All 40 tests pass (25 existing + 7 from Tasks 3/4/8 + 10 new = 42 test functions, 40 top-level test runs with sub-tests).

### Notes

Tests 040-042 use `t.TempDir()` for temp file round-trip testing. Test 042 uses an unwritable path to verify save failure behavior.
