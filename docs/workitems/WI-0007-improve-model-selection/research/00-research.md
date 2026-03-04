# Research Brief: Improve Model Selection UX

**Workitem:** WI-0007-improve-model-selection  
**Branch:** `feat/improve-model-selection`  
**Feature:** Replace the generic enum up/down selector for the `model` field with a filterable, searchable picker using `github.com/charmbracelet/bubbles/list`.

---

## Executive Summary

The `model` configuration field currently exposes 17 options through a generic enum editor that renders a static up/down scrollable list inside the detail panel. With 17 model names — many sharing prefixes (`claude-`, `gpt-5.`) — discovery and selection is slow and error-prone. The `charmbracelet/bubbles` package (already imported in `go.mod`) ships a `list` component with built-in fuzzy filtering powered by `sahilm/fuzzy` (also already in `go.sum`). The proposed change introduces a **new TUI state** (`StateModelPicker`) and a **`ModelPickerPanel`** component that presents the 17 model names in a `bubbles/list` with type-to-filter behaviour, keeping the rest of the editing flow (state machine, `cfg.Set`, `ctrl+s` save) entirely unchanged.

The full change is **contained within `internal/tui/`**. No changes are needed to `internal/config/`, `internal/copilot/`, or `cmd/`.

---

## Architecture Overview

```
cmd/ccc/main.go
    │
    ├── copilot.DetectSchema()      →  SchemaField{Name:"model", Type:"enum",
    │                                  Options: [17 model strings], Default: ""}
    ├── config.LoadConfig(path)     →  cfg.Get("model") → string or nil
    │
    └── tui.NewModel(cfg, schema, version, configPath)
            │
            ├── buildEntries()
            │     └── "model" → ConfigItem{Field: SchemaField, Value: string|nil}
            │           → categorized under "Model & AI" group header
            │
            ├── ListPanel (custom hand-rolled)       ← left panel
            │
            ├── DetailPanel                          ← right panel
            │     └── renderEditWidget() case "enum":
            │           simple up/down list (current — NO filtering)
            │
            └── [NEW] ModelPickerPanel               ← overlay / new state
                  └── bubbles/list.Model
                        └── fuzzy filtering via sahilm/fuzzy
```

### State Machine (current)

```
StateBrowsing ──Enter──▶ StateEditing ──Esc──▶ StateBrowsing
                              │
                        (writes value to cfg)
```

### State Machine (proposed)

```
StateBrowsing ──Enter (model field)──▶ StateModelPicker ──Esc──▶ StateBrowsing
StateBrowsing ──Enter (other enum)───▶ StateEditing       ──Esc──▶ StateBrowsing
```

---

## 1. Current Implementation

### 1.1 `model` Field Schema

The `model` field is parsed by `copilot.ParseSchema()` from `copilot help config` output[^1]:

```
`model`: AI model to use for Copilot CLI; can be changed with /model command or --model flag option.
  - "claude-sonnet-4.6"
  - "claude-sonnet-4.5"
  - "claude-haiku-4.5"
  - "claude-opus-4.6"
  - "claude-opus-4.6-fast"
  - "claude-opus-4.5"
  - "claude-sonnet-4"
  - "gemini-3-pro-preview"
  - "gpt-5.3-codex"
  - "gpt-5.2-codex"
  - "gpt-5.2"
  - "gpt-5.1-codex-max"
  - "gpt-5.1-codex"
  - "gpt-5.1"
  - "gpt-5.1-codex-mini"
  - "gpt-5-mini"
  - "gpt-4.1"
```

This produces a `SchemaField` with `Type: "enum"`, `Options: [17 items]`, and `Default: ""` (no "defaults to" clause exists for `model`).[^2]

The `model` field is categorized into the `"Model & AI"` group by `isModelField()` in `model.go`[^3].

### 1.2 Generic Enum Editor (Current)

When the user presses Enter on any `enum`-type field, the state transitions to `StateEditing` and the `DetailPanel` gains focus.[^4]

The enum widget is rendered by `renderEditWidget()`[^5]:

```go
// internal/tui/detail_panel.go:274-287
case "enum":
    var opts strings.Builder
    for i, opt := range d.field.Options {
        if i == d.selectIndex {
            opts.WriteString(selectedOptionStyle.Render("▶ " + opt))
        } else {
            opts.WriteString(optionStyle.Render("  " + opt))
        }
        opts.WriteString("\n")
    }
    return opts.String()
```

