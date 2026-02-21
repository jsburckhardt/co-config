# Task Breakdown: TUI Two-Panel Redesign

## Workitem

- **ID:** WI-0003-tui-two-panel-redesign
- **Action Plan:** [01-action-plan.md](./01-action-plan.md)
- **Research:** [00-research.md](../research/00-research.md)

## Overview

This task breakdown implements the transition from the current Huh-based single-column TUI to a custom two-panel layout using Bubbletea and Bubbles components, as specified in ADR-0003.

## Dependencies

- **ADRs:** ADR-0002 (Charm stack), ADR-0003 (Two-panel layout)
- **Core-Components:** CC-0004 (Configuration Management), CC-0005 (Sensitive Data Handling)
- **External:** `bubbles` v0.21.1 (to be promoted from indirect to direct dependency)

---

## Task T3.1: Enable Alt-Screen Mode and Fullscreen Setup

- **Status:** Not Started
- **Complexity:** Low (2-3 hours)
- **Dependencies:** None
- **Related ADRs:** ADR-0003
- **Related Core-Components:** None

### Description

Update the TUI initialization to run in fullscreen alt-screen mode with proper cleanup on exit. This provides the foundation for the immersive two-panel layout.

### Acceptance Criteria

- [ ] TUI launches using `tea.WithAltScreen()` program option
- [ ] Terminal switches to alternate screen buffer (fullscreen mode)
- [ ] Terminal state is correctly restored on normal exit (Ctrl+C, Esc+quit)
- [ ] Terminal state is correctly restored on panic/error (cleanup logic)
- [ ] No visual artifacts remain after TUI exits
- [ ] Works on standard 80×24 terminal minimum

### Test Coverage

- UT-TUI-005: Test that alt-screen mode is enabled in program options
- UT-TUI-006: Test cleanup logic restores terminal state on panic
- Manual: Launch TUI, verify fullscreen, exit, verify terminal clean

---

## Task T3.2: Create Base Two-Panel Layout Structure

- **Status:** Not Started
- **Complexity:** Medium (3-4 hours)
- **Dependencies:** T3.1
- **Related ADRs:** ADR-0003
- **Related Core-Components:** None

### Description

Define Lipgloss styles and basic layout structure for the two-panel design with outer frame and header.

### Acceptance Criteria

- [ ] Outer frame renders with border around entire UI using Lipgloss
- [ ] Header section displays branded title (styled text: "ccc — Copilot Config CLI")
- [ ] Left panel area defined with fixed width (30% of terminal width)
- [ ] Right panel area defined with remaining width (70% of terminal width)
- [ ] Vertical split between panels is clearly visible (border/separator)
- [ ] Layout adapts to terminal resize (minimum 80×24)
- [ ] All styles defined in updated `styles.go`

### Test Coverage

- UT-TUI-007: Test layout calculation for different terminal sizes
- UT-TUI-008: Test panel width ratios (30/70 split)
- UT-TUI-009: Test minimum terminal size (80×24) renders without overflow
- Manual: Resize terminal, verify layout adapts correctly

---

## Task T3.3: Create Config List Model (Left Panel)

- **Status:** Not Started
- **Complexity:** Medium (4-5 hours)
- **Dependencies:** T3.2
- **Related ADRs:** ADR-0003
- **Related Core-Components:** CC-0004, CC-0005

### Description

Create a new `internal/tui/configlist/` package that wraps `bubbles/list` to display config options as a navigable list with group headers and truncated current values.

### Acceptance Criteria

- [ ] New package `internal/tui/configlist/` created
- [ ] List items populated from config schema (leveraging CC-0004)
- [ ] Items grouped by category (General, Model & AI, URLs & Permissions, Display, Sensitive)
- [ ] Each item shows: field name + truncated current value (e.g., "model: gpt-4")
- [ ] Sensitive fields marked with `(read-only)` suffix (per CC-0005)
- [ ] Sensitive fields styled distinctly (dimmed or special color)
- [ ] Keyboard navigation works: Up/Down arrows, j/k vim bindings
- [ ] Group headers visually distinct from items
- [ ] Scrolling works if list exceeds panel height

### Test Coverage

- UT-TUI-010: Test list population from schema with all field types
- UT-TUI-011: Test sensitive fields marked correctly and styled
- UT-TUI-012: Test list item rendering with truncated values
- UT-TUI-013: Test group categorization logic
- UT-TUI-014: Test keyboard navigation updates selection
- Manual: Navigate list with arrow keys and vim bindings

