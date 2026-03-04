# Task Breakdown: WI-0006 — User Experience Saving and Viewing Configuration Values

> **Note:** This file should be at `plan/02-task-breakdown.md` once the `plan/` directory is created.

- **Workitem:** WI-0006-user-experience-saving
- **Action Plan:** [01-action-plan.md](01-action-plan.md)
- **Test Plan:** [03-test-plan.md](03-test-plan.md)

## Summary

This task breakdown decomposes the four UX bug fixes into 10 implementation tasks ordered by dependency. All changes are confined to `internal/tui/`. No new Go dependencies are required.

### Bug → Task Mapping

| Bug | Description | Fixing Tasks |
|-----|-------------|--------------|
| Bug 1 | Enter doesn't commit & exit editing | Task 8, Task 1, Task 2 |
| Bug 2 | No "(not-saved)" indicator | Task 3, Task 4 |
| Bug 3 | Config not re-read after Ctrl+S | Task 5 |
| Bug 4 | "✓ Saved" banner shows stale info | Task 6, Task 7 |

### Dependency Graph

```
Task 8  (CurrentFieldType accessor) ──────────┐
Task 3  (dirty tracking) ─► Task 4 (render)   │
                                               ▼
                                     Task 1  (extract commit helper)
                                       ├──► Task 2  (Enter intercept)
                                       └──► Task 6  (clear saved — in Task 1)
Task 5  (post-save reload)
  └──► Task 7  (clear modified after save — depends on Task 3 + Task 5)
Task 9  (help bar — depends on Task 8)
Task 10 (tests — depends on all above)
```

---

## Task 8: Expose CurrentFieldType accessor on DetailPanel

- **Status:** Not Started
- **Complexity:** XS (trivial — 5-line method)
- **Dependencies:** None
- **Related ADRs:** [ADR-0003](../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004](../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md)

### Description

Add a `CurrentFieldType() string` method to `DetailPanel` in `internal/tui/detail_panel.go`. This method returns `d.field.Type` when a field is loaded, or `""` when `d.field` is nil.

```go
func (d *DetailPanel) CurrentFieldType() string {
    if d.field == nil {
        return ""
    }
    return d.field.Type
}
```

This accessor is needed by Task 2 (Enter intercept logic) and Task 9 (context-aware help bar).

### Acceptance Criteria

- [ ] `DetailPanel` has a public `CurrentFieldType() string` method
- [ ] Returns `""` when no field is set (`d.field == nil`)
- [ ] Returns the field's `.Type` string (`"string"`, `"bool"`, `"enum"`, `"list"`) when a field is set
- [ ] No change to `DetailPanel.View()` or `DetailPanel.Update()` behaviour

### Test Coverage

- Unit test: `CurrentFieldType()` returns `""` on a fresh `DetailPanel` with no field set
- Unit test: `CurrentFieldType()` returns correct type after `SetField()` is called for each type (`string`, `bool`, `enum`, `list`)

---

## Task 3: Add dirty tracking to ConfigItem and ListPanel

- **Status:** Not Started
- **Complexity:** S (struct change + 2 method changes)
- **Dependencies:** None
- **Related ADRs:** [ADR-0003](../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004](../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md) — Decision 17

### Description

Modify `internal/tui/list_item.go` to support per-field dirty state tracking:

1. Add `Modified bool` field to the `ConfigItem` struct (after `Value any`)
2. In `UpdateItemValue()`, set `l.entries[i].item.Modified = true` when the value is updated
3. Add a new `ClearAllModified()` method on `ListPanel` that iterates all entries and sets `Modified = false`

This implements Decision 17 from the Decision Log: "Track per-field dirty state via `Modified` flag on `ConfigItem`".

### Acceptance Criteria

- [ ] `ConfigItem` struct has a `Modified bool` field
- [ ] `Modified` is `false` by default when entries are built via `buildEntries()`
- [ ] `UpdateItemValue()` sets `Modified = true` on the matching entry
- [ ] `ClearAllModified()` method exists on `ListPanel` and resets all `Modified` flags to `false`
- [ ] Existing `SelectedItem()` returns a copy that includes the `Modified` field

### Test Coverage

- Unit test (UT-TUI-030): `ConfigItem.Modified` is `false` by default and set to `true` after `UpdateItemValue()`
- Unit test (UT-TUI-032): `ClearAllModified()` resets all `Modified` flags to `false`

