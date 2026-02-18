package system

import (
	"context"
	"time"

	"github.com/bengrewell/aether-webui/internal/endpoint"
	"github.com/bengrewell/aether-webui/internal/provider"
)

// Config holds collection parameters for the system provider.
type Config struct {
	CollectInterval time.Duration
	Retention       time.Duration
}

// System is the provider that exposes host system info and metrics endpoints.
type System struct {
	*provider.Base
	config    Config
	endpoints []endpoint.AnyEndpoint
	cancel    context.CancelFunc
	done      chan struct{} // closed when collector goroutine exits
}

var _ provider.Provider = (*System)(nil)

// NewProvider creates a System provider with all system endpoints registered.
func NewProvider(cfg Config, opts ...provider.Option) *System {
	s := &System{
		Base:      provider.New("system", opts...),
		config:    cfg,
		endpoints: make([]endpoint.AnyEndpoint, 0, 8),
	}

	provider.Register(s.Base, endpoint.Endpoint[struct{}, CPUInfoOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "system-cpu",
			Semantics:   endpoint.Read,
			Summary:     "Get CPU information",
			Description: "Returns CPU model, core counts, frequency, cache, and flags.",
			Tags:        []string{"system"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/system/cpu"},
		},
		Handler: s.handleCPU,
	})

	provider.Register(s.Base, endpoint.Endpoint[struct{}, MemoryInfoOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "system-memory",
			Semantics:   endpoint.Read,
			Summary:     "Get memory information",
			Description: "Returns physical and swap memory usage.",
			Tags:        []string{"system"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/system/memory"},
		},
		Handler: s.handleMemory,
	})

	provider.Register(s.Base, endpoint.Endpoint[struct{}, DiskInfoOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "system-disks",
			Semantics:   endpoint.Read,
			Summary:     "Get disk information",
			Description: "Returns partition list with usage details.",
			Tags:        []string{"system"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/system/disks"},
		},
		Handler: s.handleDisks,
	})

	provider.Register(s.Base, endpoint.Endpoint[struct{}, OSInfoOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "system-os",
			Semantics:   endpoint.Read,
			Summary:     "Get OS information",
			Description: "Returns hostname, OS, platform, kernel version, arch, and uptime.",
			Tags:        []string{"system"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/system/os"},
		},
		Handler: s.handleOS,
	})

	provider.Register(s.Base, endpoint.Endpoint[struct{}, NetworkInterfacesOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "system-network-interfaces",
			Semantics:   endpoint.Read,
			Summary:     "Get network interfaces",
			Description: "Returns network interface details including addresses, MAC, MTU, and flags.",
			Tags:        []string{"system"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/system/network/interfaces"},
		},
		Handler: s.handleNetworkInterfaces,
	})

	provider.Register(s.Base, endpoint.Endpoint[struct{}, NetworkConfigOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "system-network-config",
			Semantics:   endpoint.Read,
			Summary:     "Get network configuration",
			Description: "Returns DNS servers and search domains from resolv.conf.",
			Tags:        []string{"system"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/system/network/config"},
		},
		Handler: s.handleNetworkConfig,
	})

	provider.Register(s.Base, endpoint.Endpoint[struct{}, ListeningPortsOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "system-network-ports",
			Semantics:   endpoint.Read,
			Summary:     "Get listening ports",
			Description: "Returns TCP/UDP ports in LISTEN state with process info.",
			Tags:        []string{"system"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/system/network/ports"},
		},
		Handler: s.handleListeningPorts,
	})

	provider.Register(s.Base, endpoint.Endpoint[MetricsQueryInput, MetricsQueryOutput]{
		Desc: endpoint.Descriptor{
			OperationID: "system-metrics",
			Semantics:   endpoint.Read,
			Summary:     "Query system metrics",
			Description: "Returns time-series metric data with optional time range, label filtering, and aggregation.",
			Tags:        []string{"system"},
			HTTP:        endpoint.HTTPHint{Path: "/api/v1/system/metrics"},
		},
		Handler: s.handleMetricsQuery,
	})

	return s
}

func (s *System) Endpoints() []endpoint.AnyEndpoint { return s.endpoints }
