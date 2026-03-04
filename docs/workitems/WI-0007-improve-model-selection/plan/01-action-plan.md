# Action Plan: Improve Model Selection UX

## Feature
- **ID:** WI-0007-improve-model-selection
- **Research Brief:** [docs/workitems/WI-0007-improve-model-selection/research/00-research.md](../research/00-research.md)

## ADRs Created

None. This workitem extends patterns already established in:
- **ADR-0002** (Go with Charm TUI Stack) — `bubbles/list` is already a direct dependency in `go.mod`
- **ADR-0003** (Two-Panel TUI Layout Pattern) — the state machine and two-panel layout are already defined; Decision #13 already endorses `bubbles/list`

No new architectural decisions are introduced. The implementation uses existing libraries (`bubbles/list`, `sahilm/fuzzy`) and extends the existing TUI state machine with a new state that follows the same Elm-architecture patterns.

## Core-Components Created

None. `ModelPickerPanel` is a TUI-layer component analogous to `DetailPanel` and `ListPanel` — not a cross-cutting concern.

## Architectural Approach

### Problem

The `model` config field has 17 enum options with shared prefixes (`claude-sonnet-4.`, `gpt-5.1-codex`). The current generic enum editor in `DetailPanel` renders a static up/down list with no filtering, no scrolling, and requiring up to 16 key presses to navigate. With models sharing long prefixes, visual scanning is tedious and error-prone.

### Solution

Introduce a **full-width filterable picker overlay** using `bubbles/list` with built-in fuzzy filtering (powered by `sahilm/fuzzy`, already in `go.sum`). The picker activates for any enum field with **≥5 options**, keeping small enums on the existing inline selector.

### State Machine Change

```
CURRENT:
  StateBrowsing ──Enter──▶ StateEditing ──Esc──▶ StateBrowsing

PROPOSED:
  StateBrowsing ──Enter (enum ≥5 opts)──▶ StateModelPicker ──Enter/Esc──▶ StateBrowsing
  StateBrowsing ──Enter (other fields)──▶ StateEditing      ──Esc──────▶ StateBrowsing
```

Key properties of `StateModelPicker`:
- Renders a `bubbles/list` overlay that replaces **both panels** (full inner width/height)
- Supports type-to-filter fuzzy matching (built into `bubbles/list`)
- `Enter` confirms the selection and returns to `StateBrowsing` (when NOT in filter-input mode)
- `Esc` also confirms the current selection and returns to `StateBrowsing`
- Global keys (`ctrl+s` save, `ctrl+c` quit) continue to work — they are handled before the `switch m.state` block in `handleKeyPress`

### Threshold Rule

Enum fields are routed to the picker when `len(field.Options) >= 5`. This cleanly separates:
- **Small enums** (≤4 options): `theme` (3 opts), `banner` (3), `log_level` (4) → existing inline up/down selector in `DetailPanel`
- **Large enums** (≥5 options): `model` (17 opts) → full-width filterable picker overlay

This avoids field-name-based branching (no special-casing `"model"`) and is future-proof for any new enum fields with many options.

### View Rendering in `StateModelPicker`

When `state == StateModelPicker`, the `View()` method replaces the two-panel section with a single full-width picker:

```
┌──────────────────────────────────────────────┐
│ ╭─╮╭─╮  ccc — Copilot Config CLI            │
│ ╰─╯╰─╯  Copilot CLI v0.0.412                │
│                                               │
│ ╭── Select Model ─────────────────────────╮  │
│ │ / Filter                                │  │
│ │                                         │  │
│ │ ▶ claude-sonnet-4.6                     │  │
│ │   claude-sonnet-4.5                     │  │
│ │   claude-haiku-4.5                      │  │
│ │   claude-opus-4.6                       │  │
│ │   ...                                   │  │
│ ╰─────────────────────────────────────────╯  │
│                                               │
│ ╭ / filter  •  enter confirm  •  esc back ─╮ │
│ ╰─────────────────────────────────────────╯  │
└──────────────────────────────────────────────┘
```

### Integration with Existing Components

| Component | Impact |
|-----------|--------|
| `ListPanel` (left panel) | **No change.** Continues to render the config option list. |
| `DetailPanel` (right panel) | **No change to code.** Still handles bool/string/list/small-enum editing. The `model` field simply never enters `StateEditing` — it goes to `StateModelPicker` instead. |
| `KeyMap` / help bar | **Minor addition.** New `Filter` binding (`/`); new `ShortHelp` case for `StateModelPicker`. |
| `Config.Set()` / `SaveConfig()` | **No change.** The picker writes the selected model string via the existing `cfg.Set(field.Name, value)` path. |
| Sensitive data handling | **No change.** Sensitive fields cannot enter editing or picker states — the `isSensitiveItem()` guard in `StateBrowsing` Enter handler blocks them. |

