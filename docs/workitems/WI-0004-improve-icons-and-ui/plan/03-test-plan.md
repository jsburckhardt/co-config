# Test Plan: WI-0004 — Improve Icons and UI

- **Work Item:** WI-0004-improve-icons-and-ui
- **Task Breakdown:** [02-task-breakdown.md](./02-task-breakdown.md)
- **Test File:** `internal/tui/tui_test.go`
- **Total Tests:** 8 automated + 1 manual verification
- **Run Command:** `go test ./internal/tui/... -v`

---

## Test T-01: Copilot icon constant structure

- **Type:** Unit
- **Task:** Task 1
- **Priority:** High

### Setup

No model setup needed. Tests the `copilotIcon` package-level constant directly.

```go
import (
    "strings"
    "testing"
)
```

### Steps

1. Split `copilotIcon` by `\n` and count the resulting lines
2. For each line, measure the character-cell width (use `len()` since all characters are in the BMP and are 1-cell wide in the icon)
3. Verify each line is exactly 6 cells wide

```go
func TestCopilotIconConstant(t *testing.T) {
    lines := strings.Split(copilotIcon, "\n")
    if len(lines) != 4 {
        t.Errorf("copilotIcon should have 4 lines, got %d", len(lines))
    }
    for i, line := range lines {
        // Each line should be 6 runes wide
        w := len([]rune(line))
        if w != 6 {
            t.Errorf("copilotIcon line %d width: got %d runes, want 6 (line=%q)", i, w, line)
        }
    }
}
```

### Expected Result

- `copilotIcon` has exactly 4 lines
- Each line is exactly 6 runes wide
- No empty or trailing lines

---

## Test T-02: Gear emoji removed from rendered output

- **Type:** Unit
- **Task:** Task 3
- **Priority:** High

### Setup

Create a model with a schema, set window size, and call `View()`.

```go
func TestViewNoGearEmoji(t *testing.T) {
    cfg := config.NewConfig()
    cfg.Set("model", "gpt-4")
    schema := []copilot.SchemaField{
        {Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
    }
    model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
    model.windowWidth = 100
    model.windowHeight = 30
    model.updateSizes()
```

### Steps

1. Call `model.View()` to get the rendered output string
2. Search for the `⚙` character (U+2699) in the output
3. Assert it is NOT present

```go
    view := model.View()
    if strings.Contains(view, "⚙") {
        t.Error("View() should not contain the gear emoji ⚙ after icon replacement")
    }
}
```

### Expected Result

- The rendered output does NOT contain `⚙` anywhere
- Confirms Task 3 successfully removed the old icon

---

## Test T-03: ASCII icon and title present in rendered output

- **Type:** Unit
- **Task:** Task 3
- **Priority:** High

### Setup

Same model setup as T-02.

```go
func TestViewRendersWithIcon(t *testing.T) {
    cfg := config.NewConfig()
    cfg.Set("model", "gpt-4")
    schema := []copilot.SchemaField{
        {Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
    }
    model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
    model.windowWidth = 100
    model.windowHeight = 30
    model.updateSizes()
```

### Steps

1. Call `model.View()` to get the rendered output
2. Check that the first line of the icon (`╭─╮╭─╮`) appears in the output
3. Check that the title text `ccc — Copilot Config CLI` appears in the output
4. Check that the version string appears in the output

```go
    view := model.View()
    if !strings.Contains(view, "╭─╮╭─╮") {
        t.Error("View() should contain the first line of the copilot icon")
    }
    if !strings.Contains(view, "ccc") {
        t.Error("View() should contain the title text")
    }
    if !strings.Contains(view, "0.0.412") {
        t.Error("View() should contain the version string")
    }
}
```

### Expected Result

- View output contains `╭─╮╭─╮` (icon line 1)
- View output contains the title text
- View output contains the version number
- Confirms the icon and title are rendered via horizontal join

---

## Test T-04: Framed header contains border characters

- **Type:** Unit
- **Task:** Task 4
- **Priority:** High

### Setup

