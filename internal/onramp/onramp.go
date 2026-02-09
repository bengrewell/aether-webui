package onramp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/bengrewell/aether-webui/internal/state"
)

// Config holds OnRamp manager configuration.
type Config struct {
	RepoURL       string // default: "https://github.com/opennetworkinglab/aether-onramp"
	WorkDir       string // path where OnRamp is cloned
	Branch        string // default: "main"
	EncryptionKey string // for decrypting node passwords when generating hosts.ini
}

// DefaultRepoURL is the upstream OnRamp repository.
const DefaultRepoURL = "https://github.com/opennetworkinglab/aether-onramp"

// Manager manages the OnRamp repository checkout and generates configuration files.
type Manager struct {
	config Config
	store  state.Store
	mu     sync.Mutex
}

// NewManager creates a new OnRamp manager.
func NewManager(cfg Config, store state.Store) *Manager {
	if cfg.RepoURL == "" {
		cfg.RepoURL = DefaultRepoURL
	}
	if cfg.Branch == "" {
		cfg.Branch = "main"
	}
	return &Manager{config: cfg, store: store}
}

// EnsureRepo clones the OnRamp repository if it doesn't exist, or pulls updates.
func (m *Manager) EnsureRepo(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(m.config.WorkDir), 0750); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	gitDir := filepath.Join(m.config.WorkDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		slog.Info("cloning OnRamp repository", "url", m.config.RepoURL, "branch", m.config.Branch, "dir", m.config.WorkDir)
		cmd := exec.CommandContext(ctx, "git", "clone",
			"--recurse-submodules",
			"--branch", m.config.Branch,
			m.config.RepoURL,
			m.config.WorkDir,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}
	} else {
		slog.Info("updating OnRamp repository", "dir", m.config.WorkDir)
		cmd := exec.CommandContext(ctx, "git", "-C", m.config.WorkDir, "pull", "--recurse-submodules")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git pull failed: %w", err)
		}
	}

	return EnsureVarsFile(m.config.WorkDir)
}

// GenerateHostsINI reads nodes and roles from the store and writes hosts.ini.
func (m *Manager) GenerateHostsINI(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inventoryPath := filepath.Join(m.config.WorkDir, "hosts.ini")
	return GenerateInventory(ctx, m.store, m.config.EncryptionKey, inventoryPath)
}

// GenerateVarsFile activates the named blueprint (or ensures default vars exist).
func (m *Manager) GenerateVarsFile(_ context.Context, blueprint string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if blueprint != "" {
		return ActivateBlueprint(m.config.WorkDir, blueprint)
	}
	return EnsureVarsFile(m.config.WorkDir)
}

// ListBlueprints returns available OnRamp blueprint names.
func (m *Manager) ListBlueprints() ([]string, error) {
	return ListBlueprints(m.config.WorkDir)
}

// RepoPath returns the OnRamp checkout path.
func (m *Manager) RepoPath() string {
	return m.config.WorkDir
}

// InventoryPath returns the path to the generated hosts.ini file.
func (m *Manager) InventoryPath() string {
	return filepath.Join(m.config.WorkDir, "hosts.ini")
}

// IsRepoReady returns true if the OnRamp repo has been cloned.
func (m *Manager) IsRepoReady() bool {
	gitDir := filepath.Join(m.config.WorkDir, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// EncryptionKey returns the configured encryption key.
func (m *Manager) EncryptionKey() string {
	return m.config.EncryptionKey
}
