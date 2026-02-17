package webuiapi

import (
	"context"
	"fmt"
	"time"

	"github.com/bengrewell/aether-webui/internal/crypto"
	"github.com/bengrewell/aether-webui/internal/ssh"
	"github.com/bengrewell/aether-webui/internal/state"
	"github.com/danielgtaylor/huma/v2"
)

// NodeRoutesDeps holds dependencies for node API routes.
type NodeRoutesDeps struct {
	Store         state.Store
	EncryptionKey string
}

// --- Input/Output types ---

type NodeOutput struct {
	Body struct {
		ID             string   `json:"id" doc:"Unique node identifier"`
		Name           string   `json:"name" doc:"Human-readable node name"`
		NodeType       string   `json:"node_type" doc:"Node type: local or remote"`
		Address        string   `json:"address,omitempty" doc:"SSH host address"`
		SSHPort        int      `json:"ssh_port,omitempty" doc:"SSH port number"`
		Username       string   `json:"username,omitempty" doc:"SSH username"`
		AuthMethod     string   `json:"auth_method,omitempty" doc:"Authentication method: password or private_key"`
		PrivateKeyPath string   `json:"private_key_path,omitempty" doc:"Path to SSH private key"`
		Roles          []string `json:"roles" doc:"Assigned roles/components"`
		CreatedAt      string   `json:"created_at" doc:"Creation timestamp"`
		UpdatedAt      string   `json:"updated_at" doc:"Last update timestamp"`
	}
}

type ManagedNodeListOutput struct {
	Body []NodeListItem `json:"body"`
}

