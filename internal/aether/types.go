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
	ID         string     `json:"id,omitempty" required:"false"`   // Unique identifier, generated if not provided
	Name       string     `json:"name,omitempty" required:"false"` // Friendly display name, generated if not provided
	Standalone bool       `json:"standalone" required:"false"`
	DataIface  string     `json:"data_iface" required:"false"`
	ValuesFile string     `json:"values_file,omitempty" required:"false"`
	RANSubnet  string     `json:"ran_subnet,omitempty" required:"false"`
	Helm       HelmConfig `json:"helm" required:"false"`
	UPF        UPFConfig  `json:"upf" required:"false"`
	AMF        AMFConfig  `json:"amf" required:"false"`
}

// HelmConfig represents Helm chart configuration for SD-Core.
type HelmConfig struct {
	LocalCharts  bool   `json:"local_charts" required:"false"`
	ChartRef     string `json:"chart_ref" required:"false"`
	ChartVersion string `json:"chart_version" required:"false"`
}

// UPFConfig represents User Plane Function configuration.
type UPFConfig struct {
	AccessSubnet string           `json:"access_subnet" required:"false"`
	CoreSubnet   string           `json:"core_subnet" required:"false"`
	Mode         string           `json:"mode" required:"false"`
	MultihopGNB  bool             `json:"multihop_gnb" required:"false"`
	DefaultUPF   DefaultUPFConfig `json:"default_upf" required:"false"`
}

// DefaultUPFConfig represents the default UPF instance configuration.
type DefaultUPFConfig struct {
	IP       UPFIPConfig `json:"ip" required:"false"`
	UEIPPool string      `json:"ue_ip_pool" required:"false"`
}

// UPFIPConfig represents UPF IP address configuration.
type UPFIPConfig struct {
	Access string `json:"access" required:"false"`
	Core   string `json:"core" required:"false"`
}

// AMFConfig represents Access and Mobility Management Function configuration.
type AMFConfig struct {
	IP string `json:"ip" required:"false"`
}

// CoreStatus represents the deployment status of the SD-Core.
type CoreStatus struct {
	ID           string            `json:"id"`
	Name         string            `json:"name,omitempty"`
	Host         string            `json:"host"`
	State        DeploymentState   `json:"state"`
	Message      string            `json:"message,omitempty"`
	DeployedAt   *time.Time        `json:"deployed_at,omitempty"`
	Version      string            `json:"version,omitempty"`
	Components   []ComponentStatus `json:"components,omitempty"`
}

// CoreList represents a list of SD-Core deployments.
type CoreList struct {
	Cores []CoreConfig `json:"cores"`
}

// CoreStatusList represents status for all SD-Core deployments.
type CoreStatusList struct {
	Cores []CoreStatus `json:"cores"`
}

// ComponentStatus represents the status of an individual component.
type ComponentStatus struct {
	Name   string `json:"name"`
	Ready  bool   `json:"ready"`
	Status string `json:"status"`
}

// GNBConfig represents a gNB (gNodeB) configuration.
type GNBConfig struct {
	ID         string          `json:"id" required:"false"`
	Host       string          `json:"host" required:"false"` // Target host where gNB is deployed
	Name       string          `json:"name,omitempty" required:"false"`
	Type       string          `json:"type" required:"false"` // srsran, ocudu
	IP         string          `json:"gnb_ip" required:"false"`
	ConfigFile string          `json:"gnb_conf,omitempty" required:"false"`
	Docker     GNBDockerConfig `json:"docker,omitempty" required:"false"`
	Simulation bool            `json:"simulation" required:"false"`
	UEConfig   *UEConfig       `json:"ue_config,omitempty" required:"false"` // Only for simulation mode
}

// GNBDockerConfig represents Docker configuration for gNB.
type GNBDockerConfig struct {
	Container GNBContainerConfig `json:"container" required:"false"`
	Network   GNBNetworkConfig   `json:"network" required:"false"`
}

// GNBContainerConfig represents container image configuration.
type GNBContainerConfig struct {
	GNBImage string `json:"gnb_image" required:"false"`
	UEImage  string `json:"ue_image,omitempty" required:"false"`
}

// GNBNetworkConfig represents Docker network configuration.
type GNBNetworkConfig struct {
	Name string `json:"name" required:"false"`
}

// UEConfig represents UE simulator configuration (for simulation mode).
type UEConfig struct {
	ConfigFile string `json:"ue_conf,omitempty" required:"false"`
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
	ID      string `json:"id,omitempty"`      // The generated ID for the deployed resource
	TaskID  string `json:"task_id,omitempty"` // For async operations
}
