package webuiapi

import (
	"context"
	"time"

	"github.com/bengrewell/aether-webui/internal/state"
	"github.com/danielgtaylor/huma/v2"
)

type OperationsLogInput struct {
	Limit  int    `query:"limit" doc:"Maximum entries to return" default:"50" minimum:"1" maximum:"500"`
	Offset int    `query:"offset" doc:"Number of entries to skip" default:"0" minimum:"0"`
	NodeID string `query:"node_id" doc:"Filter by node ID" default:""`
}

type OperationLogItem struct {
	ID        int    `json:"id"`
	Operation string `json:"operation"`
	NodeID    string `json:"node_id,omitempty"`
	Detail    string `json:"detail,omitempty"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at"`
}

type OperationsLogOutput struct {
	Body struct {
		Operations []OperationLogItem `json:"operations"`
		Total      int                `json:"total"`
		Limit      int                `json:"limit"`
		Offset     int                `json:"offset"`
	}
}

// RegisterOperationsRoutes registers operations log routes.
func RegisterOperationsRoutes(api huma.API, store state.Store) {
	huma.Register(api, huma.Operation{
		OperationID: "list-operations",
		Method:      "GET",
		Path:        "/api/v1/operations",
		Summary:     "List operations log",
		Description: "Returns a paginated operations audit log, optionally filtered by node",
		Tags:        []string{"Operations"},
	}, func(ctx context.Context, input *OperationsLogInput) (*OperationsLogOutput, error) {
		limit := input.Limit
		if limit == 0 {
			limit = 50
		}

		var entries []state.OperationLog
		var total int
		var err error

		if input.NodeID != "" {
			entries, total, err = store.GetOperationsLogByNode(ctx, input.NodeID, limit, input.Offset)
		} else {
			entries, total, err = store.GetOperationsLog(ctx, limit, input.Offset)
		}
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get operations log", err)
		}

		items := make([]OperationLogItem, len(entries))
		for i, e := range entries {
			items[i] = OperationLogItem{
				ID:        e.ID,
				Operation: e.Operation,
				NodeID:    e.NodeID,
				Detail:    e.Detail,
				Status:    e.Status,
				Error:     e.Error,
				CreatedAt: e.CreatedAt.Format(time.RFC3339),
			}
		}

		resp := &OperationsLogOutput{}
		resp.Body.Operations = items
		resp.Body.Total = total
		resp.Body.Limit = limit
		resp.Body.Offset = input.Offset
		return resp, nil
	})
}
