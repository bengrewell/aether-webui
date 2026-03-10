package configdefaults

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/bengrewell/aether-webui/internal/nodefacts"
	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/provider/onramp"
	"github.com/bengrewell/aether-webui/internal/store"
)

// mockGatherer returns canned facts for testing.
type mockGatherer struct {
	facts map[string]nodefacts.NodeFacts // keyed by host
}

func (m *mockGatherer) Gather(_ context.Context, host, user string, password string, sshKey []byte) (nodefacts.NodeFacts, error) {
	if f, ok := m.facts[host]; ok {
		return f, nil
	}
	return nodefacts.NodeFacts{
		AnsibleHost:  host,
		DefaultIface: "eth0",
		DefaultIP:    host,
		DefaultSubnet: "10.0.0.0/24",
		GatheredAt:   time.Now().UTC(),
	}, nil
}

func newTestProvider(t *testing.T, g nodefacts.Gatherer) (*Provider, store.Client) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(t.Context(), dbPath)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	onrampDir := t.TempDir()
	varsDir := filepath.Join(onrampDir, "vars")
	if err := os.MkdirAll(varsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a minimal vars/main.yml.
	cfg := onramp.OnRampConfig{
		Core: &onramp.CoreConfig{
			DataIface: "old_iface",
		},
	}
	data, _ := yaml.Marshal(&cfg)
	if err := os.WriteFile(filepath.Join(varsDir, "main.yml"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewProvider(Config{
		OnRampDir: onrampDir,
	}, g, provider.WithStore(st))

	return p, st
}

func TestBuildPatchMap(t *testing.T) {
	applied := []AppliedDefault{
		{Field: "core.data_iface", Value: "ens18"},
		{Field: "core.amf.ip", Value: "10.0.0.10"},
		{Field: "gnbsim.router.data_iface", Value: "ens18"},
	}

	m := buildPatchMap(applied)

	coreMap, ok := m["core"].(map[string]any)
	if !ok {
		t.Fatal("expected core to be a map")
	}
	if coreMap["data_iface"] != "ens18" {
		t.Errorf("core.data_iface = %v, want ens18", coreMap["data_iface"])
	}
	amfMap, ok := coreMap["amf"].(map[string]any)
	if !ok {
		t.Fatal("expected core.amf to be a map")
	}
	if amfMap["ip"] != "10.0.0.10" {
		t.Errorf("core.amf.ip = %v, want 10.0.0.10", amfMap["ip"])
	}
}

func TestMatchesRole(t *testing.T) {
	tests := []struct {
		nodeRoles []string
		ruleRoles []string
		want      bool
	}{
		{[]string{"master"}, []string{"master"}, true},
		{[]string{"worker"}, []string{"master"}, false},
		{[]string{"master", "gnbsim"}, []string{"gnbsim"}, true},
		{nil, []string{"master"}, false},
	}
	for _, tt := range tests {
		got := matchesRole(tt.nodeRoles, tt.ruleRoles)
		if got != tt.want {
			t.Errorf("matchesRole(%v, %v) = %v, want %v", tt.nodeRoles, tt.ruleRoles, got, tt.want)
		}
	}
}

func TestHandleApplyConfigDefaults(t *testing.T) {
	g := &mockGatherer{
		facts: map[string]nodefacts.NodeFacts{
			"10.0.0.10": {
				DefaultIface:  "ens18",
				DefaultIP:     "10.0.0.10",
				DefaultSubnet: "10.0.0.0/24",
				GatheredAt:    time.Now().UTC(),
			},
		},
	}

	p, st := newTestProvider(t, g)
	ctx := t.Context()

	// Register a node with master role.
	node := store.Node{
		ID:          "node-1",
		Name:        "node1",
		AnsibleHost: "10.0.0.10",
		AnsibleUser: "ubuntu",
		Password:    []byte("pass"),
		Roles:       []string{"master"},
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := st.UpsertNode(ctx, node); err != nil {
		t.Fatalf("UpsertNode: %v", err)
	}

	out, err := p.handleApplyConfigDefaults(ctx, &ConfigDefaultsApplyInput{Refresh: true})
	if err != nil {
		t.Fatalf("handleApplyConfigDefaults: %v", err)
	}

	result := out.Body
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if len(result.Applied) == 0 {
		t.Fatal("expected applied defaults")
	}

	// Verify specific applied fields.
	appliedFields := make(map[string]any)
	for _, a := range result.Applied {
		appliedFields[a.Field] = a.Value
	}

	if v, ok := appliedFields["core.data_iface"]; !ok || v != "ens18" {
		t.Errorf("core.data_iface = %v, want ens18", v)
	}
	if v, ok := appliedFields["core.amf.ip"]; !ok || v != "10.0.0.10" {
		t.Errorf("core.amf.ip = %v, want 10.0.0.10", v)
	}

	// Verify the returned config reflects applied values.
	if result.Config.Core == nil || result.Config.Core.DataIface != "ens18" {
		t.Errorf("config core.data_iface = %v, want ens18", result.Config.Core)
	}
	if result.Config.Core.AMF == nil || result.Config.Core.AMF.IP != "10.0.0.10" {
		t.Errorf("config core.amf.ip not set correctly")
	}

	// Verify vars/main.yml was updated on disk.
	diskCfg, err := p.readVarsFile()
	if err != nil {
		t.Fatalf("readVarsFile: %v", err)
	}
	if diskCfg.Core == nil || diskCfg.Core.DataIface != "ens18" {
		t.Errorf("disk config core.data_iface = %v, want ens18", diskCfg.Core)
	}
}

func TestHandleApplyConfigDefaultsNoNodes(t *testing.T) {
	p, _ := newTestProvider(t, &mockGatherer{})

	out, err := p.handleApplyConfigDefaults(t.Context(), &ConfigDefaultsApplyInput{})
	if err != nil {
		t.Fatalf("handleApplyConfigDefaults: %v", err)
	}
	if len(out.Body.Applied) != 0 {
		t.Errorf("expected no applied defaults, got %d", len(out.Body.Applied))
	}
	if len(out.Body.Errors) == 0 {
		t.Error("expected informational error about no nodes")
	}
}

func TestHandleApplyConfigDefaultsNoMatchingRoles(t *testing.T) {
	g := &mockGatherer{}
	p, st := newTestProvider(t, g)
	ctx := t.Context()

	// Register a node with no matching roles.
	node := store.Node{
		ID:          "node-1",
		Name:        "worker1",
		AnsibleHost: "10.0.0.20",
		AnsibleUser: "ubuntu",
		Password:    []byte("pass"),
		Roles:       []string{"worker"},
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := st.UpsertNode(ctx, node); err != nil {
		t.Fatalf("UpsertNode: %v", err)
	}

	out, err := p.handleApplyConfigDefaults(ctx, &ConfigDefaultsApplyInput{Refresh: true})
	if err != nil {
		t.Fatalf("handleApplyConfigDefaults: %v", err)
	}
	if len(out.Body.Applied) != 0 {
		t.Errorf("expected no applied defaults for worker-only node, got %d", len(out.Body.Applied))
	}
}

func TestHandleGetNodeFacts(t *testing.T) {
	g := &mockGatherer{
		facts: map[string]nodefacts.NodeFacts{
			"10.0.0.10": {
				DefaultIface:  "ens18",
				DefaultIP:     "10.0.0.10",
				DefaultSubnet: "10.0.0.0/24",
				GatheredAt:    time.Now().UTC(),
			},
		},
	}

	p, st := newTestProvider(t, g)
	ctx := t.Context()

	node := store.Node{
		ID:          "node-1",
		Name:        "node1",
		AnsibleHost: "10.0.0.10",
		AnsibleUser: "ubuntu",
		Password:    []byte("pass"),
		Roles:       []string{"master"},
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := st.UpsertNode(ctx, node); err != nil {
		t.Fatalf("UpsertNode: %v", err)
	}

	out, err := p.handleGetNodeFacts(ctx, &NodeFactsGetInput{ID: "node-1", Refresh: true})
	if err != nil {
		t.Fatalf("handleGetNodeFacts: %v", err)
	}
	if out.Body.DefaultIface != "ens18" {
		t.Errorf("DefaultIface = %q, want %q", out.Body.DefaultIface, "ens18")
	}
	if out.Body.DefaultIP != "10.0.0.10" {
		t.Errorf("DefaultIP = %q, want %q", out.Body.DefaultIP, "10.0.0.10")
	}
}

func TestHandleGetNodeFactsNotFound(t *testing.T) {
	p, _ := newTestProvider(t, &mockGatherer{})

	_, err := p.handleGetNodeFacts(t.Context(), &NodeFactsGetInput{ID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent node")
	}
}

func TestFactsCaching(t *testing.T) {
	callCount := 0
	g := &countingGatherer{
		inner: &mockGatherer{
			facts: map[string]nodefacts.NodeFacts{
				"10.0.0.10": {
					DefaultIface: "ens18",
					DefaultIP:    "10.0.0.10",
					GatheredAt:   time.Now().UTC(),
				},
			},
		},
		count: &callCount,
	}

	p, st := newTestProvider(t, g)
	ctx := t.Context()

	node := store.Node{
		ID:          "node-1",
		Name:        "node1",
		AnsibleHost: "10.0.0.10",
		AnsibleUser: "ubuntu",
		Password:    []byte("pass"),
		Roles:       []string{"master"},
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().Add(-time.Minute), // Updated before gathered
	}
	if err := st.UpsertNode(ctx, node); err != nil {
		t.Fatalf("UpsertNode: %v", err)
	}

	// First call gathers via SSH.
	_, err := p.handleGetNodeFacts(ctx, &NodeFactsGetInput{ID: "node-1", Refresh: true})
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 gather call, got %d", callCount)
	}

	// Second call should use cache.
	_, err = p.handleGetNodeFacts(ctx, &NodeFactsGetInput{ID: "node-1", Refresh: false})
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected cache hit (1 call), got %d calls", callCount)
	}

	// With refresh=true, should gather again.
	_, err = p.handleGetNodeFacts(ctx, &NodeFactsGetInput{ID: "node-1", Refresh: true})
	if err != nil {
		t.Fatalf("third call: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 gather calls after refresh, got %d", callCount)
	}
}

type countingGatherer struct {
	inner nodefacts.Gatherer
	count *int
}

func (g *countingGatherer) Gather(ctx context.Context, host, user string, password string, sshKey []byte) (nodefacts.NodeFacts, error) {
	*g.count++
	return g.inner.Gather(ctx, host, user, password, sshKey)
}

func TestDeepMergeConfig(t *testing.T) {
	base := onramp.OnRampConfig{
		Core: &onramp.CoreConfig{
			DataIface: "old_iface",
			AMF:       &onramp.AMFConfig{IP: "1.2.3.4"},
		},
	}

	patch := map[string]any{
		"core": map[string]any{
			"data_iface": "ens18",
			"amf":        map[string]any{"ip": "10.0.0.10"},
		},
	}
	patchJSON, _ := json.Marshal(patch)

	merged, err := deepMergeConfig(&base, patchJSON)
	if err != nil {
		t.Fatalf("deepMergeConfig: %v", err)
	}
	if merged.Core.DataIface != "ens18" {
		t.Errorf("DataIface = %q, want %q", merged.Core.DataIface, "ens18")
	}
	if merged.Core.AMF == nil || merged.Core.AMF.IP != "10.0.0.10" {
		t.Errorf("AMF.IP = %v, want 10.0.0.10", merged.Core.AMF)
	}
}

func TestMultiNodeMultiRole(t *testing.T) {
	g := &mockGatherer{
		facts: map[string]nodefacts.NodeFacts{
			"10.0.0.10": {
				DefaultIface:  "ens18",
				DefaultIP:     "10.0.0.10",
				DefaultSubnet: "10.0.0.0/24",
				GatheredAt:    time.Now().UTC(),
			},
			"10.0.0.20": {
				DefaultIface:  "eth0",
				DefaultIP:     "10.0.0.20",
				DefaultSubnet: "10.0.0.0/24",
				GatheredAt:    time.Now().UTC(),
			},
		},
	}

	p, st := newTestProvider(t, g)
	ctx := t.Context()

	// Master node.
	if err := st.UpsertNode(ctx, store.Node{
		ID: "node-1", Name: "node1", AnsibleHost: "10.0.0.10",
		AnsibleUser: "ubuntu", Password: []byte("pass"),
		Roles: []string{"master"}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}

	// gNBSim node.
	if err := st.UpsertNode(ctx, store.Node{
		ID: "node-2", Name: "node2", AnsibleHost: "10.0.0.20",
		AnsibleUser: "ubuntu", Password: []byte("pass"),
		Roles: []string{"gnbsim"}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}

	out, err := p.handleApplyConfigDefaults(ctx, &ConfigDefaultsApplyInput{Refresh: true})
	if err != nil {
		t.Fatal(err)
	}

	appliedFields := make(map[string]any)
	for _, a := range out.Body.Applied {
		appliedFields[a.Field] = a.Value
	}

	// Master rules should apply from node1.
	if v := appliedFields["core.data_iface"]; v != "ens18" {
		t.Errorf("core.data_iface = %v, want ens18", v)
	}
	// gNBSim rules should apply from node2.
	if v := appliedFields["gnbsim.router.data_iface"]; v != "eth0" {
		t.Errorf("gnbsim.router.data_iface = %v, want eth0", v)
	}
}
