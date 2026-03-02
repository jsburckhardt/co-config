# Task Breakdown: Show Default Values for Unset Config Keys

- **Workitem:** WI-0004-show-default-values
- **Action Plan:** [01-action-plan.md](./01-action-plan.md)

---

## Task 1: Update `formatValueCompact` to show defaults in the list panel

- **Status:** Pending
- **Complexity:** Low
- **Dependencies:** None
- **Related ADRs:** [ADR-0003: Two-Panel TUI Layout Pattern](../../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004: Configuration Management](../../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md), [CC-0005: Sensitive Data Handling](../../../architecture/core-components/CORE-COMPONENT-0005-sensitive-data-handling.md)

### Description

Modify the `formatValueCompact` function in `internal/tui/list_item.go` to accept a new `defaultVal string` parameter and display default values for unset (nil) config keys.

Changes required:

1. **Update function signature** — add `defaultVal string` as the second parameter:
   ```go
   func formatValueCompact(val any, defaultVal string, maxLen int) string {
   ```

2. **Update the `case nil` branch** — when `defaultVal` is non-empty, display `"<default> (default)"` instead of `"(not set)"`:
   ```go
   case nil:
       if defaultVal != "" {
           s = defaultVal + " (default)"
       } else {
           s = "(not set)"
       }
   ```

3. **Update the call site in `renderItem`** (line 159) — pass `item.Field.Default`:
   ```go
   val = formatValueCompact(item.Value, item.Field.Default, valWidth)
   ```

The existing truncation logic (`if len(s) > maxLen`) already applies to the composed string, so `"auto (default)"` will be truncated correctly if the column is too narrow.

**Note on sensitive fields:** The `renderItem` method short-circuits to `"🔒"` before reaching `formatValueCompact` for sensitive fields (line 152–153), so this change does not affect sensitive data display per CC-0005.

### Acceptance Criteria

- [ ] `formatValueCompact` has signature `func formatValueCompact(val any, defaultVal string, maxLen int) string`
- [ ] When `val` is `nil` and `defaultVal` is non-empty (e.g. `"auto"`), returns `"auto (default)"`
- [ ] When `val` is `nil` and `defaultVal` is empty, returns `"(not set)"` (existing behavior preserved)
- [ ] When `val` is non-nil (string, bool, list, map, etc.), behavior is unchanged regardless of `defaultVal`
- [ ] Long default+annotation strings are truncated using existing `maxLen` logic (e.g., `"very-lo..."`)
- [ ] The call site in `renderItem` passes `item.Field.Default` as the `defaultVal` argument
- [ ] Code compiles with no errors

### Test Coverage

- Covered by Task 3 (updated `TestFormatValueCompact` with new signature and new default-value test cases)

---

## Task 2: Update `renderCurrentValue` to show defaults in the detail panel

- **Status:** Pending
- **Complexity:** Low
- **Dependencies:** None (independent of Task 1 — different file)
- **Related ADRs:** [ADR-0003: Two-Panel TUI Layout Pattern](../../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004: Configuration Management](../../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md), [CC-0005: Sensitive Data Handling](../../../architecture/core-components/CORE-COMPONENT-0005-sensitive-data-handling.md)

### Description

Modify the `renderCurrentValue` method on `DetailPanel` in `internal/tui/detail_panel.go` to display the field's default value with a `"(default)"` annotation when the current value is nil and a default is available.

Use Option A from the research brief — modify `renderCurrentValue` directly, keeping `formatValueDetail` unchanged:

```go
func (d *DetailPanel) renderCurrentValue() string {
    if d.value == nil && d.field != nil && d.field.Default != "" {
        return detailValueStyle.Render(d.field.Default + " (default)")
    }
    return detailValueStyle.Render(formatValueDetail(d.value))
}
```

This approach:
- Adds a nil-with-default check before the existing `formatValueDetail` call
- Guards against `d.field` being nil (defensive programming)
- Leaves `formatValueDetail` entirely unchanged — no downstream signature changes
- Uses the existing `detailValueStyle` for consistent rendering

