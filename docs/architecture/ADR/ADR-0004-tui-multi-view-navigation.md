# ADR-0004: TUI Multi-View Tab Navigation

## Status

Accepted

## Context

ADR-0003 established a two-panel TUI layout (list panel + detail panel) for browsing and editing config options. The entire TUI currently occupies a single "view" — the config view — with only vertical navigation (up/down through the list, Enter/Esc for editing).

WI-0005 introduces a new **environment variables view** that displays the environment variables affecting GitHub Copilot CLI behaviour (sourced from `copilot help environment`). This view is logically separate from the config view: it is read-only, displays different data (process env vars vs. config file fields), and requires its own layout.

Adding this second top-level view requires a horizontal navigation model that ADR-0003 does not define. The TUI needs a way to switch between named views — a "multi-view tab navigation" pattern. This pattern must also be extensible for potential future views (e.g., a log viewer, a diff viewer).

The existing `Tab` key binding in `KeyMap` is defined but never wired in `handleKeyPress`, suggesting the codebase was already anticipating horizontal navigation.

## Decision

We will extend the TUI with a **multi-view tab navigation model** using the following rules:

### Views vs. Panels

- A **view** is a top-level screen that occupies the full TUI content area (between the header and help bar). Each view has its own state, layout, and key handling. Examples: Config View, Env Vars View.
- A **panel** is a sub-component within a view. The Config View contains two panels (list + detail) as defined by ADR-0003. The Env Vars View contains one panel (scrollable env var list).
- Views are navigated horizontally (left/right/tab). Panels within a view are navigated according to the view's own rules (e.g., up/down in lists, Enter/Esc for editing).

### New State: `StateEnvVars`

Add `StateEnvVars` to the state machine. This state represents the Env Vars View being active.

```
StateBrowsing ←──→ StateEnvVars
     │
     ↓
StateEditing
```

- `StateEditing` is only reachable from `StateBrowsing` (config view). The env vars view is entirely read-only.
- `StateSaving` and `StateExiting` remain unchanged.

### Key Bindings for View Switching

| Key | From State | To State | Behaviour |
|-----|-----------|----------|-----------|
| `right`, `l`, `tab` | `StateBrowsing` | `StateEnvVars` | Switch to Env Vars View |
| `left`, `h`, `tab` | `StateEnvVars` | `StateBrowsing` | Switch to Config View |

- In `StateBrowsing`: `left`/`right`/`h`/`l` and `tab` all transition to `StateEnvVars`.
- In `StateEnvVars`: `left`/`right`/`h`/`l` and `tab` all transition back to `StateBrowsing`.
- In `StateEditing`: `left`/`right` are forwarded to the input widget (text cursor movement). No view switching occurs during editing.
- `tab` acts as a toggle between the two views in either direction.

### Help Bar Per View

The help bar (`ShortHelp(state)`) must reflect the active view's available actions:

| State | Keys shown |
|-------|-----------|
| `StateBrowsing` | `↑/k up • ↓/j down • enter edit • →/tab env vars • ctrl+s save • ctrl+c quit` |
| `StateEditing` | `esc done • ctrl+s save • ctrl+c quit` |
| `StateEnvVars` | `↑/k up • ↓/j down • ←/tab config • ctrl+c quit` |

The `StateEnvVars` help bar omits `enter`, `ctrl+s`, and `esc` because the env vars view is read-only.

### Rendering in `View()`

The `View()` method must branch on `m.state`:
- `StateBrowsing` or `StateEditing`: render the existing two-panel config layout (ADR-0003).
- `StateEnvVars`: render the env vars panel (full-width, scrollable list of environment variable entries).

The header and outer frame remain constant across views. Only the content area between the header and help bar changes.

### Env Vars View Is Read-Only

The environment variables view displays information sourced from the process environment and `copilot help environment` metadata. It cannot modify environment variables (they are set outside the process). No editing, no saving, no Enter-to-edit. This is an intentional UX constraint, not a limitation.

## Alternatives

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| Amend ADR-0003 with horizontal navigation | Fewer ADR documents; keeps layout decisions together | ADR-0003 is about the two-panel layout pattern, not navigation between views; amending it conflates two concerns | Separation of concerns: ADR-0003 defines panel layout, ADR-0004 defines view navigation |
| Third column in existing layout | All data visible at once; no navigation needed | 3 columns at 80–120 chars is unreadably narrow; contradicts ADR-0003's 40/60 split | Terminal width constraint makes this impractical |
| Env vars as a collapsible group in the config list | No new navigation keys; consistent with existing list | Env vars are read-only and logically separate from editable config; mixing them creates UX confusion | Blurs the distinction between editable config and read-only env vars |
| Separate `ccc env` subcommand | No TUI changes needed; clean CLI separation | Doesn't fulfil the feature request for TUI-integrated navigation; loses live value display context | User explicitly requested TUI integration |

## Consequences

### Positive
- Establishes a reusable pattern for adding future views without modifying existing view code
- Clear separation of read-only (env vars) and read-write (config) concerns in the UX
- Multiple navigation affordances (arrow keys + h/l + tab) makes discovery easy
- Wires the previously-unused `Tab` binding already defined in `KeyMap`
- No changes to ADR-0003's two-panel layout — it remains intact within the Config View

### Negative
- Adds a new navigation dimension that users must discover (mitigated by help bar)
- `View()` branching adds complexity to the rendering path
- State machine grows from 4 to 5 states

### Neutral
- The env vars view's scrolling behaviour mirrors the config list panel's up/down navigation — familiar to users
- Future views (if any) would follow this same pattern: add a state, add key bindings, branch in `View()`

## Related Workitems

- [WI-0005-environment-variables](../../workitems/WI-0005-environment-variables/)

## References

- [ADR-0003: Two-Panel TUI Layout Pattern](ADR-0003-two-panel-tui-layout.md)
- [ADR-0002: Go with Charm TUI Stack](ADR-0002-go-charm-tui-stack.md)
- [CC-0004: Configuration Management](../core-components/CORE-COMPONENT-0004-configuration-management.md)
- [CC-0005: Sensitive Data Handling](../core-components/CORE-COMPONENT-0005-sensitive-data-handling.md)
