package webuiapi

import (
	"context"
	"time"

	"github.com/bengrewell/aether-webui/internal/state"
	"github.com/danielgtaylor/huma/v2"
)

// SetupStatusOutput is the response for GET /api/v1/setup/status
type SetupStatusOutput struct {
	Body struct {
		Completed   bool       `json:"completed" doc:"Whether the setup wizard has been completed"`
		CompletedAt *time.Time `json:"completed_at,omitempty" doc:"When the wizard was completed"`
		Steps       []string   `json:"steps,omitempty" doc:"List of completed setup steps"`
	}
}

// SetupCompleteInput is the request body for POST /api/v1/setup/complete
type SetupCompleteInput struct {
	Body struct {
		Steps []string `json:"steps,omitempty" doc:"Optional list of completed setup steps"`
	}
}

// SetupCompleteOutput is the response for POST /api/v1/setup/complete
type SetupCompleteOutput struct {
	Body struct {
		Success     bool      `json:"success" doc:"Whether the operation succeeded"`
		CompletedAt time.Time `json:"completed_at" doc:"When the wizard was marked complete"`
	}
}

// SetupResetOutput is the response for DELETE /api/v1/setup/status
type SetupResetOutput struct {
	Body struct {
		Success bool `json:"success" doc:"Whether the operation succeeded"`
	}
}

// RegisterSetupRoutes registers setup/wizard-related routes with the Huma API.
func RegisterSetupRoutes(api huma.API, store state.Store) {
	huma.Register(api, huma.Operation{
		OperationID: "get-setup-status",
		Method:      "GET",
		Path:        "/api/v1/setup/status",
		Summary:     "Get setup wizard status",
		Description: "Returns the completion status of the setup wizard",
		Tags:        []string{"Setup"},
	}, func(ctx context.Context, input *struct{}) (*SetupStatusOutput, error) {
		status, err := store.GetWizardStatus(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get wizard status", err)
		}
		resp := &SetupStatusOutput{}
		resp.Body.Completed = status.Completed
		resp.Body.CompletedAt = status.CompletedAt
		resp.Body.Steps = status.Steps
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "complete-setup",
		Method:      "POST",
		Path:        "/api/v1/setup/complete",
		Summary:     "Mark setup as complete",
		Description: "Marks the setup wizard as complete with optional step information",
		Tags:        []string{"Setup"},
	}, func(ctx context.Context, input *SetupCompleteInput) (*SetupCompleteOutput, error) {
		if err := store.SetWizardComplete(ctx, input.Body.Steps); err != nil {
			return nil, huma.Error500InternalServerError("failed to complete wizard", err)
		}
		resp := &SetupCompleteOutput{}
		resp.Body.Success = true
		resp.Body.CompletedAt = time.Now().UTC()
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "reset-setup-status",
		Method:      "DELETE",
		Path:        "/api/v1/setup/status",
		Summary:     "Reset setup wizard status",
		Description: "Clears the setup wizard completion status, allowing the wizard to be run again",
		Tags:        []string{"Setup"},
	}, func(ctx context.Context, input *struct{}) (*SetupResetOutput, error) {
		if err := store.ClearWizardStatus(ctx); err != nil {
			return nil, huma.Error500InternalServerError("failed to reset wizard status", err)
		}
		resp := &SetupResetOutput{}
		resp.Body.Success = true
		return resp, nil
	})
}