---

## Task T3.4: Create Detail View Component (Right Panel)

- **Status:** Not Started
- **Complexity:** Medium (3-4 hours)
- **Dependencies:** T3.2
- **Related ADRs:** ADR-0003
- **Related Core-Components:** CC-0004, CC-0005

### Description

Create a new `internal/tui/detail/` package for the right panel that displays focused field metadata and current value.

### Acceptance Criteria

- [ ] New package `internal/tui/detail/` created
- [ ] Display field name as styled header
- [ ] Display field description with text wrapping (using Lipgloss)
- [ ] Display current value with type-appropriate formatting
- [ ] Sensitive fields show masked value (per CC-0005: truncated SHA-256 hash)
- [ ] Sensitive fields show read-only indicator
- [ ] Layout fits within right panel width (70% of terminal)
- [ ] Text wraps correctly for long descriptions

### Test Coverage

- UT-TUI-015: Test detail view rendering for string field
- UT-TUI-016: Test detail view rendering for bool field
- UT-TUI-017: Test detail view rendering for enum field
- UT-TUI-018: Test detail view rendering for list field
- UT-TUI-019: Test detail view shows masked value for sensitive fields
- UT-TUI-020: Test text wrapping for long descriptions
- Manual: Navigate list, verify detail view updates correctly

---

## Task T3.5: Integrate Input Widgets for Editing

- **Status:** Not Started
- **Complexity:** High (6-8 hours)
- **Dependencies:** T3.4
- **Related ADRs:** ADR-0003
- **Related Core-Components:** CC-0004, CC-0005

### Description

Integrate Bubbles input components and create custom widgets for editing different field types in the right panel.

### Acceptance Criteria

