package tui

import (
	"testing"

	"github.com/jsburckhardt/co-config/internal/config"
	"github.com/jsburckhardt/co-config/internal/copilot"
)

// UT-TUI-001: BuildForm with a bool schema field creates a form (non-nil)
func TestBuildFormWithBoolField(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("test_bool", true)

	schema := []copilot.SchemaField{
		{
			Name:        "test_bool",
			Type:        "bool",
			Default:     "false",
			Description: "A test boolean field",
		},
	}

	form, result := BuildForm(cfg, schema)

	if form == nil {
		t.Fatal("BuildForm returned nil form")
	}
	if result == nil {
		t.Fatal("BuildForm returned nil result")
	}
	if result.Values == nil {
		t.Fatal("FormResult.Values is nil")
	}
	if _, ok := result.Values["test_bool"]; !ok {
		t.Error("Expected test_bool in result.Values")
	}
}

// UT-TUI-002: BuildForm with an enum schema field creates form with options
func TestBuildFormWithEnumField(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("test_enum", "option2")

	schema := []copilot.SchemaField{
		{
			Name:        "test_enum",
			Type:        "enum",
			Default:     "option1",
			Options:     []string{"option1", "option2", "option3"},
			Description: "A test enum field",
		},
	}

	form, result := BuildForm(cfg, schema)

	if form == nil {
		t.Fatal("BuildForm returned nil form")
	}
	if result == nil {
		t.Fatal("BuildForm returned nil result")
	}
	if _, ok := result.Values["test_enum"]; !ok {
		t.Error("Expected test_enum in result.Values")
	}
}

// UT-TUI-003: BuildForm with sensitive field excludes it from editable fields
func TestBuildFormExcludesSensitiveField(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("copilot_tokens", map[string]any{"token": "secret"})
	cfg.Set("model", "gpt-4")

	schema := []copilot.SchemaField{
		{
			Name:        "copilot_tokens",
			Type:        "string",
			Default:     "",
			Description: "Copilot authentication tokens",
		},
		{
			Name:        "model",
			Type:        "enum",
			Default:     "gpt-4",
			Options:     []string{"gpt-4", "gpt-3.5-turbo"},
			Description: "AI model to use",
		},
	}

	form, result := BuildForm(cfg, schema)

	if form == nil {
		t.Fatal("BuildForm returned nil form")
	}

	// Sensitive field should not be in editable Values (only in notes)
	if _, ok := result.Values["copilot_tokens"]; ok {
		t.Error("Expected copilot_tokens to be excluded from editable Values")
	}

	// Non-sensitive field should be present
	if _, ok := result.Values["model"]; !ok {
		t.Error("Expected model in result.Values")
	}
}

// UT-TUI-004: NewModel creates a valid model with Init()
func TestNewModel(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")

	schema := []copilot.SchemaField{
		{
			Name:        "model",
			Type:        "enum",
			Default:     "gpt-4",
			Options:     []string{"gpt-4", "gpt-3.5-turbo"},
			Description: "AI model to use",
		},
	}

	model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")

	if model.form == nil {
		t.Fatal("NewModel created model with nil form")
	}
	if model.result == nil {
		t.Fatal("NewModel created model with nil result")
	}
	if model.cfg != cfg {
		t.Error("NewModel did not store config correctly")
	}
	if model.version != "0.0.412" {
		t.Error("NewModel did not store version correctly")
	}
	if model.configPath != "/tmp/config.json" {
		t.Error("NewModel did not store configPath correctly")
	}

	// Test Init() returns a command
	cmd := model.Init()
	if cmd == nil {
		t.Error("Model.Init() returned nil command")
	}
}

// Test list field handling
func TestBuildFormWithListField(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("allowed_urls", []any{"https://example.com", "https://test.com"})

	schema := []copilot.SchemaField{
		{
			Name:        "allowed_urls",
			Type:        "list",
			Default:     "",
			Description: "List of allowed URLs",
		},
	}

	form, result := BuildForm(cfg, schema)

	if form == nil {
		t.Fatal("BuildForm returned nil form")
	}
	if _, ok := result.Values["allowed_urls"]; !ok {
		t.Error("Expected allowed_urls in result.Values")
	}

	// Check that the value is a pointer to string (multi-line text)
	if ptr, ok := result.Values["allowed_urls"].(*string); ok {
		expected := "https://example.com\nhttps://test.com"
		if *ptr != expected {
			t.Errorf("Expected list to be joined as %q, got %q", expected, *ptr)
		}
	} else {
		t.Error("Expected allowed_urls value to be *string")
	}
}

// Test string field handling
func TestBuildFormWithStringField(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("log_level", "debug")

	schema := []copilot.SchemaField{
		{
			Name:        "log_level",
			Type:        "string",
			Default:     "default",
			Description: "Log level",
		},
	}

	form, result := BuildForm(cfg, schema)

	if form == nil {
		t.Fatal("BuildForm returned nil form")
	}
	if _, ok := result.Values["log_level"]; !ok {
		t.Error("Expected log_level in result.Values")
	}

	if ptr, ok := result.Values["log_level"].(*string); ok {
		if *ptr != "debug" {
			t.Errorf("Expected log_level to be 'debug', got %q", *ptr)
		}
	} else {
		t.Error("Expected log_level value to be *string")
	}
}

// Test field categorization
func TestFieldCategorization(t *testing.T) {
	cfg := config.NewConfig()

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
		{Name: "theme", Type: "enum", Default: "auto", Options: []string{"auto", "dark", "light"}},
		{Name: "allowed_urls", Type: "list"},
		{Name: "beep", Type: "bool", Default: "true"},
	}

	form, _ := BuildForm(cfg, schema)

	if form == nil {
		t.Fatal("BuildForm returned nil form")
	}

	// This test mainly verifies that BuildForm doesn't panic
	// when categorizing different field types
}
