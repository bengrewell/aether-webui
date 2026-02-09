package onramp

import (
	"strings"
	"testing"

	"github.com/bengrewell/aether-webui/internal/crypto"
	"github.com/bengrewell/aether-webui/internal/state"
)

func TestBuildInventoryEmpty(t *testing.T) {
	content, err := buildInventory(nil, "")
	if err != nil {
		t.Fatalf("buildInventory() error = %v", err)
	}

	// All groups should be present
	for _, group := range inventoryGroups {
		if !strings.Contains(content, "["+group+"]") {
			t.Errorf("missing group [%s]", group)
		}
	}
	if !strings.Contains(content, "[all]") {
		t.Error("missing [all] group")
	}
}

func TestBuildInventoryWithNodes(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := crypto.Encrypt("mypass", key)
	if err != nil {
		t.Fatal(err)
	}

	nodes := []state.Node{
		{
			ID:                "node-1",
			Name:              "node1",
			NodeType:          state.NodeTypeRemote,
			Address:           "10.76.28.113",
			Username:          "aether",
			EncryptedPassword: encrypted,
			Roles:             []string{"k8s-master", "sd-core"},
		},
		{
			ID:       "node-2",
			Name:     "node2",
			NodeType: state.NodeTypeRemote,
			Address:  "10.76.28.115",
			Username: "aether",
			Roles:    []string{"k8s-worker", "srsran-gnb"},
		},
	}

	content, err := buildInventory(nodes, key)
	if err != nil {
		t.Fatalf("buildInventory() error = %v", err)
	}

	// Verify [all] section has both nodes
	if !strings.Contains(content, "node1 ansible_host=10.76.28.113") {
		t.Error("missing node1 in [all]")
	}
	if !strings.Contains(content, "node2 ansible_host=10.76.28.115") {
		t.Error("missing node2 in [all]")
	}

	// Verify password decryption
	if !strings.Contains(content, "ansible_password=mypass") {
		t.Error("password not decrypted in [all]")
	}
	if !strings.Contains(content, "ansible_sudo_pass=mypass") {
		t.Error("sudo_pass not in [all]")
	}

	// Verify role-based groups
	// node1 has k8s-master and sd-core, both map to master_nodes
	masterSection := extractSection(content, "master_nodes")
	if !strings.Contains(masterSection, "node1") {
		t.Error("node1 not in [master_nodes]")
	}

	workerSection := extractSection(content, "worker_nodes")
	if !strings.Contains(workerSection, "node2") {
		t.Error("node2 not in [worker_nodes]")
	}

	srsranSection := extractSection(content, "srsran_nodes")
	if !strings.Contains(srsranSection, "node2") {
		t.Error("node2 not in [srsran_nodes]")
	}

	// Empty groups should still exist
	oaiSection := extractSection(content, "oai_nodes")
	if strings.TrimSpace(oaiSection) != "" {
		t.Errorf("expected empty [oai_nodes], got %q", oaiSection)
	}
}

func TestBuildInventorySkipsLocalNodes(t *testing.T) {
	nodes := []state.Node{
		{
			ID:       "local",
			Name:     "Local",
			NodeType: state.NodeTypeLocal,
			Roles:    []string{"k8s-master"},
		},
	}

	content, err := buildInventory(nodes, "")
	if err != nil {
		t.Fatalf("buildInventory() error = %v", err)
	}

	allSection := extractSection(content, "all")
	if strings.Contains(allSection, "Local") {
		t.Error("local node should not appear in [all]")
	}
}

func TestBuildInventoryMultipleRolesSameNode(t *testing.T) {
	nodes := []state.Node{
		{
			ID:       "n1",
			Name:     "multi",
			NodeType: state.NodeTypeRemote,
			Address:  "10.0.0.1",
			Username: "user",
			Roles:    []string{"k8s-master", "k8s-worker", "gnbsim"},
		},
	}

	content, err := buildInventory(nodes, "")
	if err != nil {
		t.Fatalf("buildInventory() error = %v", err)
	}

	masterSection := extractSection(content, "master_nodes")
	if !strings.Contains(masterSection, "multi") {
		t.Error("node not in [master_nodes]")
	}

	workerSection := extractSection(content, "worker_nodes")
	if !strings.Contains(workerSection, "multi") {
		t.Error("node not in [worker_nodes]")
	}

	gnbsimSection := extractSection(content, "gnbsim_nodes")
	if !strings.Contains(gnbsimSection, "multi") {
		t.Error("node not in [gnbsim_nodes]")
	}
}

// extractSection returns the content between [name] and the next [ or end of string.
func extractSection(content, name string) string {
	marker := "[" + name + "]\n"
	idx := strings.Index(content, marker)
	if idx == -1 {
		return ""
	}
	start := idx + len(marker)
	end := strings.Index(content[start:], "[")
	if end == -1 {
		return content[start:]
	}
	return content[start : start+end]
}
