package copilot_test

import (
	"os/exec"
	"testing"

	"github.com/jsburckhardt/co-config/internal/copilot"
)

// IT-002: Live version detection
func TestDetectVersionIntegration(t *testing.T) {
	if _, err := exec.LookPath("copilot"); err != nil {
		t.Skip("copilot CLI not found, skipping integration test")
	}

	version, err := copilot.DetectVersion()
	if err != nil {
		t.Fatalf("DetectVersion() error: %v", err)
	}
	if version == "" {
		t.Error("DetectVersion() returned empty string")
	}
	t.Logf("Detected version: %s", version)
}

// IT-003: Live schema detection
func TestDetectSchemaIntegration(t *testing.T) {
	if _, err := exec.LookPath("copilot"); err != nil {
		t.Skip("copilot CLI not found, skipping integration test")
	}

	schema, err := copilot.DetectSchema()
	if err != nil {
		t.Fatalf("DetectSchema() error: %v", err)
	}
	if len(schema) < 15 {
		t.Errorf("DetectSchema() returned %d fields, want at least 15", len(schema))
	}
	t.Logf("Detected %d schema fields", len(schema))

	// Verify some known fields exist
	fieldNames := make(map[string]bool)
	for _, f := range schema {
		fieldNames[f.Name] = true
	}
	for _, name := range []string{"model", "theme", "banner"} {
		if !fieldNames[name] {
			t.Errorf("expected field %q not found in schema", name)
		}
	}
}
