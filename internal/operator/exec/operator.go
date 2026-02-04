package exec

import (
	"context"
	"fmt"

	"github.com/bengrewell/aether-webui/internal/operator"
)

// ExecOperator handles command execution on a node.
type ExecOperator interface {
	operator.Operator
	operator.Invocable

	Execute(ctx context.Context, cmd *Command) (*CommandResult, error)
	ExecuteAsync(ctx context.Context, cmd *Command) (string, error)
	GetTaskStatus(ctx context.Context, taskID string) (*TaskStatus, error)
	CancelTask(ctx context.Context, taskID string) error
}

// Operator returns "not implemented" for all methods.
type Operator struct{}

// New creates a new exec operator.
func New() *Operator {
	return &Operator{}
}

// Domain returns the operator's domain.
func (o *Operator) Domain() operator.Domain {
	return operator.DomainExec
}

// Health returns the operator's health status.
func (o *Operator) Health(_ context.Context) (*operator.OperatorHealth, error) {
	return &operator.OperatorHealth{
		Status:  "unavailable",
		Message: "not implemented",
	}, nil
}

// Execute runs a command synchronously.
func (o *Operator) Execute(_ context.Context, _ *Command) (*CommandResult, error) {
	return nil, operator.ErrNotImplemented
}

// ExecuteAsync runs a command asynchronously and returns a task ID.
func (o *Operator) ExecuteAsync(_ context.Context, _ *Command) (string, error) {
	return "", operator.ErrNotImplemented
}

// GetTaskStatus returns the status of an async task.
func (o *Operator) GetTaskStatus(_ context.Context, _ string) (*TaskStatus, error) {
	return nil, operator.ErrNotImplemented
}

// CancelTask cancels a running async task.
func (o *Operator) CancelTask(_ context.Context, _ string) error {
	return operator.ErrNotImplemented
}

// SupportedOperations returns the list of operations this operator supports.
func (o *Operator) SupportedOperations() []operator.Operation {
	return []operator.Operation{
		{Name: string(Shell), Type: operator.Action, Description: "Execute a shell command"},
		{Name: string(Ansible), Type: operator.Action, Description: "Run an Ansible playbook"},
		{Name: string(Script), Type: operator.Action, Description: "Execute a script file"},
		{Name: string(Helm), Type: operator.Action, Description: "Run a Helm command"},
		{Name: string(Kubectl), Type: operator.Action, Description: "Run a kubectl command"},
		{Name: string(Docker), Type: operator.Action, Description: "Run a Docker command"},
		{Name: string(TaskStatusOp), Type: operator.Query, Description: "Query task status"},
		{Name: string(ListTasks), Type: operator.Query, Description: "List all tasks"},
	}
}

// Invoke executes a named operation with the given arguments.
func (o *Operator) Invoke(_ context.Context, opType operator.OperationType, operation string, _ ...any) (any, error) {
	op := Op(operation)
	switch opType {
	case operator.Action:
		switch op {
		case Shell, Ansible, Script, Helm, Kubectl, Docker:
			return nil, operator.ErrNotImplemented
		default:
			return nil, fmt.Errorf("unknown action operation: %s", operation)
		}
	case operator.Query:
		switch op {
		case TaskStatusOp, ListTasks:
			return nil, operator.ErrNotImplemented
		default:
			return nil, fmt.Errorf("unknown query operation: %s", operation)
		}
	default:
		return nil, fmt.Errorf("unknown operation type: %s", opType)
	}
}
