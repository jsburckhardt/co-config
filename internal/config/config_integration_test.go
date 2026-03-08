package config_test

import (
	"encoding/json"
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
	data, _ := os.ReadFile(tmpPath) //nolint:gosec // test file path from t.TempDir()
	content := string(data)
	if len(content) > 0 && content[0] != '{' {
		t.Error("saved JSON doesn't start with {")
	}
	// Check for 2-space indent pattern
	if !strings.Contains(content, "\n  \"") {
		t.Error("saved JSON doesn't use 2-space indentation")
	}
}

// IT-004: Multi-scope config round-trip with project settings
func TestProjectSettingsRoundTripIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := config.ProjectSettingsPath(tmpDir)

	// Create and save initial project config
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-5.2")
	cfg.Set("theme", "dark")

	if err := config.SaveConfig(settingsPath, cfg); err != nil {
		t.Fatalf("saving project settings: %v", err)
	}

	// Verify .copilot/ directory was created
	copilotDir := filepath.Join(tmpDir, ".copilot")
	info, err := os.Stat(copilotDir)
	if err != nil {
		t.Fatalf(".copilot directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal(".copilot is not a directory")
	}

	// Load and verify
	loaded, err := config.LoadConfig(settingsPath)
	if err != nil {
		t.Fatalf("loading project settings: %v", err)
	}
	if got := loaded.Get("model"); got != "gpt-5.2" {
		t.Errorf("model = %v, want gpt-5.2", got)
	}
	if got := loaded.Get("theme"); got != "dark" {
		t.Errorf("theme = %v, want dark", got)
	}

	// Modify and round-trip
	loaded.Set("model", "claude-sonnet-4.5")
	if err := config.SaveConfig(settingsPath, loaded); err != nil {
		t.Fatalf("saving modified project settings: %v", err)
	}

	reloaded, err := config.LoadConfig(settingsPath)
	if err != nil {
		t.Fatalf("reloading project settings: %v", err)
	}
	if got := reloaded.Get("model"); got != "claude-sonnet-4.5" {
		t.Errorf("model after round-trip = %v, want claude-sonnet-4.5", got)
	}
	if got := reloaded.Get("theme"); got != "dark" {
		t.Errorf("theme after round-trip = %v, want dark", got)
	}
}

// IT-005: Multi-scope config round-trip with project-local settings
func TestProjectLocalSettingsRoundTripIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	localPath := config.ProjectLocalSettingsPath(tmpDir)

	// Create and save project-local config
	cfg := config.NewConfig()
	cfg.Set("stream", false)

	if err := config.SaveConfig(localPath, cfg); err != nil {
		t.Fatalf("saving project-local settings: %v", err)
	}

	// Verify .copilot/settings.local.json was created
	localFile := filepath.Join(tmpDir, ".copilot", "settings.local.json")
	if _, err := os.Stat(localFile); err != nil {
		t.Fatalf("settings.local.json not created: %v", err)
	}

	// Load and verify
	loaded, err := config.LoadConfig(localPath)
	if err != nil {
		t.Fatalf("loading project-local settings: %v", err)
	}
	if got := loaded.Get("stream"); got != false {
		t.Errorf("stream = %v, want false", got)
	}
}

// IT-007: Save to project scope creates file and directory if missing
func TestProjectScopeCreatesDirectoryIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	// No .copilot/ directory exists
	projectPath := config.ProjectSettingsPath(tmpDir)

	// Verify .copilot/ does NOT exist yet
	copilotDir := filepath.Join(tmpDir, ".copilot")
	if _, err := os.Stat(copilotDir); err == nil {
		t.Fatal(".copilot directory should not exist before save")
	}

	// Create config and save
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-5.2")
	if err := config.SaveConfig(projectPath, cfg); err != nil {
		t.Fatalf("saving to project scope: %v", err)
	}

	// Verify .copilot/ directory was created
	info, err := os.Stat(copilotDir)
	if err != nil {
		t.Fatalf(".copilot directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal(".copilot is not a directory")
	}

	// Verify settings.json exists and is valid JSON
	settingsFile := filepath.Join(copilotDir, "settings.json")
	data, err := os.ReadFile(settingsFile) //nolint:gosec // test file path from t.TempDir()
	if err != nil {
		t.Fatalf("reading settings.json: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("settings.json is not valid JSON: %v", err)
	}

	// Reload and verify value persisted
	loaded, err := config.LoadConfig(projectPath)
	if err != nil {
		t.Fatalf("loading project settings: %v", err)
	}
	if got := loaded.Get("model"); got != "gpt-5.2" {
		t.Errorf("model = %v, want gpt-5.2", got)
	}
}
