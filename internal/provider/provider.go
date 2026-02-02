package provider

import (
	"context"

	"github.com/bengrewell/aether-webui/internal/operator"
)

// NodeID represents a unique identifier for a node.
type NodeID string

// LocalNode is the constant identifier for the local node.
const LocalNode NodeID = "local"

// Provider represents a node that can dispatch operations to domain-specific operators.
type Provider interface {
	ID() NodeID
	Operator(domain operator.Domain) operator.Operator
	Operators() map[operator.Domain]operator.Operator
	Health(ctx context.Context) (*ProviderHealth, error)
	IsLocal() bool
}

// ProviderHealth represents the overall health of a provider and its operators.
type ProviderHealth struct {
	Status    string                       `json:"status"`
	Message   string                       `json:"message"`
	Operators map[operator.Domain]string `json:"operators"`
}
