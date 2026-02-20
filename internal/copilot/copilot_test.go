package copilot

import (
	"os"
	"testing"
)

// UT-COP-001: ParseVersion with captured output → "0.0.412"
func TestParseVersion(t *testing.T) {
	data, err := os.ReadFile("testdata/copilot-version.txt")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}
	
	version, err := ParseVersion(string(data))
	if err != nil {
		t.Fatalf("ParseVersion failed: %v", err)
	}
	
	expected := "0.0.412"
	if version != expected {
		t.Errorf("Expected version %q, got %q", expected, version)
	}
}

// UT-COP-002: ParseSchema field count >= 15 and verify key fields exist
func TestParseSchemaFieldCount(t *testing.T) {
	data, err := os.ReadFile("testdata/copilot-help-config.txt")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}
	
	fields, err := ParseSchema(string(data))
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	minExpected := 15
	if len(fields) < minExpected {
		t.Errorf("Expected at least %d fields, got %d", minExpected, len(fields))
		for i, f := range fields {
			t.Logf("Field %d: %s (type: %s)", i+1, f.Name, f.Type)
		}
	}
	
	// Verify key representative fields exist with expected types
	expectedFields := map[string]string{
		"model":        "enum",
		"theme":        "enum",
		"allowed_urls": "list",
	}
	
	fieldMap := make(map[string]*SchemaField)
	for i := range fields {
		fieldMap[fields[i].Name] = &fields[i]
	}
	
	for name, expectedType := range expectedFields {
		field, exists := fieldMap[name]
		if !exists {
			t.Errorf("Expected field %q not found", name)
			continue
		}
		if field.Type != expectedType {
			t.Errorf("Field %q: expected type %q, got %q", name, expectedType, field.Type)
		}
	}
}

// UT-COP-003: Bool field (alt_screen) has Type "bool", Default "false"
func TestParseSchemaBoolField(t *testing.T) {
	data, err := os.ReadFile("testdata/copilot-help-config.txt")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}
	
	fields, err := ParseSchema(string(data))
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	var altScreen *SchemaField
	for i := range fields {
		if fields[i].Name == "alt_screen" {
			altScreen = &fields[i]
			break
		}
	}
	
	if altScreen == nil {
		t.Fatal("alt_screen field not found")
	}
	
	if altScreen.Type != "bool" {
		t.Errorf("Expected Type %q, got %q", "bool", altScreen.Type)
	}
	
	if altScreen.Default != "false" {
		t.Errorf("Expected Default %q, got %q", "false", altScreen.Default)
	}
}

// UT-COP-004: Enum field (banner) has Type "enum", Options [always, never, once]
func TestParseSchemaBannerField(t *testing.T) {
	data, err := os.ReadFile("testdata/copilot-help-config.txt")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}
	
	fields, err := ParseSchema(string(data))
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	var banner *SchemaField
	for i := range fields {
		if fields[i].Name == "banner" {
			banner = &fields[i]
			break
		}
	}
	
	if banner == nil {
		t.Fatal("banner field not found")
	}
	
	if banner.Type != "enum" {
		t.Errorf("Expected Type %q, got %q", "enum", banner.Type)
	}
	
	expectedOptions := []string{"always", "never", "once"}
	if len(banner.Options) != len(expectedOptions) {
		t.Errorf("Expected %d options, got %d", len(expectedOptions), len(banner.Options))
	}
	
	for i, expected := range expectedOptions {
		if i >= len(banner.Options) {
			break
		}
		if banner.Options[i] != expected {
			t.Errorf("Option %d: expected %q, got %q", i, expected, banner.Options[i])
		}
	}
}

// UT-COP-005: Enum field (model) has Type "enum", 17 options
func TestParseSchemaModelField(t *testing.T) {
	data, err := os.ReadFile("testdata/copilot-help-config.txt")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}
	
	fields, err := ParseSchema(string(data))
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	var model *SchemaField
	for i := range fields {
		if fields[i].Name == "model" {
			model = &fields[i]
			break
		}
	}
	
	if model == nil {
		t.Fatal("model field not found")
	}
	
	if model.Type != "enum" {
		t.Errorf("Expected Type %q, got %q", "enum", model.Type)
	}
	
	expected := 17
	if len(model.Options) != expected {
		t.Errorf("Expected %d options, got %d", expected, len(model.Options))
		for i, opt := range model.Options {
			t.Logf("Option %d: %s", i+1, opt)
		}
	}
}

// UT-COP-006: Enum field (theme) has Type "enum", Options [auto, dark, light]
func TestParseSchemaThemeField(t *testing.T) {
	data, err := os.ReadFile("testdata/copilot-help-config.txt")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}
	
	fields, err := ParseSchema(string(data))
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	var theme *SchemaField
	for i := range fields {
		if fields[i].Name == "theme" {
			theme = &fields[i]
			break
		}
	}
	
	if theme == nil {
		t.Fatal("theme field not found")
	}
	
	if theme.Type != "enum" {
		t.Errorf("Expected Type %q, got %q", "enum", theme.Type)
	}
	
	expectedOptions := []string{"auto", "dark", "light"}
	if len(theme.Options) != len(expectedOptions) {
		t.Errorf("Expected %d options, got %d", len(expectedOptions), len(theme.Options))
	}
	
	for i, expected := range expectedOptions {
		if i >= len(theme.Options) {
			break
		}
		if theme.Options[i] != expected {
			t.Errorf("Option %d: expected %q, got %q", i, expected, theme.Options[i])
		}
	}
}

// UT-COP-007: List field (allowed_urls) has Type "list"
func TestParseSchemaListField(t *testing.T) {
	data, err := os.ReadFile("testdata/copilot-help-config.txt")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}
	
	fields, err := ParseSchema(string(data))
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	var allowedURLs *SchemaField
	for i := range fields {
		if fields[i].Name == "allowed_urls" {
			allowedURLs = &fields[i]
			break
		}
	}
	
	if allowedURLs == nil {
		t.Fatal("allowed_urls field not found")
	}
	
	if allowedURLs.Type != "list" {
		t.Errorf("Expected Type %q, got %q", "list", allowedURLs.Type)
	}
}

// UT-COP-008: String field (log_level) has correct default "default"
func TestParseSchemaStringField(t *testing.T) {
	data, err := os.ReadFile("testdata/copilot-help-config.txt")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}
	
	fields, err := ParseSchema(string(data))
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	var logLevel *SchemaField
	for i := range fields {
		if fields[i].Name == "log_level" {
			logLevel = &fields[i]
			break
		}
	}
	
	if logLevel == nil {
		t.Fatal("log_level field not found")
	}
	
	if logLevel.Default != "default" {
		t.Errorf("Expected Default %q, got %q", "default", logLevel.Default)
	}
}

// UT-COP-009: ParseVersion with malformed output returns ErrVersionParseFailed
func TestParseVersionMalformed(t *testing.T) {
	testCases := []string{
		"",
		"Invalid output",
		"GitHub Copilot CLI",
		"Version 1.2.3",
	}
	
	for _, tc := range testCases {
		_, err := ParseVersion(tc)
		if err != ErrVersionParseFailed {
			t.Errorf("Expected ErrVersionParseFailed for input %q, got %v", tc, err)
		}
	}
}
