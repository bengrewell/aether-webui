package _logging

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestSetupDefaultLevel(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{
		Writer:  &buf,
		NoColor: true,
	})

	slog.Info("visible")
	slog.Debug("hidden")

	out := buf.String()
	if !containsSubstring(out, "visible") {
		t.Error("expected INFO message to be logged")
	}
	if containsSubstring(out, "hidden") {
		t.Error("expected DEBUG message to be suppressed at INFO level")
	}
}

func TestSetupDebugLevel(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{
		Level:   slog.LevelDebug,
		Writer:  &buf,
		NoColor: true,
	})

	slog.Debug("debug-visible")

	if !containsSubstring(buf.String(), "debug-visible") {
		t.Error("expected DEBUG message to be logged at DEBUG level")
	}
}

func TestSetupAddSource(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{
		AddSource: true,
		Writer:    &buf,
		NoColor:   true,
	})

	slog.Info("with-source")

	if !containsSubstring(buf.String(), "logging_test.go") {
		t.Error("expected source file to appear in output")
	}
}

func TestWith(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{
		Writer:  &buf,
		NoColor: true,
	})

	logger := With("component", "test")
	logger.Info("child-log")

	out := buf.String()
	if !containsSubstring(out, "component") || !containsSubstring(out, "test") {
		t.Error("expected With attrs in output")
	}
	if !containsSubstring(out, "child-log") {
		t.Error("expected message in output")
	}
}

func TestSetupDefaultTimeFormat(t *testing.T) {
	var buf bytes.Buffer
	// Call Setup with empty Options - TimeFormat should default to time.DateTime
	Setup(Options{
		Writer:  &buf,
		NoColor: true,
		// TimeFormat intentionally omitted to test default
	})

	slog.Info("test-default-time")

	out := buf.String()
	if !containsSubstring(out, "test-default-time") {
		t.Error("expected message to be logged with default time format")
	}
}

func TestSetupDefaultWriter(t *testing.T) {
	// Call Setup with nil Writer - should default to os.Stderr without panic
	Setup(Options{
		NoColor: true,
		// Writer intentionally omitted to test default
	})

	// Just verify it doesn't panic and logger is usable
	slog.Info("test-default-writer")
}

func containsSubstring(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
