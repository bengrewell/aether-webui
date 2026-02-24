package nodes

import (
	"github.com/bengrewell/aether-webui/internal/endpoint"
	"github.com/bengrewell/aether-webui/internal/provider"
)

var _ provider.Provider = (*Nodes)(nil)

// Nodes is a provider for managing cluster nodes and their role assignments.
type Nodes struct {
	*provider.Base
	endpoints []endpoint.AnyEndpoint
}

// NewProvider creates a new Nodes provider with all CRUD endpoints registered.
func NewProvider(opts ...provider.Option) *Nodes {
	n := &Nodes{
		Base:      provider.New("nodes", opts...),
		endpoints: make([]endpoint.AnyEndpoint, 0, 5),
	}

	provider.Register(n.Base, endpoint.Endpoint[struct{}, ManagedNodeListOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "nodes-list",
			Semantics:   endpoint.Read,
			Summary:     "List all nodes",
			Description: "Returns all managed cluster nodes with role assignments.",
			Tags:        []string{"nodes"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/nodes"},
		},
		Handler: n.handleList,
	})

	provider.Register(n.Base, endpoint.Endpoint[NodeGetInput, NodeGetOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "nodes-get",
			Semantics:   endpoint.Read,
			Summary:     "Get a node",
			Description: "Returns a single node with secret-presence flags.",
			Tags:        []string{"nodes"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/nodes/{id}"},
		},
		Handler: n.handleGet,
	})

	provider.Register(n.Base, endpoint.Endpoint[NodeCreateInput, NodeCreateOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "nodes-create",
			Semantics:   endpoint.Create,
			Summary:     "Create a node",
			Description: "Creates a new node with optional roles and credentials.",
			Tags:        []string{"nodes"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/nodes"},
		},
		Handler: n.handleCreate,
	})

	provider.Register(n.Base, endpoint.Endpoint[NodeUpdateInput, NodeUpdateOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "nodes-update",
			Semantics:   endpoint.Update,
			Summary:     "Update a node",
			Description: "Partial update — merges non-nil fields; roles replaces entire set when provided.",
			Tags:        []string{"nodes"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/nodes/{id}"},
		},
		Handler: n.handleUpdate,
	})

	provider.Register(n.Base, endpoint.Endpoint[NodeDeleteInput, NodeDeleteOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "nodes-delete",
			Semantics:   endpoint.Delete,
			Summary:     "Delete a node",
			Description: "Deletes a node and its role assignments.",
			Tags:        []string{"nodes"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/nodes/{id}"},
		},
		Handler: n.handleDelete,
	})

	return n
}

// Endpoints returns all registered endpoints for the provider.
func (n *Nodes) Endpoints() []endpoint.AnyEndpoint { return n.endpoints }
