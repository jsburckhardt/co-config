# CORE-COMPONENT-0002: Error Handling

## Status

Adopted

## Purpose

Define a consistent error handling strategy across the `ccc` codebase so that errors are properly wrapped with context, presented to users in a friendly manner through the TUI, and are debuggable via logs.

## Scope

All packages in the `ccc` application. Affects how errors propagate from config parsing, CLI detection, and TUI rendering back to the user.

## Definition

### Rules
- Use `fmt.Errorf` with `%w` verb for error wrapping to preserve error chains
- Define sentinel errors in each package for expected failure modes (e.g., `ErrConfigNotFound`, `ErrCopilotNotInstalled`)
- Never expose raw Go errors to users in the TUI — always map to user-friendly messages
- Errors that prevent the tool from functioning (e.g., copilot not installed) should display a styled error screen, not a panic

### Interfaces
- Each package should define its own error types/sentinels in an `errors.go` file when the package has more than two distinct error conditions
- Use `errors.Is()` and `errors.As()` for error checking — never string comparison

### Expectations
- All public functions return `error` as the last return value following Go conventions
- Errors must include enough context to identify the failure point without a stack trace
- The TUI layer is responsible for translating errors into user-facing messages

## Rationale

Go's explicit error handling is a strength when used consistently. Wrapping with `%w` allows callers to inspect error chains. Sentinel errors make testing straightforward. User-facing error messages should be helpful, not cryptic.

## Usage Examples

```go
// Sentinel error definition
var ErrConfigNotFound = errors.New("copilot config file not found")

// Wrapping with context
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("%w: %s", ErrConfigNotFound, path)
        }
        return nil, fmt.Errorf("reading config file: %w", err)
    }
    // ...
}

// Checking in caller
cfg, err := LoadConfig(path)
if errors.Is(err, ErrConfigNotFound) {
    // Show "no config found" screen in TUI
}
```

## Integration Guidelines

- Import sentinel errors from the defining package
- Use `errors.Is()` to branch on expected errors
- Log the full error chain at debug level; show only the user-friendly message in the TUI

## Exceptions

- Quick prototype code during early development may use simpler error handling, but must be refactored before shipping

## Enforcement

- [x] Code review checklist
- [ ] Automated checks (future: linter rules)
- [x] Test coverage requirements — sentinel errors must have test cases

## Related ADRs

- [ADR-0002-go-charm-tui-stack](../ADR/ADR-0002-go-charm-tui-stack.md)
