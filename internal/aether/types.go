package aether

import "time"

// DeploymentState represents the state of a deployment.
type DeploymentState string

const (
	StateNotDeployed DeploymentState = "not_deployed"
	StateDeploying   DeploymentState = "deploying"
	StateDeployed    DeploymentState = "deployed"
	StateFailed      DeploymentState = "failed"
	StateUndeploying DeploymentState = "undeploying"
)

// CoreConfig represents the SD-Core (5G Core) configuration.
type CoreConfig struct {
	Standalone bool       `json:"standalone"`
	DataIface  string     `json:"data_iface"`
	ValuesFile string     `json:"values_file,omitempty"`
	RANSubnet  string     `json:"ran_subnet,omitempty"`
	Helm       HelmConfig `json:"helm"`
	UPF        UPFConfig  `json:"upf"`
	AMF        AMFConfig  `json:"amf"`
}

// HelmConfig represents Helm chart configuration for SD-Core.
type HelmConfig struct {
	LocalCharts  bool   `json:"local_charts"`
	ChartRef     string `json:"chart_ref"`
	ChartVersion string `json:"chart_version"`
}

// UPFConfig represents User Plane Function configuration.
type UPFConfig struct {
	AccessSubnet string        `json:"access_subnet"`
	CoreSubnet   string        `json:"core_subnet"`
	Mode         string        `json:"mode"`
	MultihopGNB  bool          `json:"multihop_gnb"`
	DefaultUPF   DefaultUPFConfig `json:"default_upf"`
}

// DefaultUPFConfig represents the default UPF instance configuration.
type DefaultUPFConfig struct {
	IP       UPFIPConfig `json:"ip"`
	UEIPPool string      `json:"ue_ip_pool"`
}

// UPFIPConfig represents UPF IP address configuration.
type UPFIPConfig struct {
	Access string `json:"access"`
	Core   string `json:"core"`
}

// AMFConfig represents Access and Mobility Management Function configuration.
type AMFConfig struct {
	IP string `json:"ip"`
}

// CoreStatus represents the deployment status of the SD-Core.
type CoreStatus struct {
	Host         string            `json:"host"`
	State        DeploymentState   `json:"state"`
	Message      string            `json:"message,omitempty"`
	DeployedAt   *time.Time        `json:"deployed_at,omitempty"`
	Version      string            `json:"version,omitempty"`
	Components   []ComponentStatus `json:"components,omitempty"`
}

// ComponentStatus represents the status of an individual component.
type ComponentStatus struct {
	Name   string `json:"name"`
	Ready  bool   `json:"ready"`
	Status string `json:"status"`
}

// GNBConfig represents a gNB (gNodeB) configuration.
type GNBConfig struct {
	ID         string          `json:"id"`
	Host       string          `json:"host"` // Target host where gNB is deployed
	Name       string          `json:"name,omitempty"`
	Type       string          `json:"type"` // srsran, ocudu
	IP         string          `json:"gnb_ip"`
	ConfigFile string          `json:"gnb_conf,omitempty"`
	Docker     GNBDockerConfig `json:"docker,omitempty"`
	Simulation bool            `json:"simulation"`
	UEConfig   *UEConfig       `json:"ue_config,omitempty"` // Only for simulation mode
}

// GNBDockerConfig represents Docker configuration for gNB.
type GNBDockerConfig struct {
	Container GNBContainerConfig `json:"container"`
	Network   GNBNetworkConfig   `json:"network"`
}

// GNBContainerConfig represents container image configuration.
type GNBContainerConfig struct {
	GNBImage string `json:"gnb_image"`
	UEImage  string `json:"ue_image,omitempty"`
}

// GNBNetworkConfig represents Docker network configuration.
type GNBNetworkConfig struct {
	Name string `json:"name"`
}

// UEConfig represents UE simulator configuration (for simulation mode).
type UEConfig struct {
	ConfigFile string `json:"ue_conf,omitempty"`
}

// GNBStatus represents the deployment status of a gNB.
type GNBStatus struct {
	ID           string          `json:"id"`
	State        DeploymentState `json:"state"`
	Message      string          `json:"message,omitempty"`
	DeployedAt   *time.Time      `json:"deployed_at,omitempty"`
	Connected    bool            `json:"connected"`
	UEsAttached  int             `json:"ues_attached"`
}

// GNBList represents a list of gNBs.
type GNBList struct {
	GNBs []GNBConfig `json:"gnbs"`
}

// GNBStatusList represents status for all gNBs.
type GNBStatusList struct {
	GNBs []GNBStatus `json:"gnbs"`
}

// DeploymentRequest represents a request to deploy a component.
type DeploymentRequest struct {
	Host string `json:"host,omitempty"` // Target host, empty or "local" for local deployment
}

// GNBCreateRequest represents a request to create/deploy a gNB.
type GNBCreateRequest struct {
	Host   string    `json:"host,omitempty"` // Target host, empty or "local" for local deployment
	Config GNBConfig `json:"config"`
}

// DeploymentResponse represents the response from a deployment action.
type DeploymentResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	TaskID  string `json:"task_id,omitempty"` // For async operations
}
