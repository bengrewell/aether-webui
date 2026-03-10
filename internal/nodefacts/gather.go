package nodefacts

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	internalssh "github.com/bengrewell/aether-webui/internal/ssh"
)

// Runner abstracts command execution on a remote host.
// The SSH gatherer uses this interface, enabling mock testing.
type Runner interface {
	Run(ctx context.Context, cmd string) (stdout, stderr []byte, exitCode int, err error)
	Close() error
}

// SSHGatherer discovers node facts by SSHing into the target.
type SSHGatherer struct {
	Timeout time.Duration // SSH dial timeout; defaults to 10s
}

// Gather connects to the node via SSH, runs discovery commands, and returns
// structured facts. Falls back to text parsing if `ip -j` is unavailable.
func (g *SSHGatherer) Gather(ctx context.Context, host, user string, password string, sshKey []byte) (NodeFacts, error) {
	timeout := g.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	client, err := internalssh.Dial(ctx, internalssh.Config{
		Host:     host,
		User:     user,
		Password: password,
		Key:      sshKey,
		Timeout:  timeout,
	})
	if err != nil {
		return NodeFacts{}, fmt.Errorf("nodefacts: dial %s: %w", host, err)
	}
	defer client.Close()

	return gatherFromRunner(ctx, client, host)
}

// gatherFromRunner runs discovery commands using the provided Runner.
func gatherFromRunner(ctx context.Context, r Runner, host string) (NodeFacts, error) {
	facts := NodeFacts{
		AnsibleHost: host,
		GatheredAt:  time.Now().UTC(),
	}

	// Discover default route interface.
	iface, ip, err := discoverDefaultRoute(ctx, r)
	if err != nil {
		facts.Error = fmt.Sprintf("default route: %v", err)
		return facts, nil
	}
	facts.DefaultIface = iface
	facts.DefaultIP = ip

	// Discover subnet for the default interface.
	subnet, err := discoverSubnet(ctx, r, iface, ip)
	if err == nil {
		facts.DefaultSubnet = subnet
	}

	// Discover all interfaces.
	ifaces, err := discoverInterfaces(ctx, r)
	if err == nil {
		facts.Interfaces = ifaces
	}

	return facts, nil
}

// --- Default route discovery ---

// ipRouteJSON is the JSON structure returned by `ip -j route show default`.
type ipRouteJSON struct {
	Dev     string `json:"dev"`
	Prefsrc string `json:"prefsrc"`
}

func discoverDefaultRoute(ctx context.Context, r Runner) (iface, ip string, err error) {
	// Try JSON output first.
	stdout, _, exitCode, err := r.Run(ctx, "ip -j route show default")
	if err != nil {
		return "", "", err
	}
	if exitCode == 0 && len(stdout) > 0 {
		var routes []ipRouteJSON
		if json.Unmarshal(stdout, &routes) == nil && len(routes) > 0 {
			return routes[0].Dev, routes[0].Prefsrc, nil
		}
	}

	// Fallback: text parsing of `ip route show default`.
	stdout, _, exitCode, err = r.Run(ctx, "ip route show default")
	if err != nil {
		return "", "", err
	}
	if exitCode != 0 {
		return "", "", fmt.Errorf("ip route exited %d", exitCode)
	}
	return parseDefaultRouteText(string(stdout))
}

// parseDefaultRouteText extracts interface and source IP from text like:
// "default via 10.0.0.1 dev ens18 proto kernel src 10.0.0.10"
func parseDefaultRouteText(output string) (iface, ip string, err error) {
	fields := strings.Fields(output)
	for i, f := range fields {
		if f == "dev" && i+1 < len(fields) {
			iface = fields[i+1]
		}
		if f == "src" && i+1 < len(fields) {
			ip = fields[i+1]
		}
	}
	if iface == "" {
		return "", "", fmt.Errorf("could not parse default route interface")
	}
	return iface, ip, nil
}

// --- Subnet discovery ---