**Note on sensitive fields:** Sensitive fields are handled by a separate branch in `View()` (lines 218–229 of `detail_panel.go`) that short-circuits before `renderCurrentValue` is ever called, so this change does not affect sensitive data display per CC-0005.

### Acceptance Criteria

- [ ] When `d.value` is `nil` and `d.field.Default` is non-empty (e.g. `"auto"`), `renderCurrentValue` returns styled `"auto (default)"`
- [ ] When `d.value` is `nil` and `d.field.Default` is empty, `renderCurrentValue` returns styled `"(not set)"` (existing behavior via `formatValueDetail`)
- [ ] When `d.value` is non-nil, `renderCurrentValue` behavior is unchanged
- [ ] `formatValueDetail` function signature and behavior are not modified
- [ ] Detail panel `View()` renders the default annotation visually when selecting an unset field with a known default
- [ ] Code compiles with no errors

### Test Coverage

- Covered by Task 4 (new `TestDetailPanelRenderUnsetWithDefault` and `TestDetailPanelRenderUnsetNoDefault` tests)

---

## Task 3: Update existing `TestFormatValueCompact` and add default-value test cases

- **Status:** Pending
- **Complexity:** Low
- **Dependencies:** Task 1 (requires the new `formatValueCompact` signature)
- **Related ADRs:** [ADR-0003: Two-Panel TUI Layout Pattern](../../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004: Configuration Management](../../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md)

### Description

Update the `TestFormatValueCompact` test (UT-TUI-012) in `internal/tui/tui_test.go` to match the new 3-argument function signature and add new test cases for default-value display behavior.

Changes required:

1. **Update test struct** — add a `defaultVal string` field:
   ```go
   tests := []struct {
       name       string
       value      any
       defaultVal string
       maxLen     int
       want       string
   }{
   ```

2. **Update all existing test cases** — add `defaultVal: ""` to preserve current behavior:
   ```go
   {"string", "test", "", 10, "test"},
   {"bool true", true, "", 10, "true"},
   // ... etc
   ```

3. **Add new test cases for default behavior:**
   ```go
   {"nil with default", nil, "auto", 20, "auto (default)"},
   {"nil with bool default", nil, "false", 20, "false (default)"},
   {"nil with long default truncated", nil, "very-long-default", 10, "very-lo..."},
   {"nil no default", nil, "", 10, "(not set)"},
   {"non-nil ignores default", "custom", "auto", 20, "custom"},
   ```

4. **Update call sites** — use the 3-argument form:
   ```go
   got := formatValueCompact(tt.value, tt.defaultVal, tt.maxLen)
   ```

### Acceptance Criteria

- [ ] Test struct includes a `defaultVal string` field
- [ ] All existing test cases pass with `defaultVal: ""` (no behavioral regression)
- [ ] New test case: `nil` with `defaultVal: "auto"` → `"auto (default)"`
- [ ] New test case: `nil` with `defaultVal: "false"` → `"false (default)"`
- [ ] New test case: `nil` with long default, short `maxLen` → truncated with `"..."`
- [ ] New test case: `nil` with `defaultVal: ""` → `"(not set)"`
- [ ] New test case: non-nil value with non-empty `defaultVal` → value displayed, default ignored
- [ ] All tests pass: `go test ./internal/tui/ -run TestFormatValueCompact`

### Test Coverage

- This task IS the test coverage for Task 1
- Tests exercise: nil-with-default, nil-without-default, non-nil-with-default, truncation of default+annotation

---

## Task 4: Add detail panel tests for default-value display

