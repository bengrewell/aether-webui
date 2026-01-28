package logging

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

func containsSubstring(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
