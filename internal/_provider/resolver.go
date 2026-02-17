package provider

import (
	"errors"
	"sync"
)

// ProviderResolver resolves node identifiers to providers.
type ProviderResolver interface {
	Resolve(node NodeID) (Provider, error)
	ListNodes() []NodeID
	RegisterNode(node NodeID, provider Provider) error
	UnregisterNode(node NodeID) error
	LocalProvider() Provider
}

// ErrNodeNotFound is returned when a node cannot be resolved.
var ErrNodeNotFound = errors.New("node not found")

// ErrNodeAlreadyRegistered is returned when attempting to register a node that already exists.
var ErrNodeAlreadyRegistered = errors.New("node already registered")

// DefaultResolver implements ProviderResolver with support for local and remote nodes.
type DefaultResolver struct {
	mu      sync.RWMutex
	local   Provider
	remotes map[NodeID]Provider
}

// NewDefaultResolver creates a new DefaultResolver with the given local provider.
func NewDefaultResolver(local Provider) *DefaultResolver {
	return &DefaultResolver{
		local:   local,
		remotes: make(map[NodeID]Provider),
	}
}

// Resolve returns the provider for the given node identifier.
// Empty string or "local" resolves to the local provider.
func (r *DefaultResolver) Resolve(node NodeID) (Provider, error) {
	if node == "" || node == LocalNode {
		return r.local, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if p, ok := r.remotes[node]; ok {
		return p, nil
	}
	return nil, ErrNodeNotFound
}

// ListNodes returns all registered node identifiers.
func (r *DefaultResolver) ListNodes() []NodeID {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]NodeID, 0, len(r.remotes)+1)
	nodes = append(nodes, LocalNode)
	for node := range r.remotes {
		nodes = append(nodes, node)
	}
	return nodes
}

// RegisterNode registers a remote provider for the given node identifier.
func (r *DefaultResolver) RegisterNode(node NodeID, provider Provider) error {
	if node == "" || node == LocalNode {
		return errors.New("cannot register local node")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.remotes[node]; exists {
		return ErrNodeAlreadyRegistered
	}
	r.remotes[node] = provider
	return nil
}

// UnregisterNode removes a remote provider registration.
func (r *DefaultResolver) UnregisterNode(node NodeID) error {
	if node == "" || node == LocalNode {
		return errors.New("cannot unregister local node")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.remotes[node]; !exists {
		return ErrNodeNotFound
	}
	delete(r.remotes, node)
	return nil
}

// LocalProvider returns the local provider.
func (r *DefaultResolver) LocalProvider() Provider {
	return r.local
}
