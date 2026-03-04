# Action Plan: User Experience — Saving and Viewing Configuration Values

> **Note:** This file should be at `plan/01-action-plan.md` once the `plan/` directory is created.

## Feature
- **ID:** WI-0006-user-experience-saving
- **Research Brief:** [00-research.md](research/00-research.md)

## ADRs Created
None — all changes are within the existing two-panel layout pattern established by [ADR-0003](../../architecture/ADR/ADR-0003-two-panel-tui-layout.md). The state machine gains a new transition (`Enter` → commit for non-`list` types) but the Browsing/Editing duality is unchanged.

## Core-Components Updated
**CORE-COMPONENT-0004 (Configuration Management)** was **updated** to document:
1. Per-field dirty tracking via `Modified` flag on `ConfigItem`
2. Post-save reload requirement (re-read from disk after `SaveConfig` via `LoadConfig`)
3. Rule that "✓ Saved" indicator must be cleared on new in-memory commits

See: [CORE-COMPONENT-0004-configuration-management.md](../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md)

## Core-Components Created
None.

## Applicable Existing Decisions
| Source | Decision |
|--------|----------|
| ADR-0003 | Two-panel TUI layout with Browsing/Editing state machine |
| CC-0004 (Decision 6) | Auto-detect config schema by running `copilot help config` at startup |
| CC-0004 (Decision 7) | Preserve unknown config fields on round-trip (no data loss) |
| CC-0004 (Decision 17) | Track per-field dirty state via `Modified` flag on `ConfigItem` |
| CC-0004 (Decision 18) | Re-read config from disk after every successful save |
| CC-0004 (Decision 19) | Clear "✓ Saved" indicator when any new in-memory change is committed |

## Constraints & Boundaries

- **Scope boundary**: All changes confined to `internal/tui/`. No changes to `internal/config/`, `internal/copilot/`, `internal/sensitive/`, or `cmd/`.
- **No new dependencies**: All required packages (`bubbles/textinput`, `bubbles/textarea`, `lipgloss`, `bubbletea`, `config.LoadConfig`) are already available.
- **`list` field exception**: `Enter` inside `textarea` (list type) is the correct UX for adding a new line. `Enter`-to-commit must be skipped for `list`-type fields. `Esc` remains the commit key for `list` fields.
- **No validation**: Input validation before commit is out of scope.
- **No auto-save**: Explicit save gesture (`Ctrl+S`) is retained as the persistence trigger.
- **No undo/revert**: Out of scope for this workitem.

## Implementation Tasks

### Task 1: Extract commit helper function
**File:** `internal/tui/model.go`

**What:** Extract the existing `Esc` commit logic (lines 203–211) into a shared `commitAndReturnToBrowsing()` method on `Model`. This method performs:
1. `newValue := m.detailPanel.StopEditing()`
2. `m.cfg.Set(item.Field.Name, newValue)`
3. `m.listPanel.UpdateItemValue(item.Field.Name, newValue)` — this also marks the field modified (see Task 3)
4. `m.saved = false` *(clear stale saved indicator)*
5. `m.err = nil`
6. `m.state = StateBrowsing`

**Why:** Eliminates duplication between `Enter` and `Esc` commit paths.

### Task 2: Intercept Enter in StateEditing for non-list fields
**File:** `internal/tui/model.go`

**What:** In the `StateEditing` branch of `handleKeyPress`, add an `Enter` handler **before** the default forwarding to `detailPanel.Update()`:
```
case StateEditing:
    if k == "esc" || (k == "enter" && currentFieldType != "list"):
        commitAndReturnToBrowsing()
    else:
        forward to detailPanel.Update(msg)
```
Use `m.listPanel.SelectedItem().Field.Type` (or `m.detailPanel.CurrentFieldType()` accessor — see Task 8) to determine whether `Enter` should commit or be forwarded.

**Why:** Fixes Gap G2 — `Enter` now commits the change and returns to browsing for `string`, `bool`, and `enum` fields.

### Task 3: Add dirty tracking to ConfigItem and ListPanel
**File:** `internal/tui/list_item.go`

**What:**
1. Add `Modified bool` field to `ConfigItem` struct
2. In `UpdateItemValue()`, set `l.entries[i].item.Modified = true` when updating the value
3. Add `ClearAllModified()` method that iterates all entries and sets `Modified = false`
4. When rendering, append ` (not-saved)` to the display string when `item.Modified` is true
5. Account for the `(not-saved)` suffix width (12 chars) when computing `valWidth` to avoid truncation of the actual value

**Why:** Fixes Gap G3 — users see which fields have in-memory changes not yet written to disk.

### Task 4: Pass Modified flag through renderItem
**File:** `internal/tui/list_item.go`

**What:** Update `renderItem()` to use `item.Modified` when computing the value display. Reduce `valWidth` by 12 characters when `item.Modified` is true so the actual value still fits alongside the `(not-saved)` suffix. Append the suffix after `formatValueCompact` returns.

**Why:** Ensures the "(not-saved)" indicator renders correctly without truncating the value display.

### Task 5: Post-save reload from disk
**File:** `internal/tui/model.go`

**What:** In the `ctrl+s` handler, after successful `SaveConfig`:
1. Call `config.LoadConfig(m.configPath)` to re-read the file
2. Replace `m.cfg` with the freshly loaded config
3. Save the current cursor field name before rebuild
4. Rebuild list entries: `entries := buildEntries(m.cfg, m.schema)`
5. Create a new `ListPanel` with the rebuilt entries (or update entries in-place)
6. Restore cursor to the same field name (by name, not index)
7. Re-sync the detail panel via `m.syncDetailPanel()`
8. Set `m.saved = true`
9. Clear `m.err` on success