Navigation in `StateEditing` for enum fields[^6]:

```go
// internal/tui/detail_panel.go:162-173
case "enum":
    switch keyMsg.String() {
    case "up", "k":
        if d.selectIndex > 0 {
            d.selectIndex--
        }
    case "down", "j":
        if d.selectIndex < len(d.field.Options)-1 {
            d.selectIndex++
        }
    }
    return nil
```

**Problems with the current approach for `model`:**
- 17 options require up to 16 key presses to navigate from top to bottom.
- No filtering — user must scan visually.
- Model names share long prefixes (`claude-sonnet-4.`, `gpt-5.1-codex`) making keyboard selection tedious.
- The detail panel height may not accommodate all 17 options without scrolling — and the panel does not currently scroll enum options.

### 1.3 State Transitions (Current)

Browsing → Editing transition[^7]:

```go
// internal/tui/model.go:196-200
case "enter":
    if item := m.listPanel.SelectedItem(); item != nil && !isSensitiveItem(*item) {
        m.state = StateEditing
        slog.Info("editing", "field", item.Field.Name)
        return m, m.detailPanel.StartEditing()
    }
```

Editing → Browsing transition[^8]:

```go
// internal/tui/model.go:203-210
if k == "esc" {
    newValue := m.detailPanel.StopEditing()
    if item := m.listPanel.SelectedItem(); item != nil {
        slog.Info("field updated", "field", item.Field.Name)
        m.cfg.Set(item.Field.Name, newValue)
        m.listPanel.UpdateItemValue(item.Field.Name, newValue)
    }
    m.state = StateBrowsing
    return m, nil
}
```

`StopEditing()` calls `GetValue()` which for `enum` returns `d.field.Options[d.selectIndex]`.[^9]

### 1.4 Left Panel (Custom List)

The left panel (`ListPanel` in `list_item.go`) is a **custom hand-rolled component** — it does NOT use `bubbles/list`.[^10] It maintains its own `cursor` and `offset` integers and does its own scrolling. This contradicts Decision #13 in the ADR decision log ("Use `bubbles/list` for left panel config option navigation"), but is the actual implementation on the branch.[^11]

---

## 2. TUI Framework Analysis

### 2.1 Dependencies

| Package | Version | Used For |
|---------|---------|----------|
| `github.com/charmbracelet/bubbletea` | `v1.3.10` | Main event loop, `tea.Model` interface |
| `github.com/charmbracelet/bubbles` | `v0.21.1-0.20250623103423-23b8fd6302d7` | `textinput`, `textarea`, `key` (and `list` — available but unused) |
| `github.com/charmbracelet/lipgloss` | `v1.1.0` | Styles, borders, layout |
| `github.com/sahilm/fuzzy` | `v0.1.1` | Fuzzy matching (**indirect** dep of bubbles; already in go.sum) |
| `github.com/spf13/cobra` | `v1.10.2` | CLI argument parsing |

`go.mod` explicitly lists `bubbles` and `bubbletea` as direct dependencies.[^12] `sahilm/fuzzy` is listed as an **indirect** dependency.[^13] This means `bubbles/list`'s filtering capability is already available — no new dependencies need to be added to `go.mod`.

### 2.2 `bubbles/list` Component

The `list` component (`github.com/charmbracelet/bubbles/list`) provides:
- A scrollable list of `list.Item` interface values
- **Built-in fuzzy filtering** triggered by typing `/` (or configurable)
- Keyboard navigation (up/down/pgup/pgdn/home/end)
- Configurable item delegates for custom rendering
- `list.Model.SelectedItem()` to get the currently highlighted item
- Configurable width/height
- Title and status bar

The fuzzy filtering uses `sahilm/fuzzy` which is already resolved in `go.sum`.[^14]

### 2.3 Bubbletea Model/Update/View Pattern

The application follows the standard Elm Architecture[^15]:

```go
// internal/tui/model.go:144-163
func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:   // resize handling
    case tea.KeyMsg:          // key input routing
    }
    // non-key messages forwarded to detail panel in StateEditing
}

func (m *Model) View() string { /* renders full screen */ }
```