---

## Task 4: Render "(not-saved)" indicator in list panel

- **Status:** Not Started
- **Complexity:** S (rendering logic change)
- **Dependencies:** Task 3
- **Related ADRs:** [ADR-0003](../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004](../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md) — Decision 17

### Description

Update `renderItem()` and/or `formatValueCompact()` in `internal/tui/list_item.go` to display a `(not-saved)` suffix when `item.Modified` is `true`.

Implementation details:
1. In `renderItem()`, when `item.Modified` is `true`, reduce `valWidth` by 12 characters (`len(" (not-saved)")`) before calling `formatValueCompact()`, so the actual value still fits alongside the suffix
2. After `formatValueCompact()` returns, append ` (not-saved)` to the value string when `Modified` is `true`
3. Ensure `valWidth` floor check (`if valWidth < 3`) still applies after the reduction
4. Test at 80-column terminal width to ensure no truncation issues

### Acceptance Criteria

- [ ] When `item.Modified == true`, the list row displays the value followed by ` (not-saved)`
- [ ] When `item.Modified == false`, the list row displays the value without any suffix (no change from current behaviour)
- [ ] The `(not-saved)` suffix does not cause the actual value to be truncated — `valWidth` is reduced to accommodate the suffix
- [ ] At 80-column terminal width, items with `(not-saved)` render without visual overflow or clipping
- [ ] `formatValueCompact()` function signature is unchanged (the suffix is appended in `renderItem()`, not in `formatValueCompact()`)

### Test Coverage

- Unit test (UT-TUI-031): `renderItem` appends `(not-saved)` when `Modified` is `true`
- Unit test: `renderItem` does NOT append `(not-saved)` when `Modified` is `false`
- Unit test: at narrow width (e.g., `width=30`), `(not-saved)` suffix and value both render without panic

---

## Task 1: Extract commitAndReturnToBrowsing helper function

- **Status:** Not Started
- **Complexity:** S (refactor — extract existing code into method)
- **Dependencies:** Task 8
- **Related ADRs:** [ADR-0003](../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004](../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md) — Decision 19

### Description

Extract the existing `Esc` commit logic (lines 203–211 of `internal/tui/model.go`) into a shared `commitAndReturnToBrowsing()` method on `*Model`. This method performs:

1. `newValue := m.detailPanel.StopEditing()`
2. Gets the selected item via `m.listPanel.SelectedItem()`
3. `m.cfg.Set(item.Field.Name, newValue)` — updates in-memory config
4. `m.listPanel.UpdateItemValue(item.Field.Name, newValue)` — updates list display (this also marks the field `Modified` per Task 3)
5. `m.saved = false` — clears stale saved indicator (Task 6 requirement)
6. `m.err = nil` — clears any stale error
7. `m.state = StateBrowsing` — returns to browsing state

The existing `Esc` branch in `handleKeyPress` is then simplified to call `m.commitAndReturnToBrowsing()`.

### Acceptance Criteria

- [ ] A `commitAndReturnToBrowsing()` method exists on `*Model`
- [ ] The `Esc` key handling in `StateEditing` calls `commitAndReturnToBrowsing()` instead of inline logic
- [ ] Existing behaviour is preserved: `Esc` in editing mode still commits the value, updates the list, and returns to `StateBrowsing`
- [ ] `m.saved` is set to `false` inside `commitAndReturnToBrowsing()` (Task 6 requirement)
- [ ] `m.err` is set to `nil` inside `commitAndReturnToBrowsing()`
- [ ] No change in user-visible behaviour for the `Esc` path (regression safety)

### Test Coverage

- Existing test UT-TUI-008 must continue to pass (Esc transitions from Editing to Browsing)
- Unit test (UT-TUI-033): After calling `commitAndReturnToBrowsing()`, `m.saved == false`
- Unit test: After calling `commitAndReturnToBrowsing()`, `m.err == nil`

---

## Task 6: Clear saved flag on new edits

- **Status:** Not Started
- **Complexity:** XS (single line — integrated into Task 1)
- **Dependencies:** Task 1
- **Related ADRs:** [ADR-0003](../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004](../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md) — Decision 19

### Description

This task is logically integrated into Task 1. The `commitAndReturnToBrowsing()` helper sets `m.saved = false` and `m.err = nil`. This ensures the "✓ Saved" banner disappears the moment the user commits a new change.

