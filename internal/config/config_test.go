package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// UT-CFG-001: Load valid config → correct fields (banner, model, etc.)
func TestLoadConfig_ValidConfig(t *testing.T) {
	cfg, err := LoadConfig("testdata/valid-config.json")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	tests := []struct {
		key      string
		expected any
	}{
		{"banner", "never"},
		{"model", "claude-sonnet-4.5"},
		{"staff", true},
		{"custom_unknown_field", "preserve_me"},
		{"log_level", "all"},
	}

	for _, tt := range tests {
		got := cfg.Get(tt.key)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.expected)
		}
	}

	// Check nested copilot_tokens
	tokens := cfg.Get("copilot_tokens")
	if tokens == nil {
		t.Fatal("copilot_tokens should not be nil")
	}
	tokensMap, ok := tokens.(map[string]any)
	if !ok {
		t.Fatalf("copilot_tokens should be a map, got %T", tokens)
	}
	if tokensMap["host:user"] != "token_value" {
		t.Errorf("copilot_tokens[host:user] = %v, want token_value", tokensMap["host:user"])
	}
}

// UT-CFG-002: Load missing file → errors.Is(err, ErrConfigNotFound)
func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("testdata/nonexistent.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !errors.Is(err, ErrConfigNotFound) {
		t.Errorf("expected ErrConfigNotFound, got %v", err)
	}
}

// UT-CFG-003: Load invalid JSON → errors.Is(err, ErrConfigInvalid)
func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("{invalid json}"), 0600); err != nil {
		t.Fatalf("failed to write invalid JSON: %v", err)
	}

	_, err := LoadConfig(invalidPath)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Errorf("expected ErrConfigInvalid, got %v", err)
	}
}

// UT-CFG-004: Round-trip: load → save to temp → load → compare all keys/values
func TestRoundTrip_AllKeysAndValues(t *testing.T) {
	// Load original
	cfg1, err := LoadConfig("testdata/valid-config.json")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Save to temp
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "config.json")
	if err := SaveConfig(tmpPath, cfg1); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load again
	cfg2, err := LoadConfig(tmpPath)
	if err != nil {
		t.Fatalf("LoadConfig (round-trip) failed: %v", err)
	}

	// Compare all keys and values
	keys1 := cfg1.Keys()
	keys2 := cfg2.Keys()

	if len(keys1) != len(keys2) {
		t.Errorf("key count mismatch: %d vs %d", len(keys1), len(keys2))
	}

	keyMap := make(map[string]bool)
	for _, k := range keys1 {
		keyMap[k] = true
	}
	for _, k := range keys2 {
		if !keyMap[k] {
			t.Errorf("key %q present in round-trip but not in original", k)
		}
	}

	for _, k := range keys1 {
		v1 := cfg1.Get(k)
		v2 := cfg2.Get(k)
		if !reflect.DeepEqual(v1, v2) {
			t.Errorf("value mismatch for key %q: %v vs %v", k, v1, v2)
		}
	}
}

// UT-CFG-005: Round-trip preserves unknown fields (custom_unknown_field)
func TestRoundTrip_PreservesUnknownFields(t *testing.T) {
	cfg1, err := LoadConfig("testdata/valid-config.json")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "config.json")
	if err := SaveConfig(tmpPath, cfg1); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	cfg2, err := LoadConfig(tmpPath)
	if err != nil {
		t.Fatalf("LoadConfig (round-trip) failed: %v", err)
	}

	unknownValue := cfg2.Get("custom_unknown_field")
	if unknownValue != "preserve_me" {
		t.Errorf("custom_unknown_field = %v, want preserve_me", unknownValue)
	}
}

