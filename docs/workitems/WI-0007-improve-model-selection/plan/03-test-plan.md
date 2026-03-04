# Test Plan: WI-0007 — Improve Model Selection UX

**Workitem:** WI-0007-improve-model-selection
**Task Breakdown:** [02-task-breakdown.md](./02-task-breakdown.md)
**Test File:** `internal/tui/tui_test.go`
**Test Runner:** `go test ./internal/tui/...`

---

## Overview

This test plan adds 8 new unit tests (UT-TUI-026 through UT-TUI-033) to the existing test suite. The existing 25 tests (UT-TUI-001 through UT-TUI-025) are unchanged. All tests reside in `internal/tui/tui_test.go`.

**Test categories:**
- **State transitions** (UT-TUI-026, UT-TUI-027): Verify the enum option count threshold routes to the correct state.
- **Component construction** (UT-TUI-028, UT-TUI-029): Verify `ModelPickerPanel` creates correctly and renders.
- **Picker interactions** (UT-TUI-030, UT-TUI-031): Verify Enter/Esc from `StateModelPicker` confirm selection and write to config.
- **View rendering** (UT-TUI-032): Verify `View()` renders without panic at various terminal sizes.
- **Global key handling** (UT-TUI-033): Verify `ctrl+s` save works from `StateModelPicker`.

---

## Test UT-TUI-026: Enter on large enum transitions to StateModelPicker

- **Type:** Unit
- **Task:** Task 4 (Wire StateModelPicker into main model)
- **Priority:** Critical

### Setup

```go
cfg := config.NewConfig()
cfg.Set("model", "gpt-4")

schema := []copilot.SchemaField{
    {Name: "model", Type: "enum", Default: "",
     Options: []string{
         "claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
         "claude-opus-4.6", "claude-opus-4.6-fast", "gpt-4",
     },
     Description: "AI model"},
}

model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
model.windowWidth = 100
model.windowHeight = 30
model.updateSizes()
```

- Schema has `model` field with 6 options (≥5 threshold).
- Navigate cursor so `model` is selected (it should be the first non-header config item).

### Steps

1. Verify initial state is `StateBrowsing`.
2. Verify `listPanel.SelectedItem()` returns the `model` field.
3. Send `tea.KeyMsg{Type: tea.KeyEnter}` via `model.Update(msg)`.
4. Cast the returned `tea.Model` to `*Model`.

### Expected Result

- `m.state == StateModelPicker` — transitioned to picker, not editing.
- `m.modelPickerPanel != nil` — picker was created.
- `m.modelPickerPanel.SelectedValue()` returns `"gpt-4"` (the current config value was pre-selected).

---

## Test UT-TUI-027: Enter on small enum transitions to StateEditing

- **Type:** Unit
- **Task:** Task 4 (Wire StateModelPicker into main model)
- **Priority:** Critical

### Setup

```go
cfg := config.NewConfig()
cfg.Set("theme", "dark")

schema := []copilot.SchemaField{
    {Name: "theme", Type: "enum", Default: "auto",
     Options: []string{"auto", "dark", "light"},
     Description: "Color theme"},
}

model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
model.windowWidth = 100
model.windowHeight = 30
model.updateSizes()
```

- Schema has `theme` field with 3 options (<5 threshold).

### Steps

1. Verify initial state is `StateBrowsing`.
2. Verify `listPanel.SelectedItem()` returns the `theme` field.
3. Send `tea.KeyMsg{Type: tea.KeyEnter}` via `model.Update(msg)`.
4. Cast the returned `tea.Model` to `*Model`.

### Expected Result

- `m.state == StateEditing` — transitioned to editing, not picker.
- `m.modelPickerPanel == nil` — no picker was created.
- The `DetailPanel` is in editing mode (existing behaviour preserved).

---

## Test UT-TUI-028: ModelPickerPanel creation and pre-selection

- **Type:** Unit
- **Task:** Task 2 (Create ModelPickerPanel component)
- **Priority:** High

### Setup

```go
options := []string{
    "claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
    "claude-opus-4.6", "claude-opus-4.6-fast", "claude-opus-4.5",
    "claude-sonnet-4", "gemini-3-pro-preview",
    "gpt-5.3-codex", "gpt-5.2-codex", "gpt-5.2",
    "gpt-5.1-codex-max", "gpt-5.1-codex", "gpt-5.1",
    "gpt-5.1-codex-mini", "gpt-5-mini", "gpt-4.1",
}
current := "claude-sonnet-4.5"
```

