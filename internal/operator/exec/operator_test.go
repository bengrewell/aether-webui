package exec

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bengrewell/aether-webui/internal/operator"
)

func TestOperatorImplementsExecOperator(t *testing.T) {
	var _ ExecOperator = (*Operator)(nil)
}

func TestOperatorImplementsInvocable(t *testing.T) {
	var _ operator.Invocable = (*Operator)(nil)
}

func TestNew(t *testing.T) {
	op := New()
	if op == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDomain(t *testing.T) {
	op := New()
	if got := op.Domain(); got != operator.DomainExec {
		t.Errorf("Domain() = %q, want %q", got, operator.DomainExec)
	}
}

func TestHealth(t *testing.T) {
	op := New()
	health, err := op.Health(context.Background())

	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health == nil {
		t.Fatal("Health() returned nil")
	}
	if health.Status != "unavailable" {
		t.Errorf("Status = %q, want %q", health.Status, "unavailable")
	}
	if health.Message != "not implemented" {
		t.Errorf("Message = %q, want %q", health.Message, "not implemented")
	}
}

func TestMethodsReturnErrNotImplemented(t *testing.T) {
	op := New()
	ctx := context.Background()

	t.Run("Execute", func(t *testing.T) {
		_, err := op.Execute(ctx, &Command{Name: "test"})
		if !errors.Is(err, operator.ErrNotImplemented) {
			t.Errorf("Execute() error = %v, want ErrNotImplemented", err)
		}
	})

	t.Run("ExecuteAsync", func(t *testing.T) {
		taskID, err := op.ExecuteAsync(ctx, &Command{Name: "test"})
		if !errors.Is(err, operator.ErrNotImplemented) {
			t.Errorf("ExecuteAsync() error = %v, want ErrNotImplemented", err)
		}
		if taskID != "" {
			t.Errorf("ExecuteAsync() taskID = %q, want empty", taskID)
		}
	})

	t.Run("GetTaskStatus", func(t *testing.T) {
		_, err := op.GetTaskStatus(ctx, "task-123")
		if !errors.Is(err, operator.ErrNotImplemented) {
			t.Errorf("GetTaskStatus() error = %v, want ErrNotImplemented", err)
		}
	})

	t.Run("CancelTask", func(t *testing.T) {
		err := op.CancelTask(ctx, "task-123")
		if !errors.Is(err, operator.ErrNotImplemented) {
			t.Errorf("CancelTask() error = %v, want ErrNotImplemented", err)
		}
	})
}

func TestSupportedOperations(t *testing.T) {
	op := New()
	ops := op.SupportedOperations()

	if len(ops) != 8 {
		t.Fatalf("SupportedOperations() returned %d operations, want 8", len(ops))
	}

	// Verify actions
	expectedActions := map[string]bool{
		"shell":   true,
		"ansible": true,
		"script":  true,
		"helm":    true,
		"kubectl": true,
		"docker":  true,
	}

	// Verify queries
	expectedQueries := map[string]bool{
		"task_status": true,
		"list_tasks":  true,
	}

	actionCount := 0
	queryCount := 0

	for _, op := range ops {
		switch op.Type {
		case operator.Action:
			if !expectedActions[op.Name] {
				t.Errorf("unexpected action: %s", op.Name)
			}
			actionCount++
		case operator.Query:
			if !expectedQueries[op.Name] {
				t.Errorf("unexpected query: %s", op.Name)
			}
			queryCount++
		default:
			t.Errorf("unexpected operation type: %s", op.Type)
		}

		if op.Description == "" {
			t.Errorf("operation %s has empty description", op.Name)
		}
	}

	if actionCount != 6 {
		t.Errorf("got %d actions, want 6", actionCount)
	}
	if queryCount != 2 {
		t.Errorf("got %d queries, want 2", queryCount)
	}
}

func TestInvokeActions(t *testing.T) {
	op := New()
	ctx := context.Background()

	tests := []struct {
		name      string
		operation Op
	}{
		{"Shell", Shell},
		{"Ansible", Ansible},
		{"Script", Script},
		{"Helm", Helm},
		{"Kubectl", Kubectl},
		{"Docker", Docker},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := op.Invoke(ctx, operator.Action, string(tc.operation))
			if !errors.Is(err, operator.ErrNotImplemented) {
				t.Errorf("Invoke(%s) error = %v, want ErrNotImplemented", tc.operation, err)
			}
		})
	}
}

func TestInvokeQueries(t *testing.T) {
	op := New()
	ctx := context.Background()

	tests := []struct {
		name      string
		operation Op
	}{
		{"TaskStatusOp", TaskStatusOp},
		{"ListTasks", ListTasks},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := op.Invoke(ctx, operator.Query, string(tc.operation))
			if !errors.Is(err, operator.ErrNotImplemented) {
				t.Errorf("Invoke(%s) error = %v, want ErrNotImplemented", tc.operation, err)
			}
		})
	}
}

func TestInvokeUnknownAction(t *testing.T) {
	op := New()
	ctx := context.Background()

	_, err := op.Invoke(ctx, operator.Action, "unknown_action")
	if err == nil {
		t.Fatal("Invoke(unknown_action) should return error")
	}
	if !strings.Contains(err.Error(), "unknown action operation") {
		t.Errorf("error = %q, should contain 'unknown action operation'", err.Error())
	}
}

func TestInvokeUnknownQuery(t *testing.T) {
	op := New()
	ctx := context.Background()

	_, err := op.Invoke(ctx, operator.Query, "unknown_query")
	if err == nil {
		t.Fatal("Invoke(unknown_query) should return error")
	}
	if !strings.Contains(err.Error(), "unknown query operation") {
		t.Errorf("error = %q, should contain 'unknown query operation'", err.Error())
	}
}

func TestInvokeUnknownOperationType(t *testing.T) {
	op := New()
	ctx := context.Background()

	_, err := op.Invoke(ctx, operator.OperationType("invalid"), "shell")
	if err == nil {
		t.Fatal("Invoke(invalid type) should return error")
	}
	if !strings.Contains(err.Error(), "unknown operation type") {
		t.Errorf("error = %q, should contain 'unknown operation type'", err.Error())
	}
}

func TestInvokeActionAsQuery(t *testing.T) {
	op := New()
	ctx := context.Background()

	// Trying to invoke an action using Query type should fail
	_, err := op.Invoke(ctx, operator.Query, string(Shell))
	if err == nil {
		t.Fatal("Invoke(Query, shell) should return error")
	}
	if !strings.Contains(err.Error(), "unknown query operation") {
		t.Errorf("error = %q, should contain 'unknown query operation'", err.Error())
	}
}

func TestInvokeQueryAsAction(t *testing.T) {
	op := New()
	ctx := context.Background()

	// Trying to invoke a query using Action type should fail
	_, err := op.Invoke(ctx, operator.Action, string(TaskStatusOp))
	if err == nil {
		t.Fatal("Invoke(Action, task_status) should return error")
	}
	if !strings.Contains(err.Error(), "unknown action operation") {
		t.Errorf("error = %q, should contain 'unknown action operation'", err.Error())
	}
}
