package executor

import (
	"testing"
)

func TestBuildShellArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     ShellOptions
		wantCmd  string
		wantArgs []string
	}{
		{
			name: "default shell",
			opts: ShellOptions{
				Command: "echo hello",
			},
			wantCmd:  "/bin/sh",
			wantArgs: []string{"-c", "echo hello"},
		},
		{
			name: "custom shell",
			opts: ShellOptions{
				Command: "echo hello",
				Shell:   "/bin/bash",
			},
			wantCmd:  "/bin/bash",
			wantArgs: []string{"-c", "echo hello"},
		},
		{
			name: "with extra args",
			opts: ShellOptions{
				Command: "echo hello",
				Shell:   "/bin/bash",
				Args:    []string{"-x"},
			},
			wantCmd:  "/bin/bash",
			wantArgs: []string{"-c", "echo hello", "-x"},
		},
		{
			name: "complex command",
			opts: ShellOptions{
				Command: "for i in 1 2 3; do echo $i; done",
			},
			wantCmd:  "/bin/sh",
			wantArgs: []string{"-c", "for i in 1 2 3; do echo $i; done"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, args := buildShellArgs(tc.opts)
			if cmd != tc.wantCmd {
				t.Errorf("buildShellArgs() cmd = %q, want %q", cmd, tc.wantCmd)
			}
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildShellArgs() args = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

// Helper function for building args (extracted for testing)
func buildShellArgs(opts ShellOptions) (string, []string) {
	shell := opts.Shell
	if shell == "" {
		shell = "/bin/sh"
	}

	args := make([]string, 0, len(opts.Args)+2)
	args = append(args, "-c", opts.Command)
	args = append(args, opts.Args...)

	return shell, args
}
