package exec

import "time"

// Op represents a specific operation the exec operator can perform.
type Op string

// Execution operations
const (
	// Shell executes a shell command directly.
	Shell Op = "shell"
	// Ansible runs an Ansible playbook.
	Ansible Op = "ansible"
	// Script executes a script file.
	Script Op = "script"
	// Helm runs a Helm command.
	Helm Op = "helm"
	// Kubectl runs a kubectl command.
	Kubectl Op = "kubectl"
	// Docker runs a Docker command.
	Docker Op = "docker"
)

// Query operations
const (
	// TaskStatusOp queries the status of an async task.
	TaskStatusOp Op = "task_status"
	// ListTasks queries all running/recent tasks.
	ListTasks Op = "list_tasks"
)

// Command represents a command to execute.
type Command struct {
	Name    string            `json:"name"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Dir     string            `json:"dir,omitempty"`
	Timeout time.Duration     `json:"timeout,omitempty"`
}

// CommandResult represents the result of a command execution.
type CommandResult struct {
	ExitCode int           `json:"exit_code"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	Duration time.Duration `json:"duration"`
}

// TaskStatus represents the status of an async task.
type TaskStatus struct {
	ID        string         `json:"id"`
	State     string         `json:"state"` // "pending", "running", "completed", "failed"
	StartedAt *time.Time     `json:"started_at,omitempty"`
	EndedAt   *time.Time     `json:"ended_at,omitempty"`
	Result    *CommandResult `json:"result,omitempty"`
}
