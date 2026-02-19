package onramp

import (
	"sync"

	"github.com/bengrewell/aether-webui/internal/endpoint"
	"github.com/bengrewell/aether-webui/internal/provider"
)

var _ provider.Provider = (*OnRamp)(nil)

// Config holds the settings for the OnRamp provider.
type Config struct {
	OnRampDir string // path to aether-onramp on disk
	RepoURL   string // git clone URL
	Version   string // tag, branch, or commit to pin
}

// OnRamp is a provider that wraps the Aether OnRamp Make/Ansible toolchain.
type OnRamp struct {
	*provider.Base
	config    Config
	endpoints []endpoint.AnyEndpoint

	mu    sync.Mutex
	tasks []*Task // most-recent first
}

// NewProvider creates a new OnRamp provider with all endpoints registered.
func NewProvider(cfg Config, opts ...provider.Option) *OnRamp {
	o := &OnRamp{
		Base:      provider.New("onramp", opts...),
		config:    cfg,
		endpoints: make([]endpoint.AnyEndpoint, 0, 12),
		tasks:     make([]*Task, 0),
	}

	// --- Components ---

	provider.Register(o.Base, endpoint.Endpoint[struct{}, ComponentListOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "onramp-list-components",
			Semantics:   endpoint.Read,
			Summary:     "List OnRamp components",
			Description: "Returns all available OnRamp components and their actions.",
			Tags:        []string{"onramp"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/onramp/components"},
		},
		Handler: o.handleListComponents,
	})

	provider.Register(o.Base, endpoint.Endpoint[ComponentGetInput, ComponentGetOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "onramp-get-component",
			Semantics:   endpoint.Read,
			Summary:     "Get OnRamp component",
			Description: "Returns a single OnRamp component by name.",
			Tags:        []string{"onramp"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/onramp/components/{component}"},
		},
		Handler: o.handleGetComponent,
	})

	provider.Register(o.Base, endpoint.Endpoint[ExecuteActionInput, ExecuteActionOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "onramp-execute-action",
			Semantics:   endpoint.Action,
			Summary:     "Execute component action",
			Description: "Runs a make target for the specified component and action.",
			Tags:        []string{"onramp"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/onramp/components/{component}/{action}"},
		},
		Handler: o.handleExecuteAction,
	})

	// --- Tasks ---

	provider.Register(o.Base, endpoint.Endpoint[struct{}, TaskListOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "onramp-list-tasks",
			Semantics:   endpoint.Read,
			Summary:     "List OnRamp tasks",
			Description: "Returns recent make target executions and their status.",
			Tags:        []string{"onramp"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/onramp/tasks"},
		},
		Handler: o.handleListTasks,
	})

	provider.Register(o.Base, endpoint.Endpoint[TaskGetInput, TaskGetOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "onramp-get-task",
			Semantics:   endpoint.Read,
			Summary:     "Get OnRamp task",
			Description: "Returns details and output for a specific task.",
			Tags:        []string{"onramp"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/onramp/tasks/{id}"},
		},
		Handler: o.handleGetTask,
	})

	// --- Config ---

	provider.Register(o.Base, endpoint.Endpoint[struct{}, ConfigGetOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "onramp-get-config",
			Semantics:   endpoint.Read,
			Summary:     "Get OnRamp configuration",
			Description: "Reads vars/main.yml and returns the parsed configuration.",
			Tags:        []string{"onramp"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/onramp/config"},
		},
		Handler: o.handleGetConfig,
	})

	provider.Register(o.Base, endpoint.Endpoint[ConfigPatchInput, ConfigPatchOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "onramp-patch-config",
			Semantics:   endpoint.Update,
			Summary:     "Patch OnRamp configuration",
			Description: "Merges the provided fields into vars/main.yml, preserving untouched values.",
			Tags:        []string{"onramp"},
			HTTP:        endpoint.HTTPHint{Method: "PATCH", Path: "/api/v1/onramp/config"},
		},
		Handler: o.handlePatchConfig,
	})

	// --- Profiles ---

	provider.Register(o.Base, endpoint.Endpoint[struct{}, ProfileListOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "onramp-list-profiles",
			Semantics:   endpoint.Read,
			Summary:     "List OnRamp config profiles",
			Description: "Lists available vars profiles (main-*.yml files).",
			Tags:        []string{"onramp"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/onramp/config/profiles"},
		},
		Handler: o.handleListProfiles,
	})

	provider.Register(o.Base, endpoint.Endpoint[ProfileGetInput, ProfileGetOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "onramp-get-profile",
			Semantics:   endpoint.Read,
			Summary:     "Get OnRamp config profile",
			Description: "Reads a specific vars profile and returns its configuration.",
			Tags:        []string{"onramp"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/onramp/config/profiles/{name}"},
		},
		Handler: o.handleGetProfile,
	})

	provider.Register(o.Base, endpoint.Endpoint[ProfileActivateInput, ProfileActivateOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "onramp-activate-profile",
			Semantics:   endpoint.Action,
			Summary:     "Activate OnRamp config profile",
			Description: "Copies the named profile to vars/main.yml, making it active.",
			Tags:        []string{"onramp"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/onramp/config/profiles/{name}/activate"},
		},
		Handler: o.handleActivateProfile,
	})

	return o
}

// Endpoints returns all registered endpoints for the provider.
func (o *OnRamp) Endpoints() []endpoint.AnyEndpoint { return o.endpoints }

// Start clones/validates the OnRamp repo and marks the provider as running.
// If repo setup fails, the provider logs the error and starts in degraded mode.
func (o *OnRamp) Start() error {
	log := o.Log()
	if err := ensureRepo(o.config, log); err != nil {
		log.Error("onramp repo setup failed; provider starting in degraded mode", "error", err)
	}
	o.SetRunning(true)
	return nil
}

// Stop marks the provider as no longer running.
func (o *OnRamp) Stop() error {
	o.SetRunning(false)
	return nil
}
