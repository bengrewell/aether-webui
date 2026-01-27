package aether

import "context"

// Provider defines the interface for managing Aether 5G components on a specific host.
type Provider interface {
	// Host returns the host this provider manages.
	Host() string

	// Core (SD-Core) management
	GetCoreConfig(ctx context.Context) (*CoreConfig, error)
	UpdateCoreConfig(ctx context.Context, config *CoreConfig) error
	DeployCore(ctx context.Context) (*DeploymentResponse, error)
	UndeployCore(ctx context.Context) (*DeploymentResponse, error)
	GetCoreStatus(ctx context.Context) (*CoreStatus, error)

	// gNB management
	ListGNBs(ctx context.Context) (*GNBList, error)
	GetGNB(ctx context.Context, id string) (*GNBConfig, error)
	CreateGNB(ctx context.Context, config *GNBConfig) (*DeploymentResponse, error)
	UpdateGNB(ctx context.Context, id string, config *GNBConfig) error
	DeleteGNB(ctx context.Context, id string) (*DeploymentResponse, error)
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
