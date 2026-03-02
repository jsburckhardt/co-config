# Action Plan: Improve Icons and UI

## Feature
- **ID:** WI-0004-improve-icons-and-ui
- **Research Brief:** [docs/workitems/WI-0004-improve-icons-and-ui/research/00-research.md](../../research/00-research.md)

## ADRs Created
- None — ADR-0003 already scoped this work. The header was described as "simple styled text for MVP" with "branded header with icon/logo" as the goal. This workitem completes that goal.

## Core-Components Created
- None — changes are purely visual, confined to `internal/tui/model.go` and `internal/tui/styles.go`. No cross-cutting behavioral contracts are introduced.

---

## Implementation Tasks

### Task 1: Define `copilotIcon` constant in `styles.go`

**File:** `internal/tui/styles.go`
**Location:** After the color variable declarations (after line 11), or at the end of the `var` block (after line 87).

**What to do:**
Add a package-level constant for the 4-line ASCII art icon:

```go
const copilotIcon = "╭─╮╭─╮\n╰─╯╰─╯\n█ ▘▝ █\n ▔▔▔▔"
```

**Why here:** The research brief recommends `styles.go` as the location (Section "Recommended Approach — Step 1"). Keeping the icon string separate from rendering logic makes it easy to swap or test. Using a `const` rather than a `var` prevents accidental mutation.

**Risks:**
- Unicode characters `▘` (U+2598), `▝` (U+259D), and `▔` (U+2594) may render inconsistently across terminal emulators (research brief: Medium severity, Medium likelihood). Mitigation: verify in Task 8; simpler block chars are a fallback.
- The icon is 6 characters wide per line. Confirm each line is exactly 6 cells wide by visual inspection.

**Acceptance Criteria:**
- [ ] `copilotIcon` is defined as a `const` in `styles.go`
- [ ] The constant contains exactly 4 lines separated by `\n`
- [ ] Each line is 6 character-cells wide
- [ ] `go build ./...` succeeds with no errors

---

### Task 2: Add `headerFrameStyle` and `helpBarFrameStyle` in `styles.go`

**File:** `internal/tui/styles.go`
**Location:** Inside the `var ( ... )` block, logically grouped near `outerFrameStyle` (lines 13–15) or near `panelStyle` (lines 29–33).

**What to do:**
Add two new Lipgloss styles for framing the header and help bar sections:

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

**Why:** The research brief (Section "Recommended Approach — Step 3") specifies these styles. They use `lipgloss.RoundedBorder()` and `borderColor` to match the existing visual language established by `panelStyle` (lines 29–33) and `focusedPanelStyle` (lines 34–37).

**Design notes:**
- `Padding(0, 1)` adds 1-char horizontal padding inside the border for readability, matching `panelStyle`.
- `.Width()` is NOT set here — it must be set dynamically in `View()` using `innerWidth - 2` to account for the outer frame's border characters (research brief footnote [^4]).

**Risks:**
- If `.Width()` is accidentally set statically here, the frame will not adapt to window resizes (research brief: Medium severity, High likelihood). Mitigation: width is set in `View()` only (Task 4).

**Acceptance Criteria:**
- [ ] `headerFrameStyle` is defined in the `var` block with rounded border, `borderColor`, and `Padding(0, 1)`
- [ ] `helpBarFrameStyle` is defined in the `var` block with rounded border, `borderColor`, and `Padding(0, 1)`
- [ ] Neither style has a static `.Width()` set
- [ ] `go build ./...` succeeds

---

### Task 3: Redesign header block in `View()` — icon + title horizontal join

**File:** `internal/tui/model.go`
**Location:** Lines 276–288 (the current `// Header: two lines` block)

**What to do:**
Replace the current 2-line header assembly:

```go
// Current (lines 277-288):
title := headerStyle.Render("⚙  ccc — Copilot Config CLI")
// ... version, saved, err ...
header := title + "\n" + version
```

