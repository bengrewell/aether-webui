package onramp

import (
	"time"

	"github.com/bengrewell/aether-webui/internal/taskrunner"
)

// ---------------------------------------------------------------------------
// Component registry
// ---------------------------------------------------------------------------

// Component describes a deployable OnRamp component and its available actions.
type Component struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Actions     []Action `json:"actions"`
}

// Action maps a human-readable operation to a Makefile target.
type Action struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Target      string `json:"target"`
}

// componentRegistry is the static set of components derived from the OnRamp Makefile.
var componentRegistry = []Component{
	{
		Name:        "k8s",
		Description: "Kubernetes (RKE2) cluster lifecycle",
		Actions: []Action{
			{Name: "install", Description: "Deploy Kubernetes (RKE2)", Target: "aether-k8s-install"},
			{Name: "uninstall", Description: "Remove Kubernetes (RKE2)", Target: "aether-k8s-uninstall"},
		},
	},
	{
		Name:        "5gc",
		Description: "5G core network (SD-Core)",
		Actions: []Action{
			{Name: "install", Description: "Deploy 5G core", Target: "aether-5gc-install"},
			{Name: "uninstall", Description: "Remove 5G core", Target: "aether-5gc-uninstall"},
			{Name: "reset", Description: "Reset 5G core state", Target: "aether-5gc-reset"},
		},
	},
	{
		Name:        "4gc",
		Description: "4G core network",
		Actions: []Action{
			{Name: "install", Description: "Deploy 4G core", Target: "aether-4gc-install"},
			{Name: "uninstall", Description: "Remove 4G core", Target: "aether-4gc-uninstall"},
			{Name: "reset", Description: "Reset 4G core state", Target: "aether-4gc-reset"},
		},
	},
	{
		Name:        "gnbsim",
		Description: "gNBSim simulated RAN",
		Actions: []Action{
			{Name: "install", Description: "Deploy gNBSim", Target: "aether-gnbsim-install"},
			{Name: "uninstall", Description: "Remove gNBSim", Target: "aether-gnbsim-uninstall"},
			{Name: "run", Description: "Run gNBSim simulation", Target: "aether-gnbsim-run"},
		},
	},
	{
		Name:        "amp",
		Description: "Aether Management Platform",
		Actions: []Action{
			{Name: "install", Description: "Deploy AMP", Target: "aether-amp-install"},
			{Name: "uninstall", Description: "Remove AMP", Target: "aether-amp-uninstall"},
		},
	},
	{
		Name:        "sdran",
		Description: "SD-RAN intelligent RAN controller",
		Actions: []Action{
			{Name: "install", Description: "Deploy SD-RAN", Target: "aether-sdran-install"},
			{Name: "uninstall", Description: "Remove SD-RAN", Target: "aether-sdran-uninstall"},
		},
	},
	{
		Name:        "ueransim",
		Description: "UERANSIM UE and gNB simulator",
		Actions: []Action{
			{Name: "install", Description: "Deploy UERANSIM", Target: "aether-ueransim-install"},
			{Name: "uninstall", Description: "Remove UERANSIM", Target: "aether-ueransim-uninstall"},
			{Name: "run", Description: "Start UERANSIM simulation", Target: "aether-ueransim-run"},
			{Name: "stop", Description: "Stop UERANSIM simulation", Target: "aether-ueransim-stop"},
		},
	},
	{
		Name:        "oai",
		Description: "OpenAirInterface RAN",
		Actions: []Action{
			{Name: "gnb-install", Description: "Deploy OAI gNB", Target: "aether-oai-gnb-install"},
			{Name: "gnb-uninstall", Description: "Remove OAI gNB", Target: "aether-oai-gnb-uninstall"},
			{Name: "uesim-start", Description: "Start OAI UE simulator", Target: "aether-oai-uesim-start"},
			{Name: "uesim-stop", Description: "Stop OAI UE simulator", Target: "aether-oai-uesim-stop"},
		},
	},
	{
		Name:        "srsran",
		Description: "srsRAN Project RAN",
		Actions: []Action{
			{Name: "gnb-install", Description: "Deploy srsRAN gNB", Target: "aether-srsran-gnb-install"},
			{Name: "gnb-uninstall", Description: "Remove srsRAN gNB", Target: "aether-srsran-gnb-uninstall"},
			{Name: "uesim-start", Description: "Start srsRAN UE simulator", Target: "aether-srsran-uesim-start"},
			{Name: "uesim-stop", Description: "Stop srsRAN UE simulator", Target: "aether-srsran-uesim-stop"},
		},
	},
	{
		Name:        "oscric",
		Description: "O-RAN SC near-RT RIC",
		Actions: []Action{
			{Name: "ric-install", Description: "Deploy OSC near-RT RIC", Target: "aether-oscric-ric-install"},
			{Name: "ric-uninstall", Description: "Remove OSC near-RT RIC", Target: "aether-oscric-ric-uninstall"},
		},
	},
	{
		Name:        "n3iwf",
		Description: "Non-3GPP Interworking Function",
		Actions: []Action{
			{Name: "install", Description: "Deploy N3IWF", Target: "aether-n3iwf-install"},
			{Name: "uninstall", Description: "Remove N3IWF", Target: "aether-n3iwf-uninstall"},
		},
	},
	{
		Name:        "cluster",
		Description: "Cluster-level operations",
		Actions: []Action{
			{Name: "pingall", Description: "Ping all cluster nodes", Target: "aether-pingall"},
			{Name: "install", Description: "Deploy full Aether stack", Target: "aether-install"},
			{Name: "uninstall", Description: "Remove full Aether stack", Target: "aether-uninstall"},
			{Name: "add-upfs", Description: "Add additional UPFs", Target: "aether-add-upfs"},
			{Name: "remove-upfs", Description: "Remove additional UPFs", Target: "aether-remove-upfs"},
		},
	},
}

