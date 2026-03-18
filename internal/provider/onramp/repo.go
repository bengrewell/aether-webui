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

// ensureSubmodules syncs submodule URLs and forces initialization. The initial
// clone uses --recurse-submodules, but that can silently leave submodules
// uninitialized on transient network errors. Running sync + update --force as
// a separate step catches those cases. When submodules are already populated
// the force-update is a fast no-op (git checks out the same commit).
//
// After the update, validateSubmodules confirms that every registered submodule
// directory actually contains files. If any are empty the function returns an
// error so the provider enters degraded mode instead of silently failing later.
func ensureSubmodules(cfg Config, log *slog.Logger) error {
	log.Info("ensuring submodules are initialized", "dir", cfg.OnRampDir)

	// Sync ensures .git/config URLs match .gitmodules (handles URL changes).
	if err := gitRun(cfg.OnRampDir, "submodule", "sync", "--recursive"); err != nil {
		return fmt.Errorf("submodule sync: %w", err)
	}

	// Force re-checkout to handle partially initialized submodules where the
	// working tree is empty but git considers the submodule "registered".
	if err := gitRun(cfg.OnRampDir, "submodule", "update", "--init", "--recursive", "--force"); err != nil {
		return fmt.Errorf("submodule update: %w", err)
	}

	return validateSubmodules(cfg.OnRampDir, log)
}

// validateSubmodules reads .gitmodules and confirms each registered submodule
// directory contains at least one file. Empty directories indicate a failed or
// incomplete submodule clone.
func validateSubmodules(dir string, log *slog.Logger) error {
	// Parse submodule paths from .gitmodules.
	gitmodules := filepath.Join(dir, ".gitmodules")
	if _, err := os.Stat(gitmodules); err != nil {
		// No .gitmodules means no submodules to validate.
		return nil
	}

	out, err := gitOutput(dir, "config", "--file", ".gitmodules", "--get-regexp", `submodule\..*\.path`)
	if err != nil || out == "" {
		return nil
	}

	for _, line := range strings.Split(out, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
		if len(parts) != 2 {
			continue
		}
		subPath := parts[1]
		absPath := filepath.Join(dir, subPath)

		entries, err := os.ReadDir(absPath)
		if err != nil {
			return fmt.Errorf("submodule %s: directory missing: %w", subPath, err)
		}
		if len(entries) == 0 {
			return fmt.Errorf("submodule %s: directory is empty (clone may have failed)", subPath)
		}
		log.Debug("submodule validated", "path", subPath, "files", len(entries))
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
