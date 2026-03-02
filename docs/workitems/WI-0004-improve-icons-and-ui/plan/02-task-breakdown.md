# Task Breakdown: WI-0004 — Improve Icons and UI

- **Work Item:** WI-0004-improve-icons-and-ui
- **Action Plan:** [01-action-plan.md](./01-action-plan.md)
- **Total Tasks:** 8
- **Estimated Total Complexity:** Low-Medium (pure presentation changes, 2 files modified, ~50 net lines)

---

## Task 1: Define `copilotIcon` constant in `styles.go`

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** None
- **Related ADRs:** ADR-0003 (header described as "simple styled text for MVP"; icon completes the branded-header goal)
- **Related Core-Components:** None

### Description

Add a package-level `const` for the 4-line Copilot-style ASCII art icon in `internal/tui/styles.go`.

**Current state:** No icon constant exists. The header uses a single `⚙` character embedded inline at `model.go:277`.

**What to do:** After the closing parenthesis of the `var ( ... )` block (after line 87 of `styles.go`), add:

```go
const copilotIcon = "╭─╮╭─╮\n╰─╯╰─╯\n█ ▘▝ █\n ▔▔▔▔"
```

Using `const` (not `var`) prevents accidental mutation. Each of the 4 lines is exactly 6 character-cells wide.

### Acceptance Criteria

- [ ] `copilotIcon` is defined as a `const` in `internal/tui/styles.go`
- [ ] The constant contains exactly 4 lines separated by `\n`
- [ ] Each line is 6 character-cells wide (verified by manual count)
- [ ] `go build ./...` succeeds with no errors

### Test Coverage

- Unit test (Task 7, T-01) validates the constant has exactly 4 lines via `strings.Count(copilotIcon, "\n") == 3`
- Unit test (T-01) validates each line is 6 cells wide via `lipgloss.Width()` or `len()`

---

## Task 2: Add `headerFrameStyle` and `helpBarFrameStyle` in `styles.go`

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** None
- **Related ADRs:** ADR-0003 (extends the existing rounded-border visual language from panels to header/help bar)
- **Related Core-Components:** None

### Description

Add two new Lipgloss styles for framing the header and help-bar sections in `internal/tui/styles.go`.

**Current state:** The `var ( ... )` block (lines 5–87) contains `panelStyle` (lines 29–33) and `focusedPanelStyle` (lines 34–37) which both use `lipgloss.RoundedBorder()` with `borderColor`. No frame styles exist for header or help bar.

**What to do:** Inside the `var ( ... )` block, after `outerFrameStyle` (line 15) and before `headerStyle` (line 17), add:

```go
headerFrameStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(borderColor).
    Padding(0, 1)

helpBarFrameStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(borderColor).
    Padding(0, 1)
```

**Critical:** Do NOT set `.Width()` statically on these styles. Width must be set dynamically in `View()` (Task 4) to respond to terminal resizes.

### Acceptance Criteria

- [ ] `headerFrameStyle` is defined inside the `var` block with `lipgloss.RoundedBorder()`, `borderColor`, and `Padding(0, 1)`
- [ ] `helpBarFrameStyle` is defined inside the `var` block with `lipgloss.RoundedBorder()`, `borderColor`, and `Padding(0, 1)`
- [ ] Neither style has a static `.Width()` call
- [ ] Visual consistency: both use `borderColor` matching `panelStyle` (line 31)
- [ ] `go build ./...` succeeds with no errors

### Test Coverage

- Compile-time verification: if styles are malformed, `go build` fails
- Integration coverage via Task 7 tests (T-04, T-05) that render View() and verify framed output contains border characters

---

## Task 3: Redesign header block in `View()` — icon + title horizontal join

- **Status:** Not Started
- **Complexity:** Medium
- **Dependencies:** Task 1 (needs `copilotIcon` constant)
- **Related ADRs:** ADR-0003 (completes "branded header with icon/logo" goal), ADR-0002 (uses Lipgloss layout primitives)
- **Related Core-Components:** None

### Description

Replace the current 2-line header assembly in `internal/tui/model.go` `View()` method with a horizontal join of the ASCII icon and title block.

**Current state (model.go lines 276–288):**
```go
// Header: two lines
title := headerStyle.Render("⚙  ccc — Copilot Config CLI")
version := ""
if m.version != "" {
    version = versionStyle.Render("   Copilot CLI v" + m.version)
}
if m.saved {
    version += "  " + savedStyle.Render("✓ Saved")
}
if m.err != nil {
    version += "  " + errorStyle.Render("✗ "+m.err.Error())
}
header := title + "\n" + version
```