Same model setup as T-02. The framed header wraps `headerContent` in `headerFrameStyle` which uses `lipgloss.RoundedBorder()`.

### Steps

1. Call `model.View()` to get the rendered output
2. Count occurrences of rounded border corner characters (`╭`, `╮`, `╰`, `╯`) in the output
3. The output should have these characters from: outer frame (1 set), header frame (1 set), help bar frame (1 set), left panel (1 set), right panel (1 set) — at least 5 sets of corners

```go
func TestViewRendersFramedHeader(t *testing.T) {
    cfg := config.NewConfig()
    cfg.Set("model", "gpt-4")
    schema := []copilot.SchemaField{
        {Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
    }
    model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
    model.windowWidth = 100
    model.windowHeight = 30
    model.updateSizes()

    view := model.View()

    // Before this change: outer frame (4 corners) + 2 panels (8 corners) = 12 corners
    // After this change: outer frame (4) + header frame (4) + help bar frame (4) + 2 panels (8) = 20 corners
    // Count ╭ (top-left corner) occurrences — should be at least 5 (outer + header + help + 2 panels)
    topLeftCount := strings.Count(view, "╭")
    if topLeftCount < 5 {
        t.Errorf("Expected at least 5 top-left corners (╭) for framed sections, got %d", topLeftCount)
    }
}
```

### Expected Result

- At least 5 `╭` characters in the output (outer frame, header frame, help bar frame, left panel, right panel)
- Confirms both header and help bar are wrapped in frame styles with rounded borders

---

## Test T-05: Framed help bar renders

- **Type:** Unit
- **Task:** Task 4
- **Priority:** Medium

### Setup

Same model setup as T-02.

### Steps

1. Call `model.View()` to get rendered output
2. Verify help bar content (key binding descriptions) appears within the output
3. Verify the help bar content is surrounded by border characters

```go
func TestViewRendersFramedHelpBar(t *testing.T) {
    cfg := config.NewConfig()
    cfg.Set("model", "gpt-4")
    schema := []copilot.SchemaField{
        {Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
    }
    model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
    model.windowWidth = 100
    model.windowHeight = 30
    model.updateSizes()

    view := model.View()
    // Help bar should contain key binding descriptions
    if !strings.Contains(view, "ctrl+c") {
        t.Error("View() should contain help bar key binding 'ctrl+c'")
    }
    if !strings.Contains(view, "ctrl+s") {
        t.Error("View() should contain help bar key binding 'ctrl+s'")
    }
}
```

### Expected Result

- Help bar text with key bindings is present in rendered output
- Key bindings `ctrl+c` and `ctrl+s` are visible

---

## Test T-06: View renders at 80×24 without panic (regression)

- **Type:** Unit (Regression)
- **Task:** Tasks 3, 4, 5, 6
- **Priority:** Critical

### Setup

Create model with schema, set window size to exactly 80×24 (the ADR-0003 target resolution).

### Steps

1. Create model with at least 2 schema fields
2. Set `windowWidth = 80`, `windowHeight = 24`
3. Call `model.updateSizes()`
4. Call `model.View()` — must not panic
5. Verify output is non-empty

```go
func TestViewRendersAt80x24(t *testing.T) {
    cfg := config.NewConfig()
    cfg.Set("model", "gpt-4")
    cfg.Set("theme", "dark")
    schema := []copilot.SchemaField{
        {Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
        {Name: "theme", Type: "enum", Default: "auto", Options: []string{"auto", "dark"}},
    }
    model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
    model.windowWidth = 80
    model.windowHeight = 24
    model.updateSizes()

    view := model.View()
    if view == "" {
        t.Error("View() at 80x24 returned empty string")
    }
}
```

### Expected Result

- `View()` returns a non-empty string at 80×24
- No panic occurs
- Confirms the overhead arithmetic doesn't cause negative panel heights or other rendering failures at the target resolution

---

## Test T-07: Panel height overhead arithmetic

- **Type:** Unit
- **Task:** Tasks 5, 6
- **Priority:** Critical

### Setup

Create a model and send `WindowSizeMsg` at known sizes to trigger `updateSizes()`.

