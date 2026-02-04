package executor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// RunScript executes a script file.
func (e *DefaultExecutor) RunScript(ctx context.Context, opts ScriptOptions) (*ExecResult, error) {
	interpreter := opts.Interpreter

	// Auto-detect interpreter if not specified
	if interpreter == "" {
		interpreter = detectInterpreter(opts.Path)
	}

	// If still no interpreter, try to execute directly (relies on shebang)
	if interpreter == "" {
		return e.runCommand(ctx, opts.BaseOptions, opts.Path, opts.Args...)
	}

	// Execute with interpreter
	args := make([]string, 0, len(opts.Args)+1)
	args = append(args, opts.Path)
	args = append(args, opts.Args...)

	return e.runCommand(ctx, opts.BaseOptions, interpreter, args...)
}

// detectInterpreter attempts to determine the appropriate interpreter for a script.
func detectInterpreter(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".sh":
		return "/bin/sh"
	case ".bash":
		return "/bin/bash"
	case ".py":
		return "python3"
	case ".rb":
		return "ruby"
	case ".pl":
		return "perl"
	case ".js":
		return "node"
	case ".ps1":
		return "pwsh"
	default:
		// Try to read shebang
		return readShebang(path)
	}
}

// readShebang reads the shebang line from a script file.
func readShebang(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	// Read first 256 bytes to find shebang
	buf := make([]byte, 256)
	n, err := f.Read(buf)
	if err != nil || n < 2 {
		return ""
	}

	// Check for shebang
	if buf[0] != '#' || buf[1] != '!' {
		return ""
	}

	// Find end of line
	line := string(buf[2:n])
	if idx := strings.IndexByte(line, '\n'); idx != -1 {
		line = line[:idx]
	}

	// Extract interpreter (handle /usr/bin/env python3 style)
	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return ""
	}

	// If using env, return the next part
	if strings.HasSuffix(parts[0], "/env") && len(parts) > 1 {
		return parts[1]
	}

	return parts[0]
}