**What to do:** Replace lines 276–288 with:
```go
// Header: icon (4 lines) + title block (2 lines), joined horizontally
iconBlock := lipgloss.NewStyle().Foreground(primaryColor).Render(copilotIcon)
title := headerStyle.Render("ccc — Copilot Config CLI")
version := ""
if m.version != "" {
    version = versionStyle.Render("Copilot CLI v" + m.version)
}
if m.saved {
    version += "  " + savedStyle.Render("✓ Saved")
}
if m.err != nil {
    version += "  " + errorStyle.Render("✗ "+m.err.Error())
}
titleBlock := lipgloss.JoinVertical(lipgloss.Left, title, version)
headerContent := lipgloss.JoinHorizontal(lipgloss.Center, iconBlock, "  ", titleBlock)
```

**Key changes:**
1. Remove the `⚙` character from the title string
2. Remove the leading `"   "` padding from the version string (the frame provides padding now)
3. Render `copilotIcon` with `primaryColor` foreground
4. Use `lipgloss.JoinHorizontal(lipgloss.Center, ...)` to vertically center the 2-line title against the 4-line icon
5. Use `"  "` (2 spaces) as separator between icon and title

**Note:** The variable name changes from `header` to `headerContent` to distinguish from the framed version created in Task 4.

### Acceptance Criteria

- [ ] The `⚙` character no longer appears anywhere in the header rendering code
- [ ] `copilotIcon` is rendered with `primaryColor` foreground
- [ ] Icon and title are joined horizontally using `lipgloss.JoinHorizontal`
- [ ] Version line (with saved/error indicators) appears below the title via `lipgloss.JoinVertical`
- [ ] The separator between icon and title is `"  "` (2 spaces)
- [ ] `go build ./...` succeeds

### Test Coverage

- Unit test (T-02): View() output does NOT contain `⚙`
- Unit test (T-03): View() output contains the first line of copilotIcon (`╭─╮╭─╮`)
- Unit test (T-03): View() output contains the title text `ccc — Copilot Config CLI`

---

## Task 4: Wrap header and help bar in frame styles in `View()`

- **Status:** Not Started
- **Complexity:** Medium
- **Dependencies:** Task 2 (needs frame styles), Task 3 (needs `headerContent` variable)
- **Related ADRs:** ADR-0003 (extends framed layout pattern from panels to header/help bar)
- **Related Core-Components:** None

### Description

Wrap the header content and help bar in their respective frame styles before the final vertical join in `View()`.

**Current state (model.go line 324):**
```go
inner := lipgloss.JoinVertical(lipgloss.Left, header, "", panels, helpBar)
```

**What to do:** Replace line 324 with:
```go
framedHeader := headerFrameStyle.Width(innerWidth - 2).Render(headerContent)
framedHelpBar := helpBarFrameStyle.Width(innerWidth - 2).Render(helpBar)
inner := lipgloss.JoinVertical(lipgloss.Left, framedHeader, "", panels, framedHelpBar)
```

**Why `innerWidth - 2`:** The framed header/help bar sit inside the outer frame (`outerFrameStyle`). The outer frame border already consumes 2 characters (1 left + 1 right). Within that, `innerWidth` is the available space. The frame styles themselves have a rounded border consuming 2 more characters — so we set `.Width(innerWidth - 2)` to account for the frame's own border characters, ensuring the content fills the available space without overflow.

**Note:** The variable `header` from the current code becomes `headerContent` (from Task 3). The old `header` variable name is no longer used.

### Acceptance Criteria

- [ ] `headerContent` is wrapped with `headerFrameStyle.Width(innerWidth - 2).Render(...)`
- [ ] `helpBar` is wrapped with `helpBarFrameStyle.Width(innerWidth - 2).Render(...)`
- [ ] The `lipgloss.JoinVertical` call uses `framedHeader` and `framedHelpBar`
- [ ] Width is set dynamically using `innerWidth - 2` (not a static value)
- [ ] The blank separator `""` between framedHeader and panels is preserved
- [ ] `go build ./...` succeeds

### Test Coverage