**Failure path:** If `LoadConfig` fails after a successful `SaveConfig`, set `m.err` to a descriptive error (e.g., "saved but reload failed: ..."), keep the in-memory state, and still set `m.saved = true` (because the save itself succeeded).

**Why:** Fixes Gap G5 — the UI reflects the actual persisted state after save.

### Task 6: Clear saved flag on new edits
**File:** `internal/tui/model.go`

**What:** In the `commitAndReturnToBrowsing()` helper (Task 1), set `m.saved = false` and `m.err = nil`. This ensures the "✓ Saved" banner disappears the moment the user commits a new change.

**Why:** Fixes Gap G6 — prevents stale "✓ Saved" banner.

### Task 7: Clear all modified flags after save
**File:** `internal/tui/model.go`

**What:** After successful save-and-reload (Task 5), call `m.listPanel.ClearAllModified()` to remove all `(not-saved)` indicators. If reload fails but save succeeded, still clear modified flags (the data is on disk).

**Why:** Fixes Gap G4 — `(not-saved)` indicators disappear after `Ctrl+S`.

### Task 8: Expose CurrentFieldType accessor
**File:** `internal/tui/detail_panel.go`

**What:** Add a `CurrentFieldType() string` method to `DetailPanel`:
```go
func (d *DetailPanel) CurrentFieldType() string {
    if d.field == nil {
        return ""
    }
    return d.field.Type
}
```

**Why:** Needed by Task 2 for field-type-aware `Enter` handling, and by Task 9 for context-aware help bar.

### Task 9: Update help bar for editing state
**Files:** `internal/tui/keys.go`, `internal/tui/model.go`

**What:**
1. Add a `Confirm key.Binding` to `KeyMap` with help text `"enter" / "confirm"`
2. Update `ShortHelp(state State)` to accept a field type parameter (or add a `ShortHelpEditing(fieldType string)` variant)
3. For `StateEditing`:
   - Non-`list` fields: show `[enter • confirm]  [esc • confirm]  [ctrl+s • save]  [ctrl+c • quit]`
   - `list` fields: show `[esc • done]  [ctrl+s • save]  [ctrl+c • quit]` (Enter = newline is implicit)
4. Update `Model.View()` to pass the field type when constructing help keys

**Why:** Users need clear guidance on how to commit their edit, especially the Enter vs Esc distinction for list fields.

### Task 10: Add/update tests
**File:** `internal/tui/tui_test.go`

**What:** Add tests for all four fixed behaviours:

| Test ID | Description |
|---------|-------------|
| UT-TUI-026 | `Enter` in `StateEditing` on `enum` field commits value and returns to `StateBrowsing` |
| UT-TUI-027 | `Enter` in `StateEditing` on `string` field commits value and returns to `StateBrowsing` |
| UT-TUI-028 | `Enter` in `StateEditing` on `bool` field commits value and returns to `StateBrowsing` |
| UT-TUI-029 | `Enter` in `StateEditing` on `list` field does NOT exit editing (forwarded to textarea) |
| UT-TUI-030 | `ConfigItem.Modified` is `false` by default and set to `true` after `UpdateItemValue` |
| UT-TUI-031 | `formatValueCompact` appends `(not-saved)` when `Modified` is true |
| UT-TUI-032 | `ClearAllModified` resets all `Modified` flags to false |
| UT-TUI-033 | `m.saved` is cleared (`false`) after committing a new edit via Enter or Esc |
| UT-TUI-034 | Help bar shows `enter • confirm` for non-list fields in editing state |
| UT-TUI-035 | Help bar shows `esc • done` (no Enter confirm) for list fields in editing state |

**Update existing test:**
- UT-TUI-008: Extend to verify that `Esc` also clears `m.saved` and marks the field modified

## Implementation Order

```
Task 3  (dirty tracking on ConfigItem/ListPanel)
  └─► Task 4  (renderItem integration)
Task 8  (CurrentFieldType accessor)
  └─► Task 1  (extract commit helper)
        ├─► Task 2  (Enter intercept — depends on Task 1 + Task 8)
        └─► Task 6  (clear saved — integrated into Task 1)
Task 5  (post-save reload)
  └─► Task 7  (clear modified after save — depends on Task 3 + Task 5)
Task 9  (help bar — depends on Task 8)
Task 10 (tests — depends on all above)
```

Tasks 3 and 8 can be done in parallel as the first step. Tasks 1+2+6 form a logical unit. Tasks 5+7 form a logical unit. Task 9 is independent of the others. Task 10 is last.

## Risk Mitigations

| Risk | Mitigation |
|------|------------|
| Cursor position lost after post-save reload | Restore cursor by field name (not index) after rebuilding entries |
| `(not-saved)` suffix truncates values in narrow terminals | Reduce `valWidth` by 12 chars when `Modified` is true; test at 80-column |
| `Enter` inside `textarea` is ambiguous for list fields | Explicitly skip Enter-commit for `list` type; document in help bar |
| Post-save `LoadConfig` fails | Non-fatal error in header; keep in-memory state; still mark as saved |
| Existing UT-TUI-008 test assumptions | Update test to also verify `m.saved` cleared and `Modified` set |
