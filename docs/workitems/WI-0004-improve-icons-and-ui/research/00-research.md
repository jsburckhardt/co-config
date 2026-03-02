# Research Brief: WI-0004 — Improve Icons and UI

## Executive Summary

WI-0004 introduces two visual enhancements to the `ccc` TUI: (1) replacing the single `⚙` gear character in the header with a 4-line Copilot-style ASCII art icon placed to the left of the title text, and (2) adding visual frames/borders around the header and help-bar sections. Both changes are **pure presentation** modifications confined to `internal/tui/model.go` and `internal/tui/styles.go`. The existing outer frame and panel borders (already provided by Lipgloss) serve as the foundation — the work is additive. The primary risks are unicode character-width variance across terminal emulators and the height-accounting arithmetic in `updateSizes()` that must be adjusted when the header grows from 2 lines to 4+ lines.

---

## Scope Classification

| Attribute | Value |
|-----------|-------|
| **Scope type** | `workitem` |
| **New ADR required?** | No — ADR-0003 explicitly describes the current header as "simple styled text for MVP" and names "branded header with icon/logo" as the goal; this work completes that goal |
| **Core-component changes?** | No — CC-0002 through CC-0005 are unrelated to visual presentation |
| **ADRs needing amendment?** | ADR-0003 (annotation only — "MVP" qualifier can be removed) |

---

## Codebase Findings

### 1. Entry Point — `View()` method

All rendering originates in a single function[^1]:

```go
// internal/tui/model.go:262-331
func (m *Model) View() string { … }
```

The header is assembled at lines 277–288:

```go
title   := headerStyle.Render("⚙  ccc — Copilot Config CLI")   // line 277
version := versionStyle.Render("   Copilot CLI v" + m.version)   // line 281
header  := title + "\n" + version                                  // line 288
```

The icon replacement must change line 277 and extend the `header` variable to accommodate 4 rows.

### 2. Layout Assembly

The inner content is assembled via[^2]:

```go
// model.go:324
inner := lipgloss.JoinVertical(lipgloss.Left, header, "", panels, helpBar)
```

Adding a frame around the header and help bar means wrapping them with a new Lipgloss style before joining.

### 3. Height Arithmetic (`updateSizes`)

```go
// model.go:228-232
innerHeight := m.windowHeight - 2          // outer frame border
panelHeight := innerHeight - 4             // Header(2) + blank(1) + help(1) = 4 overhead
```

The comment explicitly documents the 4-line overhead assumption[^3]. Changing the header to a 4-line ASCII icon block changes this constant:

| Current layout overhead | Lines |
|-------------------------|-------|
| Header title line       | 1     |
| Header version line     | 1     |
| Blank separator         | 1     |
| Help bar                | 1     |
| **Total**               | **4** |

With the new design (4-line icon, optional section borders), the overhead grows. If section borders (top+bottom) are added to the header block, the header section becomes **4 content lines + 2 border lines = 6 lines**. With a bordered help bar that is 3 lines (1 content + 2 borders), total overhead becomes **6 + 1 (blank) + 3 = 10 lines**. The constant `4` in `updateSizes` must be updated to match.

### 4. Width Arithmetic

```go
// model.go:237-239
leftWidth  := int(float64(innerWidth) * 0.40)
rightWidth := innerWidth - leftWidth - 1
```

Width allocation for panels is not affected by header height changes. However, bordered header/help sections must be rendered with `Width(innerWidth)` to avoid layout gaps[^4].

### 5. Styles

All named styles are in `styles.go`[^5]. The following styles are currently used by the header/help bar:

| Style | File:Line | Usage |
|-------|-----------|-------|
| `headerStyle` | `styles.go:17-19` | Title text (bold, primaryColor) |
| `versionStyle` | `styles.go:21-23` | Version line (muted, italic) |
| `helpStyle` | `styles.go:86` | Help bar text (muted) |
| `outerFrameStyle` | `styles.go:13-15` | Outer rounded border |
| `panelStyle` | `styles.go:29-33` | Unfocused panel (rounded border + padding) |
| `focusedPanelStyle` | `styles.go:34-37` | Focused panel (rounded border, primaryColor) |

New styles needed: `headerFrameStyle` and `helpBarFrameStyle` (both rounded borders, `borderColor`).

### 6. Existing Borders

The current layout already has borders in the right places for panels — this work extends the pattern to header and help bar, maintaining visual consistency[^6].

### 7. ASCII Icon Analysis

The proposed icon:

