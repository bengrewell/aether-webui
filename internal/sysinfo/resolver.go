package sysinfo

import "fmt"

// DefaultNodeResolver resolves node identifiers to providers.
// Currently only supports "local" or empty string, returning the mock provider.
type DefaultNodeResolver struct {
	localProvider Provider
}

// NewDefaultNodeResolver creates a new resolver with the given local provider.
func NewDefaultNodeResolver(localProvider Provider) *DefaultNodeResolver {
	return &DefaultNodeResolver{
		localProvider: localProvider,
	}
}

// Resolve returns a Provider for the given node identifier.
// Empty string or "local" returns the local provider.
// Other values will return an error until remote node support is implemented.
func (r *DefaultNodeResolver) Resolve(node string) (Provider, error) {
	if node == "" || node == "local" {
		return r.localProvider, nil
	}
	// TODO: Implement remote node resolution
	return nil, fmt.Errorf("remote node %q not supported yet", node)
}

// Ensure DefaultNodeResolver implements NodeResolver
var _ NodeResolver = (*DefaultNodeResolver)(nil)
