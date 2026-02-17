package webuiapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/bengrewell/aether-webui/internal/operator"
	"github.com/bengrewell/aether-webui/internal/operator/aether"
	"github.com/bengrewell/aether-webui/internal/operator/host"
	"github.com/bengrewell/aether-webui/internal/operator/kube"
	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

// mockHostOperator is a configurable mock for host.HostOperator.
type mockHostOperator struct {
	cpuInfo       *host.CPUInfo
	cpuInfoErr    error
	memoryInfo    *host.MemoryInfo
	memoryInfoErr error
	diskInfo      *host.DiskInfo
	diskInfoErr   error
	nicInfo       *host.NICInfo
	nicInfoErr    error
	osInfo        *host.OSInfo
	osInfoErr     error
	cpuUsage      *host.CPUUsage
	cpuUsageErr   error
	memoryUsage   *host.MemoryUsage
	memoryUsageErr error
	diskUsage     *host.DiskUsage
	diskUsageErr  error
	nicUsage      *host.NICUsage
	nicUsageErr   error
}

func (m *mockHostOperator) Domain() operator.Domain {
	return operator.DomainHost
}

func (m *mockHostOperator) Health(_ context.Context) (*operator.OperatorHealth, error) {
	return &operator.OperatorHealth{Status: "healthy", Message: "mock"}, nil
}

func (m *mockHostOperator) GetCPUInfo(_ context.Context) (*host.CPUInfo, error) {
	return m.cpuInfo, m.cpuInfoErr
}

func (m *mockHostOperator) GetMemoryInfo(_ context.Context) (*host.MemoryInfo, error) {
	return m.memoryInfo, m.memoryInfoErr
}

func (m *mockHostOperator) GetDiskInfo(_ context.Context) (*host.DiskInfo, error) {
	return m.diskInfo, m.diskInfoErr
}

func (m *mockHostOperator) GetNICInfo(_ context.Context) (*host.NICInfo, error) {
	return m.nicInfo, m.nicInfoErr
}

func (m *mockHostOperator) GetOSInfo(_ context.Context) (*host.OSInfo, error) {
	return m.osInfo, m.osInfoErr
}

func (m *mockHostOperator) GetCPUUsage(_ context.Context) (*host.CPUUsage, error) {
	return m.cpuUsage, m.cpuUsageErr
}

func (m *mockHostOperator) GetMemoryUsage(_ context.Context) (*host.MemoryUsage, error) {
	return m.memoryUsage, m.memoryUsageErr
}

func (m *mockHostOperator) GetDiskUsage(_ context.Context) (*host.DiskUsage, error) {
	return m.diskUsage, m.diskUsageErr
}

func (m *mockHostOperator) GetNICUsage(_ context.Context) (*host.NICUsage, error) {
	return m.nicUsage, m.nicUsageErr
}

// mockKubeOperator is a configurable mock for kube.KubeOperator.
type mockKubeOperator struct {
	clusterHealth    *kube.ClusterHealth
	clusterHealthErr error
	nodes            *kube.NodeList
	nodesErr         error
	namespaces       *kube.NamespaceList
	namespacesErr    error
	events           *kube.EventList
	eventsErr        error
	pods             *kube.PodList
	podsErr          error
	deployments      *kube.DeploymentList
	deploymentsErr   error
	services         *kube.ServiceList
	servicesErr      error
}

func (m *mockKubeOperator) Domain() operator.Domain {
	return operator.DomainKube
}

func (m *mockKubeOperator) Health(_ context.Context) (*operator.OperatorHealth, error) {
	return &operator.OperatorHealth{Status: "healthy", Message: "mock"}, nil
}

func (m *mockKubeOperator) GetClusterHealth(_ context.Context) (*kube.ClusterHealth, error) {
	return m.clusterHealth, m.clusterHealthErr
}

func (m *mockKubeOperator) GetNodes(_ context.Context) (*kube.NodeList, error) {
	return m.nodes, m.nodesErr
}

func (m *mockKubeOperator) GetNamespaces(_ context.Context) (*kube.NamespaceList, error) {
	return m.namespaces, m.namespacesErr
}

func (m *mockKubeOperator) GetEvents(_ context.Context, _ string, _ int) (*kube.EventList, error) {
	return m.events, m.eventsErr
}

func (m *mockKubeOperator) GetPods(_ context.Context, _ string) (*kube.PodList, error) {
	return m.pods, m.podsErr
}

func (m *mockKubeOperator) GetDeployments(_ context.Context, _ string) (*kube.DeploymentList, error) {
	return m.deployments, m.deploymentsErr
}

