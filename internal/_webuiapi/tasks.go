package webuiapi

import (
	"context"
	"errors"

	"github.com/bengrewell/aether-webui/internal/onramp"
	"github.com/bengrewell/aether-webui/internal/state"
	"github.com/danielgtaylor/huma/v2"
)

// TaskListInput is the input for listing deployment tasks.
type TaskListInput struct {
	Limit  int `query:"limit" default:"20" doc:"Maximum number of tasks to return"`
	Offset int `query:"offset" default:"0" doc:"Offset for pagination"`
}

// TaskListOutput is the response for listing deployment tasks.
type TaskListOutput struct {
	Body struct {
		Tasks []state.DeploymentTask `json:"tasks"`
		Total int                    `json:"total"`
	}
}

// TaskPathInput includes path parameter for task ID.
type TaskPathInput struct {
	ID string `path:"id" doc:"Task identifier (UUID)"`
}

// TaskOutput is the response for a single task.
type TaskOutput struct {
	Body state.DeploymentTask
}

// ComponentDeploymentListOutput is the response for listing deployment states.
type ComponentDeploymentListOutput struct {
	Body struct {
		Deployments []state.ComponentDeploymentState `json:"deployments"`
	}
}

// DeploymentPathInput includes path parameter for component name.
type DeploymentPathInput struct {
	Component string `path:"component" doc:"Component name (e.g., 5gc, srsran-gnb, k8s)"`
}

// DeploymentOutput is the response for a single deployment state.
type DeploymentStateOutput struct {
	Body state.ComponentDeploymentState
}

// TaskCancelOutput is the response for cancelling a task.
type TaskCancelOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

// RegisterTaskRoutes registers task and deployment state API routes.
func RegisterTaskRoutes(api huma.API, taskMgr *onramp.TaskManager, store state.Store) {
	huma.Register(api, huma.Operation{
		OperationID: "list-tasks",
		Method:      "GET",
		Path:        "/api/v1/tasks",
		Summary:     "List deployment tasks",
		Description: "Returns paginated list of deployment tasks ordered by creation time",
		Tags:        []string{"Tasks"},
	}, func(ctx context.Context, input *TaskListInput) (*TaskListOutput, error) {
		tasks, total, err := store.ListTasks(ctx, input.Limit, input.Offset)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list tasks", err)
		}
		output := &TaskListOutput{}
		output.Body.Tasks = tasks
		if output.Body.Tasks == nil {
			output.Body.Tasks = []state.DeploymentTask{}
		}
		output.Body.Total = total
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-task",
		Method:      "GET",
		Path:        "/api/v1/tasks/{id}",
		Summary:     "Get task details",
		Description: "Returns task details including output and error information",
		Tags:        []string{"Tasks"},
	}, func(ctx context.Context, input *TaskPathInput) (*TaskOutput, error) {
		task, err := store.GetTask(ctx, input.ID)
		if err != nil {
			if errors.Is(err, state.ErrNotFound) {
				return nil, huma.Error404NotFound("task not found")
			}
			return nil, huma.Error500InternalServerError("failed to get task", err)
		}
		return &TaskOutput{Body: *task}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "cancel-task",
		Method:      "POST",
		Path:        "/api/v1/tasks/{id}/cancel",
		Summary:     "Cancel a running task",
		Description: "Cancels a currently running deployment task",
		Tags:        []string{"Tasks"},
	}, func(_ context.Context, input *TaskPathInput) (*TaskCancelOutput, error) {
		if err := taskMgr.CancelTask(input.ID); err != nil {
			return nil, huma.Error400BadRequest("failed to cancel task", err)
		}
		output := &TaskCancelOutput{}
		output.Body.Message = "task cancelled"
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-deployments",
		Method:      "GET",
		Path:        "/api/v1/deployments",
		Summary:     "List deployment states",
		Description: "Returns the deployment state for all tracked components",
		Tags:        []string{"Deployments"},
	}, func(ctx context.Context, _ *struct{}) (*ComponentDeploymentListOutput, error) {
		states, err := store.ListDeploymentStates(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list deployment states", err)
		}
		output := &ComponentDeploymentListOutput{}
		output.Body.Deployments = states
		if output.Body.Deployments == nil {
			output.Body.Deployments = []state.ComponentDeploymentState{}
		}
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-deployment",
		Method:      "GET",
		Path:        "/api/v1/deployments/{component}",
		Summary:     "Get deployment state",
		Description: "Returns the deployment state for a specific component",
		Tags:        []string{"Deployments"},
	}, func(ctx context.Context, input *DeploymentPathInput) (*DeploymentStateOutput, error) {
		ds, err := store.GetDeploymentState(ctx, input.Component)
		if err != nil {
			if errors.Is(err, state.ErrNotFound) {
				return nil, huma.Error404NotFound("deployment state not found")
			}
			return nil, huma.Error500InternalServerError("failed to get deployment state", err)
		}
		return &DeploymentStateOutput{Body: *ds}, nil
	})
}
