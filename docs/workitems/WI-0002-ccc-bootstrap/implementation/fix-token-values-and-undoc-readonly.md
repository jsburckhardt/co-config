# Implementation Notes: Token Values & Undocumented Keys Fixes

## Tasks Completed
- `fix-token-values`: Check LooksLikeToken on config values
- `fix-undoc-readonly`: Make undocumented keys read-only

## Status: ✅ COMPLETED

## Changes Summary

### 1. Token Value Detection in Schema Fields (CC-0005)
**File:** `internal/tui/form.go`

Added token value checking in the BuildForm function's schema field loop (after line 44). Now checks if a field's current value looks like a token (e.g., starts with `gho_`, `ghp_`, `github_pat_`) even if the field name isn't in the sensitive list.

**Code Change:**
```go
// Also check if the current value looks like a token (CC-0005)
if strVal, ok := cfg.Get(sf.Name).(string); ok && sensitive.LooksLikeToken(strVal) {
    masked := sensitive.MaskValue(strVal)
    note := huh.NewNote().
        Title(sf.Name).
        Description(fmt.Sprintf("Value: %s (read-only — token detected)", masked))
    sensitiveNotes = append(sensitiveNotes, note)
    continue
}
```

### 2. Undocumented Keys Made Read-Only (CC-0004)
**File:** `internal/tui/form.go`

Changed undocumented config keys from editable `huh.NewInput()` fields to read-only `huh.NewNote()` displays. This ensures unknown fields are preserved unchanged when writing config, as required by CC-0004.

**Code Change:**
- Removed: Editable input fields for undocumented keys
- Added: Read-only note displays with "(read-only — undocumented)" description
- Also checks if undocumented values look like tokens and treats them as sensitive

**Before:**
```go
for _, key := range undocumentedKeys {
    val := cfg.Get(key)
    if strVal, ok := val.(string); ok {
        ptr := new(string)
        *ptr = strVal
        result.Values[key] = ptr
        input := huh.NewInput().
            Title(key).
            Value(ptr).
            Description("(undocumented)")
        generalFields = append(generalFields, input)
    }
}
```

**After:**
```go
for _, key := range undocumentedKeys {
    val := cfg.Get(key)
    // Also check if value looks like a token
    if strVal, ok := val.(string); ok && sensitive.LooksLikeToken(strVal) {
        masked := sensitive.MaskValue(val)
        note := huh.NewNote().
            Title(key).
            Description(fmt.Sprintf("Value: %s (read-only — token detected)", masked))
        sensitiveNotes = append(sensitiveNotes, note)
        continue
    }
    // Undocumented fields are read-only to preserve them unchanged (CC-0004)
    displayVal := fmt.Sprintf("%v", val)
    note := huh.NewNote().
        Title(key).
        Description(fmt.Sprintf("Value: %s (read-only — undocumented)", displayVal))
    generalFields = append(generalFields, note)
}
```

### 3. Test Updates
**File:** `internal/tui/tui_test.go`

#### Added New Test: `TestBuildFormTokenValueTreatedAsSensitive`
Tests that fields with token-like values are treated as sensitive and excluded from editable values:
```go
func TestBuildFormTokenValueTreatedAsSensitive(t *testing.T) {
    cfg := config.NewConfig()
    cfg.Set("custom_field", "ghp_abc123secrettoken")

    schema := []copilot.SchemaField{
        {Name: "custom_field", Type: "string", Default: "", Description: "A custom field"},
    }

    _, result := BuildForm(cfg, schema)

    // Token-like value should NOT be in editable Values
    if _, ok := result.Values["custom_field"]; ok {
        t.Error("Expected token-like value field to be excluded from editable Values")
    }
}
```