// componentIndex provides O(1) lookup by component name.
var componentIndex map[string]*Component

func init() {
	componentIndex = make(map[string]*Component, len(componentRegistry))
	for i := range componentRegistry {
		componentIndex[componentRegistry[i].Name] = &componentRegistry[i]
	}
}

// ---------------------------------------------------------------------------
// Task tracking
// ---------------------------------------------------------------------------

// OnRampTask is the API-facing representation of a make target execution.
// It preserves the JSON shape of the previous Task type and adds an
// output_offset field for incremental streaming support.
type OnRampTask struct {
	ID           string    `json:"id"`
	Component    string    `json:"component"`
	Action       string    `json:"action"`
	Target       string    `json:"target"`
	Status       string    `json:"status"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at,omitempty"`
	ExitCode     int       `json:"exit_code,omitempty"`
	Output       string    `json:"output"`
	OutputOffset int       `json:"output_offset"`
}

// toOnRampTask converts a TaskView and output chunk into the OnRamp-specific
// task representation. Component, action, and target are extracted from the
// task's labels.
func toOnRampTask(view taskrunner.TaskView, output string, outputOffset int) OnRampTask {
	return OnRampTask{
		ID:           view.ID,
		Component:    view.Labels["component"],
		Action:       view.Labels["action"],
		Target:       view.Labels["target"],
		Status:       string(view.Status),
		StartedAt:    view.StartedAt,
		FinishedAt:   view.FinishedAt,
		ExitCode:     view.ExitCode,
		Output:       output,
		OutputOffset: outputOffset,
	}
}

// ---------------------------------------------------------------------------
// Config types — typed representation of OnRamp vars/main.yml
// ---------------------------------------------------------------------------

