# WI-0006: User Experience — Saving and Viewing Configuration Values

## Status

Draft

## Scope Classification

**Workitem** — This is a focused UX bug-fix and enhancement confined to the TUI layer (`internal/tui/`). It does not introduce a new architecture, does not change the `internal/config/` or `internal/copilot/` packages, and does not require a new ADR. It does, however, necessitate **updating CORE-COMPONENT-0004** to document the "unsaved changes" state concept that this workitem introduces.

---

## Executive Summary

The TUI editing workflow has four distinct UX breakdowns. First, `Enter` in `StateEditing` does not commit the value and return to browsing — it is forwarded verbatim to the widget (inserting a newline in `textarea`, toggling a boolean, or doing nothing for enums), so users who press `Enter` to confirm a change see no feedback and their change is never committed to the in-memory config. Second, there is no visual "(not-saved)" indicator in the left-panel list to communicate that an in-memory change has not yet been persisted to disk. Third, after `Ctrl+S` saves to disk the model does not re-read the file, so the persisted state is never round-trip–verified in the UI. Fourth, the `saved` flag is never cleared when a new unsaved change is made, causing the "✓ Saved" banner to show stale information.

These four issues are all confined to `internal/tui/` and are correctable with targeted changes to `model.go`, `detail_panel.go`, and `list_item.go`.

---

## 1. Current Implementation Analysis

### 1.1 State Machine

The TUI runs a two-state machine defined in `internal/tui/state.go`[^1]:

```
StateBrowsing  ──Enter──▶  StateEditing
StateEditing   ──Esc────▶  StateBrowsing
(both states)  ──Ctrl+S──▶ (save, stay in same state)
(both states)  ──Ctrl+C──▶ tea.Quit
```

`StateSaving` and `StateExiting` are declared but never used[^1].

### 1.2 The Enter-Key Problem (Root Cause)

In `handleKeyPress` (`internal/tui/model.go:166–218`)[^2], the `StateEditing` branch is:

```go
case StateEditing:
    if k == "esc" {
        newValue := m.detailPanel.StopEditing()
        if item := m.listPanel.SelectedItem(); item != nil {
            m.cfg.Set(item.Field.Name, newValue)
            m.listPanel.UpdateItemValue(item.Field.Name, newValue)
        }
        m.state = StateBrowsing
        return m, nil
    }
    // All other keys go to detail panel
    return m, m.detailPanel.Update(msg)   // ← Enter falls here
```

When the user presses `Enter`, it is forwarded to `detailPanel.Update()`[^3], which:

| Field type | What `Enter` does |
|------------|-------------------|
| `bool`     | Toggles the boolean value (`detail_panel.go:157`) — but does **not** commit or exit |
| `enum`     | Not handled — `Enter` is silently ignored (`detail_panel.go:162–173`) |
| `string`   | Passed to `textinput.Update()` — textinput does nothing with Enter by default |
| `list`     | Passed to `textarea.Update()` — inserts a newline (correct for multi-line, but never commits) |

**Consequence**: For `bool`, `enum`, and `string` types, pressing `Enter` feels like it should commit the change but no commit occurs. The user is silently stuck in `StateEditing`. The in-memory value in `cfg` is never updated via `cfg.Set()`, so navigating away shows the old value.

### 1.3 Value Sync After Esc (Working, but Invisible)

When the user does press `Esc`, the flow works correctly[^2]:

```go
newValue := m.detailPanel.StopEditing()          // extract widget value
m.cfg.Set(item.Field.Name, newValue)              // update in-memory config
m.listPanel.UpdateItemValue(item.Field.Name, newValue)  // update list display
m.state = StateBrowsing
```

`UpdateItemValue` iterates `l.entries` and mutates the matching entry's `Value`[^4]. The list will render the new value on the next `View()` call. **This path works**, but is only reachable via `Esc`, not `Enter`.

### 1.4 No "Unsaved" State

The `Model` struct has a single `saved bool` field[^5]:

```go
type Model struct {
    ...
    saved bool    // set to true by Ctrl+S; never cleared
}
```

This is rendered in the header as `"✓ Saved"` when `m.saved == true`[^6]. There is no concept of "changes exist in memory that have not been persisted". The `saved` flag is never reset when a new edit is committed via `Esc`, so the header can show `"✓ Saved"` even after additional unsaved changes.

