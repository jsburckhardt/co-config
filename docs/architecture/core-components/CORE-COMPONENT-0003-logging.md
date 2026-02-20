# CORE-COMPONENT-0003: Logging

## Status

Adopted

## Purpose

Provide structured, leveled logging for debugging and diagnostics without polluting the TUI output. Logs should help developers and users troubleshoot issues when the TUI doesn't provide enough information.

## Scope

All packages in the `ccc` application. Logging output must not interfere with bubbletea's terminal rendering.

## Definition

### Rules
- Use Go's `log/slog` package (standard library, structured logging, available since Go 1.21)
- Log output goes to a file (`~/.copilot/ccc.log`), never to stdout/stderr (bubbletea owns the terminal)
- Default log level is `warn`; configurable via `--log-level` flag or `CCC_LOG_LEVEL` environment variable
- Log entries must include structured fields (key-value pairs), not interpolated strings
- Never log sensitive data (tokens, credentials) — even at debug level

### Interfaces
- A single `internal/logging` package initializes the global `slog.Logger`
- Other packages use `slog.Debug()`, `slog.Info()`, `slog.Warn()`, `slog.Error()` directly

### Expectations
- Log file is created on first write, not on startup
- Log rotation is out of scope for v1 — users can delete the file manually
- Debug-level logs include function names and relevant variable values
- Error-level logs include the full error chain

## Rationale

`log/slog` is the standard structured logging package in Go since 1.21. It avoids external dependencies, integrates naturally with Go idioms, and supports JSON output for machine parsing. File-based logging is required because bubbletea controls the terminal.

## Usage Examples

```go
// Initialization (in internal/logging)
func Init(level slog.Level, logPath string) error {
    f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return fmt.Errorf("opening log file: %w", err)
    }
    handler := slog.NewTextHandler(f, &slog.HandlerOptions{Level: level})
    slog.SetDefault(slog.New(handler))
    return nil
}

// Usage in any package
slog.Debug("loading config", "path", configPath)
slog.Error("failed to parse config", "error", err, "path", configPath)
```

## Integration Guidelines

- Call `logging.Init()` in `main()` before starting the TUI
- Use `slog` functions directly — no wrapper needed
- Include relevant context as structured fields

## Exceptions

- During tests, logging can be disabled or redirected to `io.Discard`

## Enforcement

- [x] Code review checklist — no `fmt.Println` or `log.Println` in non-test code
- [ ] Automated checks (future: linter rule for direct stdout writes)
- [x] Test coverage requirements — logging init must be tested

## Related ADRs

- [ADR-0002-go-charm-tui-stack](../ADR/ADR-0002-go-charm-tui-stack.md)