- [ ] Single-line string fields use `bubbles/textinput`
- [ ] Multi-line list fields use `bubbles/textarea` (one item per line)
- [ ] Boolean fields use custom toggle component (similar to Huh's confirm)
- [ ] Enum fields use custom select component (radio list of options)
- [ ] Input validation reuses existing validation logic where applicable
- [ ] Validation errors displayed inline in detail panel
- [ ] Tab/Enter activates edit mode for focused field
- [ ] Esc saves value and returns to browsing mode
- [ ] Sensitive fields cannot be edited (focus remains on list)
- [ ] Input widgets respect panel width constraints

### Test Coverage

- UT-TUI-021: Test textinput integration for string fields
- UT-TUI-022: Test textarea integration for list fields
- UT-TUI-023: Test custom toggle component for bool fields
- UT-TUI-024: Test custom select component for enum fields
- UT-TUI-025: Test input validation logic
- UT-TUI-026: Test sensitive fields reject edit attempts
- UT-TUI-027: Test field value persistence after editing
- Manual: Edit each field type, verify correct widget and validation

---

## Task T3.6: Implement State Machine and Focus Handling

- **Status:** Not Started
- **Complexity:** Medium (4-5 hours)
- **Dependencies:** T3.3, T3.5
- **Related ADRs:** ADR-0003
- **Related Core-Components:** None

### Description

Implement the state machine (Browsing ↔ Editing ↔ Saving) and focus management to route keyboard input correctly.

### Acceptance Criteria

- [ ] State enum defined: `Browsing`, `Editing`, `Saving`, `Exiting`
- [ ] `Browsing` state: arrow keys navigate list, Enter switches to `Editing`
- [ ] `Editing` state: input widget has focus, Esc saves and returns to `Browsing`
- [ ] `Saving` state: persists changes to config file
- [ ] `Exiting` state: final save (if needed) and quit
- [ ] Focused panel/widget has visual indicator (border color, cursor)
- [ ] Keyboard input routed to active component based on state
- [ ] State transitions logged for debugging
- [ ] Ctrl+C quits immediately from any state

### Test Coverage

- UT-TUI-028: Test state machine initialization (starts in Browsing)
- UT-TUI-029: Test Browsing → Editing transition on Enter
- UT-TUI-030: Test Editing → Browsing transition on Esc
- UT-TUI-031: Test Editing → Saving → Browsing flow
- UT-TUI-032: Test focus indicator updates with state changes
- UT-TUI-033: Test keyboard routing to correct component
- Manual: Navigate states, verify visual feedback and input routing

---

## Task T3.7: Add Footer Help Bar

- **Status:** Not Started
- **Complexity:** Low (2-3 hours)
- **Dependencies:** T3.6
- **Related ADRs:** ADR-0003
- **Related Core-Components:** None

### Description

Integrate `bubbles/help` component to display context-sensitive keyboard shortcuts in a footer bar.

### Acceptance Criteria

- [ ] Footer bar rendered at bottom of UI
- [ ] Shows context-sensitive hints based on current state:
  - Browsing: "↑/↓: navigate • Enter: edit • Ctrl+C: quit"
  - Editing: "Esc: save • Tab: next field • Ctrl+C: cancel"
- [ ] Help text styled consistently with Lipgloss
- [ ] Footer does not overlap with panels
- [ ] Help bar uses `bubbles/help` component

### Test Coverage

- UT-TUI-034: Test help bar content for Browsing state
- UT-TUI-035: Test help bar content for Editing state
- UT-TUI-036: Test help bar rendering (does not overlap panels)
- Manual: Verify help text updates as state changes

---

## Task T3.8: Integrate with Configuration Management

- **Status:** Not Started
- **Complexity:** Medium (3-4 hours)
- **Dependencies:** T3.6
- **Related ADRs:** ADR-0003
- **Related Core-Components:** CC-0004, CC-0005

### Description

Wire up config loading, schema detection, input validation, and persistence to leverage CC-0004 while maintaining CC-0005 constraints.

### Acceptance Criteria

- [ ] Load config schema and current values on startup (CC-0004)
- [ ] Validate input against schema constraints (type, format, allowed values)
- [ ] Save changes on transition to Saving state or Exiting state
- [ ] Preserve unknown fields during save (CC-0004)
- [ ] Sensitive fields preserved unchanged (CC-0005)
- [ ] Handle config file not found gracefully (show defaults)
- [ ] Handle malformed config file with error message
- [ ] Round-trip test: load → edit → save → reload verifies data integrity

### Test Coverage

- UT-TUI-037: Test config loading on TUI startup
- UT-TUI-038: Test schema-based validation rejects invalid input
- UT-TUI-039: Test config save preserves unknown fields
- UT-TUI-040: Test config save preserves sensitive fields unchanged
- UT-TUI-041: Test round-trip (load-edit-save-reload)
- UT-TUI-042: Test handling of missing config file
- UT-TUI-043: Test handling of malformed config file
- Manual: Edit config, save, reload, verify all changes persisted

---

## Task T3.9: Unit Testing and Edge Cases

- **Status:** Not Started
- **Complexity:** Medium (4-5 hours)
- **Dependencies:** T3.1, T3.2, T3.3, T3.4, T3.5, T3.6, T3.7, T3.8
- **Related ADRs:** ADR-0003
- **Related Core-Components:** CC-0004, CC-0005

### Description

Add comprehensive unit tests for all new components, state transitions, and edge cases.

### Acceptance Criteria

- [ ] All test coverage items from tasks T3.1-T3.8 implemented
- [ ] Test naming follows UT-XXX-### convention (UT-TUI-005 through UT-TUI-043)
- [ ] Edge cases covered:
  - Empty config (no fields)
  - All sensitive fields (no editable fields)
  - Single field config
  - Maximum number of fields (~25)
  - Very long field values (truncation)
  - Very long descriptions (wrapping)
- [ ] Snapshot tests for rendered layouts at 80×24, 120×40, 160×50
- [ ] All tests pass with `go test ./internal/tui/...`

### Test Coverage

- UT-TUI-044: Test empty config edge case
- UT-TUI-045: Test all-sensitive-fields edge case
- UT-TUI-046: Test single-field config
- UT-TUI-047: Test maximum fields (~25)
- UT-TUI-048: Test very long field value truncation
- UT-TUI-049: Test very long description wrapping
- UT-TUI-050: Snapshot test for 80×24 layout
- UT-TUI-051: Snapshot test for 120×40 layout
- UT-TUI-052: Snapshot test for 160×50 layout

---

## Task T3.10: Refactor and Remove Old Huh Implementation

- **Status:** Not Started
- **Complexity:** Low (2-3 hours)
- **Dependencies:** T3.9
- **Related ADRs:** ADR-0002, ADR-0003
- **Related Core-Components:** None

### Description

Remove the old Huh-based form implementation and update dependencies. Update internal documentation.

### Acceptance Criteria

- [ ] Old Huh form code removed from `internal/tui/form.go`
- [ ] Old `BuildForm()` function removed (replaced by new components)
- [ ] Huh imports removed from `internal/tui/model.go`
- [ ] `bubbles` promoted from indirect to direct dependency in `go.mod`
- [ ] Huh dependency removed if no other package uses it
- [ ] `internal/tui/README.md` created documenting:
  - Two-panel architecture
  - Component structure (configlist, detail packages)
  - State machine diagram
  - How to add new field types
- [ ] All existing tests updated or removed as appropriate
- [ ] No broken imports or references to old code

### Test Coverage

- UT-TUI-053: Test that all imports are valid (no Huh references)
- UT-TUI-054: Test that go.mod has bubbles as direct dependency
- Integration test: Full TUI workflow with new implementation
- Manual: Verify README.md documentation is clear and accurate

---

## Task T3.11: Manual UX Testing and Polish

- **Status:** Not Started
- **Complexity:** Medium (3-4 hours)
- **Dependencies:** T3.10
- **Related ADRs:** ADR-0003
- **Related Core-Components:** CC-0004, CC-0005

### Description

Perform comprehensive manual testing of the TUI and address any UX issues, visual glitches, or usability concerns.

### Acceptance Criteria

- [ ] All 25+ config options appear in list and are navigable
- [ ] Arrow key navigation is smooth and responsive
- [ ] Each field type (string, int, bool, list, enum) edits correctly
- [ ] Sensitive fields remain read-only and masked throughout
- [ ] Layout fits on 80×24 terminal without horizontal scrolling
- [ ] Layout fits on 80×24 terminal with vertical scrolling for long lists
- [ ] No visual artifacts during navigation or editing
- [ ] No visual artifacts on resize
- [ ] Alt-screen cleanup works on exit (Ctrl+C, normal quit, error)
- [ ] Config round-trip verified: load → edit → save → external reload shows changes
- [ ] Color scheme works in both light and dark terminals
- [ ] Keyboard shortcuts intuitive and documented in help bar

### Test Coverage

- Manual testing checklist (see acceptance criteria)
- No automated tests for this task (UX validation)

---

## Summary

| Task | Complexity | Dependencies | Estimated Hours |
|------|-----------|--------------|-----------------|
| T3.1 | Low | None | 2-3 |
| T3.2 | Medium | T3.1 | 3-4 |
| T3.3 | Medium | T3.2 | 4-5 |
| T3.4 | Medium | T3.2 | 3-4 |
| T3.5 | High | T3.4 | 6-8 |
| T3.6 | Medium | T3.3, T3.5 | 4-5 |
| T3.7 | Low | T3.6 | 2-3 |
| T3.8 | Medium | T3.6 | 3-4 |
| T3.9 | Medium | T3.1-T3.8 | 4-5 |
| T3.10 | Low | T3.9 | 2-3 |
| T3.11 | Medium | T3.10 | 3-4 |
| **Total** | - | - | **37-48 hours** |

## Task Ordering

```
T3.1 (Alt-Screen)
  ↓
T3.2 (Base Layout)
  ↓
  ├─→ T3.3 (List Model) ──→ T3.6 (State Machine)
  └─→ T3.4 (Detail View) ──→ T3.5 (Input Widgets) ──┘
                                     ↓
                                  T3.7 (Help Bar)
                                     ↓
                                  T3.8 (Config Integration)
                                     ↓
                                  T3.9 (Testing)
                                     ↓
                                  T3.10 (Refactor)
                                     ↓
                                  T3.11 (UX Polish)
```

## Verification Checklist

From [01-action-plan.md](./01-action-plan.md):

- [ ] TUI runs in alt-screen (fullscreen) mode
- [ ] Left panel shows all config options as a navigable list
- [ ] Right panel shows selected field details and edit widget
- [ ] Arrow keys navigate list; Enter focuses edit widget
- [ ] Border frame around entire UI with branded header
- [ ] Sensitive fields remain read-only and masked
- [ ] Config round-trip works (load → edit → save → reload)
- [ ] Layout fits on standard 80×24 terminal
- [ ] No visual artifacts or rendering glitches on exit