func (m *mockKubeOperator) GetServices(_ context.Context, _ string) (*kube.ServiceList, error) {
	return m.services, m.servicesErr
}

// mockAetherOperator is a configurable mock for aether.AetherOperator.
type mockAetherOperator struct {
	// Core operations
	cores           *aether.CoreList
	coresErr        error
	core            *aether.CoreConfig
	coreErr         error
	deployCore      *aether.DeploymentResponse
	deployCoreErr   error
	updateCoreErr   error
	undeployCore    *aether.DeploymentResponse
	undeployCoreErr error
	coreStatus      *aether.CoreStatus
	coreStatusErr   error
	coreStatuses    *aether.CoreStatusList
	coreStatusesErr error

	// gNB operations
	gnbs           *aether.GNBList
	gnbsErr        error
	gnb            *aether.GNBConfig
	gnbErr         error
	deployGNB      *aether.DeploymentResponse
	deployGNBErr   error
	updateGNBErr   error
	undeployGNB    *aether.DeploymentResponse
	undeployGNBErr error
	gnbStatus      *aether.GNBStatus
	gnbStatusErr   error
	gnbStatuses    *aether.GNBStatusList
	gnbStatusesErr error
}

func (m *mockAetherOperator) Domain() operator.Domain {
	return operator.DomainAether
}

func (m *mockAetherOperator) Health(_ context.Context) (*operator.OperatorHealth, error) {
	return &operator.OperatorHealth{Status: "healthy", Message: "mock"}, nil
}

func (m *mockAetherOperator) ListCores(_ context.Context) (*aether.CoreList, error) {
	return m.cores, m.coresErr
}

func (m *mockAetherOperator) GetCore(_ context.Context, _ string) (*aether.CoreConfig, error) {
	return m.core, m.coreErr
}

func (m *mockAetherOperator) DeployCore(_ context.Context, _ *aether.CoreConfig) (*aether.DeploymentResponse, error) {
	return m.deployCore, m.deployCoreErr
}

func (m *mockAetherOperator) UpdateCore(_ context.Context, _ string, _ *aether.CoreConfig) error {
	return m.updateCoreErr
}

func (m *mockAetherOperator) UndeployCore(_ context.Context, _ string) (*aether.DeploymentResponse, error) {
	return m.undeployCore, m.undeployCoreErr
}

func (m *mockAetherOperator) GetCoreStatus(_ context.Context, _ string) (*aether.CoreStatus, error) {
	return m.coreStatus, m.coreStatusErr
}

func (m *mockAetherOperator) ListCoreStatuses(_ context.Context) (*aether.CoreStatusList, error) {
	return m.coreStatuses, m.coreStatusesErr
}

func (m *mockAetherOperator) ListGNBs(_ context.Context) (*aether.GNBList, error) {
	return m.gnbs, m.gnbsErr
}

func (m *mockAetherOperator) GetGNB(_ context.Context, _ string) (*aether.GNBConfig, error) {
	return m.gnb, m.gnbErr
}

func (m *mockAetherOperator) DeployGNB(_ context.Context, _ *aether.GNBConfig) (*aether.DeploymentResponse, error) {
	return m.deployGNB, m.deployGNBErr
}

func (m *mockAetherOperator) UpdateGNB(_ context.Context, _ string, _ *aether.GNBConfig) error {
	return m.updateGNBErr
}

func (m *mockAetherOperator) UndeployGNB(_ context.Context, _ string) (*aether.DeploymentResponse, error) {
	return m.undeployGNB, m.undeployGNBErr
}

func (m *mockAetherOperator) GetGNBStatus(_ context.Context, _ string) (*aether.GNBStatus, error) {
	return m.gnbStatus, m.gnbStatusErr
}

func (m *mockAetherOperator) ListGNBStatuses(_ context.Context) (*aether.GNBStatusList, error) {
	return m.gnbStatuses, m.gnbStatusesErr
}

// mockProvider implements provider.Provider for testing.
type mockProvider struct {
	id        provider.NodeID
	operators map[operator.Domain]operator.Operator
	health    *provider.ProviderHealth
	healthErr error
	isLocal   bool
}

func (m *mockProvider) ID() provider.NodeID {
	return m.id
}

func (m *mockProvider) Operator(domain operator.Domain) operator.Operator {
	return m.operators[domain]
}

func (m *mockProvider) Operators() map[operator.Domain]operator.Operator {
	return m.operators
}

func (m *mockProvider) Health(_ context.Context) (*provider.ProviderHealth, error) {
	return m.health, m.healthErr
}

func (m *mockProvider) IsLocal() bool {
	return m.isLocal
}

// testResolver wraps a provider for testing API routes.
type testResolver struct {
	local   provider.Provider
	remotes map[provider.NodeID]provider.Provider
}

