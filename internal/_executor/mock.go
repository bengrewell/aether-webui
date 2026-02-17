package executor

import (
	"context"
	"os"
	"sync"
	"time"
)

// MockExecutor is a test double for Executor that records calls and returns configured responses.
type MockExecutor struct {
	mu sync.Mutex

	// Call tracking
	Calls []MockCall

	// Configured responses
	AnsibleResult    *ExecResult
	AnsibleError     error
	HelmInstallResult    *ExecResult
	HelmInstallError     error
	HelmUpgradeResult    *ExecResult
	HelmUpgradeError     error
	HelmUninstallResult  *ExecResult
	HelmUninstallError   error
	HelmListResult       *HelmReleaseList
	HelmListError        error
	HelmStatusResult     *HelmReleaseStatus
	HelmStatusError      error
	KubectlResult        *ExecResult
	KubectlError         error
	KubectlApplyResult   *ExecResult
	KubectlApplyError    error
	KubectlDeleteResult  *ExecResult
	KubectlDeleteError   error
	KubectlGetResult     *ExecResult
	KubectlGetError      error
	DockerResult         *ExecResult
	DockerError          error
	DockerRunResult      *ExecResult
	DockerRunError       error
	DockerStopResult     *ExecResult
	DockerStopError      error
	DockerRemoveResult   *ExecResult
	DockerRemoveError    error
	ShellResult          *ExecResult
	ShellError           error
	ScriptResult         *ExecResult
	ScriptError          error

	// File operation mocks
	FileContents      map[string][]byte
	FileExistsResults map[string]bool
	WriteFileError    error
	MkdirAllError     error
	TemplateResult    []byte
	TemplateError     error
}

// MockCall records a call to the mock.
type MockCall struct {
	Method string
	Args   []any
}

// NewMockExecutor creates a new MockExecutor with default successful responses.
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		Calls:             make([]MockCall, 0),
		FileContents:      make(map[string][]byte),
		FileExistsResults: make(map[string]bool),
	}
}

// record adds a call to the call log.
func (m *MockExecutor) record(method string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{Method: method, Args: args})
}

// CallCount returns the number of times the given method was called.
func (m *MockExecutor) CallCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, call := range m.Calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// LastCall returns the most recent call to the given method, or nil if none.
func (m *MockExecutor) LastCall(method string) *MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.Calls) - 1; i >= 0; i-- {
		if m.Calls[i].Method == method {
			return &m.Calls[i]
		}
	}
	return nil
}

// Reset clears all recorded calls.
func (m *MockExecutor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]MockCall, 0)
}

// defaultResult returns a default successful ExecResult.
func defaultResult() *ExecResult {
	return &ExecResult{
		ExitCode: 0,
		Stdout:   "",
		Stderr:   "",
		Duration: 100 * time.Millisecond,
	}
}

// RunAnsiblePlaybook records the call and returns the configured response.
func (m *MockExecutor) RunAnsiblePlaybook(_ context.Context, opts AnsibleOptions) (*ExecResult, error) {
	m.record("RunAnsiblePlaybook", opts)
	if m.AnsibleResult != nil || m.AnsibleError != nil {
		return m.AnsibleResult, m.AnsibleError
	}
	return defaultResult(), nil
}

// RunHelmInstall records the call and returns the configured response.
func (m *MockExecutor) RunHelmInstall(_ context.Context, opts HelmInstallOptions) (*ExecResult, error) {
	m.record("RunHelmInstall", opts)
	if m.HelmInstallResult != nil || m.HelmInstallError != nil {
		return m.HelmInstallResult, m.HelmInstallError
	}
	return defaultResult(), nil
}

// RunHelmUpgrade records the call and returns the configured response.
func (m *MockExecutor) RunHelmUpgrade(_ context.Context, opts HelmUpgradeOptions) (*ExecResult, error) {
	m.record("RunHelmUpgrade", opts)
	if m.HelmUpgradeResult != nil || m.HelmUpgradeError != nil {
		return m.HelmUpgradeResult, m.HelmUpgradeError
	}
	return defaultResult(), nil
}

// RunHelmUninstall records the call and returns the configured response.
func (m *MockExecutor) RunHelmUninstall(_ context.Context, opts HelmUninstallOptions) (*ExecResult, error) {
	m.record("RunHelmUninstall", opts)
	if m.HelmUninstallResult != nil || m.HelmUninstallError != nil {
		return m.HelmUninstallResult, m.HelmUninstallError
	}
	return defaultResult(), nil
}

// RunHelmList records the call and returns the configured response.
func (m *MockExecutor) RunHelmList(_ context.Context, opts HelmListOptions) (*HelmReleaseList, error) {
	m.record("RunHelmList", opts)
	if m.HelmListResult != nil || m.HelmListError != nil {
		return m.HelmListResult, m.HelmListError
	}
	return &HelmReleaseList{Releases: []HelmRelease{}}, nil
}

// RunHelmStatus records the call and returns the configured response.
func (m *MockExecutor) RunHelmStatus(_ context.Context, release, namespace string) (*HelmReleaseStatus, error) {
	m.record("RunHelmStatus", release, namespace)
	if m.HelmStatusResult != nil || m.HelmStatusError != nil {
		return m.HelmStatusResult, m.HelmStatusError
	}
	return &HelmReleaseStatus{
		Name:      release,
		Namespace: namespace,
		Status:    "deployed",
	}, nil
}

// RunKubectl records the call and returns the configured response.
func (m *MockExecutor) RunKubectl(_ context.Context, opts KubectlOptions) (*ExecResult, error) {
	m.record("RunKubectl", opts)
	if m.KubectlResult != nil || m.KubectlError != nil {
		return m.KubectlResult, m.KubectlError
	}
	return defaultResult(), nil
}

