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
	cores      map[string]*CoreConfig
	coreStatus map[string]*CoreStatus
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
		cores: map[string]*CoreConfig{
			"core-0": {
				ID:         "core-0",
				Name:       "SD-Core Primary",
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
		},
		coreStatus: map[string]*CoreStatus{
			"core-0": {
				ID:         "core-0",
				Name:       "SD-Core Primary",
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
				Name:        "srsRAN gNB 0",
				Host:        host,
				Type:        "srsran",
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

// Core (SD-Core) management

func (m *MockProvider) ListCores(ctx context.Context) (*CoreList, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cores := make([]CoreConfig, 0, len(m.cores))
	for _, core := range m.cores {
		cores = append(cores, *core)
	}

	return &CoreList{Cores: cores}, nil
}

func (m *MockProvider) GetCore(ctx context.Context, id string) (*CoreConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	core, exists := m.cores[id]
	if !exists {
		return nil, fmt.Errorf("core %s not found", id)
	}

	config := *core
	return &config, nil
}

func (m *MockProvider) DeployCore(ctx context.Context, config *CoreConfig) (*DeploymentResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Use defaults if config is nil
	if config == nil {
		config = m.defaultCoreConfig()
	}

	// Generate ID if not provided
	if config.ID == "" {
		config.ID = fmt.Sprintf("core-%d", len(m.cores))
	}

	// Generate Name if not provided
	if config.Name == "" {
		config.Name = fmt.Sprintf("SD-Core %d", len(m.cores))
	}

	if _, exists := m.cores[config.ID]; exists {
		return nil, fmt.Errorf("core %s already exists", config.ID)
	}

	// Apply defaults for empty fields
	m.applyCoreDefaults(config)
	m.cores[config.ID] = config

	now := time.Now()
	m.coreStatus[config.ID] = &CoreStatus{
		ID:         config.ID,
		Name:       config.Name,
		Host:       m.host,
		State:      StateDeployed,
		Message:    "SD-Core deployed successfully",
		DeployedAt: &now,
		Version:    config.Helm.ChartVersion,
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
		Message: fmt.Sprintf("SD-Core %s deployment initiated", config.ID),
		ID:      config.ID,
		TaskID:  fmt.Sprintf("task-core-deploy-%s", config.ID),
	}, nil
}

func (m *MockProvider) defaultCoreConfig() *CoreConfig {
	return &CoreConfig{
		Standalone: true,
		DataIface:  "ens18",
		ValuesFile: "deps/5gc/roles/core/templates/sdcore-5g-values.yaml",
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
	}
}

func (m *MockProvider) applyCoreDefaults(config *CoreConfig) {
	defaults := m.defaultCoreConfig()
	if config.DataIface == "" {
		config.DataIface = defaults.DataIface
	}
	if config.Helm.ChartRef == "" {
		config.Helm.ChartRef = defaults.Helm.ChartRef
	}
	if config.Helm.ChartVersion == "" {
		config.Helm.ChartVersion = defaults.Helm.ChartVersion
	}
	if config.UPF.AccessSubnet == "" {
		config.UPF.AccessSubnet = defaults.UPF.AccessSubnet
	}
	if config.UPF.CoreSubnet == "" {
		config.UPF.CoreSubnet = defaults.UPF.CoreSubnet
	}
	if config.UPF.Mode == "" {
		config.UPF.Mode = defaults.UPF.Mode
	}
	if config.AMF.IP == "" {
		config.AMF.IP = defaults.AMF.IP
	}
}

func (m *MockProvider) UpdateCore(ctx context.Context, id string, config *CoreConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.cores[id]; !exists {
		return fmt.Errorf("core %s not found", id)
	}

	config.ID = id // Ensure ID matches
	m.cores[id] = config

	// Update status name if changed
	if status, exists := m.coreStatus[id]; exists {
		status.Name = config.Name
	}

	return nil
}

func (m *MockProvider) UndeployCore(ctx context.Context, id string) (*DeploymentResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.cores[id]; !exists {
		return nil, fmt.Errorf("core %s not found", id)
	}

	delete(m.cores, id)
	delete(m.coreStatus, id)

	return &DeploymentResponse{
		Success: true,
		Message: fmt.Sprintf("SD-Core %s undeployment initiated", id),
		TaskID:  fmt.Sprintf("task-core-undeploy-%s", id),
	}, nil
}

func (m *MockProvider) GetCoreStatus(ctx context.Context, id string) (*CoreStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.coreStatus[id]
	if !exists {
		return nil, fmt.Errorf("core %s not found", id)
	}

	s := *status
	return &s, nil
}

func (m *MockProvider) ListCoreStatuses(ctx context.Context) (*CoreStatusList, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]CoreStatus, 0, len(m.coreStatus))
	for _, status := range m.coreStatus {
		statuses = append(statuses, *status)
	}

	return &CoreStatusList{Cores: statuses}, nil
}

// gNB management

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

func (m *MockProvider) DeployGNB(ctx context.Context, config *GNBConfig) (*DeploymentResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Use defaults if config is nil
	if config == nil {
		config = m.defaultGNBConfig()
	}

	// Generate ID if not provided
	if config.ID == "" {
		config.ID = fmt.Sprintf("gnb-%d", len(m.gnbs))
	}

	// Generate Name if not provided
	if config.Name == "" {
		config.Name = fmt.Sprintf("gNB %d", len(m.gnbs))
	}

	if _, exists := m.gnbs[config.ID]; exists {
		return nil, fmt.Errorf("gNB %s already exists", config.ID)
	}

	// Apply defaults for empty fields
	m.applyGNBDefaults(config)
	config.Host = m.host
	m.gnbs[config.ID] = config

	now := time.Now()
	m.gnbStatus[config.ID] = &GNBStatus{
		ID:          config.ID,
		Name:        config.Name,
		Host:        m.host,
		Type:        config.Type,
		State:       StateDeployed,
		Message:     "gNB deployed successfully",
		DeployedAt:  &now,
		Connected:   true,
		UEsAttached: 0,
	}

	return &DeploymentResponse{
		Success: true,
		Message: fmt.Sprintf("gNB %s deployment initiated", config.ID),
		ID:      config.ID,
		TaskID:  fmt.Sprintf("task-gnb-deploy-%s", config.ID),
	}, nil
}

func (m *MockProvider) defaultGNBConfig() *GNBConfig {
	return &GNBConfig{
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
	}
}

func (m *MockProvider) applyGNBDefaults(config *GNBConfig) {
	defaults := m.defaultGNBConfig()
	if config.Type == "" {
		config.Type = defaults.Type
	}
	if config.IP == "" {
		config.IP = defaults.IP
	}
	if config.Docker.Container.GNBImage == "" {
		config.Docker.Container.GNBImage = defaults.Docker.Container.GNBImage
	}
	if config.Docker.Container.UEImage == "" {
		config.Docker.Container.UEImage = defaults.Docker.Container.UEImage
	}
	if config.Docker.Network.Name == "" {
		config.Docker.Network.Name = defaults.Docker.Network.Name
	}
}

func (m *MockProvider) UpdateGNB(ctx context.Context, id string, config *GNBConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.gnbs[id]; !exists {
		return fmt.Errorf("gNB %s not found", id)
	}

	config.ID = id // Ensure ID matches
	config.Host = m.host
	m.gnbs[id] = config

	// Update status fields if changed
	if status, exists := m.gnbStatus[id]; exists {
		status.Name = config.Name
		status.Type = config.Type
	}

	return nil
}

func (m *MockProvider) UndeployGNB(ctx context.Context, id string) (*DeploymentResponse, error) {
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
		TaskID:  fmt.Sprintf("task-gnb-undeploy-%s", id),
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