The `Model` struct holds all state: `cfg`, `schema`, `state`, `listPanel`, `detailPanel`.[^16] Adding a `modelPickerPanel` field would follow the exact same pattern.

### 2.4 Currently Used Bubbles Components

| Component | Import | Usage |
|-----------|--------|-------|
| `bubbles/key` | `github.com/charmbracelet/bubbles/key` | `KeyMap`, `key.Binding`, `key.NewBinding` |
| `bubbles/textinput` | `github.com/charmbracelet/bubbles/textinput` | String field editing in `DetailPanel` |
| `bubbles/textarea` | `github.com/charmbracelet/bubbles/textarea` | List field editing in `DetailPanel` |
| `bubbles/list` | *(available, not used)* | Available for model picker |

---

## 3. Existing UI Patterns

### 3.1 Two-Panel Layout

The TUI uses a 40/60 split (left/right)[^17]:
- Left: `ListPanel` — config key navigation
- Right: `DetailPanel` — field info + editing widget

When `StateBrowsing`: left panel border is highlighted with `primaryColor`[^18].  
When `StateEditing`: right panel border is highlighted[^19].

### 3.2 `bool` Toggle

Simple space/enter toggle, renders `✓ Yes` or `✗ No` with color styles[^20].

### 3.3 `string` Editor

Uses `bubbles/textinput.Model`, `textinput.Blink` cmd, width auto-sized to panel[^21].

### 3.4 `list` Editor

Uses `bubbles/textarea.Model`, one URL/path per line[^22].

### 3.5 `enum` Selector (current)

Custom up/down selection from `d.field.Options`, no scroll, no filter — the target for replacement/enhancement[^23].

### 3.6 How Views Compose

`model.go:View()` composes the full screen:
1. Framed header (icon + title + version)
2. Two panels (left + right, joined horizontally)
3. Framed help bar

An overlay or full-screen replacement model picker would replace the entire `panels` line with the picker rendered at full inner width/height.

---

## 4. Proposed Approach

### 4.1 Design Decision: Overlay vs. In-panel

**Option A — In-panel (right panel)**  
Replace the detail panel contents with a `bubbles/list` when editing an enum with >N options. The model picker occupies the right panel only.

*Pros:* Minimal layout change; consistent with existing patterns.  
*Cons:* Right panel is ~60% width and limited height; fuzzy filter search bar and list must fit in a constrained space.

**Option B — Full-width overlay**  
Add `StateModelPicker` that renders the `bubbles/list` taking the full inner width/height (replacing both panels). This gives the most screen real estate for a long list.

*Pros:* Maximum UX clarity; more room for 17 items + filter bar + current value badge.  
*Cons:* Slightly more complex layout code.

**Option C — Model-specific branching on field name**  
Detect that the field being edited is `model` (or any enum with >X options) and branch into the new picker.

*Pros:* Backward-compatible; other enums (`theme`, `banner`, `log_level`) keep the simple editor.  
*Cons:* Field-name-based branching is a code smell.

**Recommendation: Option B + C combined** — introduce `StateModelPicker` triggered for any enum field with `len(Options) > threshold` (e.g., 5+), not just the `model` field specifically. This is future-proof and avoids special-casing by name.

### 4.2 Implementation Plan (High Level)

#### Step 1: Add `StateModelPicker` to state machine

```go
// internal/tui/state.go
const (
    StateBrowsing   State = iota
    StateEditing
    StateModelPicker   // NEW
    StateSaving
    StateExiting
)
```

#### Step 2: Create `ModelPickerPanel` wrapping `bubbles/list`

New file: `internal/tui/model_picker_panel.go`

