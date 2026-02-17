package executor

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestMockExecutor_ImplementsExecutor(t *testing.T) {
	var _ Executor = (*MockExecutor)(nil)
}

func TestNewMockExecutor(t *testing.T) {
	m := NewMockExecutor()
	if m == nil {
		t.Fatal("NewMockExecutor() returned nil")
	}
	if m.Calls == nil {
		t.Error("Calls should not be nil")
	}
	if m.FileContents == nil {
		t.Error("FileContents should not be nil")
	}
	if m.FileExistsResults == nil {
		t.Error("FileExistsResults should not be nil")
	}
}

func TestMockExecutor_CallTracking(t *testing.T) {
	m := NewMockExecutor()
	ctx := context.Background()

	// Make some calls
	_, _ = m.RunShell(ctx, ShellOptions{Command: "echo hello"})
	_, _ = m.RunShell(ctx, ShellOptions{Command: "echo world"})
	_, _ = m.RunHelmInstall(ctx, HelmInstallOptions{ReleaseName: "test"})

	// Check call count
	if m.CallCount("RunShell") != 2 {
		t.Errorf("CallCount(RunShell) = %d, want 2", m.CallCount("RunShell"))
	}
	if m.CallCount("RunHelmInstall") != 1 {
		t.Errorf("CallCount(RunHelmInstall) = %d, want 1", m.CallCount("RunHelmInstall"))
	}
	if m.CallCount("RunKubectl") != 0 {
		t.Errorf("CallCount(RunKubectl) = %d, want 0", m.CallCount("RunKubectl"))
	}

	// Check last call
	last := m.LastCall("RunShell")
	if last == nil {
		t.Fatal("LastCall(RunShell) returned nil")
	}
	opts, ok := last.Args[0].(ShellOptions)
	if !ok {
		t.Fatal("LastCall args[0] is not ShellOptions")
	}
	if opts.Command != "echo world" {
		t.Errorf("LastCall command = %q, want %q", opts.Command, "echo world")
	}

	// Check LastCall for non-existent method
	if m.LastCall("NonExistent") != nil {
		t.Error("LastCall(NonExistent) should return nil")
	}
}

func TestMockExecutor_Reset(t *testing.T) {
	m := NewMockExecutor()
	ctx := context.Background()

	_, _ = m.RunShell(ctx, ShellOptions{})
	_, _ = m.RunShell(ctx, ShellOptions{})

	if len(m.Calls) != 2 {
		t.Fatalf("expected 2 calls before reset, got %d", len(m.Calls))
	}

	m.Reset()

	if len(m.Calls) != 0 {
		t.Errorf("expected 0 calls after reset, got %d", len(m.Calls))
	}
}

func TestMockExecutor_ConfiguredResponses(t *testing.T) {
	m := NewMockExecutor()
	ctx := context.Background()

	// Test error response
	expectedErr := errors.New("test error")
	m.ShellError = expectedErr

	_, err := m.RunShell(ctx, ShellOptions{})
	if err != expectedErr {
		t.Errorf("RunShell() error = %v, want %v", err, expectedErr)
	}

	// Test result response
	m.ShellError = nil
	m.ShellResult = &ExecResult{
		ExitCode: 1,
		Stdout:   "custom output",
	}

	result, err := m.RunShell(ctx, ShellOptions{})
	if err != nil {
		t.Fatalf("RunShell() unexpected error: %v", err)
	}
	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", result.ExitCode)
	}
	if result.Stdout != "custom output" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "custom output")
	}
}

