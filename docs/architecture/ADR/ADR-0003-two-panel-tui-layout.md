# ADR-0003: Two-Panel TUI Layout Pattern

## Status

Accepted

## Context

The current TUI implementation (from ADR-0002) uses Huh forms in a single-column vertical scrolling layout. While functional, it has several UX limitations:

1. **Space inefficiency**: With only ~25 config options across 5 groups, the vertical layout wastes screen real estate on typical terminals
2. **No focus/detail view**: Long text fields (like `allowed_urls`, `denied_urls`) and descriptions are cramped inline with no room for contextual help
3. **No visual hierarchy**: Everything has equal visual weight; there's no clear separation of navigation vs. content editing
4. **No immersive experience**: Renders inline (not fullscreen), no frame, minimal branding, making it feel less like a dedicated tool

User feedback indicates a preference for a two-panel layout similar to modern TUI editors (e.g., k9s, lazygit):
- **Left panel**: Compact navigable list of config options with truncated current values
- **Right panel**: Detail view showing field name, description, current value, and editing widget
- **Fullscreen mode**: Alt-screen buffer with a border frame around the entire UI
- **Branded header**: Icon/logo at the top for visual identity

Huh forms are self-contained and expect to own the full-screen layout. They don't support two-panel layouts natively. However, `bubbles` (already an indirect dependency) provides lower-level components (`list`, `textinput`, `textarea`, `viewport`) that can be composed into a custom layout.

## Decision

We will **move from Huh forms to a custom Bubbletea model** using lower-level Bubbles components to implement a two-panel layout:

### Architecture

- **Left panel**: `bubbles/list.Model` — compact navigable list of config options with group headers and truncated current values
- **Right panel**: Custom detail view
  - Field name (styled header)
  - Description text (wrapped)
  - Current value display
  - Input widget for editing: `bubbles/textinput` for single-line, `bubbles/textarea` for multi-line, custom select widget for enums
- **Alt-screen mode**: Enable fullscreen rendering via `tea.WithAltScreen()` in the Bubbletea program options
- **Framed layout**: Lipgloss border styles around both panels and an outer frame for the entire UI
- **Header**: Styled branding section with icon (simple styled text for MVP) and title

### State Management

The model will use a clear state machine:

1. **Browsing**: Arrow keys navigate the left panel list; Enter switches to editing mode
2. **Editing**: Right panel input widget is focused; Esc saves and returns to browsing
3. **Saving**: On exit, persist changes to config file (leveraging CC-0004 configuration management)

### Component Breakdown

| Component | Library | Purpose |
|-----------|---------|---------|
| List navigation | `bubbles/list` | Left panel option browser |
| Single-line input | `bubbles/textinput` | Edit strings, numbers |
| Multi-line input | `bubbles/textarea` | Edit lists, long text |
| Layout styling | `lipgloss` | Borders, spacing, colors |
| Help bar | `bubbles/help` | Keyboard shortcuts in footer |

### Compatibility with Existing ADRs/Core-Components

- **ADR-0002 (Charm stack)**: Still uses Bubbletea + Lipgloss; replaces Huh with Bubbles (both are Charm libraries)
- **CC-0004 (Configuration Management)**: Schema-driven form building remains; list of fields comes from config schema
- **CC-0005 (Sensitive Data Handling)**: Sensitive fields appear in left panel marked `(read-only)`, right panel shows masked value, no editing allowed

## Alternatives

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| Keep Huh forms, add custom styling | No code rewrite needed, stays on documented path | Huh doesn't support two-panel layouts natively; would require forking or extensive monkey-patching | Too much effort for marginal gain; fundamentally incompatible with two-panel UX |
| Use third-party TUI framework (tview, cview) | Mature widget libraries with built-in panel support | Violates ADR-0002 mandate for Charm stack; different styling language | Consistency with existing ADR is critical; Charm ecosystem is already integrated |
| Build custom widgets from scratch (no Bubbles) | Maximum control, no library constraints | Reimplementing textinput/textarea is 1000+ lines of code; high maintenance burden | Bubbles already provides battle-tested input components; no need to reinvent |
| Keep inline rendering, skip fullscreen | Simpler implementation, no alt-screen handling | Misses the "immersive experience" goal; no room for framing/branding | User specifically requested fullscreen mode |

## Consequences

### Positive

- **Better space utilization**: Two-panel layout fits ~80% of config options above the fold on a standard 80×24 terminal
- **Focused editing**: Right panel provides room for descriptions, examples, and validation feedback without cluttering the list
- **Visual hierarchy**: Clear separation of "what" (left panel) vs. "how/why" (right panel)
- **Immersive UX**: Alt-screen + framing makes it feel like a dedicated tool rather than inline form output
- **Reusable pattern**: This two-panel layout can be adapted for future TUI features (e.g., diff viewer, log browser)
- **Still Charm ecosystem**: Uses Bubbletea + Lipgloss + Bubbles — all from the same maintainer and design philosophy

### Negative

- **Huh abandonment**: Contradicts ADR-0002's rationale for Huh, though the two-panel requirement necessitates it
- **More code**: Custom state management, input handling, and layout code (~500 lines) vs. Huh's declarative forms (~100 lines)
- **Accessibility parity**: Must manually replicate Huh's keyboard navigation hints with a footer help bar
- **Maintenance**: More surface area for bugs (focus handling, panel resizing, input validation)

### Neutral

- **Learning curve**: Contributors need to understand Bubbletea's Elm architecture and Bubbles components (but this was already true for ADR-0002)
- **Testing strategy**: Unit tests for model state transitions + snapshot tests for rendered layouts

## Related Workitems

- [WI-0003-tui-two-panel-redesign](../../workitems/WI-0003-tui-two-panel-redesign/)

## References

- [Bubbletea Alt-Screen Docs](https://github.com/charmbracelet/bubbletea#advanced-options)
- [Bubbles List Component](https://github.com/charmbracelet/bubbles/tree/master/list)
- [Bubbles Text Input](https://github.com/charmbracelet/bubbles/tree/master/textinput)
- [Bubbles Text Area](https://github.com/charmbracelet/bubbles/tree/master/textarea)
- [Lipgloss Layout Examples](https://github.com/charmbracelet/lipgloss#layout)
- [ADR-0002: Go with Charm TUI Stack](ADR-0002-go-charm-tui-stack.md)
- [CC-0004: Configuration Management](../core-components/CORE-COMPONENT-0004-configuration-management.md)
- [CC-0005: Sensitive Data Handling](../core-components/CORE-COMPONENT-0005-sensitive-data-handling.md)
