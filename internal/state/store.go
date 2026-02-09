package state

import (
	"context"
	"time"
)

// Store defines the interface for persistent state storage.
// Implementations handle app state, wizard status, system info caching,
// and metrics history.
type Store interface {
	// App state (key-value)
	GetState(ctx context.Context, key string) (string, error)
	SetState(ctx context.Context, key string, value string) error
	DeleteState(ctx context.Context, key string) error

	// Wizard-specific helpers
	GetWizardStatus(ctx context.Context) (*WizardStatus, error)
	SetWizardComplete(ctx context.Context, steps []string) error
	ClearWizardStatus(ctx context.Context) error

	// System info cache
	GetCachedSystemInfo(ctx context.Context, infoType string) ([]byte, time.Time, error)
	SetCachedSystemInfo(ctx context.Context, infoType string, data []byte) error

	// Metrics history
	RecordMetrics(ctx context.Context, metricType string, data []byte) error
	GetMetricsHistory(ctx context.Context, metricType string, limit int) ([]MetricsSnapshot, error)
	GetMetricsRange(ctx context.Context, metricType string, start, end time.Time) ([]MetricsSnapshot, error)
	PruneOldMetrics(ctx context.Context, olderThan time.Duration) error

	// Node management
	CreateNode(ctx context.Context, node *Node) error
	GetNode(ctx context.Context, id string) (*Node, error)
	ListNodes(ctx context.Context) ([]Node, error)
	UpdateNode(ctx context.Context, node *Node) error
	DeleteNode(ctx context.Context, id string) error
	EnsureLocalNode(ctx context.Context) (*Node, error)

	// Node roles
	AssignRole(ctx context.Context, nodeID string, role string) error
	RemoveRole(ctx context.Context, nodeID string, role string) error
	GetNodeRoles(ctx context.Context, nodeID string) ([]NodeRole, error)

	// Operations log
	LogOperation(ctx context.Context, entry *OperationLog) error
	GetOperationsLog(ctx context.Context, limit, offset int) ([]OperationLog, int, error)
	GetOperationsLogByNode(ctx context.Context, nodeID string, limit, offset int) ([]OperationLog, int, error)

	// Schema
	GetSchemaVersion() (int, error)

	// Lifecycle
	Close() error
}
