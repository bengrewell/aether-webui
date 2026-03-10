package wizard

import (
	"github.com/bengrewell/aether-webui/internal/endpoint"
	"github.com/bengrewell/aether-webui/internal/provider"
)

var _ provider.Provider = (*Wizard)(nil)

// Wizard is a provider for tracking setup wizard completion state.
type Wizard struct {
	*provider.Base
	endpoints []endpoint.AnyEndpoint
}

// NewProvider creates a new Wizard provider with all endpoints registered.
func NewProvider(opts ...provider.Option) *Wizard {
	w := &Wizard{
		Base:      provider.New("wizard", opts...),
		endpoints: make([]endpoint.AnyEndpoint, 0, 3),
	}

	provider.Register(w.Base, endpoint.Endpoint[struct{}, WizardGetOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "wizard-get",
			Semantics:   endpoint.Read,
			Summary:     "Get wizard state",
			Description: "Returns all wizard steps with completion status.",
			Tags:        []string{"wizard"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/wizard"},
		},
		Handler: w.handleGet,
	})

	provider.Register(w.Base, endpoint.Endpoint[StepCompleteInput, StepCompleteOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "wizard-complete-step",
			Semantics:   endpoint.Update,
			Summary:     "Complete a wizard step",
			Description: "Marks a wizard step as completed.",
			Tags:        []string{"wizard"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/wizard/steps/{step}"},
		},
		Handler: w.handleCompleteStep,
	})

	provider.Register(w.Base, endpoint.Endpoint[struct{}, WizardResetOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "wizard-reset",
			Semantics:   endpoint.Delete,
			Summary:     "Reset wizard",
			Description: "Clears all wizard step completions.",
			Tags:        []string{"wizard"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/wizard"},
		},
		Handler: w.handleReset,
	})

	return w
}

// Endpoints returns all registered endpoints for the provider.
func (w *Wizard) Endpoints() []endpoint.AnyEndpoint { return w.endpoints }