// OnRampConfig is the superset of all vars file sections.
type OnRampConfig struct {
	K8s      *K8sConfig      `json:"k8s,omitempty"      yaml:"k8s,omitempty"`
	Core     *CoreConfig     `json:"core,omitempty"      yaml:"core,omitempty"`
	GNBSim   *GNBSimConfig   `json:"gnbsim,omitempty"    yaml:"gnbsim,omitempty"`
	AMP      *AMPConfig      `json:"amp,omitempty"       yaml:"amp,omitempty"`
	SDRAN    *SDRANConfig    `json:"sdran,omitempty"     yaml:"sdran,omitempty"`
	UERANSIM *UERANSIMConfig `json:"ueransim,omitempty"  yaml:"ueransim,omitempty"`
	OAI      *OAIConfig      `json:"oai,omitempty"       yaml:"oai,omitempty"`
	SRSRan   *SRSRanConfig   `json:"srsran,omitempty"    yaml:"srsran,omitempty"`
	N3IWF    *N3IWFConfig    `json:"n3iwf,omitempty"     yaml:"n3iwf,omitempty"`
}

// --- K8s ---

type K8sConfig struct {
	RKE2 *RKE2Config `json:"rke2,omitempty" yaml:"rke2,omitempty"`
	Helm *HelmRef    `json:"helm,omitempty" yaml:"helm,omitempty"`
}

type RKE2Config struct {
	Version string     `json:"version" yaml:"version"`
	Config  *RKE2Inner `json:"config,omitempty" yaml:"config,omitempty"`
}

type RKE2Inner struct {
	Token      string      `json:"token" yaml:"token"`
	Port       int         `json:"port" yaml:"port"`
	ParamsFile *ParamsFile `json:"params_file,omitempty" yaml:"params_file,omitempty"`
}

type ParamsFile struct {
	Master string `json:"master" yaml:"master"`
	Worker string `json:"worker" yaml:"worker"`
}

type HelmRef struct {
	Version      string `json:"version,omitempty" yaml:"version,omitempty"`
	LocalCharts  *bool  `json:"local_charts,omitempty" yaml:"local_charts,omitempty"`
	ChartRef     string `json:"chart_ref,omitempty" yaml:"chart_ref,omitempty"`
	ChartVersion string `json:"chart_version,omitempty" yaml:"chart_version,omitempty"`
}

// --- Core ---

type CoreConfig struct {
	Standalone *bool      `json:"standalone,omitempty" yaml:"standalone,omitempty"`
	DataIface  string     `json:"data_iface,omitempty" yaml:"data_iface,omitempty"`
	ValuesFile string     `json:"values_file,omitempty" yaml:"values_file,omitempty"`
	RANSubnet  string     `json:"ran_subnet,omitempty" yaml:"ran_subnet,omitempty"`
	Helm       *HelmRef   `json:"helm,omitempty" yaml:"helm,omitempty"`
	UPF        *UPFConfig `json:"upf,omitempty" yaml:"upf,omitempty"`
	AMF        *AMFConfig `json:"amf,omitempty" yaml:"amf,omitempty"`
	MME        *MMEConfig `json:"mme,omitempty" yaml:"mme,omitempty"`
}

type UPFConfig struct {
	AccessSubnet   string                 `json:"access_subnet,omitempty" yaml:"access_subnet,omitempty"`
	CoreSubnet     string                 `json:"core_subnet,omitempty" yaml:"core_subnet,omitempty"`
	Mode           string                 `json:"mode,omitempty" yaml:"mode,omitempty"`
	MultihopGNB    *bool                  `json:"multihop_gnb,omitempty" yaml:"multihop_gnb,omitempty"`
	Helm           *HelmRef               `json:"helm,omitempty" yaml:"helm,omitempty"`
	ValuesFile     string                 `json:"values_file,omitempty" yaml:"values_file,omitempty"`
	DefaultUPF     *UPFInstance           `json:"default_upf,omitempty" yaml:"default_upf,omitempty"`
	AdditionalUPFs map[string]*UPFInstance `json:"additional_upfs,omitempty" yaml:"additional_upfs,omitempty"`
}

type UPFInstance struct {
	IP       *UPFIP `json:"ip,omitempty" yaml:"ip,omitempty"`
	UEIPPool string `json:"ue_ip_pool,omitempty" yaml:"ue_ip_pool,omitempty"`
}

type UPFIP struct {
	Access string `json:"access" yaml:"access"`
	Core   string `json:"core" yaml:"core"`
}

