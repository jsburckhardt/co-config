# ADR-0002: Go with Charm TUI Stack

## Status

Accepted

## Context

We need to build `ccc` (Copilot Config CLI), a terminal-based interactive tool for configuring GitHub Copilot CLI settings. The tool requires a rich TUI with forms, toggles, selects, and styled output. The devcontainer already includes Go.

## Decision

We will use **Go** as the implementation language with the **Charm** ecosystem for the TUI:

- **[Bubbletea](https://github.com/charmbracelet/bubbletea)** — Elm-architecture TUI framework for the application shell
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — Terminal styling and layout
- **[Huh](https://github.com/charmbracelet/huh)** — Form components (toggles, selects, text inputs) for config editing
- **go modules** — Dependency management
- **go test** — Testing

The binary will be named `ccc` (Copilot Config CLI).

## Alternatives

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| Python + Textual | Rich widget library, rapid prototyping | Requires Python runtime, slower startup, packaging complexity | Go produces a single static binary — simpler distribution |
| Rust + Ratatui | High performance, strong type safety | Steeper learning curve, slower iteration, not in devcontainer | Go already available, Charm ecosystem is more mature for forms |
| Go + tview | Stable, widget-based | Less composable, dated look, no built-in form library | Charm stack is more modern and produces better-looking UIs |

## Consequences

### Positive
- Single static binary — easy to distribute and install
- Charm ecosystem is the most popular Go TUI stack with excellent documentation
- Huh provides ready-made form components that match our config-editing use case perfectly
- Fast startup time ideal for a CLI tool

### Negative
- Elm architecture has a learning curve for contributors unfamiliar with it
- Charm libraries are actively evolving — API changes possible

### Neutral
- Go's error handling verbosity is standard for the language

## Related Workitems

- [WI-0002-ccc-bootstrap](../../workitems/WI-0002-ccc-bootstrap/)

## References

- [Bubbletea](https://github.com/charmbracelet/bubbletea)
- [Lipgloss](https://github.com/charmbracelet/lipgloss)
- [Huh](https://github.com/charmbracelet/huh)
- [Charm ecosystem](https://charm.sh/)
