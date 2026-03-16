package mcp

import (
	"context"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bengrewell/aether-webui/internal/provider/nodes"
)

func (s *Server) registerNodeTools() {
	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "nodes_list",
		Description: "List all managed cluster nodes with roles and connection info",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, _ NodesListInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.nodes.HandleList(ctx, nil)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "nodes_get",
		Description: "Get a single managed node by ID",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, args NodesGetInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.nodes.HandleGet(ctx, &nodes.NodeGetInput{ID: args.ID})
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "nodes_create",
		Description: "Create a new managed cluster node with credentials and role assignments",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, args NodesCreateInput) (*gomcp.CallToolResult, any, error) {
		in := &nodes.NodeCreateInput{}
		in.Body.Name = args.Name
		in.Body.AnsibleHost = args.AnsibleHost
		in.Body.AnsibleUser = args.AnsibleUser
		in.Body.Password = args.Password
		in.Body.SudoPassword = args.SudoPassword
		in.Body.SSHKey = args.SSHKey
		in.Body.Roles = args.Roles
		out, err := s.nodes.HandleCreate(ctx, in)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "nodes_update",
		Description: "Partial update a managed node (only provided fields are changed)",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, args NodesUpdateInput) (*gomcp.CallToolResult, any, error) {
		in := &nodes.NodeUpdateInput{ID: args.ID}
		in.Body.Name = args.Name
		in.Body.AnsibleHost = args.AnsibleHost
		in.Body.AnsibleUser = args.AnsibleUser
		in.Body.Password = args.Password
		in.Body.SudoPassword = args.SudoPassword
		in.Body.SSHKey = args.SSHKey
		in.Body.Roles = args.Roles
		out, err := s.nodes.HandleUpdate(ctx, in)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "nodes_delete",
		Description: "Delete a managed node by ID",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, args NodesDeleteInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.nodes.HandleDelete(ctx, &nodes.NodeDeleteInput{ID: args.ID})
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})
}