```go
package tui

import (
    "github.com/charmbracelet/bubbles/list"
    tea "github.com/charmbracelet/bubbletea"
)

// modelItem implements list.Item for model names
type modelItem struct{ name string }
func (m modelItem) Title() string       { return m.name }
func (m modelItem) Description() string { return "" }
func (m modelItem) FilterValue() string { return m.name }

// ModelPickerPanel wraps bubbles/list for enum option selection
type ModelPickerPanel struct {
    list     list.Model
    selected string
}

func NewModelPickerPanel(options []string, current string) ModelPickerPanel {
    items := make([]list.Item, len(options))
    for i, opt := range options {
        items[i] = modelItem{name: opt}
    }
    l := list.New(items, list.NewDefaultDelegate(), 0, 0)
    l.Title = "Select Model"
    l.SetShowStatusBar(false)
    l.SetFilteringEnabled(true)
    // pre-select current value
    for i, opt := range options {
        if opt == current { l.Select(i); break }
    }
    return ModelPickerPanel{list: l, selected: current}
}

func (p *ModelPickerPanel) SetSize(w, h int) {
    p.list.SetSize(w, h)
}

func (p *ModelPickerPanel) Update(msg tea.Msg) tea.Cmd {
    var cmd tea.Cmd
    p.list, cmd = p.list.Update(msg)
    return cmd
}

func (p *ModelPickerPanel) SelectedValue() string {
    if item, ok := p.list.SelectedItem().(modelItem); ok {
        return item.name
    }
    return p.selected
}

func (p *ModelPickerPanel) View() string {
    return p.list.View()
}
```

#### Step 3: Add `modelPickerPanel` field to `Model`

```go
// internal/tui/model.go
type Model struct {
    // ... existing fields ...
    modelPickerPanel *ModelPickerPanel   // NEW
}
```

#### Step 4: Branch in `handleKeyPress` for `StateModelPicker`

```go
// In handleKeyPress, StateBrowsing branch:
case "enter":
    if item := m.listPanel.SelectedItem(); item != nil && !isSensitiveItem(*item) {
        if item.Field.Type == "enum" && len(item.Field.Options) >= 5 {
            // Use model picker for large enums
            current := ""
            if s, ok := item.Value.(string); ok { current = s }
            picker := NewModelPickerPanel(item.Field.Options, current)
            picker.SetSize(m.windowWidth-6, m.windowHeight-14)
            m.modelPickerPanel = &picker
            m.state = StateModelPicker
            return m, nil
        }
        // Small enums and other types: existing editing flow
        m.state = StateEditing
        return m, m.detailPanel.StartEditing()
    }

// New case StateModelPicker:
case StateModelPicker:
    if k == "esc" {
        // Confirm selection and return to browsing
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
    if k == "enter" {
        // Also confirm on enter (when not in filter mode)
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
    // All other keys go to the picker
    return m, m.modelPickerPanel.Update(msg)
```

#### Step 5: Render picker overlay in `View()`

In `StateModelPicker`, replace the two-panel section with the picker taking full inner space:

```go
// In model.go View()
if m.state == StateModelPicker && m.modelPickerPanel != nil {
    // Full-width picker overlay
    pickerContent := m.modelPickerPanel.View()
    pickerPanel := focusedPanelStyle.
        Width(innerWidth - 4).
        Height(panelHeight - 2).
        Render(pickerContent)
    // ... assemble with header + picker + helpbar
}
```

#### Step 6: Update help bar for `StateModelPicker`

```go
// internal/tui/keys.go — add Filter binding
Filter: key.NewBinding(
    key.WithKeys("/"),
    key.WithHelp("/", "filter"),
),

// internal/tui/model.go ShortHelp()
case StateModelPicker:
    return []key.Binding{k.Filter, k.Enter, k.Escape, k.Save, k.Quit}
```

### 4.3 UX Behaviour Summary

| Action | Result |
|--------|--------|
| Navigate to `model` field, press `Enter` | Opens `StateModelPicker` full-screen overlay |
| Type characters | `bubbles/list` fuzzy-filters in real time |
| Press `/` | Explicitly enter filter mode |
| Arrow keys | Navigate filtered list |
| Press `Enter` | Confirm selection, return to `StateBrowsing`, value saved to in-memory config |
| Press `Esc` | Confirm current selection (same as Enter) and return |
| Press `ctrl+s` | Save config to disk (works from any state) |
| Press `ctrl+c` | Quit (works from any state) |

### 4.4 Hardcoded Model List (Fallback)

The `model` field options are dynamically parsed from `copilot help config`. If schema detection fails, `schema = []copilot.SchemaField{}`[^24] and no `model` entry will appear. However, the implementer **may** choose to embed a hardcoded fallback list of known models (matching the 17 provided) as a guard against `copilot` binary absence. This is a separate decision.