- Unit test (T-04): View() output contains rounded border characters (`╭`, `╮`, `╯`, `╰`) from the header frame
- Unit test (T-05): View() output contains rounded border characters from the help bar frame
- Unit test (T-06): View() at 80×24 renders without panic and produces non-empty output (regression guard)

---

## Task 5: Update height overhead constant in `updateSizes()` from 4 to 10

- **Status:** Not Started
- **Complexity:** Medium (high-severity if wrong)
- **Dependencies:** None (arithmetic change independent of rendering, but logically paired with Tasks 3-4)
- **Related ADRs:** ADR-0003 (80×24 is the target resolution; panel height must remain usable)
- **Related Core-Components:** None

### Description

Update the overhead constant in `updateSizes()` to account for the new framed header (6 lines) and framed help bar (3 lines).

**Current state (model.go lines 231–232):**
```go
// Header(2) + blank(1) + help(1) = 4 lines overhead
panelHeight := innerHeight - 4
```

**New overhead calculation:**
| Component          | Lines |
|--------------------|-------|
| Framed header: 4 content lines (icon height) + 2 border lines | 6 |
| Blank separator    | 1 |
| Framed help bar: 1 content line + 2 border lines | 3 |
| **Total**          | **10** |

**What to do:** Replace lines 231–232 with:
```go
// framedHeader(6) + blank(1) + framedHelpBar(3) = 10 lines overhead
panelHeight := innerHeight - 10
```

**At 80×24:** `innerHeight = 24 - 2 = 22`, `panelHeight = 22 - 10 = 12`. With panel border overhead of 2, that's 10 visible rows per panel — still usable per ADR-0003.

**Risk:** This is the highest-severity change. If wrong, panels overflow or truncate. The `panelHeight < 3` guard at line 233 provides a safety floor but doesn't prevent subtle truncation. The exact value of 10 must be empirically verified in Task 8.

### Acceptance Criteria

- [ ] `panelHeight := innerHeight - 10` in `updateSizes()` (line 232 of current code)
- [ ] Comment above accurately documents: `framedHeader(6) + blank(1) + framedHelpBar(3) = 10`
- [ ] The `panelHeight < 3` guard (line 233) remains unchanged
- [ ] At 80×24: panelHeight = 12 (verified by test T-07)
- [ ] At 120×40: panelHeight = 28 (verified by test T-07)

### Test Coverage

- Unit test (T-07): Send `WindowSizeMsg{80, 24}` and verify panelHeight is 12 (derived from model list panel height)
- Unit test (T-07): Send `WindowSizeMsg{120, 40}` and verify panelHeight is 28
- Unit test (T-08): Send `WindowSizeMsg{40, 15}` (small terminal) and verify panelHeight floors at 3

---

## Task 6: Update duplicate height calculation in `View()`

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** Task 5 (must match the same overhead value)
- **Related ADRs:** ADR-0003
- **Related Core-Components:** None

### Description

The `View()` method duplicates the height arithmetic from `updateSizes()`. Both must agree.

**Current state (model.go line 269):**
```go
panelHeight := innerHeight - 4
```

**What to do:** Update line 269 to match Task 5:
```go
panelHeight := innerHeight - 10
```

**Refactoring note:** The action plan acknowledges this duplication but explicitly states refactoring is out of scope for this workitem. Just ensure both values match.

### Acceptance Criteria

- [ ] `panelHeight := innerHeight - 10` in `View()` at line 269
- [ ] The value matches `updateSizes()` exactly (both use `- 10`)
- [ ] The `panelHeight < 3` guard at line 270 remains unchanged
- [ ] `go build ./...` succeeds

### Test Coverage

- Covered indirectly by T-06 (View renders at 80×24 without panic)
- Covered indirectly by T-07 (overhead arithmetic test validates the model-level calculation, and T-06 validates View-level consistency)

---

## Task 7: Add and update tests for new overhead and rendering

- **Status:** Not Started
- **Complexity:** Medium
- **Dependencies:** Tasks 1–6 (all implementation tasks must be complete)
- **Related ADRs:** ADR-0003 (80×24 target), ADR-0002 (Charm stack test patterns)
- **Related Core-Components:** None

### Description

Add new tests and verify existing tests pass in `internal/tui/tui_test.go`. This task covers 4 sub-areas:

**7a. Icon constant validation (new test):**
Add `TestCopilotIconConstant` to validate the icon has 4 lines, each 6 cells wide.