type NodeListItem struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	NodeType       string   `json:"node_type"`
	Address        string   `json:"address,omitempty"`
	SSHPort        int      `json:"ssh_port,omitempty"`
	Username       string   `json:"username,omitempty"`
	AuthMethod     string   `json:"auth_method,omitempty"`
	PrivateKeyPath string   `json:"private_key_path,omitempty"`
	Roles          []string `json:"roles"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

type CreateNodeInput struct {
	Body struct {
		ID             string `json:"id" doc:"Unique node identifier" minLength:"1"`
		Name           string `json:"name" doc:"Human-readable name" minLength:"1"`
		Address        string `json:"address" doc:"SSH host address" minLength:"1"`
		SSHPort        int    `json:"ssh_port,omitempty" doc:"SSH port (default 22)"`
		Username       string `json:"username" doc:"SSH username" minLength:"1"`
		AuthMethod     string `json:"auth_method" doc:"Authentication method: password or private_key" enum:"password,private_key"`
		Password       string `json:"password,omitempty" doc:"SSH password (only for auth_method=password)"`
		PrivateKeyPath string `json:"private_key_path,omitempty" doc:"Path to private key (only for auth_method=private_key)"`
	}
}

type CreateNodeOutput struct {
	Body struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
	}
}

type NodePathInput struct {
	ID string `path:"id" doc:"Node ID"`
}

type UpdateNodeInput struct {
	ID   string `path:"id" doc:"Node ID"`
	Body struct {
		Name           string `json:"name,omitempty" doc:"Human-readable name"`
		Address        string `json:"address,omitempty" doc:"SSH host address"`
		SSHPort        int    `json:"ssh_port,omitempty" doc:"SSH port"`
		Username       string `json:"username,omitempty" doc:"SSH username"`
		AuthMethod     string `json:"auth_method,omitempty" doc:"Authentication method" enum:"password,private_key,"`
		Password       string `json:"password,omitempty" doc:"SSH password (cleared if empty)"`
		PrivateKeyPath string `json:"private_key_path,omitempty" doc:"Path to private key"`
	}
}

type DeleteNodeOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

type ConnectivityTestInput struct {
	Body struct {
		Address        string `json:"address" doc:"SSH host address" minLength:"1"`
		Port           int    `json:"port,omitempty" doc:"SSH port (default 22)"`
		Username       string `json:"username" doc:"SSH username" minLength:"1"`
		Password       string `json:"password,omitempty" doc:"SSH password"`
		PrivateKeyPath string `json:"private_key_path,omitempty" doc:"Path to private key"`
	}
}

type ConnectivityTestOutput struct {
	Body struct {
		Reachable     bool   `json:"reachable"`
		Authenticated bool   `json:"authenticated"`
		LatencyMs     int64  `json:"latency_ms"`
		Error         string `json:"error,omitempty"`
		ServerVersion string `json:"server_version,omitempty"`
	}
}

type AssignRoleInput struct {
	ID   string `path:"id" doc:"Node ID"`
	Body struct {
		Role string `json:"role" doc:"Role to assign" minLength:"1"`
	}
}

type AssignRoleOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

type RemoveRoleInput struct {
	ID   string `path:"id" doc:"Node ID"`
	Role string `path:"role" doc:"Role to remove"`
}

type RemoveRoleOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// --- Helper functions ---

func nodeToListItem(n state.Node) NodeListItem {
	roles := n.Roles
	if roles == nil {
		roles = []string{}
	}
	return NodeListItem{
		ID:             n.ID,
		Name:           n.Name,
		NodeType:       n.NodeType,
		Address:        n.Address,
		SSHPort:        n.SSHPort,
		Username:       n.Username,
		AuthMethod:     n.AuthMethod,
		PrivateKeyPath: n.PrivateKeyPath,
		Roles:          roles,
		CreatedAt:      n.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      n.UpdatedAt.Format(time.RFC3339),
	}
}

func nodeToOutput(n *state.Node) *NodeOutput {
	roles := n.Roles
	if roles == nil {
		roles = []string{}
	}
	resp := &NodeOutput{}
	resp.Body.ID = n.ID
	resp.Body.Name = n.Name
	resp.Body.NodeType = n.NodeType
	resp.Body.Address = n.Address
	resp.Body.SSHPort = n.SSHPort
	resp.Body.Username = n.Username
	resp.Body.AuthMethod = n.AuthMethod
	resp.Body.PrivateKeyPath = n.PrivateKeyPath
	resp.Body.Roles = roles
	resp.Body.CreatedAt = n.CreatedAt.Format(time.RFC3339)
	resp.Body.UpdatedAt = n.UpdatedAt.Format(time.RFC3339)
	return resp
}

func (d *NodeRoutesDeps) logOp(ctx context.Context, op, nodeID, detail, status, errMsg string) {
	_ = d.Store.LogOperation(ctx, &state.OperationLog{
		Operation: op,
		NodeID:    nodeID,
		Detail:    detail,
		Status:    status,
		Error:     errMsg,
	})
}

// RegisterNodeRoutes registers all node management routes.
func RegisterNodeRoutes(api huma.API, deps NodeRoutesDeps) {
	// List all nodes
	huma.Register(api, huma.Operation{
		OperationID: "list-all-nodes",
		Method:      "GET",
		Path:        "/api/v1/nodes",
		Summary:     "List all nodes",
		Description: "Returns all configured nodes with their assigned roles",
		Tags:        []string{"Nodes"},
	}, func(ctx context.Context, _ *struct{}) (*ManagedNodeListOutput, error) {
		nodes, err := deps.Store.ListNodes(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list nodes", err)
		}
		items := make([]NodeListItem, len(nodes))
		for i, n := range nodes {
			items[i] = nodeToListItem(n)
		}
		return &ManagedNodeListOutput{Body: items}, nil
	})

	// Create a remote node
	huma.Register(api, huma.Operation{
		OperationID: "create-node",
		Method:      "POST",
		Path:        "/api/v1/nodes",
		Summary:     "Create a node",
		Description: "Add a new remote node to the cluster configuration",
		Tags:        []string{"Nodes"},
	}, func(ctx context.Context, input *CreateNodeInput) (*CreateNodeOutput, error) {
		port := input.Body.SSHPort
		if port == 0 {
			port = 22
		}

		node := &state.Node{
			ID:             input.Body.ID,
			Name:           input.Body.Name,
			NodeType:       state.NodeTypeRemote,
			Address:        input.Body.Address,
			SSHPort:        port,
			Username:       input.Body.Username,
			AuthMethod:     input.Body.AuthMethod,
			PrivateKeyPath: input.Body.PrivateKeyPath,
		}

		// Encrypt password if provided
		if input.Body.Password != "" && deps.EncryptionKey != "" {
			encrypted, err := crypto.Encrypt(input.Body.Password, deps.EncryptionKey)
			if err != nil {
				deps.logOp(ctx, state.OpCreateNode, input.Body.ID, "", state.OpStatusFailure, err.Error())
				return nil, huma.Error500InternalServerError("failed to encrypt password", err)
			}
			node.EncryptedPassword = encrypted
		}

		if err := deps.Store.CreateNode(ctx, node); err != nil {
			deps.logOp(ctx, state.OpCreateNode, input.Body.ID, "", state.OpStatusFailure, err.Error())
			return nil, huma.Error422UnprocessableEntity("failed to create node", err)
		}

		deps.logOp(ctx, state.OpCreateNode, node.ID, fmt.Sprintf("name=%s address=%s", node.Name, node.Address), state.OpStatusSuccess, "")

		resp := &CreateNodeOutput{}
		resp.Body.ID = node.ID
		resp.Body.Name = node.Name
		resp.Body.CreatedAt = node.CreatedAt.Format(time.RFC3339)
		return resp, nil
	})

	// Get a single node
	huma.Register(api, huma.Operation{
		OperationID: "get-node",
		Method:      "GET",
		Path:        "/api/v1/nodes/{id}",
		Summary:     "Get node details",
		Description: "Returns details of a single node including assigned roles",
		Tags:        []string{"Nodes"},
	}, func(ctx context.Context, input *NodePathInput) (*NodeOutput, error) {
		node, err := deps.Store.GetNode(ctx, input.ID)
		if err != nil {
			if err == state.ErrNotFound {
				return nil, huma.Error404NotFound("node not found")
			}
			return nil, huma.Error500InternalServerError("failed to get node", err)
		}
		return nodeToOutput(node), nil
	})

	// Update a node
	huma.Register(api, huma.Operation{
		OperationID: "update-node",
		Method:      "PUT",
		Path:        "/api/v1/nodes/{id}",
		Summary:     "Update a node",
		Description: "Update properties of an existing node",
		Tags:        []string{"Nodes"},
	}, func(ctx context.Context, input *UpdateNodeInput) (*NodeOutput, error) {
		existing, err := deps.Store.GetNode(ctx, input.ID)
		if err != nil {
			if err == state.ErrNotFound {
				return nil, huma.Error404NotFound("node not found")
			}
			return nil, huma.Error500InternalServerError("failed to get node", err)
		}

		// Apply updates (only non-zero values)
		if input.Body.Name != "" {
			existing.Name = input.Body.Name
		}
		if input.Body.Address != "" {
			existing.Address = input.Body.Address
		}
		if input.Body.SSHPort != 0 {
			existing.SSHPort = input.Body.SSHPort
		}
		if input.Body.Username != "" {
			existing.Username = input.Body.Username
		}
		if input.Body.AuthMethod != "" {
			existing.AuthMethod = input.Body.AuthMethod
		}
		if input.Body.PrivateKeyPath != "" {
			existing.PrivateKeyPath = input.Body.PrivateKeyPath
		}

		// Handle password update
		if input.Body.Password != "" && deps.EncryptionKey != "" {
			encrypted, err := crypto.Encrypt(input.Body.Password, deps.EncryptionKey)
			if err != nil {
				deps.logOp(ctx, state.OpUpdateNode, input.ID, "", state.OpStatusFailure, err.Error())
				return nil, huma.Error500InternalServerError("failed to encrypt password", err)
			}
			existing.EncryptedPassword = encrypted
		}

		if err := deps.Store.UpdateNode(ctx, existing); err != nil {
			deps.logOp(ctx, state.OpUpdateNode, input.ID, "", state.OpStatusFailure, err.Error())
			return nil, huma.Error500InternalServerError("failed to update node", err)
		}

		deps.logOp(ctx, state.OpUpdateNode, input.ID, fmt.Sprintf("name=%s", existing.Name), state.OpStatusSuccess, "")

		// Re-fetch to get updated roles
		updated, err := deps.Store.GetNode(ctx, input.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get updated node", err)
		}
		return nodeToOutput(updated), nil
	})

	// Delete a node
	huma.Register(api, huma.Operation{
		OperationID: "delete-node",
		Method:      "DELETE",
		Path:        "/api/v1/nodes/{id}",
		Summary:     "Delete a node",
		Description: "Remove a remote node from the cluster configuration. Cannot delete the local node.",
		Tags:        []string{"Nodes"},
	}, func(ctx context.Context, input *NodePathInput) (*DeleteNodeOutput, error) {
		if err := deps.Store.DeleteNode(ctx, input.ID); err != nil {
			if err == state.ErrNotFound {
				return nil, huma.Error404NotFound("node not found")
			}
			if err == state.ErrLocalNodeDelete {
				deps.logOp(ctx, state.OpDeleteNode, input.ID, "", state.OpStatusFailure, err.Error())
				return nil, huma.Error422UnprocessableEntity("cannot delete the local node")
			}
			deps.logOp(ctx, state.OpDeleteNode, input.ID, "", state.OpStatusFailure, err.Error())
			return nil, huma.Error500InternalServerError("failed to delete node", err)
		}

		deps.logOp(ctx, state.OpDeleteNode, input.ID, "", state.OpStatusSuccess, "")

		resp := &DeleteNodeOutput{}
		resp.Body.Success = true
		return resp, nil
	})

	// Test connectivity (standalone, no node required)
	huma.Register(api, huma.Operation{
		OperationID: "test-node-connectivity",
		Method:      "POST",
		Path:        "/api/v1/nodes/test-connectivity",
		Summary:     "Test SSH connectivity",
		Description: "Test SSH connectivity to a remote host without creating a node",
		Tags:        []string{"Nodes"},
	}, func(ctx context.Context, input *ConnectivityTestInput) (*ConnectivityTestOutput, error) {
		port := input.Body.Port
		if port == 0 {
			port = 22
		}

		password := input.Body.Password
		// Decrypt password if it looks encrypted and we have a key
		// (for standalone test, password comes in plaintext from the client)

		result := ssh.TestConnectivity(ctx, ssh.ConnectivityTestRequest{
			Address:        input.Body.Address,
			Port:           port,
			Username:       input.Body.Username,
			Password:       password,
			PrivateKeyPath: input.Body.PrivateKeyPath,
		})

		opStatus := state.OpStatusSuccess
		opErr := ""
		if result.Error != "" {
			opStatus = state.OpStatusFailure
			opErr = result.Error
		}
		deps.logOp(ctx, state.OpTestConnectivity, "", fmt.Sprintf("address=%s:%d", input.Body.Address, port), opStatus, opErr)

		resp := &ConnectivityTestOutput{}
		resp.Body.Reachable = result.Reachable
		resp.Body.Authenticated = result.Authenticated
		resp.Body.LatencyMs = result.LatencyMs
		resp.Body.Error = result.Error
		resp.Body.ServerVersion = result.ServerVersion
		return resp, nil
	})

	// Test connectivity for an existing node
	huma.Register(api, huma.Operation{
		OperationID: "test-existing-node-connectivity",
		Method:      "POST",
		Path:        "/api/v1/nodes/{id}/test-connectivity",
		Summary:     "Test existing node connectivity",
		Description: "Re-test SSH connectivity for an existing node using its stored credentials",
		Tags:        []string{"Nodes"},
	}, func(ctx context.Context, input *NodePathInput) (*ConnectivityTestOutput, error) {
		node, err := deps.Store.GetNode(ctx, input.ID)
		if err != nil {
			if err == state.ErrNotFound {
				return nil, huma.Error404NotFound("node not found")
			}
			return nil, huma.Error500InternalServerError("failed to get node", err)
		}

		// Decrypt stored password if available
		password := ""
		if node.EncryptedPassword != "" && deps.EncryptionKey != "" {
			decrypted, err := crypto.Decrypt(node.EncryptedPassword, deps.EncryptionKey)
			if err != nil {
				deps.logOp(ctx, state.OpTestConnectivity, input.ID, "", state.OpStatusFailure, "failed to decrypt password")
				return nil, huma.Error500InternalServerError("failed to decrypt stored password", err)
			}
			password = decrypted
		}

		result := ssh.TestConnectivity(ctx, ssh.ConnectivityTestRequest{
			Address:        node.Address,
			Port:           node.SSHPort,
			Username:       node.Username,
			Password:       password,
			PrivateKeyPath: node.PrivateKeyPath,
		})

		opStatus := state.OpStatusSuccess
		opErr := ""
		if result.Error != "" {
			opStatus = state.OpStatusFailure
			opErr = result.Error
		}
		deps.logOp(ctx, state.OpTestConnectivity, input.ID, fmt.Sprintf("address=%s:%d", node.Address, node.SSHPort), opStatus, opErr)

		resp := &ConnectivityTestOutput{}
		resp.Body.Reachable = result.Reachable
		resp.Body.Authenticated = result.Authenticated
		resp.Body.LatencyMs = result.LatencyMs
		resp.Body.Error = result.Error
		resp.Body.ServerVersion = result.ServerVersion
		return resp, nil
	})

	// Assign role to a node
	huma.Register(api, huma.Operation{
		OperationID: "assign-node-role",
		Method:      "POST",
		Path:        "/api/v1/nodes/{id}/roles",
		Summary:     "Assign a role to a node",
		Description: "Assign a component role to a node. Idempotent — re-assigning an existing role is a no-op.",
		Tags:        []string{"Nodes"},
	}, func(ctx context.Context, input *AssignRoleInput) (*AssignRoleOutput, error) {
		// Verify node exists
		if _, err := deps.Store.GetNode(ctx, input.ID); err != nil {
			if err == state.ErrNotFound {
				return nil, huma.Error404NotFound("node not found")
			}
			return nil, huma.Error500InternalServerError("failed to get node", err)
		}

		if err := deps.Store.AssignRole(ctx, input.ID, input.Body.Role); err != nil {
			deps.logOp(ctx, state.OpAssignRole, input.ID, fmt.Sprintf("role=%s", input.Body.Role), state.OpStatusFailure, err.Error())
			return nil, huma.Error500InternalServerError("failed to assign role", err)
		}

		deps.logOp(ctx, state.OpAssignRole, input.ID, fmt.Sprintf("role=%s", input.Body.Role), state.OpStatusSuccess, "")

		resp := &AssignRoleOutput{}
		resp.Body.Success = true
		return resp, nil
	})

	// Remove role from a node
	huma.Register(api, huma.Operation{
		OperationID: "remove-node-role",
		Method:      "DELETE",
		Path:        "/api/v1/nodes/{id}/roles/{role}",
		Summary:     "Remove a role from a node",
		Description: "Remove a component role assignment from a node",
		Tags:        []string{"Nodes"},
	}, func(ctx context.Context, input *RemoveRoleInput) (*RemoveRoleOutput, error) {
		// Verify node exists
		if _, err := deps.Store.GetNode(ctx, input.ID); err != nil {
			if err == state.ErrNotFound {
				return nil, huma.Error404NotFound("node not found")
			}
			return nil, huma.Error500InternalServerError("failed to get node", err)
		}

		if err := deps.Store.RemoveRole(ctx, input.ID, input.Role); err != nil {
			deps.logOp(ctx, state.OpRemoveRole, input.ID, fmt.Sprintf("role=%s", input.Role), state.OpStatusFailure, err.Error())
			return nil, huma.Error500InternalServerError("failed to remove role", err)
		}

		deps.logOp(ctx, state.OpRemoveRole, input.ID, fmt.Sprintf("role=%s", input.Role), state.OpStatusSuccess, "")

		resp := &RemoveRoleOutput{}
		resp.Body.Success = true
		return resp, nil
	})
}
