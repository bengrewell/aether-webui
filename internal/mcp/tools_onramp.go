package mcp

import (
	"context"
	"encoding/json"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bengrewell/aether-webui/internal/provider/onramp"
)

func (s *Server) registerOnRampTools() {
	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "components_list",
		Description: "List all deployable OnRamp components and their available actions",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, _ ComponentsListInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.onramp.HandleListComponents(ctx, nil)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "component_get",
		Description: "Get a single OnRamp component's details and available actions",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, args ComponentGetInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.onramp.HandleGetComponent(ctx, &onramp.ComponentGetInput{Component: args.Component})
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "deploy_action",
		Description: "Execute a deployment action on a component (async, returns task ID for monitoring)",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, args DeployActionInput) (*gomcp.CallToolResult, any, error) {
		in := &onramp.ExecuteActionInput{
			Component: args.Component,
			Action:    args.Action,
		}
		if len(args.Labels) > 0 || len(args.Tags) > 0 {
			in.Body = &onramp.ExecuteActionBody{
				Labels: args.Labels,
				Tags:   args.Tags,
			}
		}
		out, err := s.onramp.HandleExecuteAction(ctx, in)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "repo_status",
		Description: "Get the OnRamp git repository clone status, branch, commit, and dirty state",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, _ RepoStatusInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.onramp.HandleGetRepoStatus(ctx, nil)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "repo_refresh",
		Description: "Clone the OnRamp repository if missing, or validate and refresh it",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, _ RepoRefreshInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.onramp.HandleRefreshRepo(ctx, nil)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "config_get",
		Description: "Get the current OnRamp configuration (vars/main.yml)",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, _ ConfigGetInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.onramp.HandleGetConfig(ctx, nil)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "config_patch",
		Description: "Patch the OnRamp configuration by deep-merging provided fields into vars/main.yml",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, args ConfigPatchInput) (*gomcp.CallToolResult, any, error) {
		rawBody, err := json.Marshal(args.Config)
		if err != nil {
			return errorResult(err), nil, nil
		}
		in := &onramp.ConfigPatchInput{RawBody: rawBody}
		out, err := s.onramp.HandlePatchConfig(ctx, in)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "profiles_list",
		Description: "List available OnRamp configuration profiles (main-*.yml files)",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, _ ProfilesListInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.onramp.HandleListProfiles(ctx, nil)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})
}
