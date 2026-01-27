package webuiapi

import (
	"context"

	"github.com/bengrewell/aether-webui/internal/aether"
	"github.com/danielgtaylor/huma/v2"
)

// HostInput is the common input for endpoints that accept a host parameter.
type HostInput struct {
	Host string `query:"host" default:"local" doc:"Target host identifier. Use 'local' or empty for the local host."`
}

// GNBPathInput includes path parameter for gNB ID.
type GNBPathInput struct {
	Host string `query:"host" default:"local" doc:"Target host identifier."`
	ID   string `path:"id" doc:"gNB identifier"`
}

// CoreConfigOutput is the response for GET /api/v1/aether/core/config
type CoreConfigOutput struct {
	Body aether.CoreConfig
}

// CoreConfigInput is the request body for PUT /api/v1/aether/core/config
type CoreConfigInput struct {
	Host string `query:"host" default:"local" doc:"Target host identifier."`
	Body aether.CoreConfig
}

// CoreStatusOutput is the response for GET /api/v1/aether/core/status
type CoreStatusOutput struct {
	Body aether.CoreStatus
}

// DeploymentOutput is the response for deployment actions.
type DeploymentOutput struct {
	Body aether.DeploymentResponse
}

// GNBListOutput is the response for GET /api/v1/aether/gnb
type GNBListOutput struct {
	Body aether.GNBList
}

// GNBOutput is the response for GET /api/v1/aether/gnb/{id}
type GNBOutput struct {
	Body aether.GNBConfig
}

// GNBCreateInput is the request body for POST /api/v1/aether/gnb
type GNBCreateInput struct {
	Host string `query:"host" default:"local" doc:"Target host identifier."`
	Body aether.GNBConfig
}

// GNBUpdateInput is the request body for PUT /api/v1/aether/gnb/{id}
type GNBUpdateInput struct {
	Host string `query:"host" default:"local" doc:"Target host identifier."`
	ID   string `path:"id" doc:"gNB identifier"`
	Body aether.GNBConfig
}

// GNBStatusOutput is the response for GET /api/v1/aether/gnb/{id}/status
type GNBStatusOutput struct {
	Body aether.GNBStatus
}

// GNBStatusListOutput is the response for GET /api/v1/aether/gnb/status
type GNBStatusListOutput struct {
	Body aether.GNBStatusList
}

// HostListOutput is the response for GET /api/v1/aether/hosts
type HostListOutput struct {
	Body struct {
		Hosts []string `json:"hosts"`
	}
}

