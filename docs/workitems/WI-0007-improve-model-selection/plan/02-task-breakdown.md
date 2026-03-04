# Task Breakdown: WI-0007 — Improve Model Selection UX

**Workitem:** WI-0007-improve-model-selection
**Action Plan:** [01-action-plan.md](./01-action-plan.md)
**Total Estimated Lines:** ~260

---

## Task 1: Add `StateModelPicker` to state machine

- **Status:** Not Started
- **Complexity:** Low (~5 lines)
- **Dependencies:** None
- **Related ADRs:** ADR-0003 (Two-Panel TUI Layout — state machine extension)
- **Related Core-Components:** None

### Description

Add a new `StateModelPicker` constant to the `State` iota block in `internal/tui/state.go`. Insert it between `StateEditing` and `StateSaving` to maintain logical ordering. Add the corresponding `"ModelPicker"` case to the `String()` method.

**File:** `internal/tui/state.go` (Modify)

**Changes:**
1. Add `StateModelPicker` constant after `StateEditing` in the `const` iota block.
2. Add `case StateModelPicker: return "ModelPicker"` to the `String()` switch statement.

### Acceptance Criteria

- [ ] `StateModelPicker` constant exists in the `State` iota block, positioned between `StateEditing` and `StateSaving`.
- [ ] `StateModelPicker.String()` returns `"ModelPicker"`.
- [ ] Existing states (`StateBrowsing`, `StateEditing`, `StateSaving`, `StateExiting`) are unchanged.
- [ ] Code compiles without errors.

### Test Coverage

- Covered indirectly by UT-TUI-026 (state transition to `StateModelPicker`) and UT-TUI-027 (non-transition for small enums).
- No dedicated unit test needed for the constant itself; verified through state machine integration tests.

---

## Task 2: Create `ModelPickerPanel` component

- **Status:** Not Started
- **Complexity:** Medium (~70 lines, new file)
- **Dependencies:** Task 1 (uses `StateModelPicker` conceptually, but the component itself is state-agnostic)
- **Related ADRs:** ADR-0002 (Go with Charm TUI Stack — `bubbles/list`), ADR-0003 (Two-Panel TUI Layout — component pattern)
- **Related Core-Components:** None

### Description

Create a new file `internal/tui/model_picker_panel.go` containing the `ModelPickerPanel` component that wraps `bubbles/list.Model` with fuzzy filtering.

**File:** `internal/tui/model_picker_panel.go` (Create)

**Structs & Types:**
1. `modelItem` — implements `list.Item` interface:
   - `Title() string` → returns the model name
   - `Description() string` → returns `""` (no descriptions for model options)
   - `FilterValue() string` → returns the model name (used by `sahilm/fuzzy` for filtering)
2. `ModelPickerPanel` — wraps the bubbles list:
   - `list list.Model` — the bubbles list component
   - `selected string` — the value when the picker was opened (fallback)

**Constructor:**
- `NewModelPickerPanel(options []string, current string) ModelPickerPanel`:
  - Converts `options` to `[]list.Item` (as `modelItem` values)
  - Creates `list.New(items, list.NewDefaultDelegate(), 0, 0)`
  - Configures: `SetShowStatusBar(false)`, `SetFilteringEnabled(true)`, `SetShowHelp(false)`
  - Sets `l.Title = "Select Model"`
  - Pre-selects current value by calling `l.Select(i)` where `options[i] == current`

**Methods:**
- `SetSize(w, h int)` — delegates to `list.SetSize(w, h)`
- `Update(msg tea.Msg) tea.Cmd` — delegates to `list.Update(msg)`
- `SelectedValue() string` — returns `list.SelectedItem().(modelItem).name` or fallback `selected`
- `View() string` — returns `list.View()`

**Styling:**
- Customize the default delegate's `Styles.SelectedTitle` to use `primaryColor` from `styles.go` for visual consistency with the rest of the TUI.

### Acceptance Criteria

