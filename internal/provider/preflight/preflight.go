package preflight

import (
	"github.com/bengrewell/aether-webui/internal/endpoint"
	"github.com/bengrewell/aether-webui/internal/provider"
)

var _ provider.Provider = (*Preflight)(nil)

// Preflight is a provider for pre-deployment system checks.
type Preflight struct {
	*provider.Base
	endpoints []endpoint.AnyEndpoint
}

// NewProvider creates a new Preflight provider with all endpoints registered.
func NewProvider(opts ...provider.Option) *Preflight {
	p := &Preflight{
		Base:      provider.New("preflight", opts...),
		endpoints: make([]endpoint.AnyEndpoint, 0, 3),
	}

	provider.Register(p.Base, endpoint.Endpoint[struct{}, PreflightListOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "preflight-list",
			Semantics:   endpoint.Read,
			Summary:     "Run all preflight checks",
			Description: "Runs all preflight checks in parallel and returns aggregate results.",
			Tags:        []string{"preflight"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/preflight"},
		},
		Handler: p.handleListChecks,
	})

	provider.Register(p.Base, endpoint.Endpoint[PreflightGetInput, PreflightGetOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "preflight-get",
			Semantics:   endpoint.Read,
			Summary:     "Run a single preflight check",
			Description: "Runs a single preflight check by ID and returns the result.",
			Tags:        []string{"preflight"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/preflight/{id}"},
		},
		Handler: p.handleGetCheck,
	})

	provider.Register(p.Base, endpoint.Endpoint[PreflightFixInput, PreflightFixOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "preflight-fix",
			Semantics:   endpoint.Action,
			Summary:     "Apply a preflight fix",
			Description: "Executes the automated fix for a preflight check. Returns 422 if no fix is available.",
			Tags:        []string{"preflight"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/preflight/{id}/fix"},
		},
		Handler: p.handleFixCheck,
	})

	return p
}

// Endpoints returns all registered endpoints for the provider.
func (p *Preflight) Endpoints() []endpoint.AnyEndpoint { return p.endpoints }
