package nodefacts

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// mockRunner replays canned command responses for testing.
type mockRunner struct {
	responses map[string]mockResponse
}

type mockResponse struct {
	stdout   []byte
	stderr   []byte
	exitCode int
	err      error
}

func (m *mockRunner) Run(_ context.Context, cmd string) ([]byte, []byte, int, error) {
	resp, ok := m.responses[cmd]
	if !ok {
		return nil, []byte("command not found"), 127, nil
	}
	return resp.stdout, resp.stderr, resp.exitCode, resp.err
}

func (m *mockRunner) Close() error { return nil }

func newMockRunner() *mockRunner {
	routes := []ipRouteJSON{{Dev: "ens18", Prefsrc: "10.0.0.10"}}
	routeJSON, _ := json.Marshal(routes)

	addrs := []struct {
		Ifname   string     `json:"ifname"`
		AddrInfo []addrInfo `json:"addr_info"`
	}{
		{
			Ifname: "lo",
			AddrInfo: []addrInfo{
				{Local: "127.0.0.1", Prefixlen: 8},
			},
		},
		{
			Ifname: "ens18",
			AddrInfo: []addrInfo{
				{Local: "10.0.0.10", Prefixlen: 24},
			},
		},
	}
	addrJSON, _ := json.Marshal(addrs)

	links := []ipLinkJSON{
		{Ifname: "lo", Address: "00:00:00:00:00:00", Operstate: "UNKNOWN"},
		{Ifname: "ens18", Address: "aa:bb:cc:dd:ee:ff", Operstate: "UP"},
	}
	linkJSON, _ := json.Marshal(links)

	ens18Addr := []ipAddrJSON{
		{AddrInfo: []struct {
			Local     string `json:"local"`
			Prefixlen int    `json:"prefixlen"`
		}{{Local: "10.0.0.10", Prefixlen: 24}}},
	}
	ens18AddrJSON, _ := json.Marshal(ens18Addr)

	return &mockRunner{
		responses: map[string]mockResponse{
			"ip -j route show default": {stdout: routeJSON},
			"ip -j addr show dev ens18": {stdout: ens18AddrJSON},
			"ip -j addr show":          {stdout: addrJSON},
			"ip -j link show":          {stdout: linkJSON},
		},
	}
}

type addrInfo struct {
	Local     string `json:"local"`
	Prefixlen int    `json:"prefixlen"`
}

func TestGatherFromRunner(t *testing.T) {
	r := newMockRunner()
	facts, err := gatherFromRunner(t.Context(), r, "10.0.0.10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if facts.DefaultIface != "ens18" {
		t.Errorf("DefaultIface = %q, want %q", facts.DefaultIface, "ens18")
	}
	if facts.DefaultIP != "10.0.0.10" {
		t.Errorf("DefaultIP = %q, want %q", facts.DefaultIP, "10.0.0.10")
	}
	if facts.DefaultSubnet != "10.0.0.0/24" {
		t.Errorf("DefaultSubnet = %q, want %q", facts.DefaultSubnet, "10.0.0.0/24")
	}
	if facts.AnsibleHost != "10.0.0.10" {
		t.Errorf("AnsibleHost = %q, want %q", facts.AnsibleHost, "10.0.0.10")
	}
	if facts.Error != "" {
		t.Errorf("Error = %q, want empty", facts.Error)
	}
	if len(facts.Interfaces) != 2 {
		t.Fatalf("Interfaces count = %d, want 2", len(facts.Interfaces))
	}
	ens := facts.Interfaces[1]
	if ens.Name != "ens18" || ens.MAC != "aa:bb:cc:dd:ee:ff" || !ens.IsUp {
		t.Errorf("unexpected ens18 info: %+v", ens)
	}
}

func TestParseDefaultRouteText(t *testing.T) {
	tests := []struct {
		input     string
		wantIface string
		wantIP    string
		wantErr   bool
	}{
		{
			input:     "default via 10.0.0.1 dev ens18 proto kernel src 10.0.0.10",
			wantIface: "ens18",
			wantIP:    "10.0.0.10",
		},
		{
			input:     "default via 192.168.1.1 dev eth0",
			wantIface: "eth0",
			wantIP:    "",
		},
		{
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			iface, ip, err := parseDefaultRouteText(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if iface != tt.wantIface {
				t.Errorf("iface = %q, want %q", iface, tt.wantIface)
			}
			if ip != tt.wantIP {
				t.Errorf("ip = %q, want %q", ip, tt.wantIP)
			}
		})
	}
}

func TestParseSubnetText(t *testing.T) {
	output := `2: ens18: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500
    inet 10.0.0.10/24 brd 10.0.0.255 scope global ens18
    inet6 fe80::1/64 scope link`

	subnet, err := parseSubnetText(output, "10.0.0.10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subnet != "10.0.0.0/24" {
		t.Errorf("subnet = %q, want %q", subnet, "10.0.0.0/24")
	}
}

func TestGatherFallbackToText(t *testing.T) {
	r := &mockRunner{
		responses: map[string]mockResponse{
			"ip -j route show default": {exitCode: 1}, // JSON fails
			"ip route show default":    {stdout: []byte("default via 10.0.0.1 dev eth0 proto kernel src 10.0.0.5")},
			"ip -j addr show dev eth0": {exitCode: 1}, // JSON fails
			"ip addr show dev eth0": {stdout: []byte(`2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP>
    inet 10.0.0.5/24 brd 10.0.0.255 scope global eth0`)},
			"ip -j link show": {exitCode: 1}, // JSON fails
			"ip -j addr show": {exitCode: 1}, // JSON fails
		},
	}

	facts, err := gatherFromRunner(t.Context(), r, "10.0.0.5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if facts.DefaultIface != "eth0" {
		t.Errorf("DefaultIface = %q, want %q", facts.DefaultIface, "eth0")
	}
	if facts.DefaultIP != "10.0.0.5" {
		t.Errorf("DefaultIP = %q, want %q", facts.DefaultIP, "10.0.0.5")
	}
	if facts.DefaultSubnet != "10.0.0.0/24" {
		t.Errorf("DefaultSubnet = %q, want %q", facts.DefaultSubnet, "10.0.0.0/24")
	}
}

func TestGatherUnreachableNode(t *testing.T) {
	r := &mockRunner{
		responses: map[string]mockResponse{
			"ip -j route show default": {err: fmt.Errorf("connection refused")},
		},
	}

	facts, err := gatherFromRunner(t.Context(), r, "10.0.0.99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Error is reported in the facts, not as a return error.
	if facts.Error == "" {
		t.Error("expected non-empty Error field for unreachable node")
	}
}
