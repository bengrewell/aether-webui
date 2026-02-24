package onramp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/bengrewell/aether-webui/internal/store"
)

// roleSections defines the fixed ordering of [role_nodes] sections in hosts.ini.
var roleSections = []struct {
	role    string
	section string
}{
	{"master", "master_nodes"},
	{"worker", "worker_nodes"},
	{"gnbsim", "gnbsim_nodes"},
	{"oai", "oai_nodes"},
	{"ueransim", "ueransim_nodes"},
	{"srsran", "srsran_nodes"},
	{"oscric", "oscric_nodes"},
	{"n3iwf", "n3iwf_nodes"},
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func (o *OnRamp) handleGetInventory(_ context.Context, _ *struct{}) (*InventoryGetOutput, error) {
	path := filepath.Join(o.config.OnRampDir, "hosts.ini")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &InventoryGetOutput{Body: InventoryData{}}, nil
		}
		return nil, huma.Error500InternalServerError("failed to read hosts.ini", err)
	}
	inv := parseHostsINI(data)
	return &InventoryGetOutput{Body: inv}, nil
}

func (o *OnRamp) handleSyncInventory(ctx context.Context, _ *struct{}) (*InventorySyncOutput, error) {
	infos, err := o.Store().ListNodes(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list nodes", err)
	}

	// Fetch full node data (with decrypted secrets) for each node.
	nodes := make([]store.Node, 0, len(infos))
	for _, info := range infos {
		node, ok, err := o.Store().GetNode(ctx, info.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get node", err)
		}
		if ok {
			nodes = append(nodes, node)
		}
	}

	data := generateHostsINI(nodes)
	path := filepath.Join(o.config.OnRampDir, "hosts.ini")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, huma.Error500InternalServerError("failed to write hosts.ini", err)
	}

	out := &InventorySyncOutput{}
	out.Body.Message = fmt.Sprintf("hosts.ini written with %d nodes", len(nodes))
	out.Body.Path = path
	return out, nil
}

// ---------------------------------------------------------------------------
// Parser
// ---------------------------------------------------------------------------

// parseHostsINI parses an Ansible hosts.ini file into structured inventory data.
func parseHostsINI(data []byte) InventoryData {
	inv := InventoryData{}

	// First pass: parse [all] section for node definitions.
	nodeMap := make(map[string]*InventoryNode)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			continue
		}

		switch {
		case currentSection == "all":
			node := parseAllLine(line)
			if node.Name != "" {
				nodeMap[node.Name] = &node
			}

		default:
			// Role sections: bare node names
			role := sectionToRole(currentSection)
			if role == "" {
				continue
			}
			name := strings.Fields(line)[0]
			if n, ok := nodeMap[name]; ok {
				n.Roles = append(n.Roles, role)
			}
		}
	}

	for _, n := range nodeMap {
		inv.Nodes = append(inv.Nodes, *n)
	}
	return inv
}

// parseAllLine parses a line from the [all] section.
// Format: name ansible_host=... ansible_user=... ansible_password=... ansible_sudo_pass=...
func parseAllLine(line string) InventoryNode {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return InventoryNode{}
	}

	node := InventoryNode{Name: fields[0]}
	for _, f := range fields[1:] {
		k, v, ok := strings.Cut(f, "=")
		if !ok {
			continue
		}
		switch k {
		case "ansible_host":
			node.AnsibleHost = v
		case "ansible_user":
			node.AnsibleUser = v
		}
	}
	return node
}

// sectionToRole maps an INI section name to a node role.
func sectionToRole(section string) string {
	for _, rs := range roleSections {
		if rs.section == section {
			return rs.role
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Generator
// ---------------------------------------------------------------------------

// generateHostsINI produces an Ansible hosts.ini file from the given nodes.
func generateHostsINI(nodes []store.Node) []byte {
	var buf bytes.Buffer

	// [all] section
	buf.WriteString("[all]\n")
	for _, n := range nodes {
		buf.WriteString(n.Name)
		buf.WriteString(" ansible_host=")
		buf.WriteString(n.AnsibleHost)
		if n.AnsibleUser != "" {
			buf.WriteString(" ansible_user=")
			buf.WriteString(n.AnsibleUser)
		}
		if len(n.Password) > 0 {
			buf.WriteString(" ansible_password=")
			buf.WriteString(string(n.Password))
		}
		if len(n.SudoPassword) > 0 {
			buf.WriteString(" ansible_sudo_pass=")
			buf.WriteString(string(n.SudoPassword))
		}
		buf.WriteString("\n")
	}

	// Build role->names index
	roleNodes := make(map[string][]string)
	for _, n := range nodes {
		for _, r := range n.Roles {
			roleNodes[r] = append(roleNodes[r], n.Name)
		}
	}

	// Emit role sections in fixed order
	for _, rs := range roleSections {
		buf.WriteString("\n[")
		buf.WriteString(rs.section)
		buf.WriteString("]\n")
		for _, name := range roleNodes[rs.role] {
			buf.WriteString(name)
			buf.WriteString("\n")
		}
	}

	return buf.Bytes()
}
