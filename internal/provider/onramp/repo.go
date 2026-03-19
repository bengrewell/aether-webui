package onramp

import (
	"errors"
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

	head, err := gitOutput(cfg.OnRampDir, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("rev-parse HEAD: %w", err)
	}
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

// ensureSubmodules initializes submodules and forces a working-tree checkout.
//
// Three steps:
//  1. sync — ensures .git/config URLs match .gitmodules.
//  2. update --init — clones any submodules whose objects are missing.
//  3. foreach checkout -f — unconditionally writes the working tree for every
//     submodule. This is the critical step: if the process was killed mid-clone,
//     git may have fetched the submodule objects and recorded the correct HEAD
//     but never populated the working tree. Steps 1-2 consider such a submodule
//     "up to date" and skip it; only an explicit checkout recovers the files.
//
// After checkout, validateSubmodules confirms every submodule directory has
// content. If any are empty the provider enters degraded mode.
func ensureSubmodules(cfg Config, log *slog.Logger) error {
	log.Info("ensuring submodules are initialized", "dir", cfg.OnRampDir)

	if err := gitRun(cfg.OnRampDir, "submodule", "sync", "--recursive"); err != nil {
		return fmt.Errorf("submodule sync: %w", err)
	}

	if err := gitRun(cfg.OnRampDir, "submodule", "update", "--init", "--recursive"); err != nil {
		return fmt.Errorf("submodule update: %w", err)
	}

	if err := gitRun(cfg.OnRampDir, "submodule", "foreach", "--recursive", "git checkout -f HEAD"); err != nil {
		return fmt.Errorf("submodule foreach checkout: %w", err)
	}

	return validateSubmodules(cfg.OnRampDir, log)
}

// validateSubmodules confirms each registered submodule directory contains
// checked-out content (at least one non-.git entry).
func validateSubmodules(dir string, log *slog.Logger) error {
	gitmodules := filepath.Join(dir, ".gitmodules")
	if _, err := os.Stat(gitmodules); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("stat .gitmodules: %w", err)
	}

	out, err := gitOutput(dir, "config", "--file", ".gitmodules", "--get-regexp", `submodule\..*\.path`)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil
		}
		return fmt.Errorf("parse .gitmodules: %w", err)
	}
	if out == "" {
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

		contentCount := 0
		for _, e := range entries {
			if e.Name() != ".git" {
				contentCount++
			}
		}
		if contentCount == 0 {
			return fmt.Errorf("submodule %s: no checked-out content (clone may have failed)", subPath)
		}
		log.Debug("submodule validated", "path", subPath, "files", contentCount)
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