Known models (as of this research, from `testdata/copilot-help-config.txt`[^25]):
```
claude-sonnet-4.6, claude-sonnet-4.5, claude-haiku-4.5,
claude-opus-4.6, claude-opus-4.6-fast, claude-opus-4.5,
claude-sonnet-4, gemini-3-pro-preview,
gpt-5.3-codex, gpt-5.2-codex, gpt-5.2,
gpt-5.1-codex-max, gpt-5.1-codex, gpt-5.1,
gpt-5.1-codex-mini, gpt-5-mini, gpt-4.1
```

---

## 5. Files to Create / Modify

| File | Action | Notes |
|------|--------|-------|
| `internal/tui/state.go` | Modify | Add `StateModelPicker` constant |
| `internal/tui/model_picker_panel.go` | **Create** | `ModelPickerPanel` wrapping `bubbles/list` |
| `internal/tui/model.go` | Modify | Add `modelPickerPanel` field; branch in `handleKeyPress`; update `View()` for overlay; update `updateSizes()` |
| `internal/tui/keys.go` | Modify | Add `Filter` binding; update `ShortHelp()` for `StateModelPicker` |
| `internal/tui/tui_test.go` | Modify | Tests for `StateModelPicker` transitions; `ModelPickerPanel` rendering; enum threshold branching |

No changes required to:
- `internal/config/config.go`
- `internal/copilot/copilot.go`
- `cmd/ccc/main.go`
- `internal/sensitive/`
- `internal/logging/`

---

## 6. Scope Classification

| Dimension | Classification | Rationale |
|-----------|---------------|-----------|
| **Type** | Workitem | Contained feature improvement, no new architectural pattern |
| **ADR needed?** | Optional | If the "large enum → picker" threshold is formalised as a pattern, a new ADR could capture it. Otherwise this is an implementation detail. |
| **Core Component needed?** | No | `ModelPickerPanel` is a TUI component, not a cross-cutting concern |
| **Test scope** | Unit + integration | TUI model tests; state machine tests |

---

## 7. Next WI Number

Existing workitems in `docs/workitems/`:

| Directory | Status |
|-----------|--------|
| `WI-0001-example` | Template |
| `WI-0002-ccc-bootstrap` | Complete |
| `WI-0003-tui-two-panel-redesign` | Complete |
| `WI-0004-improve-icons-and-ui` | Complete |
| `WI-0004-show-default-values` | Complete (duplicate number) |

> ⚠️ **Note:** Both `WI-0004-improve-icons-and-ui` and `WI-0004-show-default-values` share the number `0004`. The next available sequential number is **WI-0007**.

**This workitem is WI-0007** (`WI-0007-improve-model-selection`).

---

## 8. ADRs and Core Components

### ADR Consideration

**No new ADR is strictly required.** The Charm TUI stack is already ratified (ADR-0002) and the two-panel layout is established (ADR-0003). Using `bubbles/list` is already listed as Decision #13 in the decision log (for the left panel, but it was never implemented). Using it for the model picker is consistent with the existing direction.

However, formalising the rule "enum fields with ≥5 options use `bubbles/list` picker; enum fields with <5 options use inline selector" as an update to ADR-0003 or a new ADR-0004 would document the intent clearly for future contributors.

**Recommendation:** Update the decision log to add:
> Decision #17: Use `bubbles/list` with fuzzy filtering for enum fields with ≥5 options (model picker pattern)

### Core Component Consideration

No new core component is needed. `ModelPickerPanel` is a TUI-layer component analogous to `DetailPanel`. It does not need a cross-cutting specification.

---

## 9. Confidence Assessment

| Finding | Confidence | Basis |
|---------|------------|-------|
| `model` field has 17 options, type `enum`, no default | **High** | `testdata/copilot-help-config.txt:44-62`; `ParseSchema` output |
| Current enum editor has no filtering or fuzzy search | **High** | `detail_panel.go:162-173`, `detail_panel.go:264-287` |
| `bubbles/list` is available (already in go.mod) | **High** | `go.mod:6` |
| `sahilm/fuzzy` (needed for list filtering) is already in go.sum | **High** | `go.mod:29` indirect dep; all deps already resolved |
| `bubbles/list` supports built-in filtering | **High** | bubbles v0.21.x changelog and API; `SetFilteringEnabled(true)` |
| No new `go.mod` dependencies needed | **High** | All required packages already present |
| Left panel uses custom `ListPanel`, not `bubbles/list` | **High** | `list_item.go:24-100`; no bubbles/list import anywhere |
| ADR-0003 Decision #13 contradicts actual implementation | **High** | Decision log says "Use `bubbles/list` for left panel" but code uses custom struct |
| State machine can safely accommodate new `StateModelPicker` | **High** | `state.go:6-15`; iota pattern trivially extensible |
| `ctrl+s` save works from `StateModelPicker` | **High** | Save handler in `handleKeyPress` runs before state switch statement[^26] |
| `bubbles/list` `Enter` key confirm behaviour (not in filter mode) | **Medium** | Standard bubbles/list UX: Enter on selected item; need to verify if confirm requires explicit Esc from filter back to list first |
| Threshold of 5 for large enum branching | **Low/Design** | Heuristic choice; `theme` has 3 options, `banner` has 3, `model` has 17 — threshold of 4 or 5 cleanly separates them |