## Implementation Tasks

### Task 1: Add `StateModelPicker` to state machine
**File:** `internal/tui/state.go`
**Action:** Modify
**Details:**
- Add `StateModelPicker` constant to the `State` iota block (between `StateEditing` and `StateSaving`)
- Add `"ModelPicker"` case to the `String()` method
- Estimated: ~5 lines changed

### Task 2: Create `ModelPickerPanel` component
**File:** `internal/tui/model_picker_panel.go`
**Action:** Create (new file, ~70 lines)
**Details:**
- Define `modelItem` struct implementing `list.Item` interface:
  - `Title() string` → returns the model name
  - `Description() string` → returns `""` (no description needed)
  - `FilterValue() string` → returns the model name (used by fuzzy filter)
- Define `ModelPickerPanel` struct:
  - `list list.Model` — the bubbles list component
  - `selected string` — the value when the picker was opened (for fallback)
- Constructor `NewModelPickerPanel(options []string, current string) ModelPickerPanel`:
  - Convert options to `[]list.Item`
  - Create `list.New(items, list.NewDefaultDelegate(), 0, 0)`
  - Configure: `SetShowStatusBar(false)`, `SetFilteringEnabled(true)`, `SetShowHelp(false)`
  - Set title (e.g., `l.Title = "Select Model"`)
  - Pre-select current value via `l.Select(i)` where `options[i] == current`
- Methods: `SetSize(w, h)`, `Update(msg) tea.Cmd`, `SelectedValue() string`, `View() string`
- Style the default delegate to use the project's colors from `styles.go` (set `Styles.SelectedTitle` to use `primaryColor`, etc.)

### Task 3: Add `Filter` key binding
**File:** `internal/tui/keys.go`
**Action:** Modify (~5 lines)
**Details:**
- Add `Filter key.Binding` field to `KeyMap` struct
- Initialize in `DefaultKeyMap()`:
  ```go
  Filter: key.NewBinding(
      key.WithKeys("/"),
      key.WithHelp("/", "filter"),
  ),
  ```

### Task 4: Wire `StateModelPicker` into main model
**File:** `internal/tui/model.go`
**Action:** Modify (~60 lines across several sections)

**4a. Add field to `Model` struct:**
```go
modelPickerPanel *ModelPickerPanel
```

**4b. Branch in `handleKeyPress` — `StateBrowsing` / `Enter`:**
Before the existing `StateEditing` transition, check:
```go
if item.Field.Type == "enum" && len(item.Field.Options) >= 5 {
    // Route to model picker
    current := ""
    if s, ok := item.Value.(string); ok { current = s }
    picker := NewModelPickerPanel(item.Field.Options, current)
    picker.SetSize(innerWidth-4, panelHeight-2)  // use pre-computed sizes
    m.modelPickerPanel = &picker
    m.state = StateModelPicker
    return m, nil
}
```

**4c. Add `StateModelPicker` case in `handleKeyPress`:**
```go
case StateModelPicker:
    if k == "enter" {
        // Only confirm if NOT actively filtering
        if m.modelPickerPanel != nil && m.modelPickerPanel.list.FilterState() != list.Filtering {
            newValue := m.modelPickerPanel.SelectedValue()
            if item := m.listPanel.SelectedItem(); item != nil {
                m.cfg.Set(item.Field.Name, newValue)
                m.listPanel.UpdateItemValue(item.Field.Name, newValue)
                m.detailPanel.SetField(item.Field, newValue)
            }
            m.modelPickerPanel = nil
            m.state = StateBrowsing
            return m, nil
        }
    }
    if k == "esc" {
        // Always confirm on Esc (even during filtering, cancel filter + exit)
        if m.modelPickerPanel != nil {
            newValue := m.modelPickerPanel.SelectedValue()
            if item := m.listPanel.SelectedItem(); item != nil {
                m.cfg.Set(item.Field.Name, newValue)
                m.listPanel.UpdateItemValue(item.Field.Name, newValue)
                m.detailPanel.SetField(item.Field, newValue)
            }
        }
        m.modelPickerPanel = nil
        m.state = StateBrowsing
        return m, nil
    }
    // Forward all other keys to picker
    return m, m.modelPickerPanel.Update(msg)
```

