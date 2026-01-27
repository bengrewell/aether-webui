package aether

import "fmt"

// DefaultHostResolver resolves host identifiers to providers.
type DefaultHostResolver struct {
	localProvider Provider
	remoteHosts   map[string]Provider
}

// NewDefaultHostResolver creates a new resolver with the given local provider.
func NewDefaultHostResolver(localProvider Provider) *DefaultHostResolver {
	return &DefaultHostResolver{
		localProvider: localProvider,
		remoteHosts:   make(map[string]Provider),
	}
}

// AddHost adds a remote host provider.
func (r *DefaultHostResolver) AddHost(host string, provider Provider) {
	r.remoteHosts[host] = provider
}

// Resolve returns a Provider for the given host identifier.
// Empty string or "local" returns the local provider.
func (r *DefaultHostResolver) Resolve(host string) (Provider, error) {
	if host == "" || host == "local" {
		return r.localProvider, nil
	}

	provider, exists := r.remoteHosts[host]
	if !exists {
		return nil, fmt.Errorf("host %q not configured", host)
	}
	return provider, nil
}

// ListHosts returns all configured hosts including "local".
func (r *DefaultHostResolver) ListHosts() []string {
	hosts := make([]string, 0, len(r.remoteHosts)+1)
	hosts = append(hosts, "local")
	for host := range r.remoteHosts {
		hosts = append(hosts, host)
	}
	return hosts
}

// Ensure DefaultHostResolver implements HostResolver
var _ HostResolver = (*DefaultHostResolver)(nil)