---

## Footnotes

[^1]: `internal/copilot/testdata/copilot-help-config.txt:44-62` — full `model` field entry with 17 option lines
[^2]: `internal/copilot/copilot.go:61-170` — `ParseSchema`; the `model` field has no "defaults to" phrase; `SchemaField.Default` will be `""`
[^3]: `internal/tui/model.go:106-113` — `isModelField()` includes `"model"` → categorized as `"Model & AI"`
[^4]: `internal/tui/model.go:196-200` — `case "enter":` transitions to `StateEditing` and calls `m.detailPanel.StartEditing()`
[^5]: `internal/tui/detail_panel.go:274-287` — `renderEditWidget()` case `"enum"`: hand-rendered option list with `▶` cursor
[^6]: `internal/tui/detail_panel.go:162-173` — `Update()` case `"enum"`: up/down by 1, no wrapping, no filtering
[^7]: `internal/tui/model.go:196-200` — `StateBrowsing` Enter key handler
[^8]: `internal/tui/model.go:203-210` — `StateEditing` Esc key handler; calls `StopEditing()` then `cfg.Set()`
[^9]: `internal/tui/detail_panel.go:120-146` — `GetValue()` case `"enum"`: returns `d.field.Options[d.selectIndex]`
[^10]: `internal/tui/list_item.go:24-100` — `ListPanel` struct with `cursor int`, `offset int`; no import of `bubbles/list`
[^11]: `docs/architecture/ADR/DECISION-LOG.md:38` — Decision #13: "Use `bubbles/list` for left panel config option navigation"
[^12]: `go.mod:6-9` — `bubbles v0.21.1-0.20250623103423-23b8fd6302d7`, `bubbletea v1.3.10`, `lipgloss v1.1.0` as direct deps
[^13]: `go.mod:29` — `github.com/sahilm/fuzzy v0.1.1 // indirect`
[^14]: `go.sum` — `github.com/sahilm/fuzzy v0.1.1 h1:...` entry confirming resolved version
[^15]: `internal/tui/model.go:144-163` — `Init()`, `Update()`, `View()` implementing `tea.Model`
[^16]: `internal/tui/model.go:17-32` — `Model` struct fields
[^17]: `internal/tui/model.go:236-239` — `leftWidth := int(float64(innerWidth) * 0.40)`
[^18]: `internal/tui/model.go:296-303` — `StateBrowsing`: `leftStyle = focusedPanelStyle`, `rightStyle = panelStyle`
[^19]: `internal/tui/model.go:299-303` — `StateEditing`: `leftStyle = panelStyle`, `rightStyle = focusedPanelStyle`
[^20]: `internal/tui/detail_panel.go:267-271` — bool toggle renders `✓ Yes` / `✗ No`
[^21]: `internal/tui/detail_panel.go:96-108` — `StartEditing()` for string: `d.textInput.Focus()`, returns `textinput.Blink`
[^22]: `internal/tui/detail_panel.go:103-108` — `StartEditing()` for list: `d.textArea.Focus()`, returns `textarea.Blink`
[^23]: `internal/tui/detail_panel.go:264-287` — enum case in `renderEditWidget()`
[^24]: `cmd/ccc/main.go:69-74` — `schema = []copilot.SchemaField{}` on `DetectSchema` error
[^25]: `internal/copilot/testdata/copilot-help-config.txt:45-61` — 17 model option lines
[^26]: `internal/tui/model.go:169-184` — `ctrl+c` and `ctrl+s` handled before the `switch m.state` block, so they work globally