- [ ] File `internal/tui/model_picker_panel.go` exists in the `tui` package.
- [ ] `modelItem` implements `list.Item` (has `Title()`, `Description()`, `FilterValue()` methods).
- [ ] `NewModelPickerPanel(options, current)` creates a panel with all options present in the list.
- [ ] `NewModelPickerPanel` pre-selects the `current` value when it matches an option.
- [ ] `SelectedValue()` returns the currently highlighted item's name.
- [ ] `SelectedValue()` returns the `selected` fallback when no item is highlighted.
- [ ] `SetShowHelp(false)` is set so the component's built-in help does not conflict with the app's help bar.
- [ ] `SetFilteringEnabled(true)` enables the built-in fuzzy filter.
- [ ] Delegate styling uses `primaryColor` for the selected item highlight.
- [ ] No new dependencies added to `go.mod` — uses existing `bubbles/list` and indirect `sahilm/fuzzy`.

### Test Coverage

- **UT-TUI-028:** `NewModelPickerPanel` creates panel with all options; `SelectedValue()` returns current.
- **UT-TUI-029:** `ModelPickerPanel.View()` renders a non-empty string.

---

## Task 3: Add `Filter` key binding

- **Status:** Not Started
- **Complexity:** Low (~5 lines)
- **Dependencies:** None
- **Related ADRs:** ADR-0003 (Two-Panel TUI Layout — keyboard navigation)
- **Related Core-Components:** None

### Description

Add a `Filter` key binding to the `KeyMap` struct in `internal/tui/keys.go` so it can be shown in the help bar during `StateModelPicker`.

**File:** `internal/tui/keys.go` (Modify)

**Changes:**
1. Add `Filter key.Binding` field to the `KeyMap` struct.
2. Initialize it in `DefaultKeyMap()`:
   ```go
   Filter: key.NewBinding(
       key.WithKeys("/"),
       key.WithHelp("/", "filter"),
   ),
   ```

### Acceptance Criteria

- [ ] `KeyMap` struct has a `Filter` field of type `key.Binding`.
- [ ] `DefaultKeyMap().Filter` is bound to `/` with help text `"/ filter"`.
- [ ] Existing key bindings (`Up`, `Down`, `Enter`, `Escape`, `Save`, `Quit`, `Tab`) are unchanged.
- [ ] Code compiles without errors.

### Test Coverage

- Covered indirectly through UT-TUI-033 (help bar rendering for `StateModelPicker` state includes the filter binding).
- No dedicated unit test needed for the binding definition itself.

---

## Task 4: Wire `StateModelPicker` into main model

- **Status:** Not Started
- **Complexity:** High (~60 lines across 6 sub-tasks)
- **Dependencies:** Task 1, Task 2, Task 3
- **Related ADRs:** ADR-0002 (Charm TUI stack — Elm architecture), ADR-0003 (Two-Panel TUI Layout — state machine, view rendering)
- **Related Core-Components:** CC-0004 (Configuration Management — `cfg.Set()`), CC-0005 (Sensitive Data Handling — sensitive field guard unchanged)

### Description

Integrate the `ModelPickerPanel` into the main `Model` in `internal/tui/model.go`. This is the largest task and has 6 sub-changes:

**File:** `internal/tui/model.go` (Modify)

**4a. Add field to `Model` struct:**
- Add `modelPickerPanel *ModelPickerPanel` field.

**4b. Branch in `handleKeyPress` — `StateBrowsing` / `Enter`:**
- Before the existing `StateEditing` transition for `case "enter":`, check if the field is an enum with `len(item.Field.Options) >= 5`.
- If so, create a `NewModelPickerPanel`, size it, assign to `m.modelPickerPanel`, transition to `StateModelPicker`, and return.
- If not, fall through to the existing `StateEditing` transition (no change for small enums, bools, strings, lists).
- The `isSensitiveItem()` guard must still block entry to the picker for sensitive fields (unchanged — it runs before the enum check).

**4c. Add `StateModelPicker` case in `handleKeyPress`:**
- `Enter` key: If `FilterState() != list.Filtering`, confirm the selection (read `SelectedValue()`, write via `cfg.Set()`, update `ListPanel` and `DetailPanel`, clear picker, transition to `StateBrowsing`). If filtering is active, forward Enter to the list (to confirm the filter text).
- `Esc` key: Always confirm the current selection and return to `StateBrowsing` (regardless of filter state). This is the simpler approach recommended in the action plan.
- All other keys: Forward to `m.modelPickerPanel.Update(msg)`.

**4d. Forward non-key messages in `Update()`:**
- When `m.state == StateModelPicker && m.modelPickerPanel != nil`, forward the message to `m.modelPickerPanel.Update(msg)` (analogous to the existing `StateEditing` forwarding for blink timers).

