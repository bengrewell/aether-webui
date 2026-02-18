package meta

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/bengrewell/aether-webui/internal/endpoint"
	"github.com/bengrewell/aether-webui/internal/provider"
)

// Meta is the provider that exposes application introspection endpoints.
type Meta struct {
	*provider.Base
	endpoints   []endpoint.AnyEndpoint
	versionInfo VersionInfo
	appConfig   AppConfig
	schemaVer   SchemaVersionFunc
	providersFn ProviderStatusFunc
	storeInfoFn StoreInfoFunc
	startTime   time.Time
}

var _ provider.Provider = (*Meta)(nil)

// NewProvider creates a Meta provider with all introspection endpoints registered.
func NewProvider(version VersionInfo, config AppConfig, schemaVer SchemaVersionFunc, providersFn ProviderStatusFunc, storeInfoFn StoreInfoFunc, opts ...provider.Option) *Meta {
	m := &Meta{
		Base:        provider.New("meta", opts...),
		endpoints:   make([]endpoint.AnyEndpoint, 0, 6),
		versionInfo: version,
		appConfig:   config,
		schemaVer:   schemaVer,
		providersFn: providersFn,
		storeInfoFn: storeInfoFn,
		startTime:   time.Now(),
	}

	provider.Register(m.Base, endpoint.Endpoint[struct{}, VersionOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "meta-version",
			Semantics:   endpoint.Read,
			Summary:     "Get build and version information",
			Description: "Returns the server's version string, build date, git branch, and commit hash. Useful for diagnostics, compatibility checks, and deployment verification.",
			Tags:        []string{"meta"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/meta/version"},
		},
		Handler: m.handleVersion,
	})

	provider.Register(m.Base, endpoint.Endpoint[struct{}, BuildOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "meta-build",
			Semantics:   endpoint.Read,
			Summary:     "Get build environment information",
			Description: "Returns the Go toolchain version, target OS, and architecture used to compile the server binary.",
			Tags:        []string{"meta"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/meta/build"},
		},
		Handler: m.handleBuild,
	})

	provider.Register(m.Base, endpoint.Endpoint[struct{}, RuntimeOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "meta-runtime",
			Semantics:   endpoint.Read,
			Summary:     "Get process runtime information",
			Description: "Returns the server's PID, running user/group, binary path, start time, and uptime.",
			Tags:        []string{"meta"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/meta/runtime"},
		},
		Handler: m.handleRuntime,
	})

	provider.Register(m.Base, endpoint.Endpoint[struct{}, ConfigOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "meta-config",
			Semantics:   endpoint.Read,
			Summary:     "Get active application configuration",
			Description: "Returns non-secret configuration values including listen address, storage paths, feature flags, and schema version.",
			Tags:        []string{"meta"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/meta/config"},
		},
		Handler: m.handleConfig,
	})

	provider.Register(m.Base, endpoint.Endpoint[struct{}, ProvidersOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "meta-providers",
			Semantics:   endpoint.Read,
			Summary:     "Get registered provider statuses",
			Description: "Returns the name, enabled/running state, and endpoint count for each registered provider.",
			Tags:        []string{"meta"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/meta/providers"},
		},
		Handler: m.handleProviders,
	})

	provider.Register(m.Base, endpoint.Endpoint[struct{}, StoreOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "meta-store",
			Semantics:   endpoint.Read,
			Summary:     "Get store health and metadata",
			Description: "Returns store engine, path, file size, schema version, and live diagnostic results (ping, write, read, delete).",
			Tags:        []string{"meta"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/meta/store"},
		},
		Handler: m.handleStore,
	})

	return m
}

func (m *Meta) Name() string { return "meta" }

func (m *Meta) Endpoints() []endpoint.AnyEndpoint { return m.endpoints }

func (m *Meta) handleVersion(_ context.Context, _ *struct{}) (*VersionOutput, error) {
	return &VersionOutput{Body: m.versionInfo}, nil
}

func (m *Meta) handleBuild(_ context.Context, _ *struct{}) (*BuildOutput, error) {
	return &BuildOutput{Body: BuildInfo{
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}}, nil
}

func (m *Meta) handleRuntime(_ context.Context, _ *struct{}) (*RuntimeOutput, error) {
	info := RuntimeInfo{
		PID:       os.Getpid(),
		StartTime: m.startTime.Format(time.RFC3339),
		Uptime:    fmt.Sprintf("%s", time.Since(m.startTime).Round(time.Second)),
	}

	if exe, err := os.Executable(); err == nil {
		info.BinaryPath = exe
	}

	if u, err := user.Current(); err == nil {
		info.User = UserInfo{UID: u.Uid, Name: u.Username}
		if g, err := user.LookupGroupId(u.Gid); err == nil {
			info.Group = GroupInfo{GID: g.Gid, Name: g.Name}
		} else {
			info.Group = GroupInfo{GID: u.Gid}
		}
	}

	return &RuntimeOutput{Body: info}, nil
}

func (m *Meta) handleConfig(_ context.Context, _ *struct{}) (*ConfigOutput, error) {
	schemaVersion := 0
	if m.schemaVer != nil {
		if v, err := m.schemaVer(); err == nil {
			schemaVersion = v
		}
	}
	return &ConfigOutput{Body: ConfigInfo{
		AppConfig:     m.appConfig,
		SchemaVersion: schemaVersion,
	}}, nil
}

func (m *Meta) handleStore(ctx context.Context, _ *struct{}) (*StoreOutput, error) {
	if m.storeInfoFn == nil {
		return &StoreOutput{Body: StoreInfo{
			Status:      "unhealthy",
			Diagnostics: []DiagnosticCheck{},
		}}, nil
	}
	return &StoreOutput{Body: m.storeInfoFn(ctx)}, nil
}

func (m *Meta) handleProviders(_ context.Context, _ *struct{}) (*ProvidersOutput, error) {
	var providers []ProviderStatus
	if m.providersFn != nil {
		providers = m.providersFn()
	}
	if providers == nil {
		providers = []ProviderStatus{}
	}
	return &ProvidersOutput{Body: ProvidersInfo{Providers: providers}}, nil
}
