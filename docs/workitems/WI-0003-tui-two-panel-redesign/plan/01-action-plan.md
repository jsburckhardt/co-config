# Action Plan: TUI Two-Panel Redesign

## Feature

- **ID:** WI-0003-tui-two-panel-redesign
- **Research Brief:** [00-research.md](../research/00-research.md)

## ADRs Created

- [ADR-0003: Two-Panel TUI Layout Pattern](../../../architecture/ADR/ADR-0003-two-panel-tui-layout.md) — Accepted

## Core-Components Created

None. This is a UI implementation detail that doesn't affect cross-cutting concerns.

## Implementation Tasks

### Phase 1: Foundation & Alt-Screen Setup

**Goal:** Set up fullscreen rendering and basic layout structure.

1. **Enable Alt-Screen Mode**
   - Update `internal/tui/tui.go` to initialize Bubbletea with `tea.WithAltScreen()`
   - Verify fullscreen rendering works and clears on exit
   - Add cleanup logic to ensure terminal state is restored on panic

2. **Create Base Layout Structure**
   - Define Lipgloss styles for outer frame, header, left panel, right panel
   - Implement basic two-panel layout with fixed widths (e.g., 30% left, 70% right)
   - Add branded header section with styled title (defer ASCII art to future iteration)
   - Verify layout renders correctly on 80×24 terminal

### Phase 2: Left Panel (List Navigation)

**Goal:** Replace Huh form groups with a navigable list of config options.

3. **Integrate `bubbles/list` Component**
   - Create `internal/tui/configlist/` package for list model
   - Populate list items from config schema (leverage CC-0004's field metadata)
   - Group items by category (General, Model & AI, URLs, Display, Sensitive)
   - Show truncated current value inline (e.g., `"api_key: abc...789"`)
   - Implement keyboard navigation (Up/Down arrows, j/k vim bindings)

4. **Add Sensitive Field Indicators**
   - Mark sensitive fields with `(read-only)` suffix in list (per CC-0005)
   - Style sensitive items distinctly (e.g., dimmed, lock icon)

### Phase 3: Right Panel (Detail View)

**Goal:** Display focused field details and enable editing.

5. **Create Detail View Component**
   - Create `internal/tui/detail/` package for right panel rendering
   - Display field name (styled header)
   - Display description text (wrapped with Lipgloss)
   - Display current value with type-appropriate formatting

6. **Integrate Input Widgets**
   - **Single-line fields**: Use `bubbles/textinput` for strings, integers
   - **Multi-line fields**: Use `bubbles/textarea` for lists (`allowed_urls`, `denied_urls`)
   - **Boolean fields**: Create custom toggle component (similar to Huh's confirm)
   - **Enum fields**: Create custom select component (show options as radio list)
   - Wire up input validation (reuse existing validation logic from Huh implementation if available)

7. **Handle Sensitive Fields**
   - Display masked value (per CC-0005: truncated SHA-256 hash)
   - Disable editing (focus stays on list when Enter is pressed on sensitive field)
   - Show read-only indicator in detail panel

### Phase 4: State Management & Navigation

**Goal:** Implement state machine for browsing ↔ editing transitions.

8. **Define State Machine**
   - States: `Browsing`, `Editing`, `Saving`, `Exiting`
   - `Browsing`: Arrow keys navigate list, Enter switches to `Editing`
   - `Editing`: Right panel input widget is focused, Esc saves and returns to `Browsing`
   - `Saving`: Persist changes to config file, then transition to `Browsing` or `Exiting`

9. **Implement Focus Handling**
   - Track which panel/widget has focus
   - Route keyboard input to active component
   - Visual indicator for focused panel (border color change, cursor visibility)

10. **Add Footer Help Bar**
    - Use `bubbles/help` component for keyboard shortcuts
    - Context-sensitive hints (e.g., "Enter: Edit" in browsing mode, "Esc: Save" in editing mode)

### Phase 5: Integration & Testing

**Goal:** Wire up config persistence and validate UX.

11. **Integrate with CC-0004 (Configuration Management)**
    - Load config schema and current values on startup
    - Validate input against schema constraints (type, format, allowed values)
    - Save changes on Esc or exit (preserve unknown fields per CC-0004)
    - Handle config file not found / malformed cases

12. **Testing**
    - **Unit tests**: State transitions, input validation, list population
    - **Snapshot tests**: Rendered layout at different terminal sizes
    - **Manual UX testing**:
      - All 25 config options appear in list
      - Navigating with arrow keys works smoothly
      - Editing each field type (string, int, bool, list, enum) works
      - Sensitive fields remain read-only and masked
      - Layout fits on 80×24 terminal without scrolling (for most fields)
      - Config round-trip works (load → edit → save → reload)
      - Alt-screen cleans up on exit

13. **Refactoring & Cleanup**
    - Remove old Huh form code from `internal/tui/`
    - Update imports and remove unused Huh dependency (if no other use)
    - Document component structure in `internal/tui/README.md`

### Phase 6: Polish (Optional / Future)

**Goal:** Enhance UX beyond MVP.

14. **Advanced List Editing**
    - For list fields (`allowed_urls`), implement sub-list editor instead of textarea (add/remove/edit individual items)

15. **Theming**
    - Extract colors/styles into a theme struct
    - Support light/dark themes via config or auto-detection

16. **Branding**
    - Replace styled text header with ASCII art logo (defer to avoid scope creep)

## Dependencies

- `bubbles` v0.21.1 (already indirect dependency, make direct)
- Existing ADRs: ADR-0002 (Charm stack), ADR-0003 (Two-panel layout)
- Existing Core-Components: CC-0004 (Configuration Management), CC-0005 (Sensitive Data Handling)

## Verification Checklist

- [ ] TUI runs in alt-screen (fullscreen) mode
- [ ] Left panel shows all config options as a navigable list
- [ ] Right panel shows selected field details and edit widget
- [ ] Arrow keys navigate list; Enter focuses edit widget
- [ ] Border frame around entire UI with branded header
- [ ] Sensitive fields remain read-only and masked
- [ ] Config round-trip works (load → edit → save → reload)
- [ ] Layout fits on standard 80×24 terminal
- [ ] No visual artifacts or rendering glitches on exit

## Estimated Effort

- Phase 1-2: 4-6 hours (alt-screen setup, list integration)
- Phase 3: 6-8 hours (detail view, input widgets)
- Phase 4: 4-6 hours (state machine, focus handling)
- Phase 5: 4-6 hours (integration, testing, cleanup)
- **Total MVP**: 18-26 hours

Phase 6 (polish) is optional and can be deferred to future workitems.
