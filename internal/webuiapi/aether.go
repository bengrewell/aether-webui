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

// CorePathInput includes path parameter for Core ID.
type CorePathInput struct {
	Host string `query:"host" default:"local" doc:"Target host identifier."`
	ID   string `path:"id" doc:"SD-Core identifier"`
}

// GNBPathInput includes path parameter for gNB ID.
type GNBPathInput struct {
	Host string `query:"host" default:"local" doc:"Target host identifier."`
	ID   string `path:"id" doc:"gNB identifier"`
}

// CoreListOutput is the response for GET /api/v1/aether/core
type CoreListOutput struct {
	Body aether.CoreList
}

// CoreConfigOutput is the response for GET /api/v1/aether/core/{id}
type CoreConfigOutput struct {
	Body aether.CoreConfig
}

// CoreDeployInput is the request body for POST /api/v1/aether/core
// All fields are optional - ID is always generated, Name uses default if not provided
type CoreDeployInput struct {
	Host string              `query:"host" default:"local" doc:"Target host identifier."`
	Body *aether.CoreConfig `doc:"Optional configuration. If not provided, defaults are used. ID field is ignored and always server-generated."`
}

// CoreUpdateInput is the request body for PUT /api/v1/aether/core/{id}
type CoreUpdateInput struct {
	Host string            `query:"host" default:"local" doc:"Target host identifier."`
	ID   string            `path:"id" doc:"SD-Core identifier"`
	Body aether.CoreConfig `doc:"Configuration to update. ID field in body is ignored."`
}

// CoreStatusOutput is the response for GET /api/v1/aether/core/{id}/status
type CoreStatusOutput struct {
	Body aether.CoreStatus
}

// CoreStatusListOutput is the response for GET /api/v1/aether/core/status
type CoreStatusListOutput struct {
	Body aether.CoreStatusList
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

// GNBDeployInput is the request body for POST /api/v1/aether/gnb
// All fields are optional - ID is always generated, Name uses default if not provided
type GNBDeployInput struct {
	Host string             `query:"host" default:"local" doc:"Target host identifier."`
	Body *aether.GNBConfig `doc:"Optional configuration. If not provided, defaults are used. ID field is ignored and always server-generated."`
}

// GNBUpdateInput is the request body for PUT /api/v1/aether/gnb/{id}
type GNBUpdateInput struct {
	Host string           `query:"host" default:"local" doc:"Target host identifier."`
	ID   string           `path:"id" doc:"gNB identifier"`
	Body aether.GNBConfig `doc:"Configuration to update. ID field in body is ignored."`
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

	// SD-Core endpoints
	huma.Register(api, huma.Operation{
		OperationID: "list-cores",
		Method:      "GET",
		Path:        "/api/v1/aether/core",
		Summary:     "List SD-Core deployments",
		Description: "Returns all SD-Core deployments on the specified host",
		Tags:        []string{"Aether", "Core"},
	}, func(ctx context.Context, input *HostInput) (*CoreListOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		list, err := provider.ListCores(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list cores", err)
		}
		return &CoreListOutput{Body: *list}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-core",
		Method:      "GET",
		Path:        "/api/v1/aether/core/{id}",
		Summary:     "Get SD-Core configuration",
		Description: "Returns the configuration for a specific SD-Core deployment",
		Tags:        []string{"Aether", "Core"},
	}, func(ctx context.Context, input *CorePathInput) (*CoreConfigOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		config, err := provider.GetCore(ctx, input.ID)
		if err != nil {
			return nil, huma.Error404NotFound("core not found", err)
		}
		return &CoreConfigOutput{Body: *config}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "deploy-core",
		Method:      "POST",
		Path:        "/api/v1/aether/core",
		Summary:     "Deploy SD-Core",
		Description: "Deploys a new SD-Core instance on the specified host. Configuration is optional - defaults are used if not provided. The ID is always server-generated and returned in the response.",
		Tags:        []string{"Aether", "Core"},
	}, func(ctx context.Context, input *CoreDeployInput) (*DeploymentOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		resp, err := provider.DeployCore(ctx, input.Body)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to deploy SD-Core", err)
		}
		return &DeploymentOutput{Body: *resp}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-core",
		Method:      "PUT",
		Path:        "/api/v1/aether/core/{id}",
		Summary:     "Update SD-Core configuration",
		Description: "Updates the configuration for a specific SD-Core deployment",
		Tags:        []string{"Aether", "Core"},
	}, func(ctx context.Context, input *CoreUpdateInput) (*CoreConfigOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		if err := provider.UpdateCore(ctx, input.ID, &input.Body); err != nil {
			return nil, huma.Error404NotFound("core not found", err)
		}
		input.Body.ID = input.ID
		return &CoreConfigOutput{Body: input.Body}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "undeploy-core",
		Method:      "DELETE",
		Path:        "/api/v1/aether/core/{id}",
		Summary:     "Undeploy SD-Core",
		Description: "Removes a specific SD-Core deployment from the specified host",
		Tags:        []string{"Aether", "Core"},
	}, func(ctx context.Context, input *CorePathInput) (*DeploymentOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		resp, err := provider.UndeployCore(ctx, input.ID)
		if err != nil {
			return nil, huma.Error404NotFound("core not found", err)
		}
		return &DeploymentOutput{Body: *resp}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-core-status",
		Method:      "GET",
		Path:        "/api/v1/aether/core/{id}/status",
		Summary:     "Get SD-Core status",
		Description: "Returns the deployment status of a specific SD-Core instance",
		Tags:        []string{"Aether", "Core"},
	}, func(ctx context.Context, input *CorePathInput) (*CoreStatusOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		status, err := provider.GetCoreStatus(ctx, input.ID)
		if err != nil {
			return nil, huma.Error404NotFound("core not found", err)
		}
		return &CoreStatusOutput{Body: *status}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-core-statuses",
		Method:      "GET",
		Path:        "/api/v1/aether/core/status",
		Summary:     "List all SD-Core statuses",
		Description: "Returns deployment status for all SD-Core instances on the specified host",
		Tags:        []string{"Aether", "Core"},
	}, func(ctx context.Context, input *HostInput) (*CoreStatusListOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		statuses, err := provider.ListCoreStatuses(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list core statuses", err)
		}
		return &CoreStatusListOutput{Body: *statuses}, nil
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
		OperationID: "deploy-gnb",
		Method:      "POST",
		Path:        "/api/v1/aether/gnb",
		Summary:     "Deploy gNB",
		Description: "Deploys a new gNB on the specified host. Configuration is optional - defaults are used if not provided. The ID is always server-generated and returned in the response.",
		Tags:        []string{"Aether", "gNB"},
	}, func(ctx context.Context, input *GNBDeployInput) (*DeploymentOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		resp, err := provider.DeployGNB(ctx, input.Body)
		if err != nil {
			return nil, huma.Error400BadRequest("failed to deploy gNB", err)
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
		OperationID: "undeploy-gnb",
		Method:      "DELETE",
		Path:        "/api/v1/aether/gnb/{id}",
		Summary:     "Undeploy gNB",
		Description: "Removes gNB deployment from the specified host",
		Tags:        []string{"Aether", "gNB"},
	}, func(ctx context.Context, input *GNBPathInput) (*DeploymentOutput, error) {
		provider, err := resolver.Resolve(input.Host)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid host", err)
		}
		resp, err := provider.UndeployGNB(ctx, input.ID)
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
