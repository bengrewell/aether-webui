package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bengrewell/aether-webui/internal/provider/onramp"
	"github.com/bengrewell/aether-webui/internal/taskrunner"
)

func (s *Server) registerTaskTools() {
	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "tasks_list",
		Description: "List all active and recent tasks (make target executions)",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, _ TasksListInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.onramp.HandleListTasks(ctx, nil)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "task_get",
		Description: "Get task details and output (supports incremental reads via offset)",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, args TaskGetInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.onramp.HandleGetTask(ctx, &onramp.TaskGetInput{
			ID:     args.ID,
			Offset: args.Offset,
		})
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "task_cancel",
		Description: "Cancel a pending or running task",
	}, func(_ context.Context, _ *gomcp.CallToolRequest, args TaskCancelInput) (*gomcp.CallToolResult, any, error) {
		runner := s.onramp.Runner()
		err := runner.Cancel(args.ID)
		if err != nil {
			if err == taskrunner.ErrNotFound {
				return errorResult(fmt.Errorf("task not found: %s", args.ID)), nil, nil
			}
			if err == taskrunner.ErrNotRunning {
				return errorResult(fmt.Errorf("task is not running: %s", args.ID)), nil, nil
			}
			return errorResult(err), nil, nil
		}
		return jsonResult(map[string]string{"message": fmt.Sprintf("task %s canceled", args.ID)}), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "actions_list",
		Description: "Query action execution history with optional filters for component, action, and status",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, args ActionsListInput) (*gomcp.CallToolResult, any, error) {
		limit := args.Limit
		if limit == 0 {
			limit = 50
		}
		out, err := s.onramp.HandleListActions(ctx, &onramp.ActionListInput{
			Component: args.Component,
			Action:    args.Action,
			Status:    args.Status,
			Limit:     limit,
			Offset:    args.Offset,
		})
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "action_get",
		Description: "Get a single action execution record by ID",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, args ActionGetInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.onramp.HandleGetAction(ctx, &onramp.ActionGetInput{ID: args.ID})
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})
}
