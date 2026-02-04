package executor

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	e := New(Config{})
	if e == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNew_WithConfig(t *testing.T) {
	e := New(Config{
		DefaultTimeout: 5 * time.Minute,
	})
	if e.config.DefaultTimeout != 5*time.Minute {
		t.Errorf("DefaultTimeout = %v, want %v", e.config.DefaultTimeout, 5*time.Minute)
	}
}

func TestNew_DefaultTimeout(t *testing.T) {
	e := New(Config{})
	if e.config.DefaultTimeout != 10*time.Minute {
		t.Errorf("DefaultTimeout = %v, want %v", e.config.DefaultTimeout, 10*time.Minute)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.DefaultTimeout != 10*time.Minute {
		t.Errorf("DefaultTimeout = %v, want %v", cfg.DefaultTimeout, 10*time.Minute)
	}
}

func TestDefaultExecutorImplementsExecutor(t *testing.T) {
	var _ Executor = (*DefaultExecutor)(nil)
}