func TestMockExecutor_DefaultResponses(t *testing.T) {
	m := NewMockExecutor()
	ctx := context.Background()

	// Without configured response, should return default success
	result, err := m.RunShell(ctx, ShellOptions{})
	if err != nil {
		t.Fatalf("RunShell() unexpected error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
}

func TestMockExecutor_FileOperations(t *testing.T) {
	m := NewMockExecutor()

	// Test ReadFile with configured content
	m.FileContents["/test/file.txt"] = []byte("test content")

	content, err := m.ReadFile("/test/file.txt")
	if err != nil {
		t.Fatalf("ReadFile() unexpected error: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("ReadFile() = %q, want %q", content, "test content")
	}

	// Test ReadFile with nonexistent file
	_, err = m.ReadFile("/nonexistent")
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("ReadFile() error = %v, want os.ErrNotExist", err)
	}

	// Test WriteFile
	err = m.WriteFile("/new/file.txt", []byte("new content"), 0644)
	if err != nil {
		t.Fatalf("WriteFile() unexpected error: %v", err)
	}
	if string(m.FileContents["/new/file.txt"]) != "new content" {
		t.Error("WriteFile() did not store content")
	}

	// Test FileExists
	m.FileExistsResults["/exists"] = true
	if !m.FileExists("/exists") {
		t.Error("FileExists() should return true for configured path")
	}
	if m.FileExists("/not-exists") {
		t.Error("FileExists() should return false for unconfigured path")
	}
}

func TestMockExecutor_AllMethods(t *testing.T) {
	m := NewMockExecutor()
	ctx := context.Background()

	// Test all methods return without error by default
	_, err := m.RunAnsiblePlaybook(ctx, AnsibleOptions{})
	if err != nil {
		t.Errorf("RunAnsiblePlaybook() error = %v", err)
	}

	_, err = m.RunHelmInstall(ctx, HelmInstallOptions{})
	if err != nil {
		t.Errorf("RunHelmInstall() error = %v", err)
	}

	_, err = m.RunHelmUpgrade(ctx, HelmUpgradeOptions{})
	if err != nil {
		t.Errorf("RunHelmUpgrade() error = %v", err)
	}

	_, err = m.RunHelmUninstall(ctx, HelmUninstallOptions{})
	if err != nil {
		t.Errorf("RunHelmUninstall() error = %v", err)
	}

	_, err = m.RunHelmList(ctx, HelmListOptions{})
	if err != nil {
		t.Errorf("RunHelmList() error = %v", err)
	}

	_, err = m.RunHelmStatus(ctx, "release", "namespace")
	if err != nil {
		t.Errorf("RunHelmStatus() error = %v", err)
	}

	_, err = m.RunKubectl(ctx, KubectlOptions{})
	if err != nil {
		t.Errorf("RunKubectl() error = %v", err)
	}

	_, err = m.KubectlApply(ctx, []byte("manifest"), "namespace")
	if err != nil {
		t.Errorf("KubectlApply() error = %v", err)
	}

	_, err = m.KubectlDelete(ctx, "pod", "name", "namespace")
	if err != nil {
		t.Errorf("KubectlDelete() error = %v", err)
	}

	_, err = m.KubectlGet(ctx, "pod", "name", "namespace", "json")
	if err != nil {
		t.Errorf("KubectlGet() error = %v", err)
	}

	_, err = m.RunDockerCommand(ctx, DockerOptions{})
	if err != nil {
		t.Errorf("RunDockerCommand() error = %v", err)
	}

	_, err = m.DockerRun(ctx, DockerRunOptions{})
	if err != nil {
		t.Errorf("DockerRun() error = %v", err)
	}

	_, err = m.DockerStop(ctx, "container", 10*time.Second)
	if err != nil {
		t.Errorf("DockerStop() error = %v", err)
	}

	_, err = m.DockerRemove(ctx, "container", true)
	if err != nil {
		t.Errorf("DockerRemove() error = %v", err)
	}

	_, err = m.RunScript(ctx, ScriptOptions{})
	if err != nil {
		t.Errorf("RunScript() error = %v", err)
	}

	err = m.MkdirAll("/test/dir", 0755)
	if err != nil {
		t.Errorf("MkdirAll() error = %v", err)
	}
}

func TestMockExecutor_ThreadSafety(t *testing.T) {
	m := NewMockExecutor()
	ctx := context.Background()

	// Run multiple goroutines making calls
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, _ = m.RunShell(ctx, ShellOptions{})
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check total calls
	if m.CallCount("RunShell") != 1000 {
		t.Errorf("CallCount(RunShell) = %d, want 1000", m.CallCount("RunShell"))
	}
}
