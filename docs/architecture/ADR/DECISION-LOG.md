# Decision Log

This file is the single registry of all architectural decisions and core-components in the project. Every new or modified ADR or core-component **must** be recorded here.

## ADRs

| ID | Title | Status | Date |
|----|-------|--------|------|
| ADR-0002 | [Go with Charm TUI Stack](ADR-0002-go-charm-tui-stack.md) | Accepted | 2026-02-20 |
| ADR-0003 | [Two-Panel TUI Layout Pattern](ADR-0003-two-panel-tui-layout.md) | Accepted | 2026-02-20 |

## Core-Components

| ID | Title | Status | Date |
|----|-------|--------|------|
| CC-0002 | [Error Handling](../core-components/CORE-COMPONENT-0002-error-handling.md) | Adopted | 2026-02-20 |
| CC-0003 | [Logging](../core-components/CORE-COMPONENT-0003-logging.md) | Adopted | 2026-02-20 |
| CC-0004 | [Configuration Management](../core-components/CORE-COMPONENT-0004-configuration-management.md) | Adopted | 2026-02-20 |
| CC-0005 | [Sensitive Data Handling](../core-components/CORE-COMPONENT-0005-sensitive-data-handling.md) | Adopted | 2026-02-20 |

## Decisions

Short, actionable statements derived from ADRs and core-components. More than one decision can originate from a single source.

| # | Decision | Source | Date |
|---|----------|--------|------|
| 1 | Use Go as the implementation language for `ccc` | ADR-0002 | 2026-02-20 |
| 2 | Use Bubbletea + Lipgloss + Huh (Charm stack) for the TUI | ADR-0002 | 2026-02-20 |
| 3 | Use `go test` as the test runner | ADR-0002 | 2026-02-20 |
| 4 | Wrap errors with `fmt.Errorf` and `%w`; define sentinel errors per package | CC-0002 | 2026-02-20 |
| 5 | Log to file (`~/.copilot/ccc.log`) using `log/slog`; never write to stdout | CC-0003 | 2026-02-20 |
| 6 | Auto-detect config schema by running `copilot help config` at startup | CC-0004 | 2026-02-20 |
| 7 | Preserve unknown config fields on round-trip (no data loss) | CC-0004 | 2026-02-20 |
| 8 | Mask sensitive fields with truncated SHA-256 hash; make them read-only | CC-0005 | 2026-02-20 |
| 9 | Never log sensitive data (tokens, credentials) at any level | CC-0005 | 2026-02-20 |
| 10 | Use Cobra for CLI argument parsing and subcommands | ADR-0002 | 2026-02-20 |
| 11 | Use Viper for configuration file binding and env var support | ADR-0002 | 2026-02-20 |
| 12 | Replace Huh forms with custom Bubbletea + Bubbles two-panel layout | ADR-0003 | 2026-02-20 |
| 13 | Use `bubbles/list` for left panel config option navigation | ADR-0003 | 2026-02-20 |
| 14 | Use `bubbles/textinput` and `bubbles/textarea` for right panel editing | ADR-0003 | 2026-02-20 |
| 15 | Enable fullscreen mode via `tea.WithAltScreen()` for immersive UX | ADR-0003 | 2026-02-20 |
| 16 | Frame the entire TUI with Lipgloss borders and branded header | ADR-0003 | 2026-02-20 |
