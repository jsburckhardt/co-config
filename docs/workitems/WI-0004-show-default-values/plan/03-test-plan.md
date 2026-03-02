# Test Plan: Show Default Values for Unset Config Keys

- **Workitem:** WI-0004-show-default-values
- **Task Breakdown:** [02-task-breakdown.md](./02-task-breakdown.md)

---

## Test T1: Nil value with non-empty default shows "(default)" in compact format

- **Type:** Unit
- **Task:** Task 1, Task 3
- **Priority:** High

### Setup

No external dependencies. Direct call to `formatValueCompact`.

### Steps

1. Call `formatValueCompact(nil, "auto", 20)`
2. Check the returned string

### Expected Result

Returns `"auto (default)"`

---

## Test T2: Nil value with boolean default shows "(default)" in compact format

- **Type:** Unit
- **Task:** Task 1, Task 3
- **Priority:** High

### Setup

No external dependencies. Direct call to `formatValueCompact`.

### Steps

1. Call `formatValueCompact(nil, "false", 20)`
2. Check the returned string

### Expected Result

Returns `"false (default)"`

---

## Test T3: Nil value with empty default shows "(not set)" in compact format

- **Type:** Unit
- **Task:** Task 1, Task 3
- **Priority:** High

### Setup

No external dependencies. Direct call to `formatValueCompact`.

### Steps

1. Call `formatValueCompact(nil, "", 10)`
2. Check the returned string

### Expected Result

Returns `"(not set)"` — preserves existing behavior for fields with no known default (e.g., `model`, `allowed_urls`).

---

## Test T4: Long default with annotation is truncated correctly

- **Type:** Unit
- **Task:** Task 1, Task 3
- **Priority:** High

### Setup

No external dependencies. Direct call to `formatValueCompact` with a `maxLen` shorter than the composed default+annotation string.

### Steps

1. Call `formatValueCompact(nil, "very-long-default", 10)`
2. Check the returned string

### Expected Result

Returns `"very-lo..."` — the composed string `"very-long-default (default)"` (27 chars) is truncated to 10 chars using existing `maxLen` truncation logic.

---

## Test T5: Non-nil value ignores default in compact format

- **Type:** Unit
- **Task:** Task 1, Task 3
- **Priority:** High

### Setup

No external dependencies. Direct call to `formatValueCompact`.

### Steps

1. Call `formatValueCompact("custom", "auto", 20)`
2. Check the returned string

### Expected Result

Returns `"custom"` — when a value is explicitly set, the default is irrelevant and should not appear.

---

## Test T6: Existing compact format cases still pass (regression)

- **Type:** Unit
- **Task:** Task 1, Task 3
- **Priority:** High

### Setup

No external dependencies. Direct calls to `formatValueCompact` with updated 3-argument signature, passing `defaultVal: ""` for all existing cases.

### Steps

1. Call `formatValueCompact("test", "", 10)` → `"test"`
2. Call `formatValueCompact(true, "", 10)` → `"true"`
3. Call `formatValueCompact(false, "", 10)` → `"false"`
4. Call `formatValueCompact([]any{}, "", 10)` → `"(empty)"`
5. Call `formatValueCompact([]any{"a", "b"}, "", 20)` → `"(2 items)"`
6. Call `formatValueCompact("very long string that exceeds max length", "", 10)` → `"very lo..."`

### Expected Result

All return values match the existing expected outputs — no behavioral regression from adding the `defaultVal` parameter.

---

## Test T7: Detail panel renders default annotation for unset field with default

- **Type:** Unit
- **Task:** Task 2, Task 4
- **Priority:** High

### Setup

Create a `DetailPanel` via `NewDetailPanel()`, set size to `50×20`.

### Steps

1. Create a `SchemaField` with `Name: "theme"`, `Type: "enum"`, `Default: "auto"`, `Options: ["auto", "dark", "light"]`
2. Call `detail.SetField(field, nil)` — value is nil (not set in config)
3. Call `detail.View()`
4. Check the rendered view for content

### Expected Result

The rendered view contains both `"auto"` (the default value) and `"default"` (the annotation). This confirms the detail panel shows `"auto (default)"` instead of `"(not set)"`.

---

## Test T8: Detail panel renders "(not set)" for unset field without default

- **Type:** Unit
- **Task:** Task 2, Task 4
- **Priority:** High

### Setup

Create a `DetailPanel` via `NewDetailPanel()`, set size to `50×20`.

### Steps

1. Create a `SchemaField` with `Name: "model"`, `Type: "enum"`, `Default: ""`, `Options: ["gpt-4", "claude-sonnet-4"]`
2. Call `detail.SetField(field, nil)` — value is nil, no default
3. Call `detail.View()`
4. Check the rendered view for content

### Expected Result

The rendered view contains `"not set"`. It should NOT contain `"(default)"` since there is no default value.

---

## Test T9: Detail panel renders set value without default annotation

- **Type:** Unit
- **Task:** Task 2, Task 4
- **Priority:** High

### Setup

Create a `DetailPanel` via `NewDetailPanel()`, set size to `50×20`.

### Steps

1. Create a `SchemaField` with `Name: "theme"`, `Type: "enum"`, `Default: "auto"`, `Options: ["auto", "dark", "light"]`
2. Call `detail.SetField(field, "dark")` — value is explicitly set to `"dark"`
3. Call `detail.View()`
4. Check the rendered view for content

### Expected Result

The rendered view contains `"dark"` but does NOT contain `"(default)"`. When a value is explicitly set, only the actual value is shown — no default annotation.

