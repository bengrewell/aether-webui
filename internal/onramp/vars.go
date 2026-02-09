package onramp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ListBlueprints returns the names of available blueprint files in vars/.
// Blueprint names are derived from files matching vars/main-*.yml.
func ListBlueprints(onrampPath string) ([]string, error) {
	pattern := filepath.Join(onrampPath, "vars", "main-*.yml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob blueprints: %w", err)
	}

	names := make([]string, 0, len(matches))
	for _, m := range matches {
		base := filepath.Base(m)
		name := strings.TrimPrefix(base, "main-")
		name = strings.TrimSuffix(name, ".yml")
		names = append(names, name)
	}
	return names, nil
}

// ActivateBlueprint copies vars/main-<name>.yml to vars/main.yml.
func ActivateBlueprint(onrampPath, name string) error {
	src := filepath.Join(onrampPath, "vars", "main-"+name+".yml")
	dst := filepath.Join(onrampPath, "vars", "main.yml")

	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("blueprint %q not found: %s", name, src)
	}

	return copyFile(src, dst)
}

// EnsureVarsFile ensures vars/main.yml exists. If it doesn't, copies
// vars/main-quickstart.yml as the default.
func EnsureVarsFile(onrampPath string) error {
	mainYml := filepath.Join(onrampPath, "vars", "main.yml")
	if _, err := os.Stat(mainYml); err == nil {
		return nil // already exists
	}

	defaultSrc := filepath.Join(onrampPath, "vars", "main-quickstart.yml")
	if _, err := os.Stat(defaultSrc); os.IsNotExist(err) {
		return nil // no default blueprint available
	}

	return copyFile(defaultSrc, mainYml)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return out.Close()
}