The implementation is a single line (`m.saved = false`) inside the helper created in Task 1. This task exists as a separate logical concern for traceability to Bug 4 and Decision 19.

### Acceptance Criteria

- [ ] When `m.saved == true` (i.e., after a previous `Ctrl+S`), committing a new change (via Enter or Esc) sets `m.saved = false`
- [ ] The "✓ Saved" banner in `View()` no longer appears after a new change is committed
- [ ] `m.err` is cleared to `nil` after a commit

### Test Coverage

- Unit test (UT-TUI-033): `m.saved` is `false` after committing a new edit via Enter or Esc, even when it was `true` before the commit
- Extends UT-TUI-008: Verify that `m.saved` is cleared after Esc-commit

---

## Task 2: Intercept Enter in StateEditing for non-list fields

- **Status:** Not Started
- **Complexity:** S (conditional branch + field-type check)
- **Dependencies:** Task 1, Task 8
- **Related ADRs:** [ADR-0003](../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004](../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md)

### Description

In the `StateEditing` branch of `handleKeyPress` in `internal/tui/model.go`, add an `Enter` handler **before** the default forwarding to `detailPanel.Update()`:

```go
case StateEditing:
    if k == "esc" || (k == "enter" && m.detailPanel.CurrentFieldType() != "list") {
        m.commitAndReturnToBrowsing()
        return m, nil
    }
    // All other keys go to detail panel
    return m, m.detailPanel.Update(msg)
```

Key behavioural rules:
- For `string`, `bool`, `enum` field types: `Enter` commits the value and returns to `StateBrowsing`
- For `list` field type: `Enter` is forwarded to the `textarea` widget (inserts a newline) — `Esc` remains the only commit key for lists
- `Esc` continues to commit for ALL field types (including `list`)

Uses `m.detailPanel.CurrentFieldType()` (from Task 8) to determine the field type.

### Acceptance Criteria

- [ ] `Enter` in `StateEditing` on `enum` fields commits the value and returns to `StateBrowsing`
- [ ] `Enter` in `StateEditing` on `string` fields commits the value and returns to `StateBrowsing`
- [ ] `Enter` in `StateEditing` on `bool` fields commits the value and returns to `StateBrowsing`
- [ ] `Enter` in `StateEditing` on `list` fields does NOT exit editing — the key is forwarded to `textarea`
- [ ] `Esc` still commits and exits for ALL field types (including `list`)
- [ ] The committed value is correctly stored in `m.cfg` and reflected in the list panel

### Test Coverage

- Unit test (UT-TUI-026): `Enter` on `enum` field commits value and returns to `StateBrowsing`
- Unit test (UT-TUI-027): `Enter` on `string` field commits value and returns to `StateBrowsing`
- Unit test (UT-TUI-028): `Enter` on `bool` field commits value and returns to `StateBrowsing`
- Unit test (UT-TUI-029): `Enter` on `list` field does NOT exit editing (stays in `StateEditing`)

---

## Task 5: Post-save reload from disk

- **Status:** Not Started
- **Complexity:** M (reload + rebuild + cursor restoration)
- **Dependencies:** None (can be developed in parallel with Tasks 1-4, but must be integrated after Task 3 for Task 7)
- **Related ADRs:** [ADR-0003](../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004](../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md) — Decision 18

### Description

In the `ctrl+s` handler in `internal/tui/model.go`, after a successful `SaveConfig`, reload the config from disk to verify round-trip integrity:

```go
if k == "ctrl+s" {
    if err := config.SaveConfig(m.configPath, m.cfg); err != nil {
        m.err = err
        slog.Error("save failed", "error", err)
    } else {
        // Post-save reload
        reloaded, reloadErr := config.LoadConfig(m.configPath)
        if reloadErr != nil {
            m.err = fmt.Errorf("saved but reload failed: %w", reloadErr)
            slog.Error("post-save reload failed", "error", reloadErr)
        } else {
            // Save cursor position by field name
            cursorFieldName := ""
            if item := m.listPanel.SelectedItem(); item != nil {
                cursorFieldName = item.Field.Name
            }

            // Replace config and rebuild
            m.cfg = reloaded
            entries := buildEntries(m.cfg, m.schema)
            m.listPanel = NewListPanel(entries)

            // Restore cursor by name
            m.restoreCursorByName(cursorFieldName)

            m.err = nil
        }
        m.saved = true
        m.syncDetailPanel()
    }
    return m, nil
}
```