func newTestResolver(local provider.Provider) *testResolver {
	return &testResolver{
		local:   local,
		remotes: make(map[provider.NodeID]provider.Provider),
	}
}

func (r *testResolver) Resolve(node provider.NodeID) (provider.Provider, error) {
	if node == "" || node == provider.LocalNode {
		return r.local, nil
	}
	if p, ok := r.remotes[node]; ok {
		return p, nil
	}
	return nil, provider.ErrNodeNotFound
}

func (r *testResolver) ListNodes() []provider.NodeID {
	nodes := []provider.NodeID{provider.LocalNode}
	for n := range r.remotes {
		nodes = append(nodes, n)
	}
	return nodes
}

func (r *testResolver) RegisterNode(node provider.NodeID, p provider.Provider) error {
	r.remotes[node] = p
	return nil
}

func (r *testResolver) UnregisterNode(node provider.NodeID) error {
	delete(r.remotes, node)
	return nil
}

func (r *testResolver) LocalProvider() provider.Provider {
	return r.local
}

// newSystemTestRouter creates a router with system routes for testing.
func newSystemTestRouter(t *testing.T, hostOp *mockHostOperator) http.Handler {
	t.Helper()
	localProvider := &mockProvider{
		id:      provider.LocalNode,
		isLocal: true,
		operators: map[operator.Domain]operator.Operator{
			operator.DomainHost: hostOp,
		},
	}
	resolver := newTestResolver(localProvider)

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterSystemRoutes(api, resolver)
	return router
}

// newMetricsTestRouter creates a router with metrics routes for testing.
func newMetricsTestRouter(t *testing.T, hostOp *mockHostOperator) http.Handler {
	t.Helper()
	localProvider := &mockProvider{
		id:      provider.LocalNode,
		isLocal: true,
		operators: map[operator.Domain]operator.Operator{
			operator.DomainHost: hostOp,
		},
	}
	resolver := newTestResolver(localProvider)

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterMetricsRoutes(api, resolver)
	return router
}

// newKubernetesTestRouter creates a router with kubernetes routes for testing.
func newKubernetesTestRouter(t *testing.T, kubeOp *mockKubeOperator) http.Handler {
	t.Helper()
	localProvider := &mockProvider{
		id:      provider.LocalNode,
		isLocal: true,
		operators: map[operator.Domain]operator.Operator{
			operator.DomainKube: kubeOp,
		},
	}
	resolver := newTestResolver(localProvider)

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterKubernetesRoutes(api, resolver)
	return router
}

// newAetherTestRouter creates a router with aether routes for testing.
func newAetherTestRouter(t *testing.T, aetherOp *mockAetherOperator) http.Handler {
	t.Helper()
	localProvider := &mockProvider{
		id:      provider.LocalNode,
		isLocal: true,
		operators: map[operator.Domain]operator.Operator{
			operator.DomainAether: aetherOp,
		},
	}
	resolver := newTestResolver(localProvider)

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterAetherRoutes(api, resolver)
	return router
}

// newAetherTestRouterWithResolver creates a router with aether routes and custom resolver.
func newAetherTestRouterWithResolver(t *testing.T, resolver provider.ProviderResolver) http.Handler {
	t.Helper()
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterAetherRoutes(api, resolver)
	return router
}

// newSystemTestRouterNoOperator creates a router without a host operator for testing error cases.
func newSystemTestRouterNoOperator(t *testing.T) http.Handler {
	t.Helper()
	localProvider := &mockProvider{
		id:        provider.LocalNode,
		isLocal:   true,
		operators: map[operator.Domain]operator.Operator{},
	}
	resolver := newTestResolver(localProvider)

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterSystemRoutes(api, resolver)
	RegisterMetricsRoutes(api, resolver)
	return router
}

// newKubernetesTestRouterNoOperator creates a router without a kube operator.
func newKubernetesTestRouterNoOperator(t *testing.T) http.Handler {
	t.Helper()
	localProvider := &mockProvider{
		id:        provider.LocalNode,
		isLocal:   true,
		operators: map[operator.Domain]operator.Operator{},
	}
	resolver := newTestResolver(localProvider)

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterKubernetesRoutes(api, resolver)
	return router
}

// newAetherTestRouterNoOperator creates a router without an aether operator.
func newAetherTestRouterNoOperator(t *testing.T) http.Handler {
	t.Helper()
	localProvider := &mockProvider{
		id:        provider.LocalNode,
		isLocal:   true,
		operators: map[operator.Domain]operator.Operator{},
	}
	resolver := newTestResolver(localProvider)

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterAetherRoutes(api, resolver)
	return router
}