**4d. Forward non-key messages in `Update()`:**
```go
if m.state == StateModelPicker && m.modelPickerPanel != nil {
    return m, m.modelPickerPanel.Update(msg)
}
```

**4e. Update `View()` — render picker overlay:**
When `state == StateModelPicker`, replace the two-panel section with:
```go
pickerContent := m.modelPickerPanel.View()
panels = focusedPanelStyle.
    Width(innerWidth - 4).
    Height(panelHeight - 2).
    Render(pickerContent)
```

**4f. Update `updateSizes()`:**
Add sizing for picker when present:
```go
if m.modelPickerPanel != nil {
    m.modelPickerPanel.SetSize(innerWidth-4, panelHeight-2)
}
```

### Task 5: Update help bar for `StateModelPicker`
**File:** `internal/tui/model.go` (in `ShortHelp()` method)
**Action:** Modify (~3 lines)
```go
case StateModelPicker:
    return []key.Binding{k.Filter, k.Enter, k.Escape, k.Save, k.Quit}
```

### Task 6: Add unit tests
**File:** `internal/tui/tui_test.go`
**Action:** Modify (~120 lines)

| Test ID | Description |
|---------|-------------|
| UT-TUI-026 | Pressing Enter on enum with ≥5 options transitions to `StateModelPicker` |
| UT-TUI-027 | Pressing Enter on enum with <5 options transitions to `StateEditing` (not picker) |
| UT-TUI-028 | `NewModelPickerPanel` creates panel with all options; `SelectedValue()` returns current |
| UT-TUI-029 | `ModelPickerPanel.View()` renders non-empty string |
| UT-TUI-030 | Esc from `StateModelPicker` writes selected value to config, returns to `StateBrowsing` |
| UT-TUI-031 | Enter from `StateModelPicker` (not filtering) confirms selection, returns to `StateBrowsing` |
| UT-TUI-032 | `View()` in `StateModelPicker` at various sizes renders without panic |
| UT-TUI-033 | `ctrl+s` save works from `StateModelPicker` state |

## Constraints & Risks

1. **No new dependencies.** `bubbles/list` and `sahilm/fuzzy` are already resolved in `go.mod`/`go.sum`. Zero `go get` calls needed.

2. **No changes outside `internal/tui/`.** Config, copilot detection, CLI parsing, sensitive data handling, and logging are all unaffected.

3. **`bubbles/list` Enter key ambiguity.** The list component uses Enter to confirm a filter query when in filter-input mode (`FilterState() == list.Filtering`). The `StateModelPicker` handler must check `FilterState()` before treating Enter as "confirm selection and exit". When filtering is active, Enter should be forwarded to the list (to confirm the filter text and return to the filtered results view). Only when `FilterState() != list.Filtering` should Enter trigger exit from the picker.

4. **Esc key in filter mode.** When `bubbles/list` is in filter mode, Esc clears the filter. This means our "Esc = confirm and exit" behavior conflicts. The implementation should forward Esc to the list first when `FilterState() == list.Filtering` (to clear the filter), and only exit the picker on a second Esc when `FilterState() == list.Unfiltered`. Alternatively, always exit on Esc and let the user's most recent selection stand — this is simpler and matches the research brief's recommendation.

5. **Delegate styling.** The `list.NewDefaultDelegate()` has its own styles that may clash with the project's `primaryColor`/`secondaryColor` palette in `styles.go`. The implementer should customize the delegate's `Styles` to use the project's colors for visual consistency.

6. **Decision Log discrepancy.** Decision #13 says "Use `bubbles/list` for left panel config option navigation" but the left panel is actually a custom `ListPanel`. This workitem does **not** address that discrepancy — it introduces `bubbles/list` only for the picker overlay.

7. **Backward compatibility.** Existing enum fields with <5 options are completely unaffected. The existing `DetailPanel` enum editor (up/down with `▶` cursor) continues to handle them unchanged.

## Files Summary

| File | Action | Est. Lines |
|------|--------|-----------|
| `internal/tui/state.go` | Modify | ~5 |
| `internal/tui/model_picker_panel.go` | **Create** | ~70 |
| `internal/tui/keys.go` | Modify | ~5 |
| `internal/tui/model.go` | Modify | ~60 |
| `internal/tui/tui_test.go` | Modify | ~120 |
| **Total** | | **~260** |
