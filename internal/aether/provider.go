package aether

import "context"

// Provider defines the interface for managing Aether 5G components on a specific host.
type Provider interface {
	// Host returns the host this provider manages.
	Host() string

	// Core (SD-Core) management
	ListCores(ctx context.Context) (*CoreList, error)
	GetCore(ctx context.Context, id string) (*CoreConfig, error)
	DeployCore(ctx context.Context, config *CoreConfig) (*DeploymentResponse, error)
	UpdateCore(ctx context.Context, id string, config *CoreConfig) error
	UndeployCore(ctx context.Context, id string) (*DeploymentResponse, error)
	GetCoreStatus(ctx context.Context, id string) (*CoreStatus, error)
	ListCoreStatuses(ctx context.Context) (*CoreStatusList, error)

	// gNB management
	ListGNBs(ctx context.Context) (*GNBList, error)
	GetGNB(ctx context.Context, id string) (*GNBConfig, error)
	DeployGNB(ctx context.Context, config *GNBConfig) (*DeploymentResponse, error)
	UpdateGNB(ctx context.Context, id string, config *GNBConfig) error
	UndeployGNB(ctx context.Context, id string) (*DeploymentResponse, error)
	GetGNBStatus(ctx context.Context, id string) (*GNBStatus, error)
	ListGNBStatuses(ctx context.Context) (*GNBStatusList, error)
}

// HostResolver resolves a host identifier to a Provider.
// When host is empty or "local", it returns the local provider.
type HostResolver interface {
	Resolve(host string) (Provider, error)
	// ListHosts returns all configured hosts.
	ListHosts() []string
}
