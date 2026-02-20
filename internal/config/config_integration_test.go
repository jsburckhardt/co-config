package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jsburckhardt/co-config/internal/config"
)

// IT-001: Config round-trip with real file
func TestConfigRoundTripIntegration(t *testing.T) {
	// Use the real config file
	realPath := config.DefaultPath()
	if _, err := os.Stat(realPath); os.IsNotExist(err) {
		t.Skip("no real config file found, skipping integration test")
	}

	// Load real config
	original, err := config.LoadConfig(realPath)
	if err != nil {
		t.Fatalf("loading real config: %v", err)
	}
	originalKeys := original.Keys()

	// Modify a field
	original.Set("model", "test-integration-model")

	// Save to temp
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "config.json")
	if err := config.SaveConfig(tmpPath, original); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	// Reload
	reloaded, err := config.LoadConfig(tmpPath)
	if err != nil {
		t.Fatalf("reloading config: %v", err)
	}

	// Verify model changed
	if got := reloaded.Get("model"); got != "test-integration-model" {
		t.Errorf("model = %v, want test-integration-model", got)
	}

	// Verify all original keys preserved
	reloadedKeys := make(map[string]bool)
	for _, k := range reloaded.Keys() {
		reloadedKeys[k] = true
	}
	for _, k := range originalKeys {
		if !reloadedKeys[k] {
			t.Errorf("key %q lost during round-trip", k)
		}
	}

	// Verify JSON formatting (2-space indent)
	data, _ := os.ReadFile(tmpPath)
	content := string(data)
	if len(content) > 0 && content[0] != '{' {
		t.Error("saved JSON doesn't start with {")
	}
	// Check for 2-space indent pattern
	if !strings.Contains(content, "\n  \"") {
		t.Error("saved JSON doesn't use 2-space indentation")
	}
}