type AMFConfig struct {
	IP string `json:"ip" yaml:"ip"`
}

type MMEConfig struct {
	IP string `json:"ip" yaml:"ip"`
}

// --- GNBSim ---

type GNBSimConfig struct {
	Docker  *GNBSimDocker    `json:"docker,omitempty" yaml:"docker,omitempty"`
	Router  *GNBSimRouter    `json:"router,omitempty" yaml:"router,omitempty"`
	Servers map[int][]string `json:"servers,omitempty" yaml:"servers,omitempty"`
}

type GNBSimDocker struct {
	Container *GNBSimContainer `json:"container,omitempty" yaml:"container,omitempty"`
	Network   *GNBSimNetwork   `json:"network,omitempty" yaml:"network,omitempty"`
}

type GNBSimContainer struct {
	Image  string `json:"image" yaml:"image"`
	Prefix string `json:"prefix" yaml:"prefix"`
	Count  int    `json:"count" yaml:"count"`
}

type GNBSimNetwork struct {
	Macvlan *GNBSimMacvlan `json:"macvlan,omitempty" yaml:"macvlan,omitempty"`
}

type GNBSimMacvlan struct {
	Name string `json:"name" yaml:"name"`
}

type GNBSimRouter struct {
	DataIface string              `json:"data_iface,omitempty" yaml:"data_iface,omitempty"`
	Macvlan   *GNBSimRouterMacvlan `json:"macvlan,omitempty" yaml:"macvlan,omitempty"`
}

type GNBSimRouterMacvlan struct {
	SubnetPrefix string `json:"subnet_prefix" yaml:"subnet_prefix"`
}

// --- AMP ---

type AMPConfig struct {
	ROCModels        string       `json:"roc_models,omitempty" yaml:"roc_models,omitempty"`
	MonitorDashboard string       `json:"monitor_dashboard,omitempty" yaml:"monitor_dashboard,omitempty"`
	AetherROC        *HelmBlock   `json:"aether_roc,omitempty" yaml:"aether_roc,omitempty"`
	Atomix           *HelmBlock   `json:"atomix,omitempty" yaml:"atomix,omitempty"`
	ONOSProject      *HelmBlock   `json:"onosproject,omitempty" yaml:"onosproject,omitempty"`
	Store            *StoreConfig `json:"store,omitempty" yaml:"store,omitempty"`
	Monitor          *HelmBlock   `json:"monitor,omitempty" yaml:"monitor,omitempty"`
	MonitorCRD       *HelmBlock   `json:"monitor_crd,omitempty" yaml:"monitor_crd,omitempty"`
}

type HelmBlock struct {
	Helm *HelmRef `json:"helm,omitempty" yaml:"helm,omitempty"`
}

type StoreConfig struct {
	LPP *LPPConfig `json:"lpp,omitempty" yaml:"lpp,omitempty"`
}

type LPPConfig struct {
	Version string `json:"version" yaml:"version"`
}

// --- SDRAN ---

type SDRANConfig struct {
	Platform *SDRANPlatform `json:"platform,omitempty" yaml:"platform,omitempty"`
	SDRAN    *SDRANInner    `json:"sdran,omitempty" yaml:"sdran,omitempty"`
}

type SDRANPlatform struct {
	Atomix      *HelmBlock   `json:"atomix,omitempty" yaml:"atomix,omitempty"`
	ONOSProject *HelmBlock   `json:"onosproject,omitempty" yaml:"onosproject,omitempty"`
	Store       *StoreConfig `json:"store,omitempty" yaml:"store,omitempty"`
}

type SDRANInner struct {
	Helm   *HelmRef      `json:"helm,omitempty" yaml:"helm,omitempty"`
	Import *SDRANImport  `json:"import,omitempty" yaml:"import,omitempty"`
	RANSim *RANSimConfig `json:"ransim,omitempty" yaml:"ransim,omitempty"`
}

