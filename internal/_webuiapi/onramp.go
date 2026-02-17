package webuiapi

import (
	"context"

	"github.com/bengrewell/aether-webui/internal/onramp"
	"github.com/danielgtaylor/huma/v2"
)

// OnRampStatusOutput is the response for GET /api/v1/onramp/status.
type OnRampStatusOutput struct {
	Body struct {
		Ready    bool   `json:"ready"`
		RepoPath string `json:"repo_path"`
	}
}

// OnRampSetupOutput is the response for POST /api/v1/onramp/setup.
type OnRampSetupOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

// BlueprintListOutput is the response for GET /api/v1/onramp/blueprints.
type BlueprintListOutput struct {
	Body struct {
		Blueprints []string `json:"blueprints"`
	}
}

// BlueprintActivateInput includes path parameter for blueprint name.
type BlueprintActivateInput struct {
	Name string `path:"name" doc:"Blueprint name (e.g., srsran, gnbsim, quickstart)"`
}

// BlueprintActivateOutput is the response for POST /api/v1/onramp/blueprints/{name}/activate.
type BlueprintActivateOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

// InventoryGenerateOutput is the response for POST /api/v1/onramp/inventory/generate.
type InventoryGenerateOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

// PingOutput is the response for POST /api/v1/onramp/ping.
type PingOutput struct {
	Body struct {
		TaskID  string `json:"task_id"`
		Message string `json:"message"`
	}
}

// RegisterOnRampRoutes registers OnRamp management routes.
func RegisterOnRampRoutes(api huma.API, mgr *onramp.Manager, taskMgr *onramp.TaskManager) {
	huma.Register(api, huma.Operation{
		OperationID: "onramp-status",
		Method:      "GET",
		Path:        "/api/v1/onramp/status",
		Summary:     "Check OnRamp status",
		Description: "Returns whether the OnRamp repository has been cloned and is ready",
		Tags:        []string{"OnRamp"},
	}, func(_ context.Context, _ *struct{}) (*OnRampStatusOutput, error) {
		output := &OnRampStatusOutput{}
		output.Body.Ready = mgr.IsRepoReady()
		output.Body.RepoPath = mgr.RepoPath()
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "onramp-setup",
		Method:      "POST",
		Path:        "/api/v1/onramp/setup",
		Summary:     "Clone or update OnRamp repository",
		Description: "Clones the OnRamp repository if it doesn't exist, or pulls updates",
		Tags:        []string{"OnRamp"},
	}, func(ctx context.Context, _ *struct{}) (*OnRampSetupOutput, error) {
		if err := mgr.EnsureRepo(ctx); err != nil {
			return nil, huma.Error500InternalServerError("failed to setup OnRamp", err)
		}
		output := &OnRampSetupOutput{}
		output.Body.Message = "OnRamp repository ready"
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "onramp-ping",
		Method:      "POST",
		Path:        "/api/v1/onramp/ping",
		Summary:     "Ping all hosts",
		Description: "Runs the aether-pingall playbook to verify Ansible connectivity to all configured hosts",
		Tags:        []string{"OnRamp"},
	}, func(ctx context.Context, _ *struct{}) (*PingOutput, error) {
		seq, ok := onramp.Sequences["aether-pingall"]
		if !ok {
			return nil, huma.Error500InternalServerError("pingall sequence not defined")
		}
		taskID, err := taskMgr.StartSequence(ctx, seq, "")
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to start ping", err)
		}
		output := &PingOutput{}
		output.Body.TaskID = taskID
		output.Body.Message = "ping started"
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-blueprints",
		Method:      "GET",
		Path:        "/api/v1/onramp/blueprints",
		Summary:     "List available blueprints",
		Description: "Returns the list of available OnRamp deployment blueprint files",
		Tags:        []string{"OnRamp"},
	}, func(_ context.Context, _ *struct{}) (*BlueprintListOutput, error) {
		blueprints, err := mgr.ListBlueprints()
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list blueprints", err)
		}
		output := &BlueprintListOutput{}
		output.Body.Blueprints = blueprints
		if output.Body.Blueprints == nil {
			output.Body.Blueprints = []string{}
		}
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "activate-blueprint",
		Method:      "POST",
		Path:        "/api/v1/onramp/blueprints/{name}/activate",
		Summary:     "Activate a blueprint",
		Description: "Copies the specified blueprint to vars/main.yml as the active configuration",
		Tags:        []string{"OnRamp"},
	}, func(ctx context.Context, input *BlueprintActivateInput) (*BlueprintActivateOutput, error) {
		if err := mgr.GenerateVarsFile(ctx, input.Name); err != nil {
			return nil, huma.Error400BadRequest("failed to activate blueprint", err)
		}
		output := &BlueprintActivateOutput{}
		output.Body.Message = "blueprint " + input.Name + " activated"
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "generate-inventory",
		Method:      "POST",
		Path:        "/api/v1/onramp/inventory/generate",
		Summary:     "Generate inventory file",
		Description: "Regenerates the hosts.ini inventory file from the database node state",
		Tags:        []string{"OnRamp"},
	}, func(ctx context.Context, _ *struct{}) (*InventoryGenerateOutput, error) {
		if err := mgr.GenerateHostsINI(ctx); err != nil {
			return nil, huma.Error500InternalServerError("failed to generate inventory", err)
		}
		output := &InventoryGenerateOutput{}
		output.Body.Message = "inventory generated"
		return output, nil
	})
}