- **Status:** Pending
- **Complexity:** Low
- **Dependencies:** Task 2 (requires the updated `renderCurrentValue` logic)
- **Related ADRs:** [ADR-0003: Two-Panel TUI Layout Pattern](../../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** [CC-0004: Configuration Management](../../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md), [CC-0005: Sensitive Data Handling](../../../architecture/core-components/CORE-COMPONENT-0005-sensitive-data-handling.md)

### Description

Add new tests to `internal/tui/tui_test.go` that verify the detail panel correctly renders default-value annotations for unset fields and preserves `"(not set)"` for fields without a default.

New tests to add:

1. **`TestDetailPanelRenderUnsetWithDefault`** — Create a `DetailPanel`, set size, call `SetField` with a field that has `Default: "auto"` and `nil` value. Assert the rendered view contains both the default value `"auto"` and the annotation `"default"`.

2. **`TestDetailPanelRenderUnsetNoDefault`** — Create a `DetailPanel`, call `SetField` with `Default: ""` and `nil` value. Assert the rendered view contains `"not set"`.

3. **`TestDetailPanelRenderSetValueIgnoresDefault`** — Create a `DetailPanel`, call `SetField` with `Default: "auto"` and value `"dark"`. Assert the rendered view contains `"dark"` and does NOT contain `"default"`.

### Acceptance Criteria

- [ ] `TestDetailPanelRenderUnsetWithDefault`: view contains `"auto"` AND `"default"` when value is `nil` and `Default` is `"auto"`
- [ ] `TestDetailPanelRenderUnsetNoDefault`: view contains `"not set"` when value is `nil` and `Default` is `""`
- [ ] `TestDetailPanelRenderSetValueIgnoresDefault`: view contains the set value and does NOT contain `"(default)"` annotation
- [ ] All tests pass: `go test ./internal/tui/ -run TestDetailPanel`

### Test Coverage

- This task IS the test coverage for Task 2
- Tests exercise: nil-with-default detail rendering, nil-without-default detail rendering, set-value-with-default (no annotation)

---

## Task 5 (Optional): Add styled annotation for "(default)" suffix

- **Status:** Deferred
- **Complexity:** Low
- **Dependencies:** Task 1, Task 2 (must be implemented first)
- **Related ADRs:** [ADR-0003: Two-Panel TUI Layout Pattern](../../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- **Related Core-Components:** None

### Description

Add a muted, italic style for the `" (default)"` annotation to visually distinguish it from actual values. This is optional for the initial implementation — plain text `" (default)"` is sufficient and simpler.

If implemented:

1. **Add style to `internal/tui/styles.go`:**
   ```go
   defaultAnnotationStyle = lipgloss.NewStyle().
       Foreground(mutedColor).
       Italic(true)
   ```

2. **Apply in `formatValueCompact`** (list panel):
   ```go
   s = defaultVal + defaultAnnotationStyle.Render(" (default)")
   ```

3. **Apply in `renderCurrentValue`** (detail panel):
   ```go
   return detailValueStyle.Render(d.field.Default) + defaultAnnotationStyle.Render(" (default)")
   ```

**Caveat:** Lipgloss ANSI escape sequences increase string length, which affects the `maxLen` truncation logic in `formatValueCompact`. If this task is implemented, truncation must use `lipgloss.Width()` (visual width) instead of `len()` (byte length), or the styled annotation must be appended after truncation.

### Acceptance Criteria

- [ ] `defaultAnnotationStyle` is defined in `styles.go` with `mutedColor` foreground and italic
- [ ] The `" (default)"` suffix renders in the muted/italic style in both list and detail panels
- [ ] Truncation still works correctly (visual width, not byte length)
- [ ] Existing visual tests still pass

### Test Coverage

- Visual inspection / manual testing (lipgloss styles are difficult to unit-test for appearance)
- Ensure `TestFormatValueCompact` truncation cases still pass if truncation logic changes

---

## Summary

| Task | Title | Complexity | Dependencies | Status |
|------|-------|-----------|--------------|--------|
| 1 | Update `formatValueCompact` for defaults (list panel) | Low | None | Pending |
| 2 | Update `renderCurrentValue` for defaults (detail panel) | Low | None | Pending |
| 3 | Update `TestFormatValueCompact` + new default test cases | Low | Task 1 | Pending |
| 4 | Add detail panel tests for default display | Low | Task 2 | Pending |
| 5 | (Optional) Styled "(default)" annotation | Low | Task 1, 2 | Deferred |

**Recommended execution order:** Task 1 → Task 3 → Task 2 → Task 4 → Task 5 (optional)

Tasks 1 and 2 are independent (different files) and can be implemented in parallel, but tests (3, 4) depend on their respective implementation tasks.