---

## Test T10: Sensitive field with nil value shows lock icon, not default

- **Type:** Unit
- **Task:** Task 1 (implicit — verifying no regression)
- **Priority:** Medium

### Setup

Create a `ListPanel` with a sensitive field entry that has a nil value but a non-empty default.

### Steps

1. Build entries with schema containing `{Name: "copilot_tokens", Type: "string", Default: ""}` and an empty config
2. Create a `ListPanel` with these entries, set width/height
3. Call `View()` on the list panel
4. Inspect the rendered line for "copilot_tokens"

### Expected Result

The sensitive field line shows `"🔒"` — NOT `"(default)"` or `"(not set)"`. The `renderItem` method short-circuits before calling `formatValueCompact` for sensitive fields (per CC-0005).

---

## Test T11: Bool field set to `false` shows "false", not "false (default)"

- **Type:** Unit
- **Task:** Task 1, Task 3
- **Priority:** Medium

### Setup

No external dependencies. Direct call to `formatValueCompact`.

### Steps

1. Call `formatValueCompact(false, "false", 20)` — value is `false` (explicitly set in JSON), default is also `"false"`
2. Check the returned string

### Expected Result

Returns `"false"` — NOT `"false (default)"`. The value `false` is non-nil (it's a `bool` value), so the nil branch is never reached. This correctly distinguishes "explicitly set to false" from "not set, default is false".

---

## Test T12: All schema keys appear in list entries (even with empty config)

- **Type:** Unit
- **Task:** Task 1 (implicit — verifying existing behavior)
- **Priority:** Medium

### Setup

Create an empty `Config` (no keys set) and a schema with multiple fields.

### Steps

1. Create `cfg := config.NewConfig()` — empty config
2. Define schema with 3 fields: `theme` (default: `"auto"`), `model` (default: `""`), `alt_screen` (default: `"false"`)
3. Call `buildEntries(cfg, schema)`
4. Count non-header entries

### Expected Result

All 3 schema fields appear as entries in the list, even though none are set in config. Values are nil for all entries. This confirms the existing behavior (per research: `buildEntries` iterates schema, not `cfg.Keys()`).

---

## Test T13: Detail panel with nil field does not panic

- **Type:** Unit
- **Task:** Task 2 (defensive — edge case)
- **Priority:** Low

### Setup

Create a `DetailPanel` via `NewDetailPanel()`. Do NOT call `SetField`.

### Steps

1. Call `detail.View()`
2. Check that no panic occurs

### Expected Result

Returns the placeholder text `"Select a field to view details"` — the guard `if d.field == nil` in `View()` prevents any access to `d.field.Default` in the updated `renderCurrentValue`.

---

## Test T14: Compact format with maxLen equal to annotation length

- **Type:** Unit
- **Task:** Task 1, Task 3
- **Priority:** Low

### Setup

No external dependencies. Edge case for truncation logic.

### Steps

1. Call `formatValueCompact(nil, "x", 3)` — `"x (default)"` is 11 chars, maxLen is 3

### Expected Result

Returns a truncated string of length 3 or less (the exact result depends on the truncation: with the `s[:maxLen-3] + "..."` formula and `maxLen=3`, this would be `"..."` since `maxLen-3 = 0`). The key requirement is no panic and the result length ≤ 3.

---

## Summary

| Test ID | Title | Type | Task | Priority |
|---------|-------|------|------|----------|
| T1 | Nil with default → "auto (default)" | Unit | 1, 3 | High |
| T2 | Nil with bool default → "false (default)" | Unit | 1, 3 | High |
| T3 | Nil without default → "(not set)" | Unit | 1, 3 | High |
| T4 | Long default truncated | Unit | 1, 3 | High |
| T5 | Non-nil ignores default | Unit | 1, 3 | High |
| T6 | Existing cases pass (regression) | Unit | 1, 3 | High |
| T7 | Detail panel: unset with default | Unit | 2, 4 | High |
| T8 | Detail panel: unset without default | Unit | 2, 4 | High |
| T9 | Detail panel: set value ignores default | Unit | 2, 4 | High |
| T10 | Sensitive field shows 🔒 not default | Unit | 1 | Medium |
| T11 | Bool `false` value ≠ "false (default)" | Unit | 1, 3 | Medium |
| T12 | All schema keys in entries (empty config) | Unit | 1 | Medium |
| T13 | Detail panel nil field no panic | Unit | 2 | Low |
| T14 | Compact format tiny maxLen edge case | Unit | 1, 3 | Low |

### Coverage Matrix

| Edge Case | Test(s) | Description |
|-----------|---------|-------------|
| Field with default, value unset | T1, T2, T7 | Shows `"<default> (default)"` in both list and detail panels |
| Field without default, value unset | T3, T8 | Shows `"(not set)"` — existing behavior preserved |
| Field with default, value explicitly set | T5, T9 | Shows actual value, no annotation |
| Bool `false` explicitly set (not nil) | T11 | Type-switch on `bool` fires before `nil` |
| Sensitive field (nil value, has default) | T10 | Short-circuits to `"🔒"`, default never shown |
| Truncation of default+annotation | T4, T14 | Long defaults truncated; tiny maxLen edge case |
| Empty config (all keys nil) | T12 | All schema keys appear in list |
| Detail panel with no field selected | T13 | Defensive nil guard prevents panic |
| Existing non-default behaviors | T6 | Regression suite for string, bool, list, map, truncation |