### Steps

1. Call `picker := NewModelPickerPanel(options, current)`.
2. Call `picker.SetSize(60, 20)`.
3. Read `picker.SelectedValue()`.

### Expected Result

- `picker.SelectedValue() == "claude-sonnet-4.5"` — the current value was pre-selected.
- The picker was created without panics.
- The underlying list has all 17 items.

---

## Test UT-TUI-029: ModelPickerPanel.View() renders non-empty string

- **Type:** Unit
- **Task:** Task 2 (Create ModelPickerPanel component)
- **Priority:** High

### Setup

```go
options := []string{
    "claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
    "gpt-5.1-codex", "gpt-4.1",
}
picker := NewModelPickerPanel(options, "gpt-4.1")
picker.SetSize(60, 20)
```

### Steps

1. Call `view := picker.View()`.

### Expected Result

- `view != ""` — the view renders a non-empty string.
- `view` contains `"Select Model"` (the list title).
- No panic occurs during rendering.

---

## Test UT-TUI-030: Esc from StateModelPicker confirms and returns to StateBrowsing

- **Type:** Unit
- **Task:** Task 4 (Wire StateModelPicker into main model)
- **Priority:** Critical

### Setup

```go
cfg := config.NewConfig()
cfg.Set("model", "gpt-4.1")

schema := []copilot.SchemaField{
    {Name: "model", Type: "enum", Default: "",
     Options: []string{
         "claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
         "claude-opus-4.6", "claude-opus-4.6-fast", "gpt-4.1",
     },
     Description: "AI model"},
}

model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
model.windowWidth = 100
model.windowHeight = 30
model.updateSizes()
```

### Steps

1. Send `tea.KeyMsg{Type: tea.KeyEnter}` to enter `StateModelPicker`.
2. Verify state is `StateModelPicker`.
3. Send `tea.KeyMsg{Type: tea.KeyEsc}` to exit the picker.
4. Cast the returned `tea.Model` to `*Model`.

### Expected Result

- `m.state == StateBrowsing` — returned to browsing.
- `m.modelPickerPanel == nil` — picker was cleaned up.
- `cfg.Get("model")` returns the selected value (the current value since no navigation was done in the picker, so it should still be `"gpt-4.1"`).
- `m.listPanel.SelectedItem().Value` reflects the updated value.

---

## Test UT-TUI-031: Enter from StateModelPicker (not filtering) confirms and returns

- **Type:** Unit
- **Task:** Task 4 (Wire StateModelPicker into main model)
- **Priority:** Critical

### Setup

```go
cfg := config.NewConfig()
cfg.Set("model", "gpt-4.1")

schema := []copilot.SchemaField{
    {Name: "model", Type: "enum", Default: "",
     Options: []string{
         "claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
         "claude-opus-4.6", "claude-opus-4.6-fast", "gpt-4.1",
     },
     Description: "AI model"},
}

model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
model.windowWidth = 100
model.windowHeight = 30
model.updateSizes()
```

### Steps

1. Send `tea.KeyMsg{Type: tea.KeyEnter}` to enter `StateModelPicker`.
2. Verify state is `StateModelPicker`.
3. Send `tea.KeyMsg{Type: tea.KeyEnter}` again to confirm selection.
4. Cast the returned `tea.Model` to `*Model`.

### Expected Result

- `m.state == StateBrowsing` — returned to browsing after confirmation.
- `m.modelPickerPanel == nil` — picker was cleaned up.
- `cfg.Get("model")` returns the selected value.
- Value was written via `cfg.Set()` and reflected in both `ListPanel` and `DetailPanel`.

---

## Test UT-TUI-032: View() in StateModelPicker at various sizes renders without panic

- **Type:** Unit
- **Task:** Task 4 (Wire StateModelPicker into main model)
- **Priority:** High

### Setup

```go
sizes := []struct{ width, height int }{
    {80, 24},   // minimum standard terminal
    {120, 40},  // large terminal
    {100, 30},  // medium terminal
}

cfg := config.NewConfig()
cfg.Set("model", "gpt-4.1")

schema := []copilot.SchemaField{
    {Name: "model", Type: "enum", Default: "",
     Options: []string{
         "claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
         "claude-opus-4.6", "claude-opus-4.6-fast", "gpt-4.1",
     },
     Description: "AI model"},
}
```

