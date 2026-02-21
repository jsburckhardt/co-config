# Research Brief: TUI Two-Panel Redesign

## Scope Classification

- **Scope Type:** workitem
- **Workitem ID:** WI-0003-tui-two-panel-redesign

## Problem Statement

The current TUI implementation uses a single-column vertical scrolling form with Huh groups. While functional, it has several UX limitations:

1. **Space inefficiency**: With only ~25 config options across 5 groups, the vertical layout wastes screen real estate
2. **No focus/detail view**: Long text fields (like `allowed_urls`, `denied_urls`) and descriptions are cramped inline
3. **No visual hierarchy**: Everything has equal visual weight; no clear separation of navigation vs. content
4. **No immersive experience**: Renders inline (not fullscreen), no frame, minimal branding

The user requests a redesigned TUI with:

- **Two-panel layout**: Left panel for compact option list, right panel for detail/editing
- **Compact design**: Tighter spacing given the small number of options
- **Fullscreen mode**: Alt-screen buffer with a border frame around the entire UI
- **Branded header**: Icon/logo at the top for visual identity

## Existing Context

### Current Implementation

- **Location**: `/workspaces/co-config/internal/tui/`
- **Stack**: Bubbletea (model) + Huh (forms) + Lipgloss (styling)
- **Layout**: Single-column vertical form with 5 groups (General, Model & AI, URLs & Permissions, Display, Sensitive)
- **Form fields**: `huh.NewConfirm()`, `huh.NewSelect()`, `huh.NewInput()`, `huh.NewText()`, `huh.NewNote()`
- **No alt-screen**: Renders inline in terminal
- **No framing**: Plain output

### Relevant ADRs

- **ADR-0002**: Go with Charm TUI Stack (Accepted) — mandates Bubbletea, Lipgloss, Huh. Huh's design assumptions (vertical full-screen forms) don't fit the two-panel requirement.

### Relevant Core-Components

- **CC-0004**: Configuration Management — config fields categorized into 5 groups; schema-driven form building
- **CC-0005**: Sensitive Data Handling — some fields are read-only (masked); must remain non-editable

### Technical Constraints

- Huh forms are self-contained and expect to own full-screen layout
- `bubbles v0.21.1` is available as an indirect dependency (list, textinput, textarea, viewport)
- Bubbletea supports alt-screen via `tea.WithAltScreen()`

## Proposed ADRs

### ADR-0003: Two-Panel TUI Layout Pattern

Move from Huh forms to a custom Bubbletea model using:

- **Left panel**: Bubbles `list.Model` — compact navigable list of config options with truncated current values
- **Right panel**: Custom detail view — field name, description, current value, input widget for editing
- **Alt-screen mode**: `tea.EnterAltScreen` for fullscreen
- **Framed layout**: Lipgloss border styles around panels and outer frame
- **Header**: Styled branding with icon and title

**Alternatives considered**:

| Alternative | Why Rejected |
|---|---|
| Keep Huh forms, add custom styling | Huh doesn't support two-panel layouts natively |
| Third-party TUI framework (tview) | Violates ADR-0002 (Charm stack mandate) |

## Proposed Core-Components

None required. This is a UI implementation detail that doesn't affect cross-cutting concerns.

## Risks

1. **Huh abandonment**: Contradicts ADR-0002's rationale for Huh, but the two-panel requirement necessitates it. Still uses the Charm ecosystem.
2. **Input handling complexity**: Must use `bubbles/textinput`, `bubbles/textarea`, custom select — more code than Huh forms.
3. **Accessibility**: Must replicate Huh's keyboard navigation hints with a footer help bar.
4. **State management**: Need a clear state machine (browsing → editing → saving).

## Open Questions

1. List fields: multi-line textarea (MVP) vs. sub-list editor (future)
2. Left panel: show truncated current values inline (recommended) vs. names only
3. Sensitive fields: appear in list as `(read-only)`, right panel shows masked value
4. Icon/branding: simple styled text for MVP, defer ASCII art
5. Frame/border: hardcoded for MVP, themeable later

## Verification Strategy

- TUI runs in alt-screen (fullscreen) mode
- Left panel shows all config options as a navigable list
- Right panel shows selected field details and edit widget
- Arrow keys navigate; Enter/Tab focuses edit widget
- Border frame around entire UI
- Sensitive fields remain read-only and masked
- Config round-trip works (load → edit → save → reload)
- Layout fits on standard 80×24 terminal

## Architect Handoff

- **ADR-0003** must be written to document the move from Huh forms to custom Bubbletea + Bubbles
- Action plan should cover: ADR approval, TUI package refactoring, left panel, right panel, alt-screen/framing, testing