**4e. Update `View()` — render picker overlay:**
- When `state == StateModelPicker`, replace the two-panel `panels` variable with a single full-width picker panel rendered using `focusedPanelStyle`:
  ```go
  pickerContent := m.modelPickerPanel.View()
  panels = focusedPanelStyle.
      Width(innerWidth - 4).
      Height(panelHeight - 2).
      Render(pickerContent)
  ```

**4f. Update `updateSizes()`:**
- After the existing list/detail sizing, add:
  ```go
  if m.modelPickerPanel != nil {
      m.modelPickerPanel.SetSize(innerWidth-4, panelHeight-2)
  }
  ```

### Acceptance Criteria

- [ ] `Model` struct has a `modelPickerPanel *ModelPickerPanel` field.
- [ ] Pressing Enter on an enum field with ≥5 options (and not sensitive) transitions to `StateModelPicker` and creates the picker.
- [ ] Pressing Enter on an enum field with <5 options still transitions to `StateEditing` (unchanged behaviour).
- [ ] Pressing Enter on a sensitive field does nothing (unchanged behaviour — `isSensitiveItem()` guard).
- [ ] In `StateModelPicker`, pressing Enter (when not filtering) confirms selection, writes to `cfg`, updates both panels, clears the picker, and returns to `StateBrowsing`.
- [ ] In `StateModelPicker`, pressing Enter while filtering (`FilterState() == list.Filtering`) forwards Enter to the list (confirms filter text only, does not exit picker).
- [ ] In `StateModelPicker`, pressing Esc confirms current selection and returns to `StateBrowsing`.
- [ ] In `StateModelPicker`, all other keys are forwarded to the picker's `Update()`.
- [ ] Non-key messages (e.g., blink timers) in `StateModelPicker` are forwarded to the picker's `Update()`.
- [ ] Global keys (`ctrl+s` save, `ctrl+c` quit) continue to work from `StateModelPicker` (they are handled before the `switch m.state` block — no change needed).
- [ ] `View()` in `StateModelPicker` renders a full-width picker overlay replacing both panels.
- [ ] `updateSizes()` propagates size changes to `modelPickerPanel` when it is non-nil.
- [ ] The `DetailPanel` code is unchanged — no modifications to `detail_panel.go`.
- [ ] The `ListPanel` code is unchanged — no modifications to `list_item.go`.

### Test Coverage

- **UT-TUI-026:** Enter on enum with ≥5 options transitions to `StateModelPicker`.
- **UT-TUI-027:** Enter on enum with <5 options transitions to `StateEditing` (not picker).
- **UT-TUI-030:** Esc from `StateModelPicker` writes selected value to config, returns to `StateBrowsing`.
- **UT-TUI-031:** Enter from `StateModelPicker` (not filtering) confirms selection, returns to `StateBrowsing`.
- **UT-TUI-032:** `View()` in `StateModelPicker` at various sizes renders without panic.
- **UT-TUI-033:** `ctrl+s` save works from `StateModelPicker` state.

---

## Task 5: Update help bar for `StateModelPicker`

- **Status:** Not Started
- **Complexity:** Low (~3 lines)
- **Dependencies:** Task 3 (Filter binding), Task 4 (StateModelPicker wiring)
- **Related ADRs:** ADR-0003 (Two-Panel TUI Layout — help bar pattern)
- **Related Core-Components:** None

### Description

Add a `StateModelPicker` case to the `ShortHelp()` method in `internal/tui/model.go` so the help bar displays the correct key bindings when the picker is active.

**File:** `internal/tui/model.go` (Modify — in `ShortHelp()` method)

**Change:**
Add a new case before `default`:
```go
case StateModelPicker:
    return []key.Binding{k.Filter, k.Enter, k.Escape, k.Save, k.Quit}
```

This shows: `/ filter  •  enter confirm  •  esc back  •  ctrl+s save  •  ctrl+c quit`.

### Acceptance Criteria

- [ ] `ShortHelp(StateModelPicker)` returns exactly `[Filter, Enter, Escape, Save, Quit]` bindings.
- [ ] `ShortHelp(StateBrowsing)` is unchanged: `[Up, Down, Enter, Save, Quit]`.
- [ ] `ShortHelp(StateEditing)` is unchanged: `[Escape, Save, Quit]`.
- [ ] The help bar renders correctly in the `StateModelPicker` view.

### Test Coverage