// UT-CFG-006: Save format: read back raw bytes, verify 2-space indent
func TestSaveConfig_Format(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("model", "gpt-4")
	cfg.Set("banner", "never")

	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "config.json")
	if err := SaveConfig(tmpPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	content := string(data)

	// Check for 2-space indentation
	if !strings.Contains(content, "  \"model\"") || !strings.Contains(content, "  \"banner\"") {
		t.Errorf("expected 2-space indentation, got:\n%s", content)
	}

	// Check for trailing newline
	if !strings.HasSuffix(content, "\n") {
		t.Error("expected trailing newline")
	}

	// Check file permissions
	info, err := os.Stat(tmpPath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

// UT-CFG-007: Save preserves sensitive fields unchanged (copilot_tokens value)
func TestSaveConfig_PreservesSensitiveFields(t *testing.T) {
	cfg1, err := LoadConfig("testdata/valid-config.json")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "config.json")
	if err := SaveConfig(tmpPath, cfg1); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	cfg2, err := LoadConfig(tmpPath)
	if err != nil {
		t.Fatalf("LoadConfig (round-trip) failed: %v", err)
	}

	tokens1 := cfg1.Get("copilot_tokens")
	tokens2 := cfg2.Get("copilot_tokens")

	if !reflect.DeepEqual(tokens1, tokens2) {
		t.Errorf("copilot_tokens not preserved: %v vs %v", tokens1, tokens2)
	}

	// Verify the actual token value
	tokensMap, ok := tokens2.(map[string]any)
	if !ok {
		t.Fatalf("copilot_tokens should be a map, got %T", tokens2)
	}
	if tokensMap["host:user"] != "token_value" {
		t.Errorf("token value changed: got %v, want token_value", tokensMap["host:user"])
	}
}

// UT-CFG-008: Get known field returns correct value
func TestConfig_GetKnownField(t *testing.T) {
	cfg, err := LoadConfig("testdata/minimal-config.json")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	model := cfg.Get("model")
	if model != "gpt-5.2" {
		t.Errorf("Get(\"model\") = %v, want gpt-5.2", model)
	}
}

// UT-CFG-009: Set then Get returns updated value
func TestConfig_SetAndGet(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("test_key", "test_value")

	got := cfg.Get("test_key")
	if got != "test_value" {
		t.Errorf("Get(\"test_key\") = %v, want test_value", got)
	}

	// Update value
	cfg.Set("test_key", "new_value")
	got = cfg.Get("test_key")
	if got != "new_value" {
		t.Errorf("Get(\"test_key\") after update = %v, want new_value", got)
	}
}

// UT-CFG-010: Get unknown field returns value from data
func TestConfig_GetUnknownField(t *testing.T) {
	cfg, err := LoadConfig("testdata/valid-config.json")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	value := cfg.Get("custom_unknown_field")
	if value != "preserve_me" {
		t.Errorf("Get(\"custom_unknown_field\") = %v, want preserve_me", value)
	}
}

// UT-CFG-011: Set unknown field stores in data
func TestConfig_SetUnknownField(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("my_custom_field", "custom_value")

	got := cfg.Get("my_custom_field")
	if got != "custom_value" {
		t.Errorf("Get(\"my_custom_field\") = %v, want custom_value", got)
	}

	// Verify it's in the data map
	data := cfg.Data()
	if data["my_custom_field"] != "custom_value" {
		t.Errorf("Data()[\"my_custom_field\"] = %v, want custom_value", data["my_custom_field"])
	}
}

// UT-CFG-012: DefaultPath with XDG_CONFIG_HOME set
func TestDefaultPath_WithXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")

	path := DefaultPath()
	expected := "/custom/config/copilot/config.json"
	if path != expected {
		t.Errorf("DefaultPath() = %q, want %q", path, expected)
	}
}

// UT-CFG-013: DefaultPath without XDG_CONFIG_HOME
func TestDefaultPath_WithoutXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	path := DefaultPath()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".copilot", "config.json")
	if path != expected {
		t.Errorf("DefaultPath() = %q, want %q", path, expected)
	}
}

// Additional test: NewConfig creates empty config
func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig() returned nil")
	}

	keys := cfg.Keys()
	if len(keys) != 0 {
		t.Errorf("NewConfig() should have no keys, got %d", len(keys))
	}
}

// Additional test: Keys returns all keys
func TestConfig_Keys(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("key1", "value1")
	cfg.Set("key2", "value2")
	cfg.Set("key3", "value3")

	keys := cfg.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys() returned %d keys, want 3", len(keys))
	}

	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	for _, expected := range []string{"key1", "key2", "key3"} {
		if !keyMap[expected] {
			t.Errorf("Keys() missing %q", expected)
		}
	}
}

// Additional test: SaveConfig creates directory if needed
func TestSaveConfig_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "nested", "config.json")

	cfg := NewConfig()
	cfg.Set("test", "value")

	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Verify directory was created
	dirPath := filepath.Join(tmpDir, "subdir", "nested")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Error("config directory was not created")
	}
}
