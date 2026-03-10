package nodefacts

import (
	"context"
	"time"
)

// NodeFacts holds discovered network information about a managed node.
type NodeFacts struct {
	NodeID        string          `json:"node_id"`
	NodeName      string          `json:"node_name"`
	AnsibleHost   string          `json:"ansible_host"`
	DefaultIface  string          `json:"default_iface"`
	DefaultIP     string          `json:"default_ip"`
	DefaultSubnet string          `json:"default_subnet"`
	Interfaces    []InterfaceInfo `json:"interfaces"`
	GatheredAt    time.Time       `json:"gathered_at"`
	Error         string          `json:"error,omitempty"`
}

// InterfaceInfo describes a single network interface on a node.
type InterfaceInfo struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	MAC       string   `json:"mac"`
	IsUp      bool     `json:"is_up"`
}

// Gatherer discovers network facts from a remote node.
type Gatherer interface {
	Gather(ctx context.Context, host, user string, password string, sshKey []byte) (NodeFacts, error)
}
