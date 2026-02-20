package sensitive

import (
	"testing"
)

// UT-SEN-001: IsSensitive returns true for "copilot_tokens"
func TestIsSensitive_CopilotTokens(t *testing.T) {
	if !IsSensitive("copilot_tokens") {
		t.Error("Expected 'copilot_tokens' to be sensitive")
	}
}

// UT-SEN-002: IsSensitive returns true for "logged_in_users"
func TestIsSensitive_LoggedInUsers(t *testing.T) {
	if !IsSensitive("logged_in_users") {
		t.Error("Expected 'logged_in_users' to be sensitive")
	}
}

// UT-SEN-003: IsSensitive returns true for "last_logged_in_user"
func TestIsSensitive_LastLoggedInUser(t *testing.T) {
	if !IsSensitive("last_logged_in_user") {
		t.Error("Expected 'last_logged_in_user' to be sensitive")
	}
}

// UT-SEN-004: IsSensitive returns true for "staff"
func TestIsSensitive_Staff(t *testing.T) {
	if !IsSensitive("staff") {
		t.Error("Expected 'staff' to be sensitive")
	}
}

// UT-SEN-005: IsSensitive returns false for non-sensitive field
func TestIsSensitive_NonSensitiveField(t *testing.T) {
	if IsSensitive("non_sensitive_field") {
		t.Error("Expected 'non_sensitive_field' to not be sensitive")
	}
	if IsSensitive("some_other_field") {
		t.Error("Expected 'some_other_field' to not be sensitive")
	}
}

// UT-SEN-006: MaskValue for string produces 12-char hex + "..."
func TestMaskValue_String(t *testing.T) {
	result := MaskValue("gho_abc123")

	// Should end with "..."
	if len(result) < 3 || result[len(result)-3:] != "..." {
		t.Errorf("Expected result to end with '...', got: %s", result)
	}

	// Should have exactly 12 hex chars before "..."
	hexPart := result[:len(result)-3]
	if len(hexPart) != 12 {
		t.Errorf("Expected 12 hex characters, got %d: %s", len(hexPart), hexPart)
	}

	// Verify all characters are valid hex
	for _, c := range hexPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Expected only hex characters, found invalid char '%c' in: %s", c, hexPart)
		}
	}
}

// UT-SEN-007: MaskValue is deterministic (same input = same output)
func TestMaskValue_Deterministic(t *testing.T) {
	input := "secret_value_123"
	result1 := MaskValue(input)
	result2 := MaskValue(input)

	if result1 != result2 {
		t.Errorf("Expected deterministic output, got different results: %s vs %s", result1, result2)
	}

	// Also verify different inputs produce different outputs
	differentResult := MaskValue("different_value")
	if result1 == differentResult {
		t.Error("Expected different inputs to produce different masked values")
	}
}

// UT-SEN-008: MaskValue for map produces "[redacted — N items]"
func TestMaskValue_Map(t *testing.T) {
	testMap := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	result := MaskValue(testMap)
	expected := "[redacted — 3 items]"

	if result != expected {
		t.Errorf("Expected '%s', got: %s", expected, result)
	}

	// Test with empty map
	emptyMap := map[string]any{}
	result = MaskValue(emptyMap)
	expected = "[redacted — 0 items]"

	if result != expected {
		t.Errorf("Expected '%s' for empty map, got: %s", expected, result)
	}
}

// UT-SEN-009: MaskValue for slice produces "[redacted — N items]"
func TestMaskValue_Slice(t *testing.T) {
	testSlice := []any{"item1", "item2", "item3", "item4"}

	result := MaskValue(testSlice)
	expected := "[redacted — 4 items]"

	if result != expected {
		t.Errorf("Expected '%s', got: %s", expected, result)
	}

	// Test with empty slice
	emptySlice := []any{}
	result = MaskValue(emptySlice)
	expected = "[redacted — 0 items]"

	if result != expected {
		t.Errorf("Expected '%s' for empty slice, got: %s", expected, result)
	}
}

// UT-SEN-010: LooksLikeToken returns true for "gho_" prefix
func TestLooksLikeToken_GhoPrefix(t *testing.T) {
	if !LooksLikeToken("gho_abc123xyz") {
		t.Error("Expected token starting with 'gho_' to be detected")
	}

	if !LooksLikeToken("gho_") {
		t.Error("Expected exact 'gho_' to be detected")
	}
}

// UT-SEN-011: LooksLikeToken returns true for "ghp_" prefix
func TestLooksLikeToken_GhpPrefix(t *testing.T) {
	if !LooksLikeToken("ghp_abc123xyz") {
		t.Error("Expected token starting with 'ghp_' to be detected")
	}

	if !LooksLikeToken("ghp_") {
		t.Error("Expected exact 'ghp_' to be detected")
	}
}

// UT-SEN-012: LooksLikeToken returns true for "github_pat_" prefix
func TestLooksLikeToken_GithubPatPrefix(t *testing.T) {
	if !LooksLikeToken("github_pat_abc123xyz") {
		t.Error("Expected token starting with 'github_pat_' to be detected")
	}

	if !LooksLikeToken("github_pat_") {
		t.Error("Expected exact 'github_pat_' to be detected")
	}
}

// UT-SEN-013: LooksLikeToken returns false for non-token strings
func TestLooksLikeToken_NonToken(t *testing.T) {
	testCases := []string{
		"not_a_token",
		"regular_string",
		"",
		"GHO_uppercase",
		"prefix_gho_",
		" gho_with_space",
	}

	for _, tc := range testCases {
		if LooksLikeToken(tc) {
			t.Errorf("Expected '%s' to not be detected as a token", tc)
		}
	}
}

// Additional test: MaskValue for other types
func TestMaskValue_OtherTypes(t *testing.T) {
	// Test with integer
	result := MaskValue(123)
	expected := "[redacted]"
	if result != expected {
		t.Errorf("Expected '%s' for integer, got: %s", expected, result)
	}

	// Test with boolean
	result = MaskValue(true)
	if result != expected {
		t.Errorf("Expected '%s' for boolean, got: %s", expected, result)
	}

	// Test with nil
	result = MaskValue(nil)
	if result != expected {
		t.Errorf("Expected '%s' for nil, got: %s", expected, result)
	}
}