**7b. Icon presence in rendered output (new test):**
Add `TestViewRendersWithIcon` to verify the gear emoji is gone and the ASCII icon is present in `View()` output.

**7c. Framed sections in rendered output (new test):**
Add `TestViewRendersFramedSections` to verify rounded border characters appear in View() output, confirming header and help bar frames render.

**7d. Height overhead arithmetic (new test):**
Add `TestPanelHeightOverhead` to verify panel height at known window sizes (80×24 → 12, 120×40 → 28, small terminal → floor of 3).

**7e. Existing test regression:**
- `TestViewRenders` (UT-TUI-014, line 329): Must still pass — calls `View()` and checks `view != ""`
- `TestWindowResize` (UT-TUI-006, line 124): Must still pass — checks windowWidth/windowHeight storage

### Acceptance Criteria

- [ ] `TestCopilotIconConstant` exists and validates 4 lines, each 6 cells wide
- [ ] `TestViewRendersWithIcon` exists and verifies `⚙` is absent, `╭─╮╭─╮` is present, and `ccc — Copilot Config CLI` is present
- [ ] `TestViewRendersFramedSections` exists and verifies border characters in View() output
- [ ] `TestPanelHeightOverhead` exists and validates panel height at 80×24 (12), 120×40 (28), and small terminal (floor at 3)
- [ ] `TestViewRenders` (UT-TUI-014) passes unchanged
- [ ] `TestWindowResize` (UT-TUI-006) passes unchanged
- [ ] `go test ./internal/tui/...` passes with zero failures

### Test Coverage

- This task IS the test coverage. All new tests are defined in the Test Plan (03-test-plan.md) as T-01 through T-08.

---

## Task 8: Manual visual verification at 80×24

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** Tasks 1–7 (all implementation and tests must be complete)
- **Related ADRs:** ADR-0003 (80×24 is the target resolution)
- **Related Core-Components:** None

### Description

**No file changes — verification only.**

Run the TUI at standard terminal sizes and verify visual correctness. This catches issues that unit tests cannot: unicode rendering, alignment aesthetics, and overall visual polish.

**Verification checklist:**

1. **80×24 (primary target, per ADR-0003):**
   - All 4 lines of the ASCII icon render with correct alignment
   - Title text is vertically centered relative to the 4-line icon
   - Header frame: rounded border encloses icon+title, spans full width
   - Help bar frame: rounded border encloses help text, spans full width
   - Both panels are visible and usable with ≥10 visible items
   - No content overflow or unexpected gaps

2. **120×40 (common developer terminal):**
   - Same checks as 80×24
   - Extra width/height is gracefully absorbed

3. **60×20 (edge case — minimum usable):**
   - No crash or panic
   - Layout degrades gracefully (some truncation acceptable)
   - `panelHeight < 3` guard activates correctly

4. **Unicode characters:** `▘` (U+2598), `▝` (U+259D), `▔` (U+2594) render as expected (1-cell wide, correct glyph)

5. **Resize behavior:** Dynamically resizing the terminal reflows all framed sections correctly

**If overhead value (10) is empirically incorrect:** Loop back to Tasks 5 and 6 to adjust.

### Acceptance Criteria

- [ ] TUI renders correctly at 80×24 with no visual artifacts
- [ ] TUI renders correctly at 120×40
- [ ] TUI does not crash or panic at 60×20
- [ ] Overhead value (10) is empirically confirmed correct
- [ ] Icon characters render as expected in the development terminal
- [ ] Dynamic resize works without visual breakage

### Test Coverage

- Manual verification only; no automated tests (golden-file/snapshot tests are out of scope per action plan and research brief)

---

## Task Dependency Graph

```
Task 1 (icon const) ──────────────┐
                                   ├──→ Task 3 (header redesign) ──→ Task 4 (frame wrapping) ──┐
Task 2 (frame styles) ────────────┘                                                             │
                                                                                                 ├──→ Task 7 (tests) ──→ Task 8 (visual verify)
Task 5 (updateSizes overhead) ──→ Task 6 (View overhead) ──────────────────────────────────────┘
```

**Parallel execution opportunities:**
- Tasks 1, 2, and 5 have no dependencies and can be implemented simultaneously
- Task 6 depends only on Task 5
- Task 3 depends only on Task 1
- Task 4 depends on Tasks 2 and 3
- Task 7 depends on all implementation tasks (1–6)
- Task 8 depends on Task 7
