# Action Plan: Show Default Values for Unset Config Keys

## Feature
- **ID:** WI-0004-show-default-values
- **Research Brief:** [00-research.md](../research/00-research.md)

## ADRs Created

None — this is a localized UI formatting change that follows existing architectural decisions:
- [ADR-0003: Two-Panel TUI Layout Pattern](../../../architecture/ADR/ADR-0003-two-panel-tui-layout.md) — defines the list/detail panel rendering model
- [CC-0004: Configuration Management](../../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md) — schema fields with defaults are already loaded at startup
- [CC-0005: Sensitive Data Handling](../../../architecture/core-components/CORE-COMPONENT-0005-sensitive-data-handling.md) — sensitive fields short-circuit before display and are unaffected

## Core-Components Created

None — no new cross-cutting concerns are introduced.

## Implementation Tasks

### Task 1: Update `formatValueCompact` to show defaults (list panel)

**File:** `internal/tui/list_item.go`

1. Add a `defaultVal string` parameter to `formatValueCompact`:
   ```go
   func formatValueCompact(val any, defaultVal string, maxLen int) string {
   ```
2. Update the `case nil` branch:
   ```go
   case nil:
       if defaultVal != "" {
           s = defaultVal + " (default)"
       } else {
           s = "(not set)"
       }
   ```
3. Update the call site in `renderItem` (line 159) to pass `item.Field.Default`:
   ```go
   val = formatValueCompact(item.Value, item.Field.Default, valWidth)
   ```

**Verification:** Existing truncation logic (`if len(s) > maxLen`) already applies to the composed string, so `"auto (default)"` will be truncated correctly if the column is too narrow.

### Task 2: Update `renderCurrentValue` to show defaults (detail panel)

**File:** `internal/tui/detail_panel.go`

1. Modify `renderCurrentValue` to check for a default when value is nil:
   ```go
   func (d *DetailPanel) renderCurrentValue() string {
       if d.value == nil && d.field.Default != "" {
           return detailValueStyle.Render(d.field.Default + " (default)")
       }
       return detailValueStyle.Render(formatValueDetail(d.value))
   }
   ```

This approach (Option A from research) keeps `formatValueDetail` unchanged and avoids signature changes downstream.

### Task 3: Update existing tests

**File:** `internal/tui/tui_test.go`

1. Update `TestFormatValueCompact` (UT-TUI-012):
   - Add `defaultVal string` field to the test struct
   - Update all existing test cases to pass `defaultVal: ""` (preserving current behavior)
   - Update the `nil` test case: `{val: nil, defaultVal: "", want: "(not set)"}`
   - Update call sites: `formatValueCompact(tt.value, tt.defaultVal, tt.maxLen)`

2. Add new test cases for default display:
   ```go
   {"nil with default", nil, "auto", 20, "auto (default)"},
   {"nil with bool default", nil, "false", 20, "false (default)"},
   {"nil with long default truncated", nil, "very-long-default", 10, "very-lo..."},
   {"nil no default", nil, "", 10, "(not set)"},
   ```

### Task 4: Add new tests for detail panel default display

**File:** `internal/tui/tui_test.go`

1. Add `TestDetailPanelRenderUnsetWithDefault`:
   - Create a `DetailPanel`, set size, call `SetField` with a field that has a default and `nil` value
   - Assert the rendered view contains the default value and `"default"` annotation

2. Add `TestDetailPanelRenderUnsetNoDefault`:
   - Create a `DetailPanel`, call `SetField` with a field that has `Default: ""` and `nil` value
   - Assert the rendered view contains `"not set"`

### Task 5 (Optional): Add styled annotation for "(default)" suffix

**File:** `internal/tui/styles.go`

If visual distinction is desired, add a muted style for the annotation:
```go
defaultAnnotationStyle = lipgloss.NewStyle().
    Foreground(mutedColor).
    Italic(true)
```

Then in `formatValueCompact` and `renderCurrentValue`, compose:
```go
s = defaultVal + defaultAnnotationStyle.Render(" (default)")
```

> **Note:** This is optional for the initial implementation. Plain text `" (default)"` is sufficient and simpler. Styled annotation can be added as a follow-up.

## Edge Cases to Verify

| Edge Case | Expected Behavior | Covered By |
|-----------|-------------------|------------|
| Field with no default (`model`, `allowed_urls`) | Shows `(not set)` — unchanged | Task 3 test case |
| Field with default (`theme` → `"auto"`) | Shows `auto (default)` | Task 3 test case |
| Bool field set to `false` (in JSON) | Shows `false` — not `false (default)` | Existing test; `cfg.Get` returns `false` not `nil` |
| Sensitive field with nil value | Shows `🔒` — unaffected | CC-0005; `renderItem` short-circuits |
| Narrow column truncation | `auto (default)` truncated to `auto...` etc. | Task 3 truncation test case |
| Empty schema (copilot not installed) | Empty list, no items to display | No change needed — existing behavior |

## Scope Boundaries

**In scope:**
- `internal/tui/list_item.go` — `formatValueCompact` signature + nil branch + call site
- `internal/tui/detail_panel.go` — `renderCurrentValue` nil-with-default branch
- `internal/tui/tui_test.go` — updated and new test cases

**Out of scope (no changes needed):**
- `internal/config/` — `Config` struct and `Get()` are sufficient
- `internal/copilot/` — `SchemaField.Default` is already populated by `ParseSchema`
- `cmd/ccc/main.go` — orchestration unchanged
- `internal/tui/model.go` — `buildEntries` already iterates all schema fields
- `internal/tui/styles.go` — optional, can defer styled annotation

## Ordering

Tasks 1 and 2 can be implemented in parallel (independent files). Task 3 depends on Task 1 (signature change). Task 4 depends on Task 2. Task 5 is optional and independent.

Recommended order: **Task 1 → Task 3 → Task 2 → Task 4 → Task 5 (optional)**
