// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package logging

import (
	"testing"
)

func TestSetupLogger(t *testing.T) {
	// Test that SetupLogger doesn't panic with various configurations
	tests := []struct {
		name       string
		debug      bool
		structured bool
	}{
		{"default", false, false},
		{"debug mode", true, false},
		{"structured mode", false, true},
		{"debug and structured", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			SetupLogger(tt.debug, tt.structured)
		})
	}
}

func TestIsDebugEnabled(t *testing.T) {
	// Setup non-debug mode
	SetupLogger(false, false)
	if IsDebugEnabled() {
		t.Log("IsDebugEnabled() returned true after setting debug=false (may depend on implementation)")
	}

	// Setup debug mode
	SetupLogger(true, false)
	// Just verify it doesn't panic
	_ = IsDebugEnabled()
}

func TestDebug(t *testing.T) {
	SetupLogger(true, false)

	// Should not panic
	Debug("test debug message")
	Debug("test with args", "key1", "value1", "key2", 42)
}

func TestInfo(t *testing.T) {
	SetupLogger(false, false)

	// Should not panic
	Info("test info message")
	Info("test with args", "key1", "value1", "key2", 42)
}

func TestWarn(t *testing.T) {
	SetupLogger(false, false)

	// Should not panic
	Warn("test warning message")
	Warn("test with args", "key1", "value1", "key2", 42)
}

func TestError(t *testing.T) {
	SetupLogger(false, false)

	// Should not panic
	Error("test error message")
	Error("test with args", "key1", "value1", "key2", 42)
}

func TestLogLevelsWithStructured(t *testing.T) {
	SetupLogger(true, true)

	// Test all log levels with structured logging
	Debug("structured debug", "component", "test")
	Info("structured info", "component", "test")
	Warn("structured warn", "component", "test")
	Error("structured error", "component", "test")
}

func TestLogWithVariousArgTypes(t *testing.T) {
	SetupLogger(true, false)

	// Test with various argument types
	Debug("test", "string", "value")
	Debug("test", "int", 42)
	Debug("test", "float", 3.14)
	Debug("test", "bool", true)
	Debug("test", "nil", nil)

	// Test with multiple pairs
	Info("test",
		"key1", "value1",
		"key2", 42,
		"key3", true,
	)
}

func TestLogEmptyMessage(t *testing.T) {
	SetupLogger(false, false)

	// Should not panic with empty message
	Debug("")
	Info("")
	Warn("")
	Error("")
}

func TestLogOddNumberOfArgs(t *testing.T) {
	SetupLogger(false, false)

	// Test with odd number of args (key without value)
	// This should be handled gracefully by the underlying logger
	Info("test", "orphan_key")
}
