package configdefaults

import (
	"github.com/bengrewell/aether-webui/internal/nodefacts"
)

// Rule maps a config field to a function that computes its default value
// from discovered node facts.
type Rule struct {
	Component string   // "core", "gnbsim", "ueransim", etc.
	Field     string   // dotted config path, e.g. "core.data_iface"
	Label     string   // human-readable description
	Roles     []string // node roles that trigger this rule
	ComputeFn func(facts nodefacts.NodeFacts) (value any, explanation string)
}

// defaultRules is the initial set of rules for config field defaulting.
var defaultRules = []Rule{
	{
		Component: "core",
		Field:     "core.data_iface",
		Label:     "Core data interface",
		Roles:     []string{"master"},
		ComputeFn: func(f nodefacts.NodeFacts) (any, string) {
			if f.DefaultIface == "" {
				return nil, ""
			}
			return f.DefaultIface, "default route interface on " + f.NodeName
		},
	},
	{
		Component: "core",
		Field:     "core.ran_subnet",
		Label:     "Core RAN subnet",
		Roles:     []string{"master"},
		ComputeFn: func(f nodefacts.NodeFacts) (any, string) {
			return "", "empty = auto-detect from data interface"
		},
	},
	{
		Component: "core",
		Field:     "core.amf.ip",
		Label:     "AMF IP address",
		Roles:     []string{"master"},
		ComputeFn: func(f nodefacts.NodeFacts) (any, string) {
			if f.DefaultIP == "" {
				return nil, ""
			}
			return f.DefaultIP, "primary IP on " + f.DefaultIface
		},
	},
	{
		Component: "gnbsim",
		Field:     "gnbsim.router.data_iface",
		Label:     "gNBSim router data interface",
		Roles:     []string{"gnbsim"},
		ComputeFn: func(f nodefacts.NodeFacts) (any, string) {
			if f.DefaultIface == "" {
				return nil, ""
			}
			return f.DefaultIface, "default route interface on " + f.NodeName
		},
	},
	{
		Component: "ueransim",
		Field:     "ueransim.gnb.ip",
		Label:     "UERANSIM gNB IP",
		Roles:     []string{"ueransim"},
		ComputeFn: func(f nodefacts.NodeFacts) (any, string) {
			if f.DefaultIP == "" {
				return nil, ""
			}
			return f.DefaultIP, "primary IP on " + f.DefaultIface
		},
	},
	{
		Component: "oai",
		Field:     "oai.docker.network.data_iface",
		Label:     "OAI Docker network data interface",
		Roles:     []string{"oai"},
		ComputeFn: func(f nodefacts.NodeFacts) (any, string) {
			if f.DefaultIface == "" {
				return nil, ""
			}
			return f.DefaultIface, "default route interface on " + f.NodeName
		},
	},
}

// matchesRole returns true if the node has any of the required roles.
func matchesRole(nodeRoles []string, ruleRoles []string) bool {
	for _, rr := range ruleRoles {
		for _, nr := range nodeRoles {
			if nr == rr {
				return true
			}
		}
	}
	return false
}
