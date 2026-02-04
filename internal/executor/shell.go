package executor

import (
	"context"
)

// RunShell executes a shell command.
// This is an explicit opt-in for shell execution, as opposed to direct command execution.
func (e *DefaultExecutor) RunShell(ctx context.Context, opts ShellOptions) (*ExecResult, error) {
	shell := opts.Shell
	if shell == "" {
		shell = "/bin/sh"
	}

	args := make([]string, 0, len(opts.Args)+2)
	args = append(args, "-c", opts.Command)
	args = append(args, opts.Args...)

	return e.runCommand(ctx, opts.BaseOptions, shell, args...)
}