```
╭─╮╭─╮
╰─╯╰─╯
█ ▘▝ █
 ▔▔▔▔
```

Character inventory:
- `╭ ╮ ╯ ╰ ─` — Box-drawing (U+256x range) — universally 1-cell wide in most modern terminals
- `█` — Full block (U+2588) — 1-cell wide
- `▘ ▝` — Quadrant block elements (U+2598, U+259D) — typically 1-cell wide but rendering varies
- `▔` — Upper one-eighth block (U+2594) — typically 1-cell wide

The icon is 6 characters wide per line. Using `lipgloss.JoinHorizontal(lipgloss.Top, iconBlock, " ", titleBlock)` places it to the left of the title with one-space separator.

### 8. Test Coverage

Existing tests that will be affected[^7]:

| Test ID | Test Name | Impact |
|---------|-----------|--------|
| UT-TUI-006 | `TestWindowResize` | No direct rendering check; passes if width/height stored correctly |
| UT-TUI-014 | `TestViewRenders` | Calls `model.View()` and checks `view != ""`; will pass with any non-empty output |

No snapshot/golden-file tests exist; visual regression must be verified manually. The test suite is minimal enough that the changes are unlikely to break any tests, but the overhead constant change in `updateSizes` should be covered by a new or updated test.

---

## Risks & Constraints

| Risk | Severity | Likelihood | Mitigation |
|------|----------|------------|------------|
| **Unicode width variance** — `▘`, `▝`, `▔` may render as 0-width or 2-width in some fonts (e.g., Windows Terminal with certain CJK fonts) | Medium | Medium | Test in multiple terminals; fall back to simpler block chars if misaligned |
| **Height arithmetic drift** — `updateSizes()` hardcodes `panelHeight := innerHeight - 4`; changing header height without updating this constant causes panels to overflow or be truncated | High | Certain | Must update the constant and its comment; add a test that validates panel height at known window sizes |
| **Lipgloss `.Width()` on framed header** — If the header frame is not set to `innerWidth`, it collapses to content width and creates a ragged border look | Medium | High | Always call `.Width(innerWidth)` on header and help-bar frame styles in `View()` |
| **Minimum terminal width** — The icon is 6 chars wide; with 1 separator + title ("⚙  ccc — Copilot Config CLI" is 27 chars visible) the minimum readable width is ~40 chars. Below that the icon may wrap | Low | Low | Add a minimum width guard (already handled by `if listContentW < 1 { listContentW = 1 }` pattern) |
| **Minimum terminal height** — Adding borders to header (2 lines) + help bar (2 lines) eats 4 more lines. On an 80×24 terminal the panel height drops from ~18 to ~14 lines | Medium | Medium | Verify the TUI is still usable at 80×24; the ADR-0003 target was 80×24 |
| **No snapshot tests** — Visual regressions are undetectable without golden-file tests | Low | High | Acceptable for MVP; document as known gap |
| **`▔` rendering** — U+2594 (UPPER ONE EIGHTH BLOCK) renders as a thin horizontal line in most terminals but may be invisible or wrong in some | Low | Low | Can substitute `▁` (lower block) or `─` if needed |

---

## Recommended Approach

### Step 1 — Add the ASCII icon constant

In `styles.go` or a new `internal/tui/icon.go`, define the icon as a constant:

```go
const copilotIcon = "╭─╮╭─╮\n╰─╯╰─╯\n█ ▘▝ █\n ▔▔▔▔"
```

This keeps the icon string separate from logic and makes it easy to swap.

### Step 2 — Redesign the header block in `View()`

Replace the current 2-line header assembly with a horizontal join:

```go
// Replace model.go:277-288
iconBlock  := lipgloss.NewStyle().Foreground(primaryColor).Render(copilotIcon)
titleBlock := lipgloss.JoinVertical(lipgloss.Left,
    headerStyle.Render("ccc — Copilot Config CLI"),
    versionStyle.Render("Copilot CLI v"+m.version)+savedIndicator+errIndicator,
)
headerContent := lipgloss.JoinHorizontal(lipgloss.Center, iconBlock, "  ", titleBlock)
```

### Step 3 — Add section frame styles in `styles.go`

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

### Step 4 — Wrap header and help bar in frames in `View()`

```go
framedHeader  := headerFrameStyle.Width(innerWidth - 2).Render(headerContent)
framedHelpBar := helpBarFrameStyle.Width(innerWidth - 2).Render(helpBarContent)
inner := lipgloss.JoinVertical(lipgloss.Left, framedHeader, "", panels, framedHelpBar)
```

