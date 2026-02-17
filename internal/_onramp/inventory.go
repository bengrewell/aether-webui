package onramp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bengrewell/aether-webui/internal/crypto"
	"github.com/bengrewell/aether-webui/internal/state"
)

// inventoryGroups defines all Ansible inventory group names in order.
// All groups are always written, even if empty.
var inventoryGroups = []string{
	"master_nodes",
	"worker_nodes",
	"gnbsim_nodes",
	"srsran_nodes",
	"oai_nodes",
	"ueransim_nodes",
	"oscric_nodes",
	"n3iwf_nodes",
}

// roleToGroup maps node role names to inventory group names.
var roleToGroup = map[string]string{
	"k8s-master": "master_nodes",
	"k8s-worker": "worker_nodes",
	"sd-core":    "master_nodes",
	"gnbsim":     "gnbsim_nodes",
	"srsran-gnb": "srsran_nodes",
	"oai-gnb":    "oai_nodes",
	"ueransim":   "ueransim_nodes",
	"oscric":     "oscric_nodes",
	"n3iwf":      "n3iwf_nodes",
}

// GenerateInventory reads nodes and roles from the store, decrypts passwords,
// and writes a hosts.ini file to the given path.
func GenerateInventory(ctx context.Context, store state.Store, encryptionKey, outputPath string) error {
	nodes, err := store.ListNodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	content, err := buildInventory(nodes, encryptionKey)
	if err != nil {
		return err
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create inventory directory: %w", err)
	}

	return os.WriteFile(outputPath, []byte(content), 0640)
}

// buildInventory constructs the INI-format inventory string.
func buildInventory(nodes []state.Node, encryptionKey string) (string, error) {
	var b strings.Builder

	// [all] section — every remote node with connection details
	b.WriteString("[all]\n")
	for _, n := range nodes {
		if n.NodeType == state.NodeTypeLocal {
			continue
		}
		line, err := allLine(n, encryptionKey)
		if err != nil {
			return "", fmt.Errorf("node %s: %w", n.ID, err)
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	b.WriteByte('\n')

	// Build role → node name map
	groupMembers := make(map[string][]string)
	for _, n := range nodes {
		for _, role := range n.Roles {
			group, ok := roleToGroup[role]
			if !ok {
				continue
			}
			groupMembers[group] = append(groupMembers[group], n.Name)
		}
	}

	// Write each inventory group (always present, even if empty)
	for _, group := range inventoryGroups {
		b.WriteString("[" + group + "]\n")
		for _, name := range groupMembers[group] {
			b.WriteString(name)
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}

	return b.String(), nil
}

// allLine builds the [all] inventory line for a node.
func allLine(n state.Node, encryptionKey string) (string, error) {
	password := ""
	if n.EncryptedPassword != "" && encryptionKey != "" {
		decrypted, err := crypto.Decrypt(n.EncryptedPassword, encryptionKey)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt password: %w", err)
		}
		password = decrypted
	}

	parts := []string{n.Name}
	if n.Address != "" {
		parts = append(parts, "ansible_host="+n.Address)
	}
	if n.Username != "" {
		parts = append(parts, "ansible_user="+n.Username)
	}
	if password != "" {
		parts = append(parts, "ansible_password="+password)
		parts = append(parts, "ansible_sudo_pass="+password)
	}

	return strings.Join(parts, " "), nil
}
