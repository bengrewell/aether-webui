package configdefaults

import (
	"github.com/bengrewell/aether-webui/internal/endpoint"
	"github.com/bengrewell/aether-webui/internal/nodefacts"
	"github.com/bengrewell/aether-webui/internal/provider"
)

var _ provider.Provider = (*Provider)(nil)

// Provider exposes node fact discovery and config defaults endpoints.
type Provider struct {
	*provider.Base
	onRampDir string
	gatherer  nodefacts.Gatherer
}

// Config holds settings for the configdefaults provider.
type Config struct {
	OnRampDir string // path to aether-onramp on disk
}

// NewProvider creates a new configdefaults provider.
func NewProvider(cfg Config, gatherer nodefacts.Gatherer, opts ...provider.Option) *Provider {
	base := provider.New("configdefaults", opts...)
	p := &Provider{
		Base:      base,
		onRampDir: cfg.OnRampDir,
		gatherer:  gatherer,
	}

	provider.Register(base, endpoint.Endpoint[NodeFactsGetInput, NodeFactsGetOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "get-node-facts",
			Semantics:   endpoint.Read,
			Summary:     "Get node network facts",
			Description: "Returns discovered network facts for a single node. Uses cached facts unless refresh=true.",
			Tags:        []string{"configdefaults"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/nodes/{id}/facts"},
		},
		Handler: p.handleGetNodeFacts,
	})

	provider.Register(base, endpoint.Endpoint[ConfigDefaultsApplyInput, ConfigDefaultsApplyOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "apply-config-defaults",
			Semantics:   endpoint.Action,
			Summary:     "Apply config defaults from node facts",
			Description: "Gathers facts for all registered nodes, computes config defaults based on node roles, and merges them into vars/main.yml.",
			Tags:        []string{"configdefaults"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/onramp/config/defaults"},
		},
		Handler: p.handleApplyConfigDefaults,
	})

	return p
}

// Endpoints returns all registered endpoints.
func (p *Provider) Endpoints() []endpoint.AnyEndpoint { return nil }

// Start marks the provider as running.
func (p *Provider) Start() error {
	p.SetRunning(true)
	return nil
}

// Stop marks the provider as no longer running.
func (p *Provider) Stop() error {
	p.SetRunning(false)
	return nil
}