### Step 5 — Update `updateSizes()` overhead constant

Recalculate the overhead to account for new border lines:

```go
// New overhead:
// framedHeader: 4 icon lines = 4 content, +2 border = 6 lines
// blank separator: 1 line
// framedHelpBar: 1 content, +2 border = 3 lines
// Total: 10 lines
panelHeight := innerHeight - 10
```

The constant should be named or commented explicitly to avoid future drift.

### Step 6 — Update test for `updateSizes`

Add or update UT-TUI-006 to assert that at `windowHeight=30`, `innerHeight=28`, `panelHeight=18` (28-10).

---

## ADRs / Core-Components Impacted

| Document | Impact | Action Required |
|----------|--------|-----------------|
| **ADR-0003** — Two-Panel TUI Layout | The "simple styled text for MVP" qualifier for the header is now superseded; section framing extends the existing border pattern | Add a note/amendment acknowledging WI-0004 completes the icon and section-framing goals; no re-decision needed |
| **ADR-0002** — Go + Charm TUI Stack | No change; Lipgloss is already the mandated layout library | None |
| **CC-0002** — Error Handling | None | None |
| **CC-0003** — Logging | None | None |
| **CC-0004** — Configuration Management | None | None |
| **CC-0005** — Sensitive Data Handling | None | None |

---

## Open Questions

1. **Icon color** — Should the icon inherit `primaryColor` (purple/indigo tones) or use a dedicated `brandColor`? The existing `primaryColor` is `#7C3AED` (dark) / `#5A67D8` (light) — visually consistent with the header title but may not read well on some backgrounds.

2. **`lipgloss.JoinHorizontal` vertical alignment** — The icon is 4 lines; the title block (title + version) is 2 lines. `lipgloss.Center` aligns them vertically by center. Is that the desired look, or should the title be top-aligned?

3. **Help bar content width** — The current help bar is a single `strings.Join(parts, "  •  ")` line. On narrow terminals this may exceed the frame width. Should it wrap or truncate?

4. **Minimum terminal size gate** — At very small terminals (< 60 columns or < 20 rows), should the ASCII icon be hidden to preserve panel space? ADR-0003 targets 80×24 as the standard, but no minimum size enforcement currently exists.

5. **Test strategy** — Are golden-file / snapshot tests in scope for this workitem, or is manual visual verification acceptable?

---

## Confidence Assessment

| Finding | Confidence | Basis |
|---------|------------|-------|
| Header is rendered at `model.go:277` as a single `⚙` character | **High** | Direct source reading |
| Height overhead constant is `4` at `model.go:232` | **High** | Direct source reading with inline comment |
| No snapshot tests exist | **High** | Full `tui_test.go` reviewed |
| Unicode rendering of `▘`/`▝`/`▔` varies across terminals | **Medium** | Known terminal behavior; not tested in this environment |
| `lipgloss.JoinHorizontal` alignment behavior with mixed-height blocks | **Medium** | Lipgloss docs behavior; not tested experimentally |
| New overhead value of `10` | **Medium** | Calculated from design intent; exact value depends on final border/padding choices |
| No new ADR required | **High** | ADR-0003 already scopes this as continuation of "MVP" header |

---

## Footnotes

[^1]: `internal/tui/model.go:262-331` — `View()` function, the sole rendering entry point for the full TUI.

[^2]: `internal/tui/model.go:324` — `lipgloss.JoinVertical` assembles header, blank, panels, helpBar into `inner`, which is then wrapped by `outerFrameStyle`.

[^3]: `internal/tui/model.go:231` — Comment reads `// Header(2) + blank(1) + help(1) = 4 lines overhead`.

[^4]: `internal/tui/model.go:326-330` — `outerFrameStyle.Width(innerWidth).Height(innerHeight).Render(inner)` already uses `innerWidth`; framed sub-sections must use `innerWidth - 2` (subtracting their own border chars).

[^5]: `internal/tui/styles.go:1-87` — All 18 named Lipgloss styles; `headerStyle` at line 17, `helpStyle` at line 86.

[^6]: `internal/tui/styles.go:29-37` — `panelStyle` and `focusedPanelStyle` both use `lipgloss.RoundedBorder()` with `borderColor`/`primaryColor` respectively, establishing the visual language this workitem extends.

[^7]: `internal/tui/tui_test.go:124-142` (UT-TUI-006) and `tui_test.go:329-349` (UT-TUI-014) — The two tests most likely to interact with rendering changes.