// KubectlApply records the call and returns the configured response.
func (m *MockExecutor) KubectlApply(_ context.Context, manifest []byte, namespace string) (*ExecResult, error) {
	m.record("KubectlApply", manifest, namespace)
	if m.KubectlApplyResult != nil || m.KubectlApplyError != nil {
		return m.KubectlApplyResult, m.KubectlApplyError
	}
	return defaultResult(), nil
}

// KubectlDelete records the call and returns the configured response.
func (m *MockExecutor) KubectlDelete(_ context.Context, resource, name, namespace string) (*ExecResult, error) {
	m.record("KubectlDelete", resource, name, namespace)
	if m.KubectlDeleteResult != nil || m.KubectlDeleteError != nil {
		return m.KubectlDeleteResult, m.KubectlDeleteError
	}
	return defaultResult(), nil
}

// KubectlGet records the call and returns the configured response.
func (m *MockExecutor) KubectlGet(_ context.Context, resource, name, namespace string, output string) (*ExecResult, error) {
	m.record("KubectlGet", resource, name, namespace, output)
	if m.KubectlGetResult != nil || m.KubectlGetError != nil {
		return m.KubectlGetResult, m.KubectlGetError
	}
	return defaultResult(), nil
}

// RunDockerCommand records the call and returns the configured response.
func (m *MockExecutor) RunDockerCommand(_ context.Context, opts DockerOptions) (*ExecResult, error) {
	m.record("RunDockerCommand", opts)
	if m.DockerResult != nil || m.DockerError != nil {
		return m.DockerResult, m.DockerError
	}
	return defaultResult(), nil
}

// DockerRun records the call and returns the configured response.
func (m *MockExecutor) DockerRun(_ context.Context, opts DockerRunOptions) (*ExecResult, error) {
	m.record("DockerRun", opts)
	if m.DockerRunResult != nil || m.DockerRunError != nil {
		return m.DockerRunResult, m.DockerRunError
	}
	return defaultResult(), nil
}

// DockerStop records the call and returns the configured response.
func (m *MockExecutor) DockerStop(_ context.Context, container string, timeout time.Duration) (*ExecResult, error) {
	m.record("DockerStop", container, timeout)
	if m.DockerStopResult != nil || m.DockerStopError != nil {
		return m.DockerStopResult, m.DockerStopError
	}
	return defaultResult(), nil
}

// DockerRemove records the call and returns the configured response.
func (m *MockExecutor) DockerRemove(_ context.Context, container string, force bool) (*ExecResult, error) {
	m.record("DockerRemove", container, force)
	if m.DockerRemoveResult != nil || m.DockerRemoveError != nil {
		return m.DockerRemoveResult, m.DockerRemoveError
	}
	return defaultResult(), nil
}

// RunShell records the call and returns the configured response.
func (m *MockExecutor) RunShell(_ context.Context, opts ShellOptions) (*ExecResult, error) {
	m.record("RunShell", opts)
	if m.ShellResult != nil || m.ShellError != nil {
		return m.ShellResult, m.ShellError
	}
	return defaultResult(), nil
}

// RunScript records the call and returns the configured response.
func (m *MockExecutor) RunScript(_ context.Context, opts ScriptOptions) (*ExecResult, error) {
	m.record("RunScript", opts)
	if m.ScriptResult != nil || m.ScriptError != nil {
		return m.ScriptResult, m.ScriptError
	}
	return defaultResult(), nil
}

// ReadFile returns the configured file contents or an error.
func (m *MockExecutor) ReadFile(path string) ([]byte, error) {
	m.record("ReadFile", path)
	if content, ok := m.FileContents[path]; ok {
		return content, nil
	}
	return nil, os.ErrNotExist
}

// WriteFile records the call and returns the configured error.
func (m *MockExecutor) WriteFile(path string, data []byte, perm os.FileMode) error {
	m.record("WriteFile", path, data, perm)
	if m.WriteFileError != nil {
		return m.WriteFileError
	}
	m.mu.Lock()
	m.FileContents[path] = data
	m.mu.Unlock()
	return nil
}

// RenderTemplate returns the configured template result or renders a simple template.
func (m *MockExecutor) RenderTemplate(tmpl string, data any) ([]byte, error) {
	m.record("RenderTemplate", tmpl, data)
	if m.TemplateResult != nil || m.TemplateError != nil {
		return m.TemplateResult, m.TemplateError
	}
	return []byte(tmpl), nil
}

// RenderTemplateFile returns the configured template result.
func (m *MockExecutor) RenderTemplateFile(tmplPath string, data any) ([]byte, error) {
	m.record("RenderTemplateFile", tmplPath, data)
	if m.TemplateResult != nil || m.TemplateError != nil {
		return m.TemplateResult, m.TemplateError
	}
	if content, ok := m.FileContents[tmplPath]; ok {
		return content, nil
	}
	return nil, os.ErrNotExist
}

// FileExists returns the configured result or checks the FileContents map.
func (m *MockExecutor) FileExists(path string) bool {
	m.record("FileExists", path)
	if result, ok := m.FileExistsResults[path]; ok {
		return result
	}
	_, ok := m.FileContents[path]
	return ok
}

// MkdirAll records the call and returns the configured error.
func (m *MockExecutor) MkdirAll(path string, perm os.FileMode) error {
	m.record("MkdirAll", path, perm)
	return m.MkdirAllError
}

// Ensure MockExecutor implements Executor.
var _ Executor = (*MockExecutor)(nil)
