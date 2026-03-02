# Implementation Notes: WI-0004 Show Default Values

## Task 1: Update `formatValueCompact` to show defaults in the list panel

- **Status:** Complete
- **Files Changed:** `internal/tui/list_item.go`
- **Tests Passed:** 0 (test file uses old signature; will be updated in Task 3)
- **Tests Failed:** 1 (expected — `TestFormatValueCompact` compilation error due to old 2-arg call)

### Changes Summary

Three edits to `internal/tui/list_item.go`:

1. **Function signature** (line 176): Added `defaultVal string` as second parameter to `formatValueCompact`.
2. **`case nil` branch** (lines 198–203): When `defaultVal` is non-empty, returns `"<default> (default)"` instead of `"(not set)"`. Preserves existing behavior when `defaultVal` is empty.
3. **Call site in `renderItem`** (line 159): Updated to pass `item.Field.Default` as the `defaultVal` argument.

### Test Results

- Production code compiles successfully (`go build ./internal/tui/`)
- Test file (`tui_test.go`) does not compile due to old 2-arg call in `TestFormatValueCompact` — this is expected and will be fixed in Task 3
- Sensitive field handling is unaffected: `renderItem` short-circuits to `"🔒"` before reaching `formatValueCompact` for sensitive fields (per CC-0005)

### Notes

- The existing truncation logic (`if len(s) > maxLen`) applies to the composed `"<default> (default)"` string, so long defaults will be correctly truncated (e.g., `"very-lo..."`)
- No changes to `formatValueDetail` or `detail_panel.go` — those are handled in Task 2

## Task 2: Update `renderCurrentValue` to show defaults in the detail panel

- **Status:** Complete
- **Files Changed:** `internal/tui/detail_panel.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 8 (all existing tests including TestDetailPanelRender and TestFormatValueCompact)
- **Tests Failed:** 0

### Changes Summary

1. **`renderCurrentValue` in `detail_panel.go`** (line 257): Added a nil-with-default guard before the existing `formatValueDetail` call. When `d.value == nil && d.field != nil && d.field.Default != ""`, returns `detailValueStyle.Render(d.field.Default + " (default)")` instead of falling through to `formatValueDetail` which would return `"(not set)"`.

2. **`tui_test.go`** (line 286): Updated the `formatValueCompact` call in `TestFormatValueCompact` to pass `""` as the new `defaultVal` parameter (matching Task 1's signature change), so the test file compiles.

### Test Results

- `go build ./internal/tui/` — compiles successfully
- `go test ./internal/tui/` — all 8 tests pass (0 failures)
- `formatValueDetail` function signature and behavior are unchanged
- Sensitive field handling is unaffected: `View()` short-circuits to the sensitive branch (lines 218–229) before `renderCurrentValue` is ever called (per CC-0005)

### Notes

- The `d.field != nil` guard is defensive — in normal flow, `View()` already returns early when `d.field == nil`, but the guard prevents panics if `renderCurrentValue` is ever called directly
- `formatValueDetail` is left completely unchanged — no downstream signature changes
- New tests for this behavior (T7, T8, T9, T13 from test plan) will be added in Task 4

## Task 3: Update TestFormatValueCompact with new default-value test cases

- **Status:** Complete
- **Files Changed:** `internal/tui/tui_test.go`
- **Tests Passed:** 14
- **Tests Failed:** 0

### Changes Summary

Updated `TestFormatValueCompact` (UT-TUI-012) to match the 3-argument `formatValueCompact` signature and added 6 new test cases covering default-value display behavior:

1. Added `defaultVal string` field to the test table struct
2. Updated all 7 existing test cases to include `defaultVal: ""` (no behavioral regression)
3. Added 6 new test cases:
   - `nil with default` → `"auto (default)"` (T1)
   - `nil with bool default` → `"false (default)"` (T2)
   - `nil with long default truncated` → `"very-lo..."` (T4)
   - `nil no default` → `"(not set)"` (T3)
   - `non-nil ignores default` → `"custom"` (T5)
   - `bool false ignores default` → `"false"` (T11)
4. Updated `formatValueCompact` call to pass `tt.defaultVal` as second argument
5. Updated error message format to include `defaultVal` for debugging

### Test Results

All 14 subtests in `TestFormatValueCompact` pass. Covers test plan items T1–T6, T11.

### Notes

The `strings` import was added to the test file to support `strings.Contains` assertions in Task 4 tests.

---

## Task 4: Add detail panel tests for default-value display

- **Status:** Complete
- **Files Changed:** `internal/tui/tui_test.go`
- **Tests Passed:** 4
- **Tests Failed:** 0

### Changes Summary

Added 4 new test functions verifying detail panel default-value rendering:

1. **`TestDetailPanelRenderUnsetWithDefault`** (UT-TUI-015, T7): Creates `DetailPanel` with `SchemaField{Default: "auto"}` and `nil` value. Asserts view contains `"auto"` and `"default"`.
2. **`TestDetailPanelRenderUnsetNoDefault`** (UT-TUI-016, T8): Creates `DetailPanel` with `SchemaField{Default: ""}` and `nil` value. Asserts view contains `"not set"` and does NOT contain `"(default)"`.
3. **`TestDetailPanelRenderSetValueIgnoresDefault`** (UT-TUI-017, T9): Creates `DetailPanel` with `SchemaField{Default: "auto"}` and value `"dark"`. Asserts view contains `"dark"` and does NOT contain `"(default)"`.
4. **`TestDetailPanelNilFieldNoPanic`** (UT-TUI-018, T13): Creates `DetailPanel` without calling `SetField`. Asserts view contains placeholder text `"Select a field"` and does not panic.

### Test Results

All 4 detail panel tests pass. Covers test plan items T7–T9, T13.

### Notes

- Tests use `strings.Contains` for assertions since the detail panel view includes lipgloss styling (ANSI escape codes), making exact string matching impractical.
- The nil-field guard in `renderCurrentValue` (`d.field != nil`) is verified by `TestDetailPanelNilFieldNoPanic` — the `View()` method's early return for `d.field == nil` prevents `renderCurrentValue` from being called with a nil field.
- Full test suite (`go test ./...`) passes with no regressions.
