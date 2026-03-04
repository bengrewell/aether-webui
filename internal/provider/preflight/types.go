package preflight

import (
	"context"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"os/user"
	"time"

	"github.com/bengrewell/aether-webui/internal/store"
)

// Severity indicates how critical a check failure is.
type Severity string

const (
	SeverityRequired Severity = "required"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// Category groups checks by domain.
type Category string

const (
	CategoryTooling Category = "tooling"
	CategoryAccess  Category = "access"
	CategoryNetwork Category = "network"
)

// CheckDeps provides injectable dependencies for testability.
type CheckDeps struct {
	Store       store.Client
	Log         *slog.Logger
	LookPath    func(string) (string, error)
	ReadFile    func(string) ([]byte, error)
	ReadDir     func(string) ([]fs.DirEntry, error)
	LookupUser  func(string) (*user.User, error)
	RunCommand  func(ctx context.Context, name string, args ...string) ([]byte, error)
	DialTimeout func(network, addr string, timeout time.Duration) (net.Conn, error)
	Stat        func(string) (os.FileInfo, error)
}

// DefaultDeps returns production-wired dependencies.
func DefaultDeps(st store.Client, log *slog.Logger) CheckDeps {
	return CheckDeps{
		Store:      st,
		Log:        log,
		LookPath:   exec.LookPath,
		ReadFile:   os.ReadFile,
		ReadDir:    os.ReadDir,
		LookupUser: user.Lookup,
		RunCommand: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).CombinedOutput()
		},
		DialTimeout: net.DialTimeout,
		Stat:        os.Stat,
	}
}

// Check is a self-contained preflight check definition.
type Check struct {
	ID          string
	Name        string
	Description string
	Severity    Severity
	Category    Category
	FixWarning  string
	RunCheck    func(ctx context.Context, deps CheckDeps) CheckResult
	RunFix      func(ctx context.Context, deps CheckDeps) FixResult // nil = no fix available
}

// CheckResult is the API-facing result of running a single check.
type CheckResult struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	Category    Category `json:"category"`
	Passed      bool     `json:"passed"`
	Message     string   `json:"message"`
	Details     string   `json:"details,omitempty"`
	Notes       string   `json:"notes,omitempty"`
	CanFix      bool     `json:"can_fix"`
	FixWarning  string   `json:"fix_warning,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// FixResult is the API-facing result of running a fix.
type FixResult struct {
	ID      string `json:"id"`
	Applied bool   `json:"applied"`
	Message string `json:"message"`
	Warning string `json:"warning,omitempty"`
	Error   string `json:"error,omitempty"`
}

// PreflightSummary is the aggregate result of running all checks.
type PreflightSummary struct {
	Passed  int           `json:"passed"`
	Failed  int           `json:"failed"`
	Total   int           `json:"total"`
	Results []CheckResult `json:"results"`
}

// ---------------------------------------------------------------------------
// Huma I/O types
// ---------------------------------------------------------------------------

// PreflightListOutput wraps the summary for the list endpoint.
type PreflightListOutput struct {
	Body PreflightSummary
}

// PreflightGetInput captures the check ID from the path.
type PreflightGetInput struct {
	ID string `path:"id" doc:"Check ID"`
}

// PreflightGetOutput wraps a single check result.
type PreflightGetOutput struct {
	Body CheckResult
}

// PreflightFixInput captures the check ID from the path.
type PreflightFixInput struct {
	ID string `path:"id" doc:"Check ID"`
}

// PreflightFixOutput wraps a fix result.
type PreflightFixOutput struct {
	Body FixResult
}