type SDRANImport struct {
	E2T    *bool `json:"e2t,omitempty" yaml:"e2t,omitempty"`
	A1T    *bool `json:"a1t,omitempty" yaml:"a1t,omitempty"`
	UENIB  *bool `json:"uenib,omitempty" yaml:"uenib,omitempty"`
	Topo   *bool `json:"topo,omitempty" yaml:"topo,omitempty"`
	Config *bool `json:"config,omitempty" yaml:"config,omitempty"`
	RANSim *bool `json:"ransim,omitempty" yaml:"ransim,omitempty"`
	KPIMon *bool `json:"kpimon,omitempty" yaml:"kpimon,omitempty"`
	PCI    *bool `json:"pci,omitempty" yaml:"pci,omitempty"`
	MHO    *bool `json:"mho,omitempty" yaml:"mho,omitempty"`
	MLB    *bool `json:"mlb,omitempty" yaml:"mlb,omitempty"`
	TS     *bool `json:"ts,omitempty" yaml:"ts,omitempty"`
}

type RANSimConfig struct {
	Model  string `json:"model,omitempty" yaml:"model,omitempty"`
	Metric string `json:"metric,omitempty" yaml:"metric,omitempty"`
}

// --- UERANSIM ---

type UERANSIMConfig struct {
	GNB     *UERANSIMGnb                `json:"gnb,omitempty" yaml:"gnb,omitempty"`
	Servers map[int]*UERANSIMServer `json:"servers,omitempty" yaml:"servers,omitempty"`
}

type UERANSIMGnb struct {
	IP string `json:"ip" yaml:"ip"`
}

type UERANSIMServer struct {
	GNB string `json:"gnb" yaml:"gnb"`
	UE  string `json:"ue" yaml:"ue"`
}

// --- OAI ---

type OAIConfig struct {
	Docker     *OAIDocker         `json:"docker,omitempty" yaml:"docker,omitempty"`
	Simulation *bool              `json:"simulation,omitempty" yaml:"simulation,omitempty"`
	Servers    map[int]*OAIServer `json:"servers,omitempty" yaml:"servers,omitempty"`
}

type OAIDocker struct {
	Container *OAIContainer `json:"container,omitempty" yaml:"container,omitempty"`
	Network   *OAINetwork   `json:"network,omitempty" yaml:"network,omitempty"`
}

type OAIContainer struct {
	GNBImage string `json:"gnb_image" yaml:"gnb_image"`
	UEImage  string `json:"ue_image" yaml:"ue_image"`
}

type OAINetwork struct {
	DataIface string     `json:"data_iface,omitempty" yaml:"data_iface,omitempty"`
	Name      string     `json:"name,omitempty" yaml:"name,omitempty"`
	Subnet    string     `json:"subnet,omitempty" yaml:"subnet,omitempty"`
	Bridge    *OAIBridge `json:"bridge,omitempty" yaml:"bridge,omitempty"`
}

type OAIBridge struct {
	Name string `json:"name" yaml:"name"`
}

type OAIServer struct {
	GNBConf string `json:"gnb_conf" yaml:"gnb_conf"`
	GNBIP   string `json:"gnb_ip" yaml:"gnb_ip"`
	UEConf  string `json:"ue_conf" yaml:"ue_conf"`
}

// --- srsRAN ---

type SRSRanConfig struct {
	Docker     *SRSRanDocker         `json:"docker,omitempty" yaml:"docker,omitempty"`
	Simulation *bool                 `json:"simulation,omitempty" yaml:"simulation,omitempty"`
	Servers    map[int]*SRSRanServer `json:"servers,omitempty" yaml:"servers,omitempty"`
}

type SRSRanDocker struct {
	Container *SRSRanContainer `json:"container,omitempty" yaml:"container,omitempty"`
	Network   *SRSRanNetwork   `json:"network,omitempty" yaml:"network,omitempty"`
}

type SRSRanContainer struct {
	GNBImage string `json:"gnb_image" yaml:"gnb_image"`
	UEImage  string `json:"ue_image" yaml:"ue_image"`
}

type SRSRanNetwork struct {
	Name string `json:"name" yaml:"name"`
}