### Steps

1. Create model with a single schema field
2. Send `tea.WindowSizeMsg{Width: 80, Height: 24}` — expect panelHeight = 12
3. Send `tea.WindowSizeMsg{Width: 120, Height: 40}` — expect panelHeight = 28
4. Assert panel height by inspecting list panel content height (which is `panelHeight - 2` for border overhead)

```go
func TestPanelHeightOverhead(t *testing.T) {
    tests := []struct {
        name           string
        width, height  int
        wantPanelH     int  // expected panelHeight = (height - 2) - 10
    }{
        {"80x24", 80, 24, 12},    // innerHeight=22, panelHeight=22-10=12
        {"120x40", 120, 40, 28},  // innerHeight=38, panelHeight=38-10=28
        {"100x30", 100, 30, 18},  // innerHeight=28, panelHeight=28-10=18
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := config.NewConfig()
            schema := []copilot.SchemaField{
                {Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
            }
            model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")

            msg := tea.WindowSizeMsg{Width: tt.width, Height: tt.height}
            newModel, _ := model.Update(msg)
            m := newModel.(*Model)

            // The list panel content height = panelHeight - 2 (panel border overhead)
            // So panelHeight = listContentH + 2
            // We can verify by checking that View() at this size doesn't panic
            // and produces non-empty output
            _ = m  // panelHeight is not directly exported; verify via View()

            view := m.View()
            if view == "" {
                t.Errorf("View() at %dx%d returned empty string", tt.width, tt.height)
            }
        })
    }
}
```

**Note:** `panelHeight` is a local variable in `updateSizes()`, not an exported field. The test validates behavior indirectly: if the overhead is wrong, `View()` will either panic (negative height) or produce visibly broken output. For a more precise assertion, the implementer may choose to expose `panelHeight` via a test helper or inspect `m.listPanel` dimensions.

### Expected Result

- View renders successfully at all three sizes
- No panics from negative heights or dimension overflows
- At 80×24, the layout is usable (12 rows for panels)

---

## Test T-08: Small terminal floor guard

- **Type:** Unit (Edge Case)
- **Task:** Tasks 5, 6
- **Priority:** Medium

### Setup

Create a model and send a small `WindowSizeMsg` to trigger the `panelHeight < 3` floor guard.

### Steps

1. Create model with a single schema field
2. Send `tea.WindowSizeMsg{Width: 40, Height: 15}` — innerHeight=13, calculated panelHeight=3 (13-10=3, just at the floor)
3. Send `tea.WindowSizeMsg{Width: 30, Height: 12}` — innerHeight=10, calculated panelHeight=0 → floored to 3
4. Call `View()` — must not panic

```go
func TestSmallTerminalFloor(t *testing.T) {
    tests := []struct {
        name          string
        width, height int
    }{
        {"40x15 - at floor", 40, 15},
        {"30x12 - below floor", 30, 12},
        {"20x10 - very small", 20, 10},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := config.NewConfig()
            schema := []copilot.SchemaField{
                {Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
            }
            model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")

            msg := tea.WindowSizeMsg{Width: tt.width, Height: tt.height}
            newModel, _ := model.Update(msg)
            m := newModel.(*Model)

            // Must not panic
            view := m.View()
            if view == "" {
                t.Errorf("View() at %dx%d returned empty string", tt.width, tt.height)
            }
        })
    }
}
```

### Expected Result

- `View()` does not panic at any small terminal size
- The `panelHeight < 3` guard ensures panels never go below 3 rows
- Output is non-empty (may be degraded but not broken)

---

## Test T-09: Existing tests pass (regression)

- **Type:** Regression
- **Task:** Task 7
- **Priority:** Critical

### Setup

No new code; run existing test suite.

### Steps

