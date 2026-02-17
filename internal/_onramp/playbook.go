package onramp

import "strings"

// PlaybookStep represents a single ansible-playbook invocation.
type PlaybookStep struct {
	Name     string   // human-readable step name
	Playbook string   // relative path from OnRamp root
	Tags     []string // e.g., ["install"]
}

// PlaybookSequence is an ordered list of playbook steps for an operation.
type PlaybookSequence struct {
	Name  string
	Steps []PlaybookStep
}

// TagString returns the tags joined with commas, or empty string if none.
func (s PlaybookStep) TagString() string {
	return strings.Join(s.Tags, ",")
}

// Sequences maps operation names to their playbook step sequences.
// Verified against the OnRamp Makefiles.
var Sequences = map[string]PlaybookSequence{
	"aether-pingall": {
		Name: "Ping all hosts",
		Steps: []PlaybookStep{
			{Name: "Ping all hosts", Playbook: "pingall.yml"},
		},
	},
	"k8s-install": {
		Name: "Install Kubernetes (RKE2 + Helm)",
		Steps: []PlaybookStep{
			{Name: "Install RKE2", Playbook: "deps/k8s/rke2.yml", Tags: []string{"install"}},
			{Name: "Install Helm", Playbook: "deps/k8s/helm.yml", Tags: []string{"install"}},
		},
	},
	"k8s-uninstall": {
		Name: "Uninstall Kubernetes",
		Steps: []PlaybookStep{
			{Name: "Uninstall Helm", Playbook: "deps/k8s/helm.yml", Tags: []string{"uninstall"}},
			{Name: "Uninstall RKE2", Playbook: "deps/k8s/rke2.yml", Tags: []string{"uninstall"}},
		},
	},
	"5gc-install": {
		Name: "Install 5G Core (SD-Core)",
		Steps: []PlaybookStep{
			{Name: "Install 5GC Router", Playbook: "deps/5gc/router.yml", Tags: []string{"install"}},
			{Name: "Install 5GC Core", Playbook: "deps/5gc/core.yml", Tags: []string{"install"}},
		},
	},
	"5gc-uninstall": {
		Name: "Uninstall 5G Core",
		Steps: []PlaybookStep{
			{Name: "Uninstall 5GC Core", Playbook: "deps/5gc/core.yml", Tags: []string{"uninstall"}},
			{Name: "Uninstall 5GC Router", Playbook: "deps/5gc/router.yml", Tags: []string{"uninstall"}},
		},
	},
	"srsran-gnb-install": {
		Name: "Install srsRAN gNB",
		Steps: []PlaybookStep{
			{Name: "Install Docker", Playbook: "deps/srsran/docker.yml", Tags: []string{"install"}},
			{Name: "Install srsRAN Router", Playbook: "deps/srsran/router.yml", Tags: []string{"install"}},
			{Name: "Start srsRAN gNB", Playbook: "deps/srsran/gNB.yml", Tags: []string{"start"}},
		},
	},
	"srsran-gnb-uninstall": {
		Name: "Uninstall srsRAN gNB",
		Steps: []PlaybookStep{
			{Name: "Stop srsRAN gNB", Playbook: "deps/srsran/gNB.yml", Tags: []string{"stop"}},
			{Name: "Uninstall srsRAN Router", Playbook: "deps/srsran/router.yml", Tags: []string{"uninstall"}},
		},
	},
}