type ipAddrJSON struct {
	AddrInfo []struct {
		Local     string `json:"local"`
		Prefixlen int    `json:"prefixlen"`
	} `json:"addr_info"`
}

func discoverSubnet(ctx context.Context, r Runner, iface, ipAddr string) (string, error) {
	stdout, _, exitCode, err := r.Run(ctx, "ip -j addr show dev "+iface)
	if err != nil {
		return "", err
	}
	if exitCode == 0 && len(stdout) > 0 {
		var addrs []ipAddrJSON
		if json.Unmarshal(stdout, &addrs) == nil {
			for _, a := range addrs {
				for _, ai := range a.AddrInfo {
					if ai.Local == ipAddr {
						_, cidr, err := net.ParseCIDR(fmt.Sprintf("%s/%d", ai.Local, ai.Prefixlen))
						if err == nil {
							return cidr.String(), nil
						}
					}
				}
			}
		}
	}

	// Fallback: text parsing.
	stdout, _, exitCode, err = r.Run(ctx, "ip addr show dev "+iface)
	if err != nil {
		return "", err
	}
	if exitCode != 0 {
		return "", fmt.Errorf("ip addr exited %d", exitCode)
	}
	return parseSubnetText(string(stdout), ipAddr)
}

func parseSubnetText(output, ipAddr string) (string, error) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "inet ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		cidrStr := fields[1] // e.g. "10.0.0.10/24"
		ip, cidr, err := net.ParseCIDR(cidrStr)
		if err != nil {
			continue
		}
		if ip.String() == ipAddr {
			return cidr.String(), nil
		}
	}
	return "", fmt.Errorf("subnet not found for %s", ipAddr)
}

// --- Interface discovery ---

type ipLinkJSON struct {
	Ifname   string `json:"ifname"`
	Address  string `json:"address"`
	Operstate string `json:"operstate"`
}

func discoverInterfaces(ctx context.Context, r Runner) ([]InterfaceInfo, error) {
	// Gather link info (MAC, state).
	links, err := discoverLinks(ctx, r)
	if err != nil {
		return nil, err
	}

	// Gather addresses for all interfaces.
	addrMap, err := discoverAllAddresses(ctx, r)
	if err != nil {
		return nil, err
	}

	var result []InterfaceInfo
	for _, link := range links {
		info := InterfaceInfo{
			Name:      link.Ifname,
			MAC:       link.Address,
			IsUp:      strings.EqualFold(link.Operstate, "up"),
			Addresses: addrMap[link.Ifname],
		}
		result = append(result, info)
	}
	return result, nil
}

func discoverLinks(ctx context.Context, r Runner) ([]ipLinkJSON, error) {
	stdout, _, exitCode, err := r.Run(ctx, "ip -j link show")
	if err != nil {
		return nil, err
	}
	if exitCode == 0 && len(stdout) > 0 {
		var links []ipLinkJSON
		if json.Unmarshal(stdout, &links) == nil {
			return links, nil
		}
	}
	return nil, fmt.Errorf("failed to parse link info")
}

func discoverAllAddresses(ctx context.Context, r Runner) (map[string][]string, error) {
	stdout, _, exitCode, err := r.Run(ctx, "ip -j addr show")
	if err != nil {
		return nil, err
	}
	result := make(map[string][]string)
	if exitCode == 0 && len(stdout) > 0 {
		var addrs []struct {
			Ifname   string `json:"ifname"`
			AddrInfo []struct {
				Local     string `json:"local"`
				Prefixlen int    `json:"prefixlen"`
			} `json:"addr_info"`
		}
		if json.Unmarshal(stdout, &addrs) == nil {
			for _, a := range addrs {
				for _, ai := range a.AddrInfo {
					result[a.Ifname] = append(result[a.Ifname], fmt.Sprintf("%s/%d", ai.Local, ai.Prefixlen))
				}
			}
			return result, nil
		}
	}
	return result, nil
}
