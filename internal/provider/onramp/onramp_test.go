package onramp

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/store"
	"github.com/bengrewell/aether-webui/internal/taskrunner"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// newTestProvider creates an OnRamp provider backed by a temp directory with
// optional vars/main.yml content pre-populated.
func newTestProvider(t *testing.T, mainYML string) *OnRamp {
	t.Helper()
	dir := t.TempDir()

	if mainYML != "" {
		varsDir := filepath.Join(dir, "vars")
		if err := os.MkdirAll(varsDir, 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(filepath.Join(varsDir, "main.yml"), []byte(mainYML), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}

	return NewProvider(Config{
		OnRampDir: dir,
		RepoURL:   "https://example.com/fake.git",
		Version:   "main",
	})
}

// writeProfile creates a vars/main-{name}.yml file in the provider's OnRampDir.
func writeProfile(t *testing.T, o *OnRamp, name, content string) {
	t.Helper()
	varsDir := filepath.Join(o.config.OnRampDir, "vars")
	if err := os.MkdirAll(varsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(varsDir, "main-"+name+".yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// initGitRepo creates a minimal git repo in dir with one commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
		{"commit", "--allow-empty", "-m", "init"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", args[0], err, out)
		}
	}
}

// submitEchoTask submits a fast-completing task via the provider's runner
// and waits for it to finish.
func submitEchoTask(t *testing.T, o *OnRamp, msg string) taskrunner.TaskView {
	t.Helper()
	view, err := o.runner.Submit(taskrunner.TaskSpec{
		Command:     "echo",
		Args:        []string{"-n", msg},
		Description: "test task",
		Labels: map[string]string{
			"component": "cluster",
			"action":    "pingall",
			"target":    "aether-pingall",
		},
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	waitForTask(t, o.runner, view.ID, 5*time.Second)
	return view
}

// waitForTask polls until the task reaches a terminal state.
func waitForTask(t *testing.T, r *taskrunner.Runner, id string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		v, err := r.Get(id)
		if err != nil {
			t.Fatalf("Get(%s): %v", id, err)
		}
		switch v.Status {
		case taskrunner.StatusSucceeded, taskrunner.StatusFailed, taskrunner.StatusCanceled:
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("task %s did not complete within %v", id, timeout)
}

// ---------------------------------------------------------------------------
// Constructor / registration tests
// ---------------------------------------------------------------------------

func TestNewProvider_ImplementsInterface(t *testing.T) {
	var _ provider.Provider = newTestProvider(t, "")
}

func TestNewProvider_EndpointCount(t *testing.T) {
	p := newTestProvider(t, "")
	descs := p.Base.Descriptors()
	if len(descs) != 14 {
		t.Errorf("registered %d endpoints, want 14", len(descs))
	}
}

func TestNewProvider_EndpointPaths(t *testing.T) {
	p := newTestProvider(t, "")

	wantOps := map[string]string{
		"onramp-get-repo-status": "/api/v1/onramp/repo",
		"onramp-refresh-repo":   "/api/v1/onramp/repo/refresh",
		"onramp-list-components": "/api/v1/onramp/components",
		"onramp-get-component":  "/api/v1/onramp/components/{component}",
		"onramp-execute-action": "/api/v1/onramp/components/{component}/{action}",
		"onramp-list-tasks":     "/api/v1/onramp/tasks",
		"onramp-get-task":       "/api/v1/onramp/tasks/{id}",
		"onramp-get-config":     "/api/v1/onramp/config",
		"onramp-patch-config":   "/api/v1/onramp/config",
		"onramp-list-profiles":  "/api/v1/onramp/config/profiles",
		"onramp-get-profile":    "/api/v1/onramp/config/profiles/{name}",
		"onramp-activate-profile": "/api/v1/onramp/config/profiles/{name}/activate",
		"onramp-get-inventory":    "/api/v1/onramp/inventory",
		"onramp-sync-inventory":   "/api/v1/onramp/inventory/sync",
	}

	descs := p.Base.Descriptors()
	for _, d := range descs {
		want, ok := wantOps[d.OperationID]
		if !ok {
			t.Errorf("unexpected operation %q", d.OperationID)
			continue
		}
		if d.HTTP.Path != want {
			t.Errorf("operation %q path = %q, want %q", d.OperationID, d.HTTP.Path, want)
		}
		delete(wantOps, d.OperationID)
	}
	for op := range wantOps {
		t.Errorf("missing operation %q", op)
	}
}

// ---------------------------------------------------------------------------
// Component handlers
// ---------------------------------------------------------------------------

func TestHandleListComponents(t *testing.T) {
	p := newTestProvider(t, "")
	out, err := p.handleListComponents(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleListComponents: %v", err)
	}
	if len(out.Body) != len(componentRegistry) {
		t.Fatalf("got %d components, want %d", len(out.Body), len(componentRegistry))
	}
	// Verify a known component is present.
	found := false
	for _, c := range out.Body {
		if c.Name == "k8s" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected k8s component in list")
	}
}

func TestHandleGetComponent(t *testing.T) {
	p := newTestProvider(t, "")
	out, err := p.handleGetComponent(t.Context(), &ComponentGetInput{Component: "cluster"})
	if err != nil {
		t.Fatalf("handleGetComponent: %v", err)
	}
	if out.Body.Name != "cluster" {
		t.Errorf("name = %q, want %q", out.Body.Name, "cluster")
	}
	if len(out.Body.Actions) == 0 {
		t.Error("expected at least one action for cluster")
	}
}

func TestHandleGetComponent_NotFound(t *testing.T) {
	p := newTestProvider(t, "")
	_, err := p.handleGetComponent(t.Context(), &ComponentGetInput{Component: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for unknown component")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, should mention 'not found'", err)
	}
}

func TestHandleGetComponent_AllRegistered(t *testing.T) {
	p := newTestProvider(t, "")
	for _, comp := range componentRegistry {
		out, err := p.handleGetComponent(t.Context(), &ComponentGetInput{Component: comp.Name})
		if err != nil {
			t.Errorf("handleGetComponent(%q): %v", comp.Name, err)
			continue
		}
		if out.Body.Name != comp.Name {
			t.Errorf("name = %q, want %q", out.Body.Name, comp.Name)
		}
	}
}

// ---------------------------------------------------------------------------
// Execute action handler
// ---------------------------------------------------------------------------

func TestHandleExecuteAction_UnknownComponent(t *testing.T) {
	p := newTestProvider(t, "")
	_, err := p.handleExecuteAction(t.Context(), &ExecuteActionInput{
		Component: "nonexistent",
		Action:    "install",
	})
	if err == nil {
		t.Fatal("expected error for unknown component")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, should mention 'not found'", err)
	}
}

func TestHandleExecuteAction_UnknownAction(t *testing.T) {
	p := newTestProvider(t, "")
	_, err := p.handleExecuteAction(t.Context(), &ExecuteActionInput{
		Component: "k8s",
		Action:    "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, should mention 'not found'", err)
	}
}

func TestHandleExecuteAction_Success(t *testing.T) {
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make not available on PATH")
	}

	p := newTestProvider(t, "")
	out, err := p.handleExecuteAction(t.Context(), &ExecuteActionInput{
		Component: "cluster",
		Action:    "pingall",
	})
	if err != nil {
		t.Fatalf("handleExecuteAction: %v", err)
	}
	if out.Body.ID == "" {
		t.Error("expected non-empty task ID")
	}
	if out.Body.Status != "running" {
		t.Errorf("status = %q, want %q", out.Body.Status, "running")
	}
	if out.Body.Component != "cluster" {
		t.Errorf("component = %q, want %q", out.Body.Component, "cluster")
	}
	if out.Body.Action != "pingall" {
		t.Errorf("action = %q, want %q", out.Body.Action, "pingall")
	}
	if out.Body.Target != "aether-pingall" {
		t.Errorf("target = %q, want %q", out.Body.Target, "aether-pingall")
	}

	// Wait for the task to finish (will fail since there's no Makefile, but it should complete).
	waitForTask(t, p.runner, out.Body.ID, 5*time.Second)
}

func TestHandleExecuteAction_ConcurrencyLimit(t *testing.T) {
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make not available on PATH")
	}

	p := newTestProvider(t, "")

	// Submit a long-running task via the runner to occupy the slot.
	view, err := p.runner.Submit(taskrunner.TaskSpec{
		Command:     "sleep",
		Args:        []string{"10"},
		Description: "blocker",
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	defer p.runner.Cancel(view.ID)

	// Second submission via the handler should hit the concurrency limit.
	_, err = p.handleExecuteAction(t.Context(), &ExecuteActionInput{
		Component: "cluster",
		Action:    "pingall",
	})
	if err == nil {
		t.Fatal("expected concurrency limit error")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("error = %q, should mention 'already running'", err)
	}
}

// ---------------------------------------------------------------------------
// Task handlers
// ---------------------------------------------------------------------------

func TestHandleListTasks_Empty(t *testing.T) {
	p := newTestProvider(t, "")
	out, err := p.handleListTasks(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleListTasks: %v", err)
	}
	if len(out.Body) != 0 {
		t.Errorf("got %d tasks, want 0", len(out.Body))
	}
}

func TestHandleListTasks_WithTasks(t *testing.T) {
	p := newTestProvider(t, "")
	v1 := submitEchoTask(t, p, "first")
	time.Sleep(10 * time.Millisecond) // ensure different CreatedAt
	v2 := submitEchoTask(t, p, "second")

	out, err := p.handleListTasks(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleListTasks: %v", err)
	}
	if len(out.Body) != 2 {
		t.Fatalf("got %d tasks, want 2", len(out.Body))
	}
	// Most recent first.
	if out.Body[0].ID != v2.ID {
		t.Errorf("first task = %s, want %s (most recent)", out.Body[0].ID, v2.ID)
	}
	if out.Body[1].ID != v1.ID {
		t.Errorf("second task = %s, want %s", out.Body[1].ID, v1.ID)
	}
	// Output should be included in list.
	if out.Body[1].Output != "first" {
		t.Errorf("task output = %q, want %q", out.Body[1].Output, "first")
	}
}

func TestHandleGetTask_NotFound(t *testing.T) {
	p := newTestProvider(t, "")
	_, err := p.handleGetTask(t.Context(), &TaskGetInput{ID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for unknown task")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, should mention 'not found'", err)
	}
}

func TestHandleGetTask_Found(t *testing.T) {
	p := newTestProvider(t, "")
	view := submitEchoTask(t, p, "hello")

	out, err := p.handleGetTask(t.Context(), &TaskGetInput{ID: view.ID})
	if err != nil {
		t.Fatalf("handleGetTask: %v", err)
	}
	if out.Body.ID != view.ID {
		t.Errorf("id = %q, want %q", out.Body.ID, view.ID)
	}
	if out.Body.Status != "succeeded" {
		t.Errorf("status = %q, want %q", out.Body.Status, "succeeded")
	}
	if out.Body.Output != "hello" {
		t.Errorf("output = %q, want %q", out.Body.Output, "hello")
	}
	if out.Body.Component != "cluster" {
		t.Errorf("component = %q, want %q", out.Body.Component, "cluster")
	}
}

func TestHandleGetTask_WithOffset(t *testing.T) {
	p := newTestProvider(t, "")
	view := submitEchoTask(t, p, "hello world")

	// Read from offset 6.
	out, err := p.handleGetTask(t.Context(), &TaskGetInput{ID: view.ID, Offset: 6})
	if err != nil {
		t.Fatalf("handleGetTask: %v", err)
	}
	if out.Body.Output != "world" {
		t.Errorf("output from offset 6 = %q, want %q", out.Body.Output, "world")
	}
	if out.Body.OutputOffset != 11 {
		t.Errorf("output_offset = %d, want 11", out.Body.OutputOffset)
	}
}

// ---------------------------------------------------------------------------
// Config handlers
// ---------------------------------------------------------------------------

const testMainYML = `k8s:
  rke2:
    version: v1.31.4+rke2r1
core:
  data_iface: ens18
  ran_subnet: 192.168.251.0/24
`

func TestHandleGetConfig(t *testing.T) {
	p := newTestProvider(t, testMainYML)
	out, err := p.handleGetConfig(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleGetConfig: %v", err)
	}
	if out.Body.K8s == nil || out.Body.K8s.RKE2 == nil {
		t.Fatal("expected k8s.rke2 to be populated")
	}
	if out.Body.K8s.RKE2.Version != "v1.31.4+rke2r1" {
		t.Errorf("rke2 version = %q, want %q", out.Body.K8s.RKE2.Version, "v1.31.4+rke2r1")
	}
	if out.Body.Core == nil {
		t.Fatal("expected core to be populated")
	}
	if out.Body.Core.DataIface != "ens18" {
		t.Errorf("data_iface = %q, want %q", out.Body.Core.DataIface, "ens18")
	}
}

func TestHandleGetConfig_MissingFile(t *testing.T) {
	p := newTestProvider(t, "") // no vars/main.yml
	_, err := p.handleGetConfig(t.Context(), nil)
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestHandlePatchConfig(t *testing.T) {
	p := newTestProvider(t, testMainYML)

	newDataIface := "ens20"
	out, err := p.handlePatchConfig(t.Context(), &ConfigPatchInput{
		Body: OnRampConfig{
			Core: &CoreConfig{DataIface: newDataIface},
		},
	})
	if err != nil {
		t.Fatalf("handlePatchConfig: %v", err)
	}
	if out.Body.Core.DataIface != newDataIface {
		t.Errorf("data_iface = %q, want %q", out.Body.Core.DataIface, newDataIface)
	}
	// K8s should be preserved from the original config.
	if out.Body.K8s == nil || out.Body.K8s.RKE2 == nil {
		t.Fatal("expected k8s.rke2 to be preserved")
	}
	if out.Body.K8s.RKE2.Version != "v1.31.4+rke2r1" {
		t.Errorf("rke2 version = %q, want %q", out.Body.K8s.RKE2.Version, "v1.31.4+rke2r1")
	}
}

func TestHandlePatchConfig_PersistsToDisk(t *testing.T) {
	p := newTestProvider(t, testMainYML)

	newDataIface := "ens20"
	_, err := p.handlePatchConfig(t.Context(), &ConfigPatchInput{
		Body: OnRampConfig{
			Core: &CoreConfig{DataIface: newDataIface},
		},
	})
	if err != nil {
		t.Fatalf("handlePatchConfig: %v", err)
	}

	// Re-read from disk to verify persistence.
	out, err := p.handleGetConfig(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleGetConfig: %v", err)
	}
	if out.Body.Core.DataIface != newDataIface {
		t.Errorf("persisted data_iface = %q, want %q", out.Body.Core.DataIface, newDataIface)
	}
}

func TestHandlePatchConfig_NilFieldsPreserved(t *testing.T) {
	p := newTestProvider(t, testMainYML)

	// Patch with only a nil-irrelevant field — all existing sections should survive.
	out, err := p.handlePatchConfig(t.Context(), &ConfigPatchInput{
		Body: OnRampConfig{}, // all nil
	})
	if err != nil {
		t.Fatalf("handlePatchConfig: %v", err)
	}
	if out.Body.K8s == nil {
		t.Error("expected k8s to be preserved when patch is empty")
	}
	if out.Body.Core == nil {
		t.Error("expected core to be preserved when patch is empty")
	}
}

// ---------------------------------------------------------------------------
// Profile handlers
// ---------------------------------------------------------------------------

const testProfileYML = `k8s:
  rke2:
    version: v1.30.0+rke2r1
core:
  data_iface: ens19
`

func TestHandleListProfiles_Empty(t *testing.T) {
	p := newTestProvider(t, testMainYML)
	out, err := p.handleListProfiles(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleListProfiles: %v", err)
	}
	if len(out.Body) != 0 {
		t.Errorf("got %d profiles, want 0", len(out.Body))
	}
}

func TestHandleListProfiles(t *testing.T) {
	p := newTestProvider(t, testMainYML)
	writeProfile(t, p, "staging", testProfileYML)
	writeProfile(t, p, "production", testProfileYML)

	out, err := p.handleListProfiles(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleListProfiles: %v", err)
	}
	if len(out.Body) != 2 {
		t.Fatalf("got %d profiles, want 2", len(out.Body))
	}
	sort.Strings(out.Body)
	if out.Body[0] != "production" || out.Body[1] != "staging" {
		t.Errorf("profiles = %v, want [production staging]", out.Body)
	}
}

func TestHandleGetProfile(t *testing.T) {
	p := newTestProvider(t, testMainYML)
	writeProfile(t, p, "staging", testProfileYML)

	out, err := p.handleGetProfile(t.Context(), &ProfileGetInput{Name: "staging"})
	if err != nil {
		t.Fatalf("handleGetProfile: %v", err)
	}
	if out.Body.Core == nil {
		t.Fatal("expected core to be populated")
	}
	if out.Body.Core.DataIface != "ens19" {
		t.Errorf("data_iface = %q, want %q", out.Body.Core.DataIface, "ens19")
	}
}

func TestHandleGetProfile_NotFound(t *testing.T) {
	p := newTestProvider(t, testMainYML)
	_, err := p.handleGetProfile(t.Context(), &ProfileGetInput{Name: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, should mention 'not found'", err)
	}
}

func TestHandleActivateProfile(t *testing.T) {
	p := newTestProvider(t, testMainYML)
	writeProfile(t, p, "staging", testProfileYML)

	out, err := p.handleActivateProfile(t.Context(), &ProfileActivateInput{Name: "staging"})
	if err != nil {
		t.Fatalf("handleActivateProfile: %v", err)
	}
	if !strings.Contains(out.Body.Message, "staging") {
		t.Errorf("message = %q, should mention 'staging'", out.Body.Message)
	}

	// Verify main.yml was replaced with the profile content.
	cfg, err := p.handleGetConfig(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleGetConfig after activate: %v", err)
	}
	if cfg.Body.Core == nil || cfg.Body.Core.DataIface != "ens19" {
		t.Errorf("expected active config to reflect staging profile (data_iface=ens19)")
	}
}

func TestHandleActivateProfile_NotFound(t *testing.T) {
	p := newTestProvider(t, testMainYML)
	_, err := p.handleActivateProfile(t.Context(), &ProfileActivateInput{Name: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, should mention 'not found'", err)
	}
}

// ---------------------------------------------------------------------------
// Repo handlers
// ---------------------------------------------------------------------------

func TestHandleGetRepoStatus_NoGitDir(t *testing.T) {
	p := newTestProvider(t, "")
	out, err := p.handleGetRepoStatus(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleGetRepoStatus: %v", err)
	}
	if out.Body.Cloned {
		t.Error("expected Cloned=false for directory without .git")
	}
	if out.Body.Dir != p.config.OnRampDir {
		t.Errorf("dir = %q, want %q", out.Body.Dir, p.config.OnRampDir)
	}
	if out.Body.RepoURL != p.config.RepoURL {
		t.Errorf("repo_url = %q, want %q", out.Body.RepoURL, p.config.RepoURL)
	}
}

func TestHandleGetRepoStatus_WithGitRepo(t *testing.T) {
	p := newTestProvider(t, "")
	initGitRepo(t, p.config.OnRampDir)

	out, err := p.handleGetRepoStatus(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleGetRepoStatus: %v", err)
	}
	if !out.Body.Cloned {
		t.Error("expected Cloned=true for initialized git repo")
	}
	if out.Body.Commit == "" {
		t.Error("expected non-empty Commit")
	}
	if out.Body.Branch == "" {
		t.Error("expected non-empty Branch")
	}
	if out.Body.Dirty {
		t.Error("expected Dirty=false for clean repo")
	}
}

func TestHandleGetRepoStatus_DirtyRepo(t *testing.T) {
	p := newTestProvider(t, "")
	initGitRepo(t, p.config.OnRampDir)

	// Create an untracked file to make the repo dirty.
	if err := os.WriteFile(filepath.Join(p.config.OnRampDir, "dirty.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	out, err := p.handleGetRepoStatus(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleGetRepoStatus: %v", err)
	}
	if !out.Body.Dirty {
		t.Error("expected Dirty=true for repo with untracked files")
	}
}

func TestHandleRefreshRepo_InvalidDir(t *testing.T) {
	// Refresh with a non-clonable URL returns a status with an error message
	// rather than a handler error (degraded mode).
	p := newTestProvider(t, "")
	out, err := p.handleRefreshRepo(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleRefreshRepo: %v", err)
	}
	// ensureRepo will fail (no valid git repo to clone), but the handler
	// should return a status with an error string, not a handler error.
	if out.Body.Error == "" {
		t.Error("expected non-empty error in repo status for failed refresh")
	}
}

// ---------------------------------------------------------------------------
// validateRepo
// ---------------------------------------------------------------------------

func TestValidateRepo_Valid(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "vars"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("all:\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "vars", "main.yml"), []byte(""), 0o644)

	if err := validateRepo(dir); err != nil {
		t.Errorf("validateRepo: %v", err)
	}
}

func TestValidateRepo_MissingMakefile(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "vars"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	os.WriteFile(filepath.Join(dir, "vars", "main.yml"), []byte(""), 0o644)

	if err := validateRepo(dir); err == nil {
		t.Error("expected error for missing Makefile")
	}
}

func TestValidateRepo_MissingVarsFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("all:\n"), 0o644)

	if err := validateRepo(dir); err == nil {
		t.Error("expected error for missing vars/main.yml")
	}
}

// ---------------------------------------------------------------------------
// mergeConfig
// ---------------------------------------------------------------------------

func TestMergeConfig_NilFieldsPreserved(t *testing.T) {
	base := OnRampConfig{
		K8s:  &K8sConfig{RKE2: &RKE2Config{Version: "v1"}},
		Core: &CoreConfig{DataIface: "ens18"},
	}
	patch := OnRampConfig{} // all nil

	mergeConfig(&base, &patch)

	if base.K8s == nil || base.K8s.RKE2.Version != "v1" {
		t.Error("K8s should be preserved when patch.K8s is nil")
	}
	if base.Core == nil || base.Core.DataIface != "ens18" {
		t.Error("Core should be preserved when patch.Core is nil")
	}
}

func TestMergeConfig_NonNilFieldsOverwrite(t *testing.T) {
	base := OnRampConfig{
		K8s:  &K8sConfig{RKE2: &RKE2Config{Version: "v1"}},
		Core: &CoreConfig{DataIface: "ens18"},
	}
	patch := OnRampConfig{
		Core: &CoreConfig{DataIface: "ens20"},
	}

	mergeConfig(&base, &patch)

	if base.Core.DataIface != "ens20" {
		t.Errorf("data_iface = %q, want %q", base.Core.DataIface, "ens20")
	}
	// K8s should be untouched.
	if base.K8s == nil || base.K8s.RKE2.Version != "v1" {
		t.Error("K8s should be preserved when patch.K8s is nil")
	}
}

func TestMergeConfig_AllSections(t *testing.T) {
	base := OnRampConfig{}
	patch := OnRampConfig{
		K8s:      &K8sConfig{},
		Core:     &CoreConfig{},
		GNBSim:   &GNBSimConfig{},
		AMP:      &AMPConfig{},
		SDRAN:    &SDRANConfig{},
		UERANSIM: &UERANSIMConfig{},
		OAI:      &OAIConfig{},
		SRSRan:   &SRSRanConfig{},
		N3IWF:    &N3IWFConfig{},
	}

	mergeConfig(&base, &patch)

	if base.K8s == nil {
		t.Error("K8s should be set")
	}
	if base.Core == nil {
		t.Error("Core should be set")
	}
	if base.GNBSim == nil {
		t.Error("GNBSim should be set")
	}
	if base.AMP == nil {
		t.Error("AMP should be set")
	}
	if base.SDRAN == nil {
		t.Error("SDRAN should be set")
	}
	if base.UERANSIM == nil {
		t.Error("UERANSIM should be set")
	}
	if base.OAI == nil {
		t.Error("OAI should be set")
	}
	if base.SRSRan == nil {
		t.Error("SRSRan should be set")
	}
	if base.N3IWF == nil {
		t.Error("N3IWF should be set")
	}
}

// ---------------------------------------------------------------------------
// toOnRampTask
// ---------------------------------------------------------------------------

func TestToOnRampTask(t *testing.T) {
	now := time.Now().UTC()
	view := taskrunner.TaskView{
		ID:     "test-id",
		Status: taskrunner.StatusSucceeded,
		Labels: map[string]string{
			"component": "k8s",
			"action":    "install",
			"target":    "aether-k8s-install",
		},
		StartedAt:  now,
		FinishedAt: now.Add(5 * time.Second),
		ExitCode:   0,
	}

	task := toOnRampTask(view, "output data", 11)

	if task.ID != "test-id" {
		t.Errorf("ID = %q, want %q", task.ID, "test-id")
	}
	if task.Component != "k8s" {
		t.Errorf("Component = %q, want %q", task.Component, "k8s")
	}
	if task.Action != "install" {
		t.Errorf("Action = %q, want %q", task.Action, "install")
	}
	if task.Target != "aether-k8s-install" {
		t.Errorf("Target = %q, want %q", task.Target, "aether-k8s-install")
	}
	if task.Status != "succeeded" {
		t.Errorf("Status = %q, want %q", task.Status, "succeeded")
	}
	if task.Output != "output data" {
		t.Errorf("Output = %q, want %q", task.Output, "output data")
	}
	if task.OutputOffset != 11 {
		t.Errorf("OutputOffset = %d, want 11", task.OutputOffset)
	}
}

func TestToOnRampTask_MissingLabels(t *testing.T) {
	view := taskrunner.TaskView{
		ID:     "test-id",
		Status: taskrunner.StatusRunning,
		Labels: nil,
	}

	task := toOnRampTask(view, "", 0)

	if task.Component != "" {
		t.Errorf("Component = %q, want empty for nil labels", task.Component)
	}
	if task.Action != "" {
		t.Errorf("Action = %q, want empty for nil labels", task.Action)
	}
}

// ---------------------------------------------------------------------------
// Component registry
// ---------------------------------------------------------------------------

func TestComponentRegistryConsistency(t *testing.T) {
	if len(componentRegistry) != len(componentIndex) {
		t.Fatalf("registry has %d components but index has %d",
			len(componentRegistry), len(componentIndex))
	}

	for _, comp := range componentRegistry {
		if comp.Name == "" {
			t.Error("component with empty name")
		}
		if len(comp.Actions) == 0 {
			t.Errorf("component %q has no actions", comp.Name)
		}
		for _, action := range comp.Actions {
			if action.Name == "" {
				t.Errorf("component %q has action with empty name", comp.Name)
			}
			if action.Target == "" {
				t.Errorf("component %q action %q has empty target", comp.Name, action.Name)
			}
		}
		// Verify index lookup.
		indexed, ok := componentIndex[comp.Name]
		if !ok {
			t.Errorf("component %q missing from index", comp.Name)
		} else if indexed.Name != comp.Name {
			t.Errorf("index[%q].Name = %q", comp.Name, indexed.Name)
		}
	}
}

// ---------------------------------------------------------------------------
// Start / Stop
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Inventory: parseHostsINI
// ---------------------------------------------------------------------------

func TestParseHostsINI_Full(t *testing.T) {
	data := []byte(`[all]
node1 ansible_host=10.0.0.1 ansible_user=ubuntu ansible_password=secret ansible_sudo_pass=sudo
node2 ansible_host=10.0.0.2 ansible_user=root

[master_nodes]
node1

[worker_nodes]
node2
`)
	inv := parseHostsINI(data)
	if len(inv.Nodes) != 2 {
		t.Fatalf("got %d nodes, want 2", len(inv.Nodes))
	}

	nodeMap := make(map[string]InventoryNode)
	for _, n := range inv.Nodes {
		nodeMap[n.Name] = n
	}

	n1 := nodeMap["node1"]
	if n1.AnsibleHost != "10.0.0.1" {
		t.Errorf("node1 host = %q, want %q", n1.AnsibleHost, "10.0.0.1")
	}
	if n1.AnsibleUser != "ubuntu" {
		t.Errorf("node1 user = %q, want %q", n1.AnsibleUser, "ubuntu")
	}
	if len(n1.Roles) != 1 || n1.Roles[0] != "master" {
		t.Errorf("node1 roles = %v, want [master]", n1.Roles)
	}

	n2 := nodeMap["node2"]
	if len(n2.Roles) != 1 || n2.Roles[0] != "worker" {
		t.Errorf("node2 roles = %v, want [worker]", n2.Roles)
	}
}

func TestParseHostsINI_Empty(t *testing.T) {
	inv := parseHostsINI([]byte(""))
	if len(inv.Nodes) != 0 {
		t.Errorf("expected empty nodes, got %d", len(inv.Nodes))
	}
}

func TestParseHostsINI_CommentsAndBlanks(t *testing.T) {
	data := []byte(`# This is a comment
[all]
# Another comment
node1 ansible_host=10.0.0.1

[master_nodes]
node1
`)
	inv := parseHostsINI(data)
	if len(inv.Nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(inv.Nodes))
	}
	if inv.Nodes[0].Name != "node1" {
		t.Errorf("name = %q, want %q", inv.Nodes[0].Name, "node1")
	}
}

// ---------------------------------------------------------------------------
// Inventory: generateHostsINI
// ---------------------------------------------------------------------------

func TestGenerateHostsINI(t *testing.T) {
	nodes := []store.Node{
		{
			Name:         "node1",
			AnsibleHost:  "10.0.0.1",
			AnsibleUser:  "ubuntu",
			Password:     []byte("pass1"),
			SudoPassword: []byte("sudo1"),
			Roles:        []string{"master"},
		},
		{
			Name:         "node2",
			AnsibleHost:  "10.0.0.2",
			AnsibleUser:  "root",
			Roles:        []string{"worker"},
		},
	}

	data := generateHostsINI(nodes)
	content := string(data)

	// Verify [all] section entries.
	if !strings.Contains(content, "node1 ansible_host=10.0.0.1 ansible_user=ubuntu ansible_password=pass1 ansible_sudo_pass=sudo1") {
		t.Errorf("missing node1 in [all] section:\n%s", content)
	}
	if !strings.Contains(content, "node2 ansible_host=10.0.0.2 ansible_user=root") {
		t.Errorf("missing node2 in [all] section:\n%s", content)
	}

	// Verify role sections.
	if !strings.Contains(content, "[master_nodes]\nnode1") {
		t.Errorf("missing node1 in [master_nodes]:\n%s", content)
	}
	if !strings.Contains(content, "[worker_nodes]\nnode2") {
		t.Errorf("missing node2 in [worker_nodes]:\n%s", content)
	}
}

func TestGenerateHostsINI_EmptySections(t *testing.T) {
	// No nodes: all role sections should still be emitted.
	data := generateHostsINI(nil)
	content := string(data)

	for _, section := range []string{"[master_nodes]", "[worker_nodes]", "[gnbsim_nodes]"} {
		if !strings.Contains(content, section) {
			t.Errorf("missing section %s:\n%s", section, content)
		}
	}
}

// ---------------------------------------------------------------------------
// Inventory: handler tests
// ---------------------------------------------------------------------------

func TestHandleGetInventory_MissingFile(t *testing.T) {
	p := newTestProvider(t, "")
	out, err := p.handleGetInventory(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleGetInventory: %v", err)
	}
	if len(out.Body.Nodes) != 0 {
		t.Errorf("expected empty inventory, got %d nodes", len(out.Body.Nodes))
	}
}

func TestHandleGetInventory_ExistingFile(t *testing.T) {
	p := newTestProvider(t, "")
	hostsINI := `[all]
node1 ansible_host=10.0.0.1

[master_nodes]
node1
`
	if err := os.WriteFile(filepath.Join(p.config.OnRampDir, "hosts.ini"), []byte(hostsINI), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	out, err := p.handleGetInventory(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleGetInventory: %v", err)
	}
	if len(out.Body.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(out.Body.Nodes))
	}
	if out.Body.Nodes[0].Name != "node1" {
		t.Errorf("name = %q, want %q", out.Body.Nodes[0].Name, "node1")
	}
}

func TestHandleSyncInventory(t *testing.T) {
	p := newTestProvider(t, "")

	// Sync requires a store — the test provider doesn't have one,
	// so calling handleSyncInventory will fail on store access.
	// This is tested via integration tests with a real store.
	// Here we just verify the handler exists and is wired.
	descs := p.Base.Descriptors()
	found := false
	for _, d := range descs {
		if d.OperationID == "onramp-sync-inventory" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected onramp-sync-inventory endpoint to be registered")
	}
}

// ---------------------------------------------------------------------------
// Start / Stop
// ---------------------------------------------------------------------------

func TestStartStop(t *testing.T) {
	p := newTestProvider(t, "")

	// Start should succeed (even without a valid repo — degraded mode).
	if err := p.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	status := p.StatusInfo()
	if !status.Running {
		t.Error("expected Running=true after Start")
	}

	if err := p.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	status = p.StatusInfo()
	if status.Running {
		t.Error("expected Running=false after Stop")
	}
}