A helper method `restoreCursorByName(name string)` on `*Model` (or on `*ListPanel`) scans entries by field name and sets the cursor position.

**Failure path**: If `LoadConfig` fails after a successful `SaveConfig`, set `m.err` to a descriptive error, keep the in-memory state, and still set `m.saved = true` (because the save itself succeeded).

### Acceptance Criteria

- [ ] After `Ctrl+S`, `config.LoadConfig(m.configPath)` is called
- [ ] On successful reload: `m.cfg` is replaced with the reloaded config
- [ ] On successful reload: list entries are rebuilt from the reloaded config via `buildEntries()`
- [ ] On successful reload: cursor is restored to the same field (by name, not index)
- [ ] On successful reload: detail panel is re-synced via `syncDetailPanel()`
- [ ] On successful reload: `m.saved = true` and `m.err = nil`
- [ ] On reload failure: `m.err` is set to `"saved but reload failed: <error>"`
- [ ] On reload failure: `m.saved = true` (save succeeded), in-memory state is preserved
- [ ] On save failure: existing behaviour preserved — `m.err` is set, `m.saved` is NOT set to `true`

### Test Coverage

- Unit test: After a successful save+reload cycle, `m.cfg` reflects the persisted state
- Unit test: Cursor position is preserved (by field name) after save+reload
- Unit test: Detail panel is synced to the correct field after save+reload
- Integration-style test: Save followed by reload produces a round-trip-consistent config
- Unit test: If post-save reload fails, `m.saved == true` and `m.err` contains "reload failed"

---

## Task 7: Clear all modified flags after save

- **Status:** Not Started
- **Complexity:** XS (single method call after save)
- **Dependencies:** Task 3, Task 5
- **Related ADRs:** [ADR-0003](../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004](../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md) — Decision 17

### Description

After a successful save-and-reload (Task 5), call `m.listPanel.ClearAllModified()` (from Task 3) to remove all `(not-saved)` indicators. If reload fails but save succeeded, still clear modified flags (the data is on disk).

The call should be placed after `m.saved = true` in the `ctrl+s` handler, so it runs regardless of whether reload succeeded or failed (as long as save succeeded).

### Acceptance Criteria

- [ ] After a successful `Ctrl+S`, all `(not-saved)` indicators are removed from the list panel
- [ ] `ClearAllModified()` is called after successful save, even if reload fails
- [ ] `ClearAllModified()` is NOT called if the save itself fails
- [ ] After clearing, subsequent `View()` calls render items without `(not-saved)` suffix

### Test Coverage

- Unit test: After save, `listPanel` entries all have `Modified == false`
- Unit test: If save fails, `Modified` flags remain unchanged

---

## Task 9: Update help bar for editing state

- **Status:** Not Started
- **Complexity:** S (keybinding + conditional rendering)
- **Dependencies:** Task 8
- **Related ADRs:** [ADR-0003](../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004](../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md)

### Description

Update `internal/tui/keys.go` and `internal/tui/model.go` to show context-aware help text in the editing state:

1. **keys.go**: Add a `Confirm key.Binding` to `KeyMap` with `key.WithKeys("enter")` and `key.WithHelp("enter", "confirm")`

2. **model.go — `ShortHelp` method**: Update to accept field type information. Two approaches (choose one):
   - Option A: Change signature to `ShortHelp(state State, fieldType string)`
   - Option B: Add a `ShortHelpEditing(fieldType string) []key.Binding` variant

   For `StateEditing`:
   - Non-`list` fields: return `[Confirm, Escape, Save, Quit]` → renders as `enter • confirm  •  esc • done  •  ctrl+s • save  •  ctrl+c • quit`
   - `list` fields: return `[Escape, Save, Quit]` → renders as `esc • done  •  ctrl+s • save  •  ctrl+c • quit` (Enter = newline is implicit for textarea)

3. **model.go — `View` method**: Pass the current field type to `ShortHelp` when constructing help keys. Use `m.detailPanel.CurrentFieldType()` (from Task 8).

### Acceptance Criteria

- [ ] A `Confirm key.Binding` exists in `KeyMap` with help text `"enter" / "confirm"`
- [ ] In `StateEditing` for non-`list` fields, help bar shows `enter • confirm` alongside `esc • done`
- [ ] In `StateEditing` for `list` fields, help bar shows `esc • done` but NOT `enter • confirm`
- [ ] In `StateBrowsing`, help bar is unchanged from current behaviour
- [ ] Help bar renders correctly at 80-column width without truncation