1. Run `go test ./internal/tui/... -v`
2. Verify all existing tests pass, specifically:
   - `TestNewModel` (UT-TUI-001)
   - `TestStateMachineInitialization` (UT-TUI-002)
   - `TestListPopulation` (UT-TUI-003)
   - `TestSensitiveFieldsInList` (UT-TUI-004)
   - `TestAltScreenCompatibility` (UT-TUI-005)
   - `TestWindowResize` (UT-TUI-006) — stores width/height; overhead change does not affect this
   - `TestBrowsingToEditingTransition` (UT-TUI-007) — sets windowWidth=100, windowHeight=30 and calls updateSizes(); must still work with new overhead
   - `TestEditingToBrowsingTransition` (UT-TUI-008)
   - `TestTokenLikeValueTreatedAsSensitive` (UT-TUI-009)
   - `TestFieldCategorization` (UT-TUI-010)
   - `TestDetailPanelRender` (UT-TUI-011)
   - `TestFormatValueCompact` (UT-TUI-012)
   - `TestListPanelSkipsHeaders` (UT-TUI-013)
   - `TestViewRenders` (UT-TUI-014) — calls View() at 100×30; must still produce non-empty output

### Expected Result

- All 14 existing tests pass with zero failures
- No test modifications required (the changes are additive — styles and constants are new; overhead change doesn't affect tests that don't check panel dimensions)
- `go test ./internal/tui/... -v` exits with code 0

---

## Test T-10: Manual visual verification

- **Type:** Manual
- **Task:** Task 8
- **Priority:** High

### Setup

1. Build the application: `go build ./cmd/ccc`
2. Set terminal to exactly 80 columns × 24 rows
3. Prepare a test config file with at least 5 schema fields across multiple categories

### Steps

1. **At 80×24:**
   - Launch the TUI
   - Verify icon renders on 4 lines with correct character alignment
   - Verify title is horizontally joined to the right of the icon
   - Verify header frame: rounded border encloses icon+title, spans full width
   - Verify help bar frame: rounded border encloses help text, spans full width
   - Verify both panels are visible with ≥10 visible items in the list
   - Verify no content overflow, no unexpected blank lines
   - Navigate up/down, enter editing mode, exit — all framed sections remain stable

2. **At 120×40:**
   - Resize terminal to 120×40
   - Verify same checks as 80×24
   - Verify extra space is gracefully distributed

3. **At 60×20 (edge case):**
   - Resize terminal to 60×20
   - Verify no crash or panic
   - Verify layout degrades gracefully (truncation acceptable)

4. **Unicode characters:**
   - Verify `▘` (U+2598), `▝` (U+259D), `▔` (U+2594) render correctly (1-cell wide, visible glyph)

5. **Dynamic resize:**
   - While TUI is running, resize terminal between 80×24 and 120×40
   - Verify all framed sections reflow without visual artifacts

### Expected Result

- All visual checks pass at 80×24 and 120×40
- No crash at 60×20
- Unicode characters render correctly in the development terminal
- Dynamic resize produces correct reflow
- If overhead value (10) is empirically incorrect, document the correct value and loop back to Tasks 5/6

---

## Coverage Summary

| Task | Tests | Coverage Type |
|------|-------|---------------|
| Task 1 (icon const) | T-01 | Structure validation |
| Task 2 (frame styles) | T-04, T-05 | Integration (rendered output) |
| Task 3 (header redesign) | T-02, T-03 | Content validation |
| Task 4 (frame wrapping) | T-04, T-05, T-06 | Integration + regression |
| Task 5 (updateSizes overhead) | T-07, T-08 | Arithmetic + edge case |
| Task 6 (View overhead) | T-06, T-07 | Consistency |
| Task 7 (tests) | T-09 | Regression |
| Task 8 (visual verify) | T-10 | Manual visual |

## Risk Coverage

| Risk | Test(s) | Mitigation |
|------|---------|------------|
| Height arithmetic drift | T-07, T-08 | Arithmetic validated at multiple sizes + floor guard |
| Frame width not set dynamically | T-04, T-06 | Border characters verified; 80×24 regression |
| Unicode character-width variance | T-01, T-10 | Constant structure test + manual visual |
| Minimum terminal size usability | T-08, T-10 | Floor guard test + manual at 60×20 |
| Existing test regression | T-09 | Full regression run of all 14 existing tests |
