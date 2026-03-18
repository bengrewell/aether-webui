package onramp

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ensureRepo clones the OnRamp repo if it does not exist, checks out the
// pinned version, ensures submodules are initialized, and validates that
// expected files are present.
func ensureRepo(cfg Config, log *slog.Logger) error {
	if err := cloneIfMissing(cfg, log); err != nil {
		return err
	}
	if err := checkoutVersion(cfg, log); err != nil {
		return err
	}
	if err := ensureSubmodules(cfg, log); err != nil {
		return err
	}
	return validateRepo(cfg.OnRampDir)
}

// cloneIfMissing clones the OnRamp repo when the target directory is absent
// or does not contain a .git directory.
func cloneIfMissing(cfg Config, log *slog.Logger) error {
	gitDir := filepath.Join(cfg.OnRampDir, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		log.Info("onramp repo already present", "dir", cfg.OnRampDir)
		return nil
	}

	log.Info("cloning onramp repo", "url", cfg.RepoURL, "dir", cfg.OnRampDir)
	cmd := exec.Command("git", "clone", "--recurse-submodules", cfg.RepoURL, cfg.OnRampDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	return nil
}

// checkoutVersion switches the repo to the configured version (tag, branch, or
// commit). Skipped when Version is empty or "main".
func checkoutVersion(cfg Config, log *slog.Logger) error {
	if cfg.Version == "" || cfg.Version == "main" {
		return nil
	}

	// Determine current HEAD ref to avoid unnecessary checkout.
	head, err := gitOutput(cfg.OnRampDir, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("rev-parse HEAD: %w", err)
	}
	// Resolve the desired version to a commit hash for comparison.
	desired, err := gitOutput(cfg.OnRampDir, "rev-parse", cfg.Version)
	if err == nil && head == desired {
		log.Info("onramp repo already at desired version", "version", cfg.Version)
		return nil
	}

	log.Info("checking out onramp version", "version", cfg.Version)
	if err := gitRun(cfg.OnRampDir, "checkout", cfg.Version); err != nil {
		return fmt.Errorf("git checkout %s: %w", cfg.Version, err)
	}
	return nil
}

// ensureSubmodules runs submodule init/update unconditionally. The initial
// clone uses --recurse-submodules, but that can silently leave submodules
// uninitialized on transient network errors. Running this as a separate step
// catches those cases and is a no-op when submodules are already up to date.
func ensureSubmodules(cfg Config, log *slog.Logger) error {
	log.Info("ensuring submodules are initialized", "dir", cfg.OnRampDir)
	if err := gitRun(cfg.OnRampDir, "submodule", "update", "--init", "--recursive"); err != nil {
		return fmt.Errorf("submodule update: %w", err)
	}
	return nil
}

// validateRepo confirms the OnRamp directory contains the expected Makefile
// and default vars file.
func validateRepo(dir string) error {
	for _, rel := range []string{"Makefile", filepath.Join("vars", "main.yml")} {
		p := filepath.Join(dir, rel)
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("onramp validation: missing %s: %w", rel, err)
		}
	}
	return nil
}

// gitRun executes a git command inside dir.
func gitRun(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// gitOutput executes a git command inside dir and returns trimmed stdout.
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