- Covered by UT-TUI-033 (verifying save works from `StateModelPicker` and help bar includes expected keys).
- Visual verification through UT-TUI-032 (`View()` rendering in `StateModelPicker` includes the help bar).

---

## Task 6: Add unit tests

- **Status:** Not Started
- **Complexity:** Medium (~120 lines)
- **Dependencies:** Task 1, Task 2, Task 3, Task 4, Task 5
- **Related ADRs:** ADR-0002 (Go with Charm TUI Stack — `go test`), ADR-0003 (Two-Panel TUI Layout — state machine)
- **Related Core-Components:** CC-0004 (Configuration Management — verifying `cfg.Set()` writes)

### Description

Add 8 new unit tests to `internal/tui/tui_test.go` following the existing `UT-TUI-###` naming pattern. Tests continue from UT-TUI-026 (the last existing test is UT-TUI-025).

**File:** `internal/tui/tui_test.go` (Modify)

| Test ID | Test Function | Description |
|---------|--------------|-------------|
| UT-TUI-026 | `TestEnterOnLargeEnumTransitionsToModelPicker` | Enter on enum ≥5 options → `StateModelPicker` |
| UT-TUI-027 | `TestEnterOnSmallEnumTransitionsToEditing` | Enter on enum <5 options → `StateEditing` (not picker) |
| UT-TUI-028 | `TestModelPickerPanelCreation` | `NewModelPickerPanel` creates panel; `SelectedValue()` returns current |
| UT-TUI-029 | `TestModelPickerPanelView` | `ModelPickerPanel.View()` renders non-empty string |
| UT-TUI-030 | `TestEscFromModelPickerConfirmsAndReturns` | Esc writes value to config, returns to `StateBrowsing` |
| UT-TUI-031 | `TestEnterFromModelPickerConfirmsAndReturns` | Enter (not filtering) confirms, returns to `StateBrowsing` |
| UT-TUI-032 | `TestViewInModelPickerAtVariousSizes` | `View()` in `StateModelPicker` at 80×24, 120×40, 100×30 renders without panic |
| UT-TUI-033 | `TestCtrlSSaveFromModelPicker` | `ctrl+s` save works from `StateModelPicker` state |

**Test setup pattern** (consistent with existing tests):
- Create a `config.NewConfig()` with `cfg.Set("model", "gpt-4")`.
- Use a schema with `model` field having ≥5 options for picker tests, or <5 options for small-enum tests.
- Set `model.windowWidth` and `model.windowHeight` and call `model.updateSizes()` before interactions.
- Send `tea.KeyMsg` via `model.Update(msg)` and assert on the resulting `*Model` state.

### Acceptance Criteria

- [ ] All 8 tests are added to `internal/tui/tui_test.go`.
- [ ] Each test has a comment with its `UT-TUI-###` ID following the existing pattern.
- [ ] All tests pass with `go test ./internal/tui/...`.
- [ ] No existing tests (UT-TUI-001 through UT-TUI-025) are modified or broken.
- [ ] Tests verify state transitions, config writes, view rendering, and global key handling.

### Test Coverage

- This task IS the test coverage. See the Test Plan (`03-test-plan.md`) for detailed test specifications.

---

## Dependency Graph

```
Task 1 (State constant)  ──┐
Task 2 (ModelPickerPanel) ──┼──▶ Task 4 (Wire into model.go) ──▶ Task 5 (Help bar) ──▶ Task 6 (Tests)
Task 3 (Filter binding)  ──┘
```

Tasks 1, 2, and 3 have no dependencies on each other and can be implemented in parallel. Task 4 depends on all three. Task 5 depends on Tasks 3 and 4. Task 6 depends on all previous tasks.

## Files Summary

| File | Action | Task(s) | Est. Lines |
|------|--------|---------|-----------|
| `internal/tui/state.go` | Modify | Task 1 | ~5 |
| `internal/tui/model_picker_panel.go` | **Create** | Task 2 | ~70 |
| `internal/tui/keys.go` | Modify | Task 3 | ~5 |
| `internal/tui/model.go` | Modify | Task 4, Task 5 | ~63 |
| `internal/tui/tui_test.go` | Modify | Task 6 | ~120 |
| **Total** | | | **~263** |

**Files NOT modified:** `detail_panel.go`, `list_item.go`, `styles.go`, `internal/config/`, `internal/copilot/`, `cmd/`, `internal/sensitive/`, `internal/logging/`.
