package aether

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockProvider provides mock implementation for Aether component management.
type MockProvider struct {
	mu         sync.RWMutex
	host       string
	coreConfig *CoreConfig
	coreStatus *CoreStatus
	gnbs       map[string]*GNBConfig
	gnbStatus  map[string]*GNBStatus
}

// NewMockProvider creates a new MockProvider with default configuration for the given host.
func NewMockProvider(host string) *MockProvider {
	if host == "" {
		host = "local"
	}
	deployedAt := time.Now().Add(-24 * time.Hour)

	return &MockProvider{
		host: host,
		coreConfig: &CoreConfig{
			Standalone: true,
			DataIface:  "ens18",
			ValuesFile: "deps/5gc/roles/core/templates/sdcore-5g-values.yaml",
			RANSubnet:  "",
			Helm: HelmConfig{
				LocalCharts:  false,
				ChartRef:     "oci://ghcr.io/omec-project/sd-core",
				ChartVersion: "3.1.3",
			},
			UPF: UPFConfig{
				AccessSubnet: "192.168.252.1/24",
				CoreSubnet:   "192.168.250.1/24",
				Mode:         "af_packet",
				MultihopGNB:  false,
				DefaultUPF: DefaultUPFConfig{
					IP: UPFIPConfig{
						Access: "192.168.252.3",
						Core:   "192.168.250.3",
					},
					UEIPPool: "192.168.100.0/24",
				},
			},
			AMF: AMFConfig{
				IP: "10.76.28.113",
			},
		},
		coreStatus: &CoreStatus{
			Host:       host,
			State:      StateDeployed,
			Message:    "SD-Core is running",
			DeployedAt: &deployedAt,
			Version:    "3.1.3",
			Components: []ComponentStatus{
				{Name: "amf", Ready: true, Status: "Running"},
				{Name: "smf", Ready: true, Status: "Running"},
				{Name: "upf", Ready: true, Status: "Running"},
				{Name: "nrf", Ready: true, Status: "Running"},
				{Name: "ausf", Ready: true, Status: "Running"},
				{Name: "nssf", Ready: true, Status: "Running"},
				{Name: "pcf", Ready: true, Status: "Running"},
				{Name: "udm", Ready: true, Status: "Running"},
				{Name: "udr", Ready: true, Status: "Running"},
				{Name: "webui", Ready: true, Status: "Running"},
				{Name: "mongodb", Ready: true, Status: "Running"},
			},
		},
		gnbs: map[string]*GNBConfig{
			"gnb-0": {
				ID:         "gnb-0",
				Host:       host,
				Name:       "srsRAN gNB 0",
				Type:       "srsran",
				IP:         "10.76.28.115",
				ConfigFile: "deps/srsran/roles/gNB/templates/gnb_zmq.yaml",
				Docker: GNBDockerConfig{
					Container: GNBContainerConfig{
						GNBImage: "aetherproject/srsran-gnb:rel-0.4.0",
						UEImage:  "aetherproject/srsran-ue:rel-0.4.0",
					},
					Network: GNBNetworkConfig{
						Name: "host",
					},
				},
				Simulation: true,
				UEConfig: &UEConfig{
					ConfigFile: "deps/srsran/roles/uEsimulator/templates/ue_zmq.conf",
				},
			},
		},
		gnbStatus: map[string]*GNBStatus{
			"gnb-0": {
				ID:          "gnb-0",
				State:       StateDeployed,
				Message:     "gNB is running in simulation mode",
				DeployedAt:  &deployedAt,
				Connected:   true,
				UEsAttached: 3,
			},
		},
	}
}

func (m *MockProvider) Host() string {
	return m.host
}

func (m *MockProvider) GetCoreConfig(ctx context.Context) (*CoreConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	config := *m.coreConfig
	return &config, nil
}

func (m *MockProvider) UpdateCoreConfig(ctx context.Context, config *CoreConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.coreConfig = config
	return nil
}