### Test Coverage

- Unit test (UT-TUI-034): Help bar includes `enter • confirm` for non-list fields in editing state
- Unit test (UT-TUI-035): Help bar does NOT include `enter • confirm` for list fields in editing state
- Unit test: Help bar in browsing state is unchanged

---

## Task 10: Add and update unit tests

- **Status:** Not Started
- **Complexity:** M (10+ new test functions, 1 test update)
- **Dependencies:** All previous tasks (1-9)
- **Related ADRs:** [ADR-0002](../../architecture/ADR/ADR-0002-go-charm-tui-stack.md) (Decision 3 — use `go test`), [ADR-0003](../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004](../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md) — Decisions 17, 18, 19

### Description

Add tests to `internal/tui/tui_test.go` for all four fixed behaviours. Each test should follow the existing test patterns (create `config.NewConfig()`, build schema, construct model, simulate key presses, assert state).

**New tests:**

| Test ID | Description | Verifying Task |
|---------|-------------|----------------|
| UT-TUI-026 | `Enter` in `StateEditing` on `enum` field commits and returns to `StateBrowsing` | Task 2 |
| UT-TUI-027 | `Enter` in `StateEditing` on `string` field commits and returns to `StateBrowsing` | Task 2 |
| UT-TUI-028 | `Enter` in `StateEditing` on `bool` field commits and returns to `StateBrowsing` | Task 2 |
| UT-TUI-029 | `Enter` in `StateEditing` on `list` field does NOT exit editing | Task 2 |
| UT-TUI-030 | `ConfigItem.Modified` is `false` by default, `true` after `UpdateItemValue` | Task 3 |
| UT-TUI-031 | `renderItem` appends `(not-saved)` when `Modified` is `true` | Task 4 |
| UT-TUI-032 | `ClearAllModified()` resets all `Modified` flags to `false` | Task 3 |
| UT-TUI-033 | `m.saved` is cleared (`false`) after committing a new edit via Enter or Esc | Task 1/6 |
| UT-TUI-034 | Help bar shows `enter • confirm` for non-list fields in editing state | Task 9 |
| UT-TUI-035 | Help bar shows `esc • done` (no Enter confirm) for list fields in editing state | Task 9 |

**Updated test:**

| Test ID | Change | Verifying Task |
|---------|--------|----------------|
| UT-TUI-008 | Extend to verify `m.saved` is cleared and `Modified` is set after Esc commit | Tasks 1, 3, 6 |

### Acceptance Criteria

- [ ] All 10 new test functions are added to `internal/tui/tui_test.go`
- [ ] UT-TUI-008 is extended with additional assertions
- [ ] All tests pass with `go test ./internal/tui/...`
- [ ] No existing tests are broken
- [ ] Tests follow the naming convention `TestXxx` and include the UT-TUI-NNN identifier in a comment
- [ ] Each test is self-contained (creates its own model, schema, config)

### Test Coverage

- Self-referential: This task IS the test coverage for all other tasks
- Regression: All existing 25 tests (UT-TUI-001 through UT-TUI-025) must continue to pass
- Coverage target: All four bug-fix behaviours have at least one dedicated test

---

## Implementation Checklist

| Order | Task | Complexity | Dependencies | Status |
|-------|------|-----------|--------------|--------|
| 1 | Task 8: CurrentFieldType accessor | XS | None | Not Started |
| 2 | Task 3: Dirty tracking on ConfigItem | S | None | Not Started |
| 3 | Task 4: Render "(not-saved)" indicator | S | Task 3 | Not Started |
| 4 | Task 1: Extract commit helper | S | Task 8 | Not Started |
| 5 | Task 6: Clear saved flag (in Task 1) | XS | Task 1 | Not Started |
| 6 | Task 2: Intercept Enter for non-list | S | Task 1, Task 8 | Not Started |
| 7 | Task 5: Post-save reload from disk | M | None | Not Started |
| 8 | Task 7: Clear modified after save | XS | Task 3, Task 5 | Not Started |
| 9 | Task 9: Update help bar | S | Task 8 | Not Started |
| 10 | Task 10: Add/update tests | M | All above | Not Started |