type SRSRanServer struct {
	GNBIP   string `json:"gnb_ip" yaml:"gnb_ip"`
	GNBConf string `json:"gnb_conf" yaml:"gnb_conf"`
	UEConf  string `json:"ue_conf" yaml:"ue_conf"`
}

// --- N3IWF ---

type N3IWFConfig struct {
	Docker  *N3IWFDocker         `json:"docker,omitempty" yaml:"docker,omitempty"`
	Servers map[int]*N3IWFServer `json:"servers,omitempty" yaml:"servers,omitempty"`
}

type N3IWFDocker struct {
	Image   string         `json:"image,omitempty" yaml:"image,omitempty"`
	Network *N3IWFNetwork  `json:"network,omitempty" yaml:"network,omitempty"`
}

type N3IWFNetwork struct {
	Name string `json:"name" yaml:"name"`
}

type N3IWFServer struct {
	ConfFile string `json:"conf_file" yaml:"conf_file"`
	N3IWFIP  string `json:"n3iwf_ip" yaml:"n3iwf_ip"`
	N2IP     string `json:"n2_ip" yaml:"n2_ip"`
	N3IP     string `json:"n3_ip" yaml:"n3_ip"`
	NWUIP    string `json:"nwu_ip" yaml:"nwu_ip"`
}

// ---------------------------------------------------------------------------
// Repo status
// ---------------------------------------------------------------------------

// RepoStatus describes the current state of the cloned OnRamp repository.
type RepoStatus struct {
	Cloned    bool   `json:"cloned"`
	Dir       string `json:"dir"`
	RepoURL   string `json:"repo_url"`
	Version   string `json:"version"`
	Commit    string `json:"commit,omitempty"`
	Branch    string `json:"branch,omitempty"`
	Tag       string `json:"tag,omitempty"`
	Dirty     bool   `json:"dirty"`
	Error     string `json:"error,omitempty"`
}

// ---------------------------------------------------------------------------
// Huma I/O types
// ---------------------------------------------------------------------------

// --- Repo ---

type RepoStatusOutput struct {
	Body RepoStatus
}

type RepoRefreshOutput struct {
	Body RepoStatus
}

// --- Components ---

type ComponentListOutput struct {
	Body []Component
}

type ComponentGetInput struct {
	Component string `path:"component" doc:"Component name"`
}

type ComponentGetOutput struct {
	Body Component
}

type ExecuteActionInput struct {
	Component string `path:"component" doc:"Component name"`
	Action    string `path:"action" doc:"Action name"`
}

type ExecuteActionOutput struct {
	Body OnRampTask
}

// --- Tasks ---

type TaskListOutput struct {
	Body []OnRampTask
}

type TaskGetInput struct {
	ID     string `path:"id" doc:"Task ID"`
	Offset int    `query:"offset" default:"0" doc:"Byte offset for incremental output reads"`
}

type TaskGetOutput struct {
	Body OnRampTask
}

// --- Config ---

type ConfigGetOutput struct {
	Body OnRampConfig
}

type ConfigPatchInput struct {
	Body OnRampConfig
}

type ConfigPatchOutput struct {
	Body OnRampConfig
}

// --- Profiles ---

type ProfileListOutput struct {
	Body []string
}

type ProfileGetInput struct {
	Name string `path:"name" doc:"Profile name"`
}

type ProfileGetOutput struct {
	Body OnRampConfig
}

type ProfileActivateInput struct {
	Name string `path:"name" doc:"Profile name"`
}

type ProfileActivateOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

// ---------------------------------------------------------------------------
// Inventory types
// ---------------------------------------------------------------------------

type InventoryData struct {
	Nodes []InventoryNode `json:"nodes"`
}

type InventoryNode struct {
	Name        string   `json:"name"`
	AnsibleHost string   `json:"ansible_host"`
	AnsibleUser string   `json:"ansible_user"`
	Roles       []string `json:"roles"`
}

type InventoryGetOutput struct {
	Body InventoryData
}

type InventorySyncOutput struct {
	Body struct {
		Message string `json:"message"`
		Path    string `json:"path"`
	}
}
