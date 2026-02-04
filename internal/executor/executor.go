package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/bgrewell/go-execute"
)

// Executor provides methods for executing external commands and file operations.
// It wraps go-execute and provides domain-specific functions for common tools.
type Executor interface {
	// Ansible
	RunAnsiblePlaybook(ctx context.Context, opts AnsibleOptions) (*ExecResult, error)

	// Helm
	RunHelmInstall(ctx context.Context, opts HelmInstallOptions) (*ExecResult, error)
	RunHelmUpgrade(ctx context.Context, opts HelmUpgradeOptions) (*ExecResult, error)
	RunHelmUninstall(ctx context.Context, opts HelmUninstallOptions) (*ExecResult, error)
	RunHelmList(ctx context.Context, opts HelmListOptions) (*HelmReleaseList, error)
	RunHelmStatus(ctx context.Context, release, namespace string) (*HelmReleaseStatus, error)

	// Kubectl
	RunKubectl(ctx context.Context, opts KubectlOptions) (*ExecResult, error)
	KubectlApply(ctx context.Context, manifest []byte, namespace string) (*ExecResult, error)
	KubectlDelete(ctx context.Context, resource, name, namespace string) (*ExecResult, error)
	KubectlGet(ctx context.Context, resource, name, namespace string, output string) (*ExecResult, error)

	// Docker
	RunDockerCommand(ctx context.Context, opts DockerOptions) (*ExecResult, error)
	DockerRun(ctx context.Context, opts DockerRunOptions) (*ExecResult, error)
	DockerStop(ctx context.Context, container string, timeout time.Duration) (*ExecResult, error)
	DockerRemove(ctx context.Context, container string, force bool) (*ExecResult, error)

	// Shell/Script (explicit opt-in)
	RunShell(ctx context.Context, opts ShellOptions) (*ExecResult, error)
	RunScript(ctx context.Context, opts ScriptOptions) (*ExecResult, error)

	// File operations
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	RenderTemplate(tmpl string, data any) ([]byte, error)
	RenderTemplateFile(tmplPath string, data any) ([]byte, error)
	FileExists(path string) bool
	MkdirAll(path string, perm os.FileMode) error
}

// Config holds configuration for the DefaultExecutor.
type Config struct {
	DefaultTimeout time.Duration // Default timeout for commands
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		DefaultTimeout: 10 * time.Minute,
	}
}

// DefaultExecutor is the standard implementation of Executor.
type DefaultExecutor struct {
	config Config
}

// New creates a new DefaultExecutor with the given configuration.
func New(config Config) *DefaultExecutor {
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = DefaultConfig().DefaultTimeout
	}
	return &DefaultExecutor{config: config}
}

// runCommand executes a command using go-execute and returns the result.
func (e *DefaultExecutor) runCommand(ctx context.Context, base BaseOptions, name string, args ...string) (*ExecResult, error) {
	timeout := base.Timeout
	if timeout == 0 {
		timeout = e.config.DefaultTimeout
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build executor with options
	executor := execute.NewExecutor()

	if base.WorkingDir != "" {
		executor.SetWorkingDir(base.WorkingDir)
	}

	if len(base.Env) > 0 {
		// Convert map to slice of KEY=VALUE strings
		envSlice := os.Environ() // Start with current environment
		for k, v := range base.Env {
			envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
		}
		executor.SetEnvironment(envSlice)
	}

	// Execute command
	result, err := executor.ExecuteContext(ctx, name, args...)
	if err != nil {
		// Check if it's a context error (timeout)
		if ctx.Err() != nil {
			return &ExecResult{
				ExitCode: -1,
				Stderr:   ctx.Err().Error(),
			}, nil
		}
		// For other errors, we still try to return the result
		if result != nil {
			return &ExecResult{
				ExitCode: result.ExitCode,
				Stdout:   result.Stdout,
				Stderr:   result.Stderr,
				Duration: result.Duration(),
			}, nil
		}
		return nil, err
	}

	return &ExecResult{
		ExitCode: result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		Duration: result.Duration(),
	}, nil
}

// runCommandWithStdin executes a command with stdin input.
// Note: go-execute doesn't directly support stdin, so we use os/exec for this.
func (e *DefaultExecutor) runCommandWithStdin(ctx context.Context, base BaseOptions, stdin []byte, name string, args ...string) (*ExecResult, error) {
	timeout := base.Timeout
	if timeout == 0 {
		timeout = e.config.DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)

	if base.WorkingDir != "" {
		cmd.Dir = base.WorkingDir
	}

	if len(base.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range base.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	cmd.Stdin = bytes.NewReader(stdin)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if ctx.Err() != nil {
			return &ExecResult{
				ExitCode: -1,
				Stderr:   ctx.Err().Error(),
				Duration: duration,
			}, nil
		} else {
			return nil, err
		}
	}

	return &ExecResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}, nil
}