There is also no per-item "dirty" marker. `listEntry` holds `item ConfigItem` with `Value any` — no `modified bool`, no `originalValue any`[^4]. The list `formatValueCompact` function has no "(not-saved)" code path[^7].

### 1.5 No Post-Save Reload

The `Ctrl+S` save handler[^2]:

```go
if k == "ctrl+s" {
    if err := config.SaveConfig(m.configPath, m.cfg); err != nil {
        m.err = err
    } else {
        m.saved = true
    }
    return m, nil
}
```

`SaveConfig` writes `m.cfg.data` to disk[^8]. After this, `m.cfg` is **not** reloaded from disk. The in-memory `cfg` continues to be the source of truth. This means:

- Round-trip integrity is not verified (e.g., a partial write would not surface in the UI)
- If the file is externally modified between load and save, those external changes are silently overwritten without the user knowing
- The UI doesn't reflect the actual persisted state; it reflects the in-memory state that was written

### 1.6 `SelectedItem()` Returns a Copy

`ListPanel.SelectedItem()` returns a pointer to a stack-allocated copy[^4]:

```go
func (l *ListPanel) SelectedItem() *ConfigItem {
    item := l.entries[l.cursor].item   // copy
    return &item                        // pointer to copy
}
```

This is safe for the current read-only usage (`item.Field.Name`, `item.Value`), but must not be relied upon for mutation. `UpdateItemValue` correctly uses a name-based scan to mutate entries in-place.

---

## 2. Gap Analysis

| # | Desired Behaviour | Current Behaviour | Gap |
|---|-------------------|-------------------|-----|
| G1 | `Enter` on a config item enters editing mode | ✅ Works — `StateBrowsing` + `Enter` → `StateEditing` | No gap |
| G2 | `Enter` while editing commits the change and returns to the list | ❌ `Enter` is forwarded to widget; no commit; no state transition | `Enter` must trigger the same `StopEditing` + `cfg.Set` + `UpdateItemValue` flow that `Esc` currently triggers |
| G3 | List shows "(not-saved)" indicator for changed-but-not-persisted items | ❌ No dirty state exists anywhere | Need dirty tracking per field; `formatValueCompact` must render the indicator |
| G4 | `Ctrl+S` removes "(not-saved)" indicators | ❌ No dirty state to clear | After successful save, clear dirty markers |
| G5 | After `Ctrl+S`, the config file is re-read so the UI reflects persisted state | ❌ No reload; in-memory state continues unchanged | Call `config.LoadConfig` after `SaveConfig`; rebuild list entries or sync values |
| G6 | "✓ Saved" banner disappears when a new unsaved change is made | ❌ `saved` flag is never cleared | Clear `m.saved` (and set dirty marker) when `cfg.Set` is called |

### 2.1 `list` Type Edge Case

For `list`-type fields, `Enter` inside `textarea` is the correct UX for adding a new line. Treating `Enter` as "commit and exit" for `list` fields would make it impossible to enter multi-item lists. The resolution for `list` fields should use a **different commit key**: `Esc` (existing), possibly also `Ctrl+Enter` or a dedicated "Done" help-bar item. This is a known UX asymmetry that the implementer must handle explicitly per field type.

---

## 3. Proposed Approach

### 3.1 Commit on Enter (Fix G2)

In `handleKeyPress`, within `StateEditing`, intercept `Enter` **before** forwarding to the detail panel, but only for field types where `Enter` is unambiguous (i.e., not `list`):

```
StateEditing + Enter:
  if field.Type != "list" → commitAndReturnToBrowsing()
  else                    → forward to detail panel (textarea newline)
```

`commitAndReturnToBrowsing()` is the same three-step sequence currently used by `Esc`:
1. `newValue := m.detailPanel.StopEditing()`
2. `m.cfg.Set(field.Name, newValue)`
3. `m.listPanel.UpdateItemValue(field.Name, newValue)`
4. `m.state = StateBrowsing`
5. Mark field as dirty (see §3.2)
6. Clear `m.saved`

For `list` fields, `Esc` remains the commit key and should be surfaced in the help bar.

### 3.2 Dirty Tracking (Fix G3, G4, G6)

Introduce a minimal dirty-state mechanism. There are two design options:

