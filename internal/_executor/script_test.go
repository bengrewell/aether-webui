package executor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectInterpreter(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"script.sh", "/bin/sh"},
		{"script.bash", "/bin/bash"},
		{"script.py", "python3"},
		{"script.rb", "ruby"},
		{"script.pl", "perl"},
		{"script.js", "node"},
		{"script.ps1", "pwsh"},
		{"script.unknown", ""},
		{"script", ""},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			got := detectInterpreter(tc.path)
			if got != tc.want {
				t.Errorf("detectInterpreter(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestReadShebang(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "bash shebang",
			content: "#!/bin/bash\necho hello",
			want:    "/bin/bash",
		},
		{
			name:    "sh shebang",
			content: "#!/bin/sh\necho hello",
			want:    "/bin/sh",
		},
		{
			name:    "env python",
			content: "#!/usr/bin/env python3\nprint('hello')",
			want:    "python3",
		},
		{
			name:    "env ruby",
			content: "#!/usr/bin/env ruby\nputs 'hello'",
			want:    "ruby",
		},
		{
			name:    "no shebang",
			content: "echo hello",
			want:    "",
		},
		{
			name:    "empty file",
			content: "",
			want:    "",
		},
		{
			name:    "shebang with args",
			content: "#!/usr/bin/perl -w\nprint 'hello';",
			want:    "/usr/bin/perl",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "script")
			if err := os.WriteFile(path, []byte(tc.content), 0755); err != nil {
				t.Fatalf("failed to write test script: %v", err)
			}

			got := readShebang(path)
			if got != tc.want {
				t.Errorf("readShebang() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestReadShebang_FileNotExists(t *testing.T) {
	got := readShebang("/nonexistent/script.sh")
	if got != "" {
		t.Errorf("readShebang() = %q, want empty string", got)
	}
}

func TestBuildScriptArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     ScriptOptions
		wantCmd  string
		wantArgs []string
	}{
		{
			name: "with interpreter",
			opts: ScriptOptions{
				Path:        "/scripts/deploy.sh",
				Interpreter: "/bin/bash",
				Args:        []string{"--env", "prod"},
			},
			wantCmd:  "/bin/bash",
			wantArgs: []string{"/scripts/deploy.sh", "--env", "prod"},
		},
		{
			name: "no interpreter (direct exec)",
			opts: ScriptOptions{
				Path: "/scripts/deploy.sh",
				Args: []string{"arg1"},
			},
			wantCmd:  "/scripts/deploy.sh",
			wantArgs: []string{"arg1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, args := buildScriptArgs(tc.opts)
			if cmd != tc.wantCmd {
				t.Errorf("buildScriptArgs() cmd = %q, want %q", cmd, tc.wantCmd)
			}
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildScriptArgs() args = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

// Helper function for building args (extracted for testing)
func buildScriptArgs(opts ScriptOptions) (string, []string) {
	if opts.Interpreter != "" {
		args := make([]string, 0, len(opts.Args)+1)
		args = append(args, opts.Path)
		args = append(args, opts.Args...)
		return opts.Interpreter, args
	}
	return opts.Path, opts.Args
}
