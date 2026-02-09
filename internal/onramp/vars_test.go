package onramp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListBlueprintsEmpty(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "vars"), 0750)

	names, err := ListBlueprints(dir)
	if err != nil {
		t.Fatalf("ListBlueprints() error = %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected 0 blueprints, got %d", len(names))
	}
}

func TestListBlueprintsFindsFiles(t *testing.T) {
	dir := t.TempDir()
	varsDir := filepath.Join(dir, "vars")
	os.MkdirAll(varsDir, 0750)

	// Create some blueprint files
	for _, name := range []string{"main-quickstart.yml", "main-srsran.yml", "main-gnbsim.yml"} {
		os.WriteFile(filepath.Join(varsDir, name), []byte("test"), 0640)
	}
	// Non-matching files should be ignored
	os.WriteFile(filepath.Join(varsDir, "main.yml"), []byte("test"), 0640)

	names, err := ListBlueprints(dir)
	if err != nil {
		t.Fatalf("ListBlueprints() error = %v", err)
	}
	if len(names) != 3 {
		t.Fatalf("expected 3 blueprints, got %d: %v", len(names), names)
	}

	found := make(map[string]bool)
	for _, n := range names {
		found[n] = true
	}
	for _, want := range []string{"quickstart", "srsran", "gnbsim"} {
		if !found[want] {
			t.Errorf("missing blueprint %q", want)
		}
	}
}

func TestActivateBlueprint(t *testing.T) {
	dir := t.TempDir()
	varsDir := filepath.Join(dir, "vars")
	os.MkdirAll(varsDir, 0750)

	os.WriteFile(filepath.Join(varsDir, "main-srsran.yml"), []byte("srsran-config"), 0640)

	if err := ActivateBlueprint(dir, "srsran"); err != nil {
		t.Fatalf("ActivateBlueprint() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(varsDir, "main.yml"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "srsran-config" {
		t.Errorf("main.yml = %q, want %q", string(content), "srsran-config")
	}
}

func TestActivateBlueprintNotFound(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "vars"), 0750)

	err := ActivateBlueprint(dir, "nonexistent")
	if err == nil {
		t.Error("ActivateBlueprint() should return error for missing blueprint")
	}
}

func TestEnsureVarsFileCreatesDefault(t *testing.T) {
	dir := t.TempDir()
	varsDir := filepath.Join(dir, "vars")
	os.MkdirAll(varsDir, 0750)

	os.WriteFile(filepath.Join(varsDir, "main-quickstart.yml"), []byte("default-config"), 0640)

	if err := EnsureVarsFile(dir); err != nil {
		t.Fatalf("EnsureVarsFile() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(varsDir, "main.yml"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "default-config" {
		t.Errorf("main.yml = %q, want %q", string(content), "default-config")
	}
}

func TestEnsureVarsFilePreservesExisting(t *testing.T) {
	dir := t.TempDir()
	varsDir := filepath.Join(dir, "vars")
	os.MkdirAll(varsDir, 0750)

	os.WriteFile(filepath.Join(varsDir, "main.yml"), []byte("existing-config"), 0640)
	os.WriteFile(filepath.Join(varsDir, "main-quickstart.yml"), []byte("default-config"), 0640)

	if err := EnsureVarsFile(dir); err != nil {
		t.Fatalf("EnsureVarsFile() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(varsDir, "main.yml"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "existing-config" {
		t.Errorf("main.yml = %q, want %q (should not be overwritten)", string(content), "existing-config")
	}
}