**Option A — Dirty set on `Model`**
Add `dirtyFields map[string]bool` to `Model`. When `cfg.Set` is called for a field, add `fieldName → true`. When `Ctrl+S` succeeds, clear the map.

**Option B — `modified bool` on `ConfigItem`**
Add `Modified bool` to `ConfigItem` in `list_item.go`. `UpdateItemValue` sets it to `true`. After a successful save, iterate all entries and reset `Modified = false`.

Option A is simpler (no struct change to `ConfigItem`). Option B makes the dirty state co-located with the item, which simplifies rendering. **Option B is recommended** because `renderItem` and `formatValueCompact` operate on `ConfigItem` directly, and co-location avoids a secondary map lookup at render time.

**`formatValueCompact` change (list_item.go)**:
```go
// After computing `s`:
if item.Modified {
    return s + " (not-saved)"  // or truncate to fit maxLen
}
```

The `(not-saved)` suffix must be taken into account when computing `maxLen` to avoid truncating the actual value. The implementer should reduce `valWidth` by `len(" (not-saved)")` (12 chars) whenever any item in the panel is potentially dirty, or truncate intelligently.

### 3.3 Post-Save Reload (Fix G5)

After a successful `config.SaveConfig`, call `config.LoadConfig(m.configPath)` to re-read the persisted file. Then:

1. Replace `m.cfg` with the freshly loaded config
2. Rebuild the list entries by calling `buildEntries(m.cfg, m.schema)`
3. Restore the cursor to the same field name (not the same index, since order could theoretically change)
4. Re-sync the detail panel

This ensures the UI is an accurate reflection of what is on disk. The reload is a file read + JSON unmarshal and is fast enough to be synchronous (no `tea.Cmd` needed).

**Failure path**: if `LoadConfig` fails after a successful `SaveConfig`, show a non-fatal error in the header (using the existing `m.err` field) and keep the in-memory state. This is an edge case (e.g., file permissions changed after save) and should not be silent.

### 3.4 Clear `saved` on New Edit (Fix G6)

When `cfg.Set` is called (i.e., inside `commitAndReturnToBrowsing` and the `Esc` path), set `m.saved = false`. The "✓ Saved" banner will only show immediately after a successful `Ctrl+S` and will disappear the moment the user commits a new change.

### 3.5 Help Bar Update

The help bar in `StateEditing` must reflect the new commit key. Concretely:
- For non-`list` fields: `enter • confirm` should appear alongside `esc • cancel/confirm`
- For `list` fields: help should read `esc • done` (Enter = newline is implicit)

This requires a field-type-aware `ShortHelp()` or passing the current field type to `ShortHelp()`. The simplest approach is adding a `currentFieldType string` accessor to `DetailPanel` and using it in `Model.View()` when constructing the help keys.

---

## 4. Files to Be Changed

| File | Change |
|------|--------|
| `internal/tui/model.go` | (1) Intercept `Enter` in `StateEditing` for non-`list` fields; (2) reload from disk after `Ctrl+S`; (3) clear `m.saved` on new commit |
| `internal/tui/list_item.go` | (4) Add `Modified bool` to `ConfigItem`; (5) `UpdateItemValue` sets `Modified = true`; (6) `formatValueCompact` appends `(not-saved)` when `Modified`; (7) new `ClearModified()` method |
| `internal/tui/detail_panel.go` | (8) Expose `CurrentFieldType() string` accessor for help-bar rendering |
| `internal/tui/keys.go` | (9) Add `Confirm key.Binding` for Enter-to-commit; update `ShortHelp` to be field-type-aware |
| `internal/tui/tui_test.go` | (10) Add/update tests for all four fixed behaviours |

No changes are required to `internal/config/`, `internal/copilot/`, `internal/sensitive/`, or `cmd/`.

---

## 5. ADRs and Core-Components

### ADRs

No new ADR is needed. All changes are within the existing two-panel layout pattern established by ADR-0003[^9]. The state machine gains a new transition (`Enter` → commit) but the overall Browsing/Editing duality is unchanged.

### Core-Components

**CORE-COMPONENT-0004 (Configuration Management)** should be **updated** to document:

1. The concept of "unsaved changes" (in-memory `cfg` vs. persisted disk state)
2. The rule that the UI must re-read the config from disk after each successful save to maintain round-trip integrity
3. The `Modified` dirty marker on `ConfigItem` as the mechanism for surfacing the "unsaved" state to the user

