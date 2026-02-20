# CORE-COMPONENT-0005: Sensitive Data Handling

## Status

Adopted

## Purpose

Ensure that sensitive data in the Copilot CLI config (tokens, credentials, user identities) is never exposed in the TUI in plain text, never logged, and never accidentally modified.

## Scope

Affects the `internal/config` package (field classification), the TUI layer (display masking), and the `internal/logging` package (log sanitization).

## Definition

### Rules
- The following config fields are classified as **sensitive**: `copilot_tokens`, `logged_in_users`, `last_logged_in_user`, `staff`
- Sensitive fields are **read-only** in the TUI — users can view them but not edit them
- Sensitive string values are displayed as a SHA-256 hash truncated to 12 characters (e.g., `a1b2c3d4e5f6...`)
- Sensitive object/array values are displayed as `[redacted — N items]` or similar summary
- Sensitive values are **never logged** at any log level
- When saving config, sensitive fields are written back exactly as they were read — no transformation

### Interfaces
- `IsSensitive(fieldName string) bool` — checks if a field is classified as sensitive
- `MaskValue(value any) string` — returns a display-safe representation of a sensitive value
- `SensitiveFields` — a package-level list of known sensitive field names

### Expectations
- The sensitive field list is maintained in code and can be extended as new sensitive fields appear in future copilot versions
- If a field not in the sensitive list contains a value that looks like a token (e.g., starts with `gho_`, `ghp_`, `github_pat_`), it should be treated as sensitive
- Hash-based masking ensures users can verify "this is the same token" without seeing the actual value

## Rationale

The config file contains OAuth tokens and user identity information. Displaying these in plain text in a TUI creates a risk of shoulder-surfing, screen recording exposure, and accidental copy-paste. Hashing provides verifiability without exposure.

## Usage Examples

```go
// Check if a field is sensitive
if sensitive.IsSensitive("copilot_tokens") {
    display = sensitive.MaskValue(value)
    // render as read-only in TUI
}

// Mask a token value
masked := sensitive.MaskValue("gho_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
// Returns: "a1b2c3d4e5f6..."

// Auto-detect token-like values
if sensitive.LooksLikeToken("gho_something") {
    // treat as sensitive even if field name is not in the list
}
```

## Integration Guidelines

- The config loader should annotate each field with `IsSensitive` metadata
- The TUI form builder should check sensitivity before creating editable vs. read-only fields
- The config writer should use the original raw values for sensitive fields, never the masked display values

## Exceptions

- During development/testing with dummy tokens, sensitivity checks can be bypassed via a test flag — never in production builds

## Enforcement

- [x] Code review checklist — any new config field must be evaluated for sensitivity
- [x] Test coverage requirements — masking functions must be tested, round-trip must verify sensitive values unchanged
- [ ] Automated checks (future: scanner for token patterns in log output)

## Related ADRs

- [ADR-0002-go-charm-tui-stack](../ADR/ADR-0002-go-charm-tui-stack.md)
