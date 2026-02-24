package nodes

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/danielgtaylor/huma/v2"

	"github.com/bengrewell/aether-webui/internal/store"
)

func (n *Nodes) handleList(ctx context.Context, _ *struct{}) (*ManagedNodeListOutput, error) {
	infos, err := n.Store().ListNodes(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list nodes", err)
	}
	out := make([]ManagedNode, len(infos))
	for i, info := range infos {
		out[i] = managedNodeFromInfo(info)
	}
	return &ManagedNodeListOutput{Body: out}, nil
}

func (n *Nodes) handleGet(ctx context.Context, in *NodeGetInput) (*NodeGetOutput, error) {
	node, ok, err := n.Store().GetNode(ctx, in.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get node", err)
	}
	if !ok {
		return nil, huma.Error404NotFound("node not found", fmt.Errorf("no node with id %s", in.ID))
	}
	return &NodeGetOutput{Body: managedNodeFromNode(node)}, nil
}

func (n *Nodes) handleCreate(ctx context.Context, in *NodeCreateInput) (*NodeCreateOutput, error) {
	if in.Body.Name == "" {
		return nil, huma.Error422UnprocessableEntity("name is required")
	}
	if in.Body.AnsibleHost == "" {
		return nil, huma.Error422UnprocessableEntity("ansible_host is required")
	}
	if err := validateRoles(in.Body.Roles); err != nil {
		return nil, err
	}

	id, err := generateID()
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to generate ID", err)
	}

	node := store.Node{
		ID:           id,
		Name:         in.Body.Name,
		AnsibleHost:  in.Body.AnsibleHost,
		AnsibleUser:  in.Body.AnsibleUser,
		Password:     []byte(in.Body.Password),
		SudoPassword: []byte(in.Body.SudoPassword),
		SSHKey:       []byte(in.Body.SSHKey),
		Roles:        in.Body.Roles,
	}

	if err := n.Store().UpsertNode(ctx, node); err != nil {
		return nil, huma.Error500InternalServerError("failed to create node", err)
	}

	created, ok, err := n.Store().GetNode(ctx, id)
	if err != nil || !ok {
		return nil, huma.Error500InternalServerError("failed to read back node", err)
	}
	return &NodeCreateOutput{Body: managedNodeFromNode(created)}, nil
}

func (n *Nodes) handleUpdate(ctx context.Context, in *NodeUpdateInput) (*NodeUpdateOutput, error) {
	existing, ok, err := n.Store().GetNode(ctx, in.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get node", err)
	}
	if !ok {
		return nil, huma.Error404NotFound("node not found", fmt.Errorf("no node with id %s", in.ID))
	}

	if in.Body.Name != nil {
		existing.Name = *in.Body.Name
	}
	if in.Body.AnsibleHost != nil {
		existing.AnsibleHost = *in.Body.AnsibleHost
	}
	if in.Body.AnsibleUser != nil {
		existing.AnsibleUser = *in.Body.AnsibleUser
	}
	if in.Body.Password != nil {
		existing.Password = []byte(*in.Body.Password)
	}
	if in.Body.SudoPassword != nil {
		existing.SudoPassword = []byte(*in.Body.SudoPassword)
	}
	if in.Body.SSHKey != nil {
		existing.SSHKey = []byte(*in.Body.SSHKey)
	}
	if in.Body.Roles != nil {
		if err := validateRoles(in.Body.Roles); err != nil {
			return nil, err
		}
		existing.Roles = in.Body.Roles
	}

	if err := n.Store().UpsertNode(ctx, existing); err != nil {
		return nil, huma.Error500InternalServerError("failed to update node", err)
	}

	updated, ok, err := n.Store().GetNode(ctx, in.ID)
	if err != nil || !ok {
		return nil, huma.Error500InternalServerError("failed to read back node", err)
	}
	return &NodeUpdateOutput{Body: managedNodeFromNode(updated)}, nil
}

func (n *Nodes) handleDelete(ctx context.Context, in *NodeDeleteInput) (*NodeDeleteOutput, error) {
	if err := n.Store().DeleteNode(ctx, in.ID); err != nil {
		return nil, huma.Error500InternalServerError("failed to delete node", err)
	}
	out := &NodeDeleteOutput{}
	out.Body.Message = fmt.Sprintf("node %s deleted", in.ID)
	return out, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func managedNodeFromNode(n store.Node) ManagedNode {
	return ManagedNode{
		ID:              n.ID,
		Name:            n.Name,
		AnsibleHost:     n.AnsibleHost,
		AnsibleUser:     n.AnsibleUser,
		HasPassword:     len(n.Password) > 0,
		HasSudoPassword: len(n.SudoPassword) > 0,
		HasSSHKey:       len(n.SSHKey) > 0,
		Roles:           n.Roles,
		CreatedAt:       n.CreatedAt,
		UpdatedAt:       n.UpdatedAt,
	}
}

func managedNodeFromInfo(info store.NodeInfo) ManagedNode {
	return ManagedNode{
		ID:          info.ID,
		Name:        info.Name,
		AnsibleHost: info.AnsibleHost,
		AnsibleUser: info.AnsibleUser,
		Roles:       info.Roles,
		CreatedAt:   info.CreatedAt,
		UpdatedAt:   info.UpdatedAt,
	}
}

func validateRoles(roles []string) error {
	for _, r := range roles {
		if !ValidRoles[r] {
			return huma.Error422UnprocessableEntity(fmt.Sprintf("invalid role %q", r))
		}
	}
	return nil
}

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