No new core-component is needed — this is an extension of the existing configuration management contract.

---

## 6. Dependencies and Risks

### Dependencies

- No new Go module dependencies are required. All needed packages (`bubbles/textinput`, `bubbles/textarea`, `lipgloss`, `bubbletea`) are already in `go.mod`.
- `config.LoadConfig` is already implemented and tested[^8].

### Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Cursor position lost after post-save reload | Medium | Low UX annoyance | Restore cursor by field name after rebuilding entries |
| `(not-saved)` suffix causes truncation of long values in narrow terminals | Medium | Cosmetic | Compute `valWidth` accounting for the suffix; test at 80-column |
| `Enter` inside `textarea` (list type) is ambiguous | High | Medium | Explicitly skip Enter-commit for `list` type; document in help bar |
| Post-save reload fails (race / permissions) | Low | Medium | Non-fatal error in header; keep in-memory state |
| `saved` true + unsaved changes at same time | Low (fixed) | Medium | Cleared by clearing `m.saved` on every `cfg.Set` call |
| Existing test UT-TUI-008 ("Esc saves value") may overlap with new Enter behaviour | Certain | Low | Update test to cover both `Enter` and `Esc` commit paths |

---

## 7. Scope Boundary

This workitem is **NOT** proposing:

- Validation before commit (e.g., URL format, enum membership) — that is a separate concern
- Auto-save on navigation — explicit save gesture (Ctrl+S) is retained as the persistence trigger
- Undo / revert-to-saved — out of scope
- Changes to the `list` field multi-line editing UX beyond the `Esc`-to-commit clarification
- Any change to `internal/config/`, `internal/copilot/`, or the CLI bootstrap in `cmd/`

---

## Confidence Assessment

| Finding | Confidence | Basis |
|---------|-----------|-------|
| `Enter` in `StateEditing` does not exit editing | **High** | Direct code trace in `model.go:202–214`[^2] |
| `Esc` path correctly updates list and in-memory config | **High** | Direct code trace in `model.go:203–210`[^2]; `UpdateItemValue` verified in `list_item.go:78–85`[^4] |
| No dirty/unsaved state exists | **High** | `Model` struct has only `saved bool`; `ConfigItem` has no `Modified` field[^4][^5] |
| No post-save disk reload | **High** | `Ctrl+S` handler ends at `m.saved = true`; no `LoadConfig` call[^2] |
| `saved` flag never cleared on new edit | **High** | No `m.saved = false` assignment anywhere in `model.go`[^2] |
| Post-save reload is safe to do synchronously | **Medium** | Config file is small JSON; `LoadConfig` is a simple `os.ReadFile` + `json.Unmarshal`[^8]; no observed latency concerns, but not benchmarked |
| Option B (dirty bit on `ConfigItem`) is preferred | **Medium** | Architectural preference based on co-location principle; implementer may choose Option A |

---

## Footnotes

[^1]: `internal/tui/state.go` — `State` enum with `StateBrowsing`, `StateEditing`, `StateSaving`, `StateExiting`
[^2]: `internal/tui/model.go:166–218` — `handleKeyPress` function; save handler at lines 174–184; `StateEditing` branch at lines 202–214
[^3]: `internal/tui/detail_panel.go:149–184` — `DetailPanel.Update()` handling `bool` (toggle on Enter), `enum` (up/down only), `string`/`list` (forwarded to Bubbles widgets)
[^4]: `internal/tui/list_item.go:12–85` — `ConfigItem`, `listEntry`, `ListPanel` types; `SelectedItem()` returns copy at line 71–75; `UpdateItemValue()` at lines 78–85
[^5]: `internal/tui/model.go:17–32` — `Model` struct definition; `saved bool` field at line 31
[^6]: `internal/tui/model.go:283–285` — `View()` rendering `"✓ Saved"` when `m.saved`
[^7]: `internal/tui/list_item.go:176–211` — `formatValueCompact` function; no dirty/modified code path
[^8]: `internal/config/config.go` — `LoadConfig` and `SaveConfig` functions
[^9]: `docs/architecture/ADR/ADR-0003-two-panel-tui-layout.md` — Two-Panel TUI Layout Pattern ADR