// RegisterAetherRoutes registers Aether 5G management routes.
func RegisterAetherRoutes(api huma.API, resolver aether.HostResolver) {
	// Host management
	huma.Register(api, huma.Operation{
		OperationID: "list-hosts",
		Method:      "GET",
		Path:        "/api/v1/aether/hosts",
		Summary:     "List configured hosts",
		Description: "Returns list of all configured hosts for Aether deployments",
		Tags:        []string{"Aether"},
	}, func(ctx context.Context, input *struct{}) (*HostListOutput, error) {
		output := &HostListOutput{}
		output.Body.Hosts = resolver.ListHosts()
		return output, nil
	})

	// Core configuration endpoints
	huma.Register(api, huma.Operation{
		OperationID: "get-core-config",
		Method:      "GET",
		Path:        "/api/v1/aether/core/config",
		Summary:     "Get SD-Core configuration",
		Description: "Returns the current SD-Core configuration for the specified host",
		Tags:        []string{"Aether", "Core"},
	}, func(ctx context.Context, input *HostInput) (*CoreConfigOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		config, err := provider.GetCoreConfig(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get core config", err)
		}
		return &CoreConfigOutput{Body: *config}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-core-config",
		Method:      "PUT",
		Path:        "/api/v1/aether/core/config",
		Summary:     "Update SD-Core configuration",
		Description: "Updates the SD-Core configuration for the specified host",
		Tags:        []string{"Aether", "Core"},
	}, func(ctx context.Context, input *CoreConfigInput) (*CoreConfigOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		if err := provider.UpdateCoreConfig(ctx, &input.Body); err != nil {
			return nil, huma.Error500InternalServerError("failed to update core config", err)
		}
		return &CoreConfigOutput{Body: input.Body}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-core-status",
		Method:      "GET",
		Path:        "/api/v1/aether/core/status",
		Summary:     "Get SD-Core status",
		Description: "Returns the deployment status of SD-Core on the specified host",
		Tags:        []string{"Aether", "Core"},
	}, func(ctx context.Context, input *HostInput) (*CoreStatusOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		status, err := provider.GetCoreStatus(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get core status", err)
		}
		return &CoreStatusOutput{Body: *status}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "deploy-core",
		Method:      "POST",
		Path:        "/api/v1/aether/core/deploy",
		Summary:     "Deploy SD-Core",
		Description: "Initiates deployment of SD-Core on the specified host",
		Tags:        []string{"Aether", "Core"},
	}, func(ctx context.Context, input *HostInput) (*DeploymentOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		resp, err := provider.DeployCore(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to deploy core", err)
		}
		return &DeploymentOutput{Body: *resp}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "undeploy-core",
		Method:      "DELETE",
		Path:        "/api/v1/aether/core",
		Summary:     "Undeploy SD-Core",
		Description: "Initiates removal of SD-Core from the specified host",
		Tags:        []string{"Aether", "Core"},
	}, func(ctx context.Context, input *HostInput) (*DeploymentOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		resp, err := provider.UndeployCore(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to undeploy core", err)
		}
		return &DeploymentOutput{Body: *resp}, nil
	})

	// gNB endpoints
	huma.Register(api, huma.Operation{
		OperationID: "list-gnbs",
		Method:      "GET",
		Path:        "/api/v1/aether/gnb",
		Summary:     "List gNBs",
		Description: "Returns all configured gNBs on the specified host",
		Tags:        []string{"Aether", "gNB"},
	}, func(ctx context.Context, input *HostInput) (*GNBListOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		list, err := provider.ListGNBs(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list gNBs", err)
		}
		return &GNBListOutput{Body: *list}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-gnb",
		Method:      "GET",
		Path:        "/api/v1/aether/gnb/{id}",
		Summary:     "Get gNB configuration",
		Description: "Returns configuration for a specific gNB",
		Tags:        []string{"Aether", "gNB"},
	}, func(ctx context.Context, input *GNBPathInput) (*GNBOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		config, err := provider.GetGNB(ctx, input.ID)
		if err != nil {
			return nil, huma.Error404NotFound("gNB not found", err)
		}
		return &GNBOutput{Body: *config}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-gnb",
		Method:      "POST",
		Path:        "/api/v1/aether/gnb",
		Summary:     "Create and deploy gNB",
		Description: "Creates and deploys a new gNB on the specified host",
		Tags:        []string{"Aether", "gNB"},
	}, func(ctx context.Context, input *GNBCreateInput) (*DeploymentOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		resp, err := provider.CreateGNB(ctx, &input.Body)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to create gNB", err)
		}
		return &DeploymentOutput{Body: *resp}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-gnb",
		Method:      "PUT",
		Path:        "/api/v1/aether/gnb/{id}",
		Summary:     "Update gNB configuration",
		Description: "Updates configuration for a specific gNB",
		Tags:        []string{"Aether", "gNB"},
	}, func(ctx context.Context, input *GNBUpdateInput) (*GNBOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		if err := provider.UpdateGNB(ctx, input.ID, &input.Body); err != nil {
			return nil, huma.Error404NotFound("gNB not found", err)
		}
		input.Body.ID = input.ID
		return &GNBOutput{Body: input.Body}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-gnb",
		Method:      "DELETE",
		Path:        "/api/v1/aether/gnb/{id}",
		Summary:     "Delete gNB",
		Description: "Removes a gNB from the specified host",
		Tags:        []string{"Aether", "gNB"},
	}, func(ctx context.Context, input *GNBPathInput) (*DeploymentOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		resp, err := provider.DeleteGNB(ctx, input.ID)
		if err != nil {
			return nil, huma.Error404NotFound("gNB not found", err)
		}
		return &DeploymentOutput{Body: *resp}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-gnb-status",
		Method:      "GET",
		Path:        "/api/v1/aether/gnb/{id}/status",
		Summary:     "Get gNB status",
		Description: "Returns deployment status for a specific gNB",
		Tags:        []string{"Aether", "gNB"},
	}, func(ctx context.Context, input *GNBPathInput) (*GNBStatusOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		status, err := provider.GetGNBStatus(ctx, input.ID)
		if err != nil {
			return nil, huma.Error404NotFound("gNB not found", err)
		}
		return &GNBStatusOutput{Body: *status}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-gnb-statuses",
		Method:      "GET",
		Path:        "/api/v1/aether/gnb/status",
		Summary:     "List all gNB statuses",
		Description: "Returns deployment status for all gNBs on the specified host",
		Tags:        []string{"Aether", "gNB"},
	}, func(ctx context.Context, input *HostInput) (*GNBStatusListOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		statuses, err := provider.ListGNBStatuses(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list gNB statuses", err)
		}
		return &GNBStatusListOutput{Body: *statuses}, nil
	})
}
