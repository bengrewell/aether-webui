package mcp

import (
	"context"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bengrewell/aether-webui/internal/provider/onramp"
)

func (s *Server) registerMetaTools() {
	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "server_status",
		Description: "Get server version, provider statuses, and store health",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, _ ServerStatusInput) (*gomcp.CallToolResult, any, error) {
		type status struct {
			Version   any `json:"version,omitempty"`
			Providers any `json:"providers,omitempty"`
			Store     any `json:"store,omitempty"`
		}
		var result status

		if ver, err := s.meta.HandleVersion(ctx, nil); err == nil {
			result.Version = ver.Body
		}
		if prov, err := s.meta.HandleProviders(ctx, nil); err == nil {
			result.Providers = prov.Body
		}
		if st, err := s.meta.HandleStore(ctx, nil); err == nil {
			result.Store = st.Body
		}

		return jsonResult(result), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "component_states_list",
		Description: "List the current deployment state of all components",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, _ ComponentStatesListInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.onramp.HandleListComponentStates(ctx, nil)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "component_state_get",
		Description: "Get the current deployment state of a single component",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, args ComponentStateGetInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.onramp.HandleGetComponentState(ctx, &onramp.ComponentStateGetInput{Component: args.Component})
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})
}
