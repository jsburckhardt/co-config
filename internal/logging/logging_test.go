package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// UT-LOG-001: Init creates log file at given path
func TestInit_CreatesLogFile(t *testing.T) {
	// Use t.TempDir() for test isolation
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	// Save original default logger and restore after test
	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
		Shutdown()
	})

	err := Init(slog.LevelWarn, logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", logPath)
	}
}

// UT-LOG-002: slog.Warn writes entry at warn level
func TestInit_WarnLevelWritesWarnEntry(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
		Shutdown()
	})

	err := Init(slog.LevelWarn, logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Write a warn log entry
	slog.Warn("test warning message", "key", "value")

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "test warning message") {
		t.Errorf("Log file does not contain warn message. Content: %s", logContent)
	}
	if !strings.Contains(logContent, "level=WARN") {
		t.Errorf("Log file does not contain WARN level. Content: %s", logContent)
	}
	if !strings.Contains(logContent, "key=value") {
		t.Errorf("Log file does not contain structured fields. Content: %s", logContent)
	}
}

// UT-LOG-003: slog.Info is filtered at warn level (not written)
func TestInit_InfoFilteredAtWarnLevel(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
		Shutdown()
	})

	err := Init(slog.LevelWarn, logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Write an info log entry (should be filtered out)
	slog.Info("test info message")

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if strings.Contains(logContent, "test info message") {
		t.Errorf("Info message should be filtered at warn level. Content: %s", logContent)
	}
}

// UT-LOG-004: slog.Debug writes at debug level
func TestInit_DebugLevelWritesDebugEntry(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
		Shutdown()
	})

	err := Init(slog.LevelDebug, logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Write a debug log entry
	slog.Debug("test debug message", "debug_key", "debug_value")

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "test debug message") {
		t.Errorf("Log file does not contain debug message. Content: %s", logContent)
	}
	if !strings.Contains(logContent, "level=DEBUG") {
		t.Errorf("Log file does not contain DEBUG level. Content: %s", logContent)
	}
	if !strings.Contains(logContent, "debug_key=debug_value") {
		t.Errorf("Log file does not contain structured fields. Content: %s", logContent)
	}
}

// UT-LOG-005: Init with invalid path returns error
func TestInit_InvalidPathReturnsError(t *testing.T) {
	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
		Shutdown()
	})

	// Try to create log file in a path where we cannot create parent directory
	// Use a path that would require root permissions
	invalidPath := "/root/forbidden/test.log"

	err := Init(slog.LevelWarn, invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

// Test ParseLevel for all cases
func TestParseLevel_AllCases(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelWarn},  // default
		{"", slog.LevelWarn},          // default
		{"DEBUG", slog.LevelDebug},    // case-insensitive
		{"Info", slog.LevelInfo},      // case-insensitive
		{"WARN", slog.LevelWarn},      // case-insensitive
		{"  debug  ", slog.LevelDebug}, // with whitespace
		{" INFO ", slog.LevelInfo},     // with whitespace
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Test Shutdown closes log file
func TestShutdown_ClosesLogFile(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	// Initialize logger
	err := Init(slog.LevelWarn, logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Shutdown should close the file without error
	err = Shutdown()
	if err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}

	// Calling Shutdown again should not error (file already closed)
	err = Shutdown()
	if err != nil {
		t.Errorf("Second Shutdown returned error: %v", err)
	}
}

// Test Init creates parent directories
func TestInit_CreatesParentDirectories(t *testing.T) {
	tempDir := t.TempDir()
	// Create a nested path that doesn't exist yet
	logPath := filepath.Join(tempDir, "subdir1", "subdir2", "test.log")

	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
		Shutdown()
	})

	err := Init(slog.LevelWarn, logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", logPath)
	}

	// Verify parent directories were created
	parentDir := filepath.Dir(logPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Errorf("Parent directory was not created at %s", parentDir)
	}
}

// Test file permissions are 0600
func TestInit_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
		Shutdown()
	})

	err := Init(slog.LevelWarn, logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("Failed to stat log file: %v", err)
	}

	// File should have 0600 permissions (owner read/write only)
	expectedPerms := os.FileMode(0600)
	actualPerms := info.Mode().Perm()
	if actualPerms != expectedPerms {
		t.Errorf("File permissions = %v, want %v", actualPerms, expectedPerms)
	}
}

// Test that calling Init twice closes the previous log file
func TestInit_ClosesExistingLogFile(t *testing.T) {
	tempDir := t.TempDir()
	logPath1 := filepath.Join(tempDir, "test1.log")
	logPath2 := filepath.Join(tempDir, "test2.log")

	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
		Shutdown()
	})

	// Initialize first logger
	err := Init(slog.LevelWarn, logPath1)
	if err != nil {
		t.Fatalf("First Init failed: %v", err)
	}

	slog.Warn("message to first log")

	// Initialize second logger (should close first file)
	err = Init(slog.LevelWarn, logPath2)
	if err != nil {
		t.Fatalf("Second Init failed: %v", err)
	}

	slog.Warn("message to second log")

	// Both files should exist
	if _, err := os.Stat(logPath1); os.IsNotExist(err) {
		t.Errorf("First log file was not created")
	}
	if _, err := os.Stat(logPath2); os.IsNotExist(err) {
		t.Errorf("Second log file was not created")
	}

	// Second log should contain the second message
	content2, err := os.ReadFile(logPath2)
	if err != nil {
		t.Fatalf("Failed to read second log file: %v", err)
	}
	if !strings.Contains(string(content2), "message to second log") {
		t.Errorf("Second log file does not contain expected message")
	}
}
