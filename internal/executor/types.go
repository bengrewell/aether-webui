package executor

import (
	"os"
	"time"
)

// ExecResult represents the result of a command execution.
type ExecResult struct {
	ExitCode int           `json:"exit_code"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	Duration time.Duration `json:"duration"`
}

// Success returns true if the command exited with code 0.
func (r *ExecResult) Success() bool {
	return r.ExitCode == 0
}

// BaseOptions contains common options for all command executions.
type BaseOptions struct {
	WorkingDir string
	Env        map[string]string
	Timeout    time.Duration
}

// AnsibleOptions configures an Ansible playbook execution.
type AnsibleOptions struct {
	BaseOptions
	Playbook   string            // Path to playbook file
	Inventory  string            // Inventory file or host pattern
	ExtraVars  map[string]string // Extra variables to pass
	Limit      string            // Limit to specific hosts
	Tags       []string          // Only run tasks with these tags
	SkipTags   []string          // Skip tasks with these tags
	Become     bool              // Run with privilege escalation
	BecomeUser string            // User to become (default: root)
	Verbosity  int               // Verbosity level (0-4)
	Check      bool              // Run in check mode (dry-run)
	Diff       bool              // Show differences
	Forks      int               // Number of parallel processes
}

// HelmInstallOptions configures a Helm install operation.
type HelmInstallOptions struct {
	BaseOptions
	ReleaseName string            // Name of the release
	Chart       string            // Chart name or path
	Namespace   string            // Kubernetes namespace
	Values      map[string]any    // Values to pass (--set)
	ValuesFiles []string          // Values files (--values)
	Version     string            // Chart version
	Repo        string            // Chart repository URL
	Wait        bool              // Wait for resources to be ready
	WaitTimeout time.Duration     // Timeout for --wait
	CreateNS    bool              // Create namespace if not exists
	Atomic      bool              // Roll back on failure
	DryRun      bool              // Simulate install
	Description string            // Release description
	Replace     bool              // Replace if exists
}

// HelmUpgradeOptions configures a Helm upgrade operation.
type HelmUpgradeOptions struct {
	BaseOptions
	ReleaseName   string            // Name of the release
	Chart         string            // Chart name or path
	Namespace     string            // Kubernetes namespace
	Values        map[string]any    // Values to pass (--set)
	ValuesFiles   []string          // Values files (--values)
	Version       string            // Chart version
	Repo          string            // Chart repository URL
	Wait          bool              // Wait for resources to be ready
	WaitTimeout   time.Duration     // Timeout for --wait
	Install       bool              // Install if not exists
	Atomic        bool              // Roll back on failure
	DryRun        bool              // Simulate upgrade
	Description   string            // Release description
	ReuseValues   bool              // Reuse existing values
	ResetValues   bool              // Reset values to chart defaults
	Force         bool              // Force resource updates
	CleanupOnFail bool              // Delete new resources on failure
}

// HelmUninstallOptions configures a Helm uninstall operation.
type HelmUninstallOptions struct {
	BaseOptions
	ReleaseName string // Name of the release
	Namespace   string // Kubernetes namespace
	KeepHistory bool   // Keep release history
	DryRun      bool   // Simulate uninstall
	Wait        bool   // Wait for deletion
	WaitTimeout time.Duration
	Description string
}

// HelmListOptions configures a Helm list operation.
type HelmListOptions struct {
	BaseOptions
	Namespace   string   // Namespace to filter (empty for all)
	AllNS       bool     // List across all namespaces
	Filter      string   // Regex filter for release names
	Deployed    bool     // Show deployed releases
	Failed      bool     // Show failed releases
	Pending     bool     // Show pending releases
	Superseded  bool     // Show superseded releases
	Uninstalled bool     // Show uninstalled releases (if history kept)
	All         bool     // Show all releases
	Offset      int      // Offset for pagination
	Max         int      // Maximum number of releases
	SortBy      string   // Sort by: name, date
	Reverse     bool     // Reverse sort order
}

// HelmRelease represents a Helm release in list output.
type HelmRelease struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Revision   int       `json:"revision"`
	Status     string    `json:"status"`
	Chart      string    `json:"chart"`
	AppVersion string    `json:"app_version"`
	Updated    time.Time `json:"updated"`
}

// HelmReleaseList is a list of Helm releases.
type HelmReleaseList struct {
	Releases []HelmRelease `json:"releases"`
}

// HelmReleaseStatus represents the status of a Helm release.
type HelmReleaseStatus struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Revision   int       `json:"revision"`
	Status     string    `json:"status"`
	Chart      string    `json:"chart"`
	AppVersion string    `json:"app_version"`
	Updated    time.Time `json:"updated"`
	Notes      string    `json:"notes,omitempty"`
	Manifest   string    `json:"manifest,omitempty"`
}

// KubectlOptions configures a kubectl command execution.
type KubectlOptions struct {
	BaseOptions
	Args       []string // Command arguments
	Namespace  string   // Kubernetes namespace
	Context    string   // Kubeconfig context
	Kubeconfig string   // Path to kubeconfig file
	Output     string   // Output format (json, yaml, wide, name)
}

// DockerOptions configures a generic Docker command execution.
type DockerOptions struct {
	BaseOptions
	Args []string // Command arguments
	Host string   // Docker host (DOCKER_HOST)
}

// DockerRunOptions configures a Docker run operation.
type DockerRunOptions struct {
	BaseOptions
	Image       string            // Image to run
	Name        string            // Container name
	Command     []string          // Command and arguments
	Detach      bool              // Run in background
	Remove      bool              // Remove container on exit
	Interactive bool              // Keep STDIN open
	TTY         bool              // Allocate pseudo-TTY
	Privileged  bool              // Give extended privileges
	Network     string            // Network to connect to
	Ports       map[string]string // Port mappings (host:container)
	Volumes     map[string]string // Volume mounts (host:container)
	EnvVars     map[string]string // Environment variables
	Labels      map[string]string // Container labels
	User        string            // Username or UID
	Hostname    string            // Container hostname
	RestartPolicy string          // Restart policy
	Memory      string            // Memory limit
	CPUs        string            // CPU limit
	Host        string            // Docker host
}

// ShellOptions configures a shell command execution.
type ShellOptions struct {
	BaseOptions
	Command string   // Command to execute
	Shell   string   // Shell to use (default: /bin/sh)
	Args    []string // Additional shell arguments
}

// ScriptOptions configures a script execution.
type ScriptOptions struct {
	BaseOptions
	Path        string   // Path to script file
	Args        []string // Arguments to pass to script
	Interpreter string   // Interpreter to use (auto-detect if empty)
}

// FileWriteOptions configures a file write operation.
type FileWriteOptions struct {
	Path    string
	Data    []byte
	Perm    os.FileMode
	MkdirAll bool // Create parent directories if needed
}

// TemplateData is a type alias for template data.
type TemplateData = any
