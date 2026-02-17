package meta

import (
	"context"
	"fmt"

	"github.com/bengrewell/aether-webui/internal/endpoint"
	"github.com/bengrewell/aether-webui/internal/provider"
)

type VersionInfo struct {
	Version    string `json:"version"`
	BuildDate  string `json:"buildDate"`
	Branch     string `json:"branch"`
	CommitHash string `json:"commitHash"`
}

type VersionOutput struct {
	Body VersionInfo
}

type Meta struct {
	*provider.Base
	enabled     bool
	running     bool
	endpoints   []endpoint.AnyEndpoint
	versionInfo VersionInfo
}

var _ provider.Provider = (*Meta)(nil)

func NewProvider(version VersionInfo, opts ...provider.Option) provider.Provider {
	m := &Meta{
		Base:        provider.New("meta", opts...),
		enabled:     true,
		running:     false,
		endpoints:   make([]endpoint.AnyEndpoint, 0, 4),
		versionInfo: version,
	}

	ver := endpoint.Endpoint[struct{}, VersionOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "version",
			Semantics:   endpoint.Read,
			Summary:     "Get build and version information",
			Description: "Returns the server's version string, build date, git branch, and commit hash. Useful for diagnostics, compatibility checks, and deployment verification.",
			Tags:        []string{"meta", "version"},
			HTTP: endpoint.HTTPHint{
				Path: "/api/v1/version",
			},
		},
		Handler: m.handleVersion,
	}

	// Single line registers w/ Huma if enabled, otherwise just records descriptor
	provider.Register(m.Base, ver)

	return m
}

func (m *Meta) Name() string { return "meta" }

func (m *Meta) Endpoints() []endpoint.AnyEndpoint { return m.endpoints }

func (m *Meta) handleVersion(ctx context.Context, _ *struct{}) (*VersionOutput, error) {
	fmt.Println(m.versionInfo)
	return &VersionOutput{Body: m.versionInfo}, nil
}
