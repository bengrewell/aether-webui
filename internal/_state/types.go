package state

import "time"

// WizardStatus represents the completion status of the setup wizard.
type WizardStatus struct {
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Steps       []string   `json:"steps,omitempty"`
}

// MetricsSnapshot represents a point-in-time metrics recording.
type MetricsSnapshot struct {
	MetricType string    `json:"metric_type"`
	Data       []byte    `json:"data"`
	RecordedAt time.Time `json:"recorded_at"`
}

// CachedInfo represents cached system information with its collection time.
type CachedInfo struct {
	InfoType    string    `json:"info_type"`
	Data        []byte    `json:"data"`
	CollectedAt time.Time `json:"collected_at"`
}

// Node represents a managed node in the cluster.
type Node struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	NodeType          string    `json:"node_type"`
	Address           string    `json:"address,omitempty"`
	SSHPort           int       `json:"ssh_port,omitempty"`
	Username          string    `json:"username,omitempty"`
	AuthMethod        string    `json:"auth_method,omitempty"`
	PrivateKeyPath    string    `json:"private_key_path,omitempty"`
	Password          string    `json:"password,omitempty"`
	EncryptedPassword string    `json:"-"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Roles             []string  `json:"roles,omitempty"`
}

// NodeRole represents a role assigned to a node.
type NodeRole struct {
	ID        int       `json:"id"`
	NodeID    string    `json:"node_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// OperationLog represents an entry in the operations audit log.
type OperationLog struct {
	ID        int       `json:"id"`
	Operation string    `json:"operation"`
	NodeID    string    `json:"node_id,omitempty"`
	Detail    string    `json:"detail,omitempty"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Node type constants.
const (
	NodeTypeLocal  = "local"
	NodeTypeRemote = "remote"
)

// Auth method constants.
const (
	AuthMethodPassword   = "password"
	AuthMethodPrivateKey = "private_key"
)

// Operation name constants.
const (
	OpCreateNode         = "create_node"
	OpUpdateNode         = "update_node"
	OpDeleteNode         = "delete_node"
	OpTestConnectivity   = "test_connectivity"
	OpAssignRole         = "assign_role"
	OpRemoveRole         = "remove_role"
)

// Operation status constants.
const (
	OpStatusSuccess = "success"
	OpStatusFailure = "failure"
)

// Local node ID is always "local".
const LocalNodeID = "local"

// DeploymentTask represents an async playbook execution task.
type DeploymentTask struct {
	ID          string     `json:"id"`
	Operation   string     `json:"operation"`
	Status      string     `json:"status"`
	Output      string     `json:"output,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// ComponentDeploymentState tracks what's currently deployed for a component.
type ComponentDeploymentState struct {
	Component  string     `json:"component"`
	Status     string     `json:"status"`
	TaskID     string     `json:"task_id,omitempty"`
	DeployedAt *time.Time `json:"deployed_at,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Task status constants.
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
	TaskStatusCancelled = "cancelled"
)

// Deployment state constants.
const (
	DeployStateNotDeployed = "not_deployed"
	DeployStateDeploying   = "deploying"
	DeployStateDeployed    = "deployed"
	DeployStateFailed      = "failed"
	DeployStateUndeploying = "undeploying"
)

// Common state keys used in the app_state table.
const (
	KeyWizardCompleted   = "wizard_completed"
	KeyWizardCompletedAt = "wizard_completed_at"
	KeyWizardSteps       = "wizard_steps"
)