With a horizontal join of the icon block and a title block:

```go
iconBlock := lipgloss.NewStyle().Foreground(primaryColor).Render(copilotIcon)
titleBlock := lipgloss.JoinVertical(lipgloss.Left,
    headerStyle.Render("ccc — Copilot Config CLI"),
    version,  // version line already includes saved/error indicators
)
headerContent := lipgloss.JoinHorizontal(lipgloss.Center, iconBlock, "  ", titleBlock)
```

**Key design decisions:**
- The `⚙` gear emoji is removed from the title text (replaced by the ASCII icon).
- `lipgloss.JoinHorizontal(lipgloss.Center, ...)` vertically centers the 2-line title block against the 4-line icon block.
- The `"  "` separator provides 2 spaces between icon and title for visual breathing room.
- The `version` variable assembly (lines 278–287) remains unchanged — it still appends saved/error indicators.

**Risks:**
- `lipgloss.JoinHorizontal` vertical alignment behavior with mixed-height blocks: research brief rates confidence as Medium. If `Center` alignment looks wrong, `Top` is a safe fallback (research brief Open Question #2).
- The icon is rendered with `primaryColor` to match the title. Research brief Open Question #1 asks about a dedicated `brandColor` — for now, `primaryColor` is consistent.

**Acceptance Criteria:**
- [ ] The `⚙` character no longer appears in the header
- [ ] The `copilotIcon` constant is rendered with `primaryColor` foreground
- [ ] Icon and title are joined horizontally using `lipgloss.JoinHorizontal`
- [ ] Version line (with saved/error indicators) appears below the title
- [ ] `go build ./...` succeeds

---

### Task 4: Wrap header and help bar in frame styles in `View()`

**File:** `internal/tui/model.go`
**Location:** Lines 321–324 (help bar rendering and inner assembly)

**What to do:**
After constructing `headerContent` (Task 3) and `helpBar` (line 321), wrap both in their frame styles before the vertical join:

```go
framedHeader := headerFrameStyle.Width(innerWidth - 2).Render(headerContent)
framedHelpBar := helpBarFrameStyle.Width(innerWidth - 2).Render(helpBar)
inner := lipgloss.JoinVertical(lipgloss.Left, framedHeader, "", panels, framedHelpBar)
```

**Why `innerWidth - 2`:** The outer frame already consumes 2 characters (1 left border + 1 right border). The framed header/help bar sit inside the outer frame, so their content width must be `innerWidth` minus their own 2 border characters. See research brief footnote [^4].

**Risks:**
- If width is not set to `innerWidth - 2`, the frame collapses to content width creating a ragged border (research brief: Medium severity, High likelihood). This is the single most common visual bug.
- The blank separator `""` between framed header and panels may need to be removed if the framed header's bottom border already provides sufficient visual separation. Verify visually in Task 8.

**Acceptance Criteria:**
- [ ] Header content is wrapped in `headerFrameStyle.Width(innerWidth - 2).Render(...)`
- [ ] Help bar content is wrapped in `helpBarFrameStyle.Width(innerWidth - 2).Render(...)`
- [ ] The `lipgloss.JoinVertical` call uses `framedHeader` and `framedHelpBar` instead of raw `header`/`helpBar`
- [ ] Framed sections span the full available width (no ragged edges)
- [ ] `go build ./...` succeeds

---

### Task 5: Update height overhead constant in `updateSizes()` from 4 to ~10

**File:** `internal/tui/model.go`
**Location:** Lines 231–232 in `updateSizes()`

**What to do:**
Update the overhead constant and its comment to reflect the new layout geometry:

```go
// Current (line 231-232):
// Header(2) + blank(1) + help(1) = 4 lines overhead
panelHeight := innerHeight - 4

// New:
// framedHeader: 4 icon lines + 2 border = 6, blank(1), framedHelpBar: 1 content + 2 border = 3
// Total overhead: 6 + 1 + 3 = 10
panelHeight := innerHeight - 10
```

**Why:** The research brief (Section "Height Arithmetic") calculates the new overhead as 10 lines. The old header was 2 lines (title + version) with no borders. The new header is 4 content lines (icon height) + 2 border lines = 6 lines. The help bar gains 2 border lines (1 content + 2 border = 3). With the blank separator, total overhead is 10.

**Important:** The exact value of 10 is based on the design with `Padding(0, 1)` and `RoundedBorder()`. If padding values change in Task 2, this constant must be adjusted. The research brief rates confidence as Medium on the exact value.

**Risks:**
- This is rated **High severity, Certain likelihood** in the research brief. If this constant is wrong, panels will overflow or be truncated. This is the single most critical change.
- The `panelHeight < 3` guard (line 233) provides a safety floor but does not prevent truncation at normal terminal sizes.

**Acceptance Criteria:**
- [ ] `panelHeight := innerHeight - 10` in `updateSizes()` (or the empirically correct value after Task 8 verification)
- [ ] Comment above the line explains the breakdown: `framedHeader(6) + blank(1) + framedHelpBar(3) = 10`
- [ ] The `panelHeight < 3` guard remains unchanged
- [ ] At 80×24 terminal: `innerHeight = 22`, `panelHeight = 12` — panels are usable with ≥10 visible items

---

### Task 6: Update duplicate height calculation in `View()`

**File:** `internal/tui/model.go`
**Location:** Line 269 in `View()`

**What to do:**
The `View()` method duplicates the height arithmetic from `updateSizes()`:

```go
// View() line 269:
panelHeight := innerHeight - 4   // ← must match updateSizes()
```

Update this to match the new value from Task 5:

```go
panelHeight := innerHeight - 10
```

**Why this duplication exists:** `updateSizes()` sets panel dimensions on the model struct for list sizing, but `View()` recalculates them locally for rendering. Both must agree. The research brief does not call out this duplication as a separate risk, but getting them out of sync would cause mismatched layout.

**Refactoring note:** Consider whether this duplication can be eliminated by reading from model fields set in `updateSizes()`. However, that refactoring is out of scope for this workitem — just ensure both values match.

**Acceptance Criteria:**
- [ ] `panelHeight := innerHeight - 10` in `View()` matches the value in `updateSizes()`
- [ ] If refactored to eliminate duplication, both paths produce the same value
- [ ] `go build ./...` succeeds

---

### Task 7: Update/add tests for new overhead and rendering

**File:** `internal/tui/tui_test.go`
**Location:** After `TestWindowResize` (line 124) and/or `TestViewRenders` (line 329)

**What to do:**

**7a. Add or update test for overhead constant:**
Add a test that validates `panelHeight` at a known window size, ensuring the overhead is correctly applied:

```go
func TestPanelHeightOverhead(t *testing.T) {
    cfg := config.NewConfig()
    schema := []copilot.SchemaField{
        {Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
    }
    model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")

    // Simulate 80x24 terminal
    msg := tea.WindowSizeMsg{Width: 80, Height: 24}
    newModel, _ := model.Update(msg)
    m := newModel.(*Model)

    // innerHeight = 24 - 2 = 22; panelHeight = 22 - 10 = 12
    expectedPanelHeight := 12
    // Assert panelHeight via list height or by inspecting model fields
    // (exact assertion depends on how panelHeight is exposed)
}
```

**7b. Verify `TestViewRenders` still passes:**
The existing `TestViewRenders` (UT-TUI-014, line 329) calls `model.View()` and checks `view != ""`. This test should still pass with the new rendering. Run it to confirm.

**7c. Verify `TestWindowResize` still passes:**
The existing `TestWindowResize` (UT-TUI-006, line 124) only checks that `windowWidth`/`windowHeight` are stored. It does not check panel height, so it should pass. Run it to confirm.

**7d. Add test for icon presence in rendered output:**
Optionally, add a test that the rendered `View()` output contains the icon characters:

```go
if !strings.Contains(view, "╭─╮╭─╮") {
    t.Error("View() should contain the copilot icon")
}
```

**Risks:**
- No snapshot/golden-file tests exist (research brief: Low severity, High likelihood). This is accepted for MVP. The tests added here are structural (height arithmetic) not visual.

**Acceptance Criteria:**
- [ ] A test exists that validates `panelHeight` equals `innerHeight - 10` at a known window size (e.g., 80×24 → panelHeight=12)
- [ ] `TestViewRenders` (UT-TUI-014) passes with the new rendering
- [ ] `TestWindowResize` (UT-TUI-006) passes unchanged
- [ ] `go test ./internal/tui/...` passes with zero failures
- [ ] Test coverage for `updateSizes()` overhead calculation is explicit (not incidental)

---

### Task 8: Manual visual verification at 80×24

**No file changes — verification only.**

**What to do:**
Run the TUI at exactly 80 columns × 24 rows and verify the following:

1. **Icon renders correctly:** All 4 lines of the ASCII icon display with correct alignment. No character-width issues with `▘`, `▝`, or `▔`.
2. **Horizontal join alignment:** The title text is vertically centered (or top-aligned, per design choice) relative to the 4-line icon.
3. **Header frame:** Rounded border encloses the icon+title block. Frame spans the full width of the outer frame.
4. **Help bar frame:** Rounded border encloses the help text. Frame spans the full width.
5. **Panel height:** Both panels are visible and usable with ≥10 items visible in the list panel.
6. **No overflow:** No content overflows the outer frame. No blank lines or gaps appear unexpectedly.
7. **No truncation:** Help bar text is fully visible within its frame at 80 columns.
8. **Resize behavior:** Resizing the terminal dynamically reflows all framed sections correctly.

**Terminal targets (from research brief risk table):**
- Primary: 80×24 (ADR-0003 target resolution)
- Secondary: 120×40 (common developer terminal)
- Edge case: 60×20 (minimum usable — icon may be hidden or layout degraded, but no crash)

**Acceptance Criteria:**
- [ ] TUI renders correctly at 80×24 with no visual artifacts
- [ ] TUI renders correctly at 120×40
- [ ] TUI does not crash or panic at 60×20
- [ ] `panelHeight` overhead value (Task 5/6) is empirically confirmed correct — if not, loop back and adjust
- [ ] Icon characters render as expected in the development terminal

---

## Task Dependency Graph

```
Task 1 (icon const) ──────────────┐
                                   ├──→ Task 3 (header redesign) ──→ Task 4 (frame wrapping) ──┐
Task 2 (frame styles) ────────────┘                                                             │
                                                                                                 ├──→ Task 8 (visual verify)
Task 5 (updateSizes overhead) ──→ Task 6 (View overhead) ──→ Task 7 (tests) ────────────────────┘
```

- Tasks 1, 2, and 5 can be done in parallel (no dependencies).
- Task 3 depends on Task 1 (needs `copilotIcon` constant).
- Task 4 depends on Tasks 2 and 3 (needs frame styles and header content).
- Task 6 depends on Task 5 (must match the same overhead value).
- Task 7 depends on Tasks 5 and 6 (tests the overhead arithmetic).
- Task 8 depends on all prior tasks (end-to-end visual verification).

## Risk Summary

| Risk | Severity | Mitigation Task |
|------|----------|-----------------|
| Height arithmetic drift (overhead constant wrong) | **High** | Tasks 5, 6, 7, 8 |
| Frame width not set dynamically (ragged borders) | **Medium** | Task 4 |
| Unicode character-width variance | **Medium** | Task 8 |
| Minimum terminal size usability (80×24) | **Medium** | Task 8 |
| `JoinHorizontal` alignment mismatch | **Medium** | Tasks 3, 8 |