#### Updated Test: `TestUndocumentedKeysSorted`
Modified to verify that undocumented keys are NOT in editable `result.Values` (since they're now read-only):
```go
// Undocumented keys should NOT be in editable result.Values (they're read-only)
undocKeys := []string{"apple_key", "banana_key", "middle_key", "zebra_key"}
for _, key := range undocKeys {
    if _, ok := result.Values[key]; ok {
        t.Errorf("Undocumented key %q should not be in editable result.Values", key)
    }
}

// model should still be in result.Values
if _, ok := result.Values["model"]; !ok {
    t.Error("Expected model in result.Values")
}
```

### 4. Documentation Update
**File:** `docs/architecture/core-components/CORE-COMPONENT-0004-configuration-management.md`

Updated line 22 to reflect the new behavior:

**Before:**
```
- When writing config, only known editable fields are updated; sensitive and unknown fields are preserved unchanged
```

**After:**
```
- When writing config, only known editable fields are updated; sensitive, token-like, and unknown fields are preserved unchanged and displayed as read-only in the TUI
```

## Test Results

All tests passed successfully:

```
=== RUN   TestBuildFormWithBoolField
--- PASS: TestBuildFormWithBoolField (0.01s)
=== RUN   TestBuildFormWithEnumField
--- PASS: TestBuildFormWithEnumField (0.00s)
=== RUN   TestBuildFormExcludesSensitiveField
--- PASS: TestBuildFormExcludesSensitiveField (0.00s)
=== RUN   TestNewModel
--- PASS: TestNewModel (0.00s)
=== RUN   TestBuildFormWithListField
--- PASS: TestBuildFormWithListField (0.00s)
=== RUN   TestBuildFormWithStringField
--- PASS: TestBuildFormWithStringField (0.00s)
=== RUN   TestFieldCategorization
--- PASS: TestFieldCategorization (0.00s)
=== RUN   TestBuildFormTokenValueTreatedAsSensitive
--- PASS: TestBuildFormTokenValueTreatedAsSensitive (0.00s)
=== RUN   TestUndocumentedKeysSorted
--- PASS: TestUndocumentedKeysSorted (0.00s)
PASS
ok      github.com/jsburckhardt/co-config/internal/tui  0.028s
```

`go vet ./internal/tui/...` also passed with no issues.

## Files Changed
1. `/workspaces/co-config/internal/tui/form.go` - Added token detection and made undocumented keys read-only
2. `/workspaces/co-config/internal/tui/tui_test.go` - Added token test and updated undocumented keys test
3. `/workspaces/co-config/docs/architecture/core-components/CORE-COMPONENT-0004-configuration-management.md` - Updated documentation

## Commit
All changes were committed in commit `fd10cb05d1aa24151583cf2ac7d6b537677305e8`:
```
fix: Replace fragile /root/forbidden test with chmod-based approach

- Replaced hardcoded /root/forbidden path with a temporary read-only directory
- Test now works reliably in CI containers running as root
- Test is cross-platform compatible (will work on Windows and Unix-like systems)
- Uses t.TempDir() for proper cleanup
- Restores directory permissions to allow cleanup to succeed

Fixes: fix-forbidden-test
```

Note: The commit message mentions the logging test fix, but this commit also includes the TUI form.go and tui_test.go changes for token values and undocumented keys.

## Architectural Compliance

✅ **CC-0004 (Configuration Management)**: Undocumented fields are now properly preserved unchanged by making them read-only in the TUI.

✅ **CC-0005 (Sensitive Data Handling)**: Token-like values are now detected and treated as sensitive regardless of field name, properly masked in the UI.

## Notes

The implementation correctly handles two related security/integrity concerns:

1. **Token Detection**: Values that look like GitHub tokens (or other sensitive tokens) are automatically detected and treated as sensitive, even if the field name isn't in the predefined sensitive list. This provides defense-in-depth against accidental token exposure.

2. **Undocumented Field Preservation**: By making undocumented config keys read-only instead of editable, we ensure that:
   - Unknown fields aren't accidentally modified by users
   - Config round-tripping maintains data integrity (no data loss)
   - Future copilot features are preserved even if ccc doesn't understand them yet

Both changes align with the principle of "do no harm" - the tool shows what it doesn't understand but doesn't allow modification, ensuring safe config management.