### Steps

For each size `(width, height)`:
1. Create a fresh `NewModel(cfg, schema, ...)`.
2. Send `tea.WindowSizeMsg{Width: width, Height: height}`.
3. Send `tea.KeyMsg{Type: tea.KeyEnter}` to enter `StateModelPicker`.
4. Verify state is `StateModelPicker`.
5. Call `m.View()`.

### Expected Result

- `m.View() != ""` — renders a non-empty string at each size.
- No panic occurs at any size.
- The output contains at least one model name (e.g., `"claude-sonnet-4.6"` or `"gpt-4.1"`).
- The rendered output contains the help bar frame character `"╭"` (outer frame border is present).

---

## Test UT-TUI-033: ctrl+s save works from StateModelPicker

- **Type:** Unit
- **Task:** Task 4 (Wire StateModelPicker into main model)
- **Priority:** High

### Setup

```go
cfg := config.NewConfig()
cfg.Set("model", "gpt-4.1")

schema := []copilot.SchemaField{
    {Name: "model", Type: "enum", Default: "",
     Options: []string{
         "claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
         "claude-opus-4.6", "claude-opus-4.6-fast", "gpt-4.1",
     },
     Description: "AI model"},
}

model := NewModel(cfg, schema, "0.0.412", "/tmp/nonexistent/config.json")
model.windowWidth = 100
model.windowHeight = 30
model.updateSizes()
```

Note: Using a non-existent path ensures save will set `m.err` (save fails gracefully), which is sufficient to verify the `ctrl+s` handler was reached.

### Steps

1. Send `tea.KeyMsg{Type: tea.KeyEnter}` to enter `StateModelPicker`.
2. Verify state is `StateModelPicker`.
3. Send `tea.KeyMsg{Type: tea.KeyCtrlS}` to trigger save.
4. Cast the returned `tea.Model` to `*Model`.

### Expected Result

- `m.state == StateModelPicker` — state did NOT change (save does not exit the picker).
- `m.err != nil` OR `m.saved == true` — save handler was executed (either succeeded or failed, proving it was reached).
- The `ctrl+s` handler runs before the `switch m.state` block, confirming global key handling works from any state.

---

## Traceability Matrix

| Test ID | Task | Category | Verifies |
|---------|------|----------|----------|
| UT-TUI-026 | Task 4 | State transition | Large enum → `StateModelPicker` |
| UT-TUI-027 | Task 4 | State transition | Small enum → `StateEditing` (unchanged) |
| UT-TUI-028 | Task 2 | Component | `ModelPickerPanel` creation and pre-selection |
| UT-TUI-029 | Task 2 | Component | `ModelPickerPanel.View()` rendering |
| UT-TUI-030 | Task 4 | Interaction | Esc confirms selection, returns to browsing |
| UT-TUI-031 | Task 4 | Interaction | Enter confirms selection, returns to browsing |
| UT-TUI-032 | Task 4, 5 | Rendering | `View()` at multiple terminal sizes |
| UT-TUI-033 | Task 4 | Global keys | `ctrl+s` save works from picker state |

## Regression Coverage

The following existing tests ensure no regressions:

| Test ID | What it guards |
|---------|---------------|
| UT-TUI-007 | `StateBrowsing` → `StateEditing` transition (must still work for all non-picker fields) |
| UT-TUI-008 | `StateEditing` → `StateBrowsing` transition (unchanged) |
| UT-TUI-011 | `DetailPanel` renders field information (unchanged) |
| UT-TUI-018 | `View()` at 80×24 (unchanged for `StateBrowsing`) |
| UT-TUI-019 | Panel height overhead arithmetic (unchanged) |
| UT-TUI-020 | Small terminal floor guard (unchanged) |
| UT-TUI-025 | `View()` renders without panic (unchanged for `StateBrowsing`) |

## Running the Tests

```bash
# Run all TUI tests
go test ./internal/tui/... -v

# Run only the new model picker tests
go test ./internal/tui/... -v -run "TestEnterOnLargeEnum|TestEnterOnSmallEnum|TestModelPickerPanel|TestEscFromModelPicker|TestEnterFromModelPicker|TestViewInModelPicker|TestCtrlSSaveFromModelPicker"

# Run with race detector
go test ./internal/tui/... -race -v
```