func (m *MockProvider) DeployCore(ctx context.Context) (*DeploymentResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.coreStatus = &CoreStatus{
		Host:       m.host,
		State:      StateDeployed,
		Message:    "SD-Core deployed successfully",
		DeployedAt: &now,
		Version:    m.coreConfig.Helm.ChartVersion,
		Components: []ComponentStatus{
			{Name: "amf", Ready: true, Status: "Running"},
			{Name: "smf", Ready: true, Status: "Running"},
			{Name: "upf", Ready: true, Status: "Running"},
			{Name: "nrf", Ready: true, Status: "Running"},
			{Name: "ausf", Ready: true, Status: "Running"},
			{Name: "nssf", Ready: true, Status: "Running"},
			{Name: "pcf", Ready: true, Status: "Running"},
			{Name: "udm", Ready: true, Status: "Running"},
			{Name: "udr", Ready: true, Status: "Running"},
			{Name: "webui", Ready: true, Status: "Running"},
			{Name: "mongodb", Ready: true, Status: "Running"},
		},
	}

	return &DeploymentResponse{
		Success: true,
		Message: "SD-Core deployment initiated",
		TaskID:  "task-core-deploy-001",
	}, nil
}

func (m *MockProvider) UndeployCore(ctx context.Context) (*DeploymentResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.coreStatus = &CoreStatus{
		Host:    m.host,
		State:   StateNotDeployed,
		Message: "SD-Core is not deployed",
	}

	return &DeploymentResponse{
		Success: true,
		Message: "SD-Core undeployment initiated",
		TaskID:  "task-core-undeploy-001",
	}, nil
}

func (m *MockProvider) GetCoreStatus(ctx context.Context) (*CoreStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := *m.coreStatus
	return &status, nil
}

func (m *MockProvider) ListGNBs(ctx context.Context) (*GNBList, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gnbs := make([]GNBConfig, 0, len(m.gnbs))
	for _, gnb := range m.gnbs {
		gnbs = append(gnbs, *gnb)
	}

	return &GNBList{GNBs: gnbs}, nil
}

func (m *MockProvider) GetGNB(ctx context.Context, id string) (*GNBConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gnb, exists := m.gnbs[id]
	if !exists {
		return nil, fmt.Errorf("gNB %s not found", id)
	}

	config := *gnb
	return &config, nil
}

func (m *MockProvider) CreateGNB(ctx context.Context, config *GNBConfig) (*DeploymentResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.gnbs[config.ID]; exists {
		return nil, fmt.Errorf("gNB %s already exists", config.ID)
	}

	config.Host = m.host
	m.gnbs[config.ID] = config

	now := time.Now()
	m.gnbStatus[config.ID] = &GNBStatus{
		ID:          config.ID,
		State:       StateDeployed,
		Message:     "gNB deployed successfully",
		DeployedAt:  &now,
		Connected:   true,
		UEsAttached: 0,
	}

	return &DeploymentResponse{
		Success: true,
		Message: fmt.Sprintf("gNB %s deployment initiated", config.ID),
		TaskID:  fmt.Sprintf("task-gnb-deploy-%s", config.ID),
	}, nil
}

func (m *MockProvider) UpdateGNB(ctx context.Context, id string, config *GNBConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.gnbs[id]; !exists {
		return fmt.Errorf("gNB %s not found", id)
	}

	config.ID = id // Ensure ID matches
	m.gnbs[id] = config
	return nil
}

func (m *MockProvider) DeleteGNB(ctx context.Context, id string) (*DeploymentResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.gnbs[id]; !exists {
		return nil, fmt.Errorf("gNB %s not found", id)
	}

	delete(m.gnbs, id)
	delete(m.gnbStatus, id)

	return &DeploymentResponse{
		Success: true,
		Message: fmt.Sprintf("gNB %s removal initiated", id),
		TaskID:  fmt.Sprintf("task-gnb-delete-%s", id),
	}, nil
}

func (m *MockProvider) GetGNBStatus(ctx context.Context, id string) (*GNBStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.gnbStatus[id]
	if !exists {
		return nil, fmt.Errorf("gNB %s not found", id)
	}

	s := *status
	return &s, nil
}

func (m *MockProvider) ListGNBStatuses(ctx context.Context) (*GNBStatusList, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]GNBStatus, 0, len(m.gnbStatus))
	for _, status := range m.gnbStatus {
		statuses = append(statuses, *status)
	}

	return &GNBStatusList{GNBs: statuses}, nil
}

// Ensure MockProvider implements Provider
var _ Provider = (*MockProvider)(nil)
