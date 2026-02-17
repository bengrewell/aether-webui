package provider

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/bengrewell/aether-webui/internal/operator"
)

// mockProvider implements Provider for testing.
type mockProvider struct {
	id        NodeID
	operators map[operator.Domain]operator.Operator
	health    *ProviderHealth
	healthErr error
	isLocal   bool
}

func (m *mockProvider) ID() NodeID {
	return m.id
}

func (m *mockProvider) Operator(domain operator.Domain) operator.Operator {
	return m.operators[domain]
}

func (m *mockProvider) Operators() map[operator.Domain]operator.Operator {
	return m.operators
}

func (m *mockProvider) Health(ctx context.Context) (*ProviderHealth, error) {
	return m.health, m.healthErr
}

func (m *mockProvider) IsLocal() bool {
	return m.isLocal
}

func newTestResolver(t *testing.T) *DefaultResolver {
	t.Helper()
	local := &mockProvider{id: LocalNode, isLocal: true}
	return NewDefaultResolver(local)
}

func TestNewDefaultResolver(t *testing.T) {
	local := &mockProvider{id: LocalNode, isLocal: true}
	r := NewDefaultResolver(local)

	if r == nil {
		t.Fatal("NewDefaultResolver() returned nil")
	}
	if r.LocalProvider() != local {
		t.Error("LocalProvider() returned wrong provider")
	}
}

func TestResolveLocalNodeEmptyString(t *testing.T) {
	r := newTestResolver(t)

	p, err := r.Resolve("")
	if err != nil {
		t.Fatalf("Resolve(\"\") error = %v", err)
	}
	if p == nil {
		t.Error("Resolve(\"\") returned nil provider")
	}
	if p.ID() != LocalNode {
		t.Errorf("Resolve(\"\").ID() = %q, want %q", p.ID(), LocalNode)
	}
}

func TestResolveLocalNodeExplicit(t *testing.T) {
	r := newTestResolver(t)

	p, err := r.Resolve(LocalNode)
	if err != nil {
		t.Fatalf("Resolve(LocalNode) error = %v", err)
	}
	if p == nil {
		t.Error("Resolve(LocalNode) returned nil provider")
	}
	if p.ID() != LocalNode {
		t.Errorf("Resolve(LocalNode).ID() = %q, want %q", p.ID(), LocalNode)
	}
}

func TestResolveRemoteNode(t *testing.T) {
	r := newTestResolver(t)
	remote := &mockProvider{id: "remote-1"}

	if err := r.RegisterNode("remote-1", remote); err != nil {
		t.Fatalf("RegisterNode() error = %v", err)
	}

	p, err := r.Resolve("remote-1")
	if err != nil {
		t.Fatalf("Resolve(\"remote-1\") error = %v", err)
	}
	if p != remote {
		t.Error("Resolve(\"remote-1\") returned wrong provider")
	}
}

func TestResolveUnknownNode(t *testing.T) {
	r := newTestResolver(t)

	_, err := r.Resolve("unknown")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("Resolve(\"unknown\") error = %v, want ErrNodeNotFound", err)
	}
}

func TestListNodes(t *testing.T) {
	r := newTestResolver(t)
	remote1 := &mockProvider{id: "remote-1"}
	remote2 := &mockProvider{id: "remote-2"}

	r.RegisterNode("remote-1", remote1)
	r.RegisterNode("remote-2", remote2)

	nodes := r.ListNodes()
	if len(nodes) != 3 {
		t.Fatalf("ListNodes() returned %d nodes, want 3", len(nodes))
	}

	// Check that local is included
	hasLocal := false
	hasRemote1 := false
	hasRemote2 := false
	for _, n := range nodes {
		switch n {
		case LocalNode:
			hasLocal = true
		case "remote-1":
			hasRemote1 = true
		case "remote-2":
			hasRemote2 = true
		}
	}
	if !hasLocal {
		t.Error("ListNodes() missing LocalNode")
	}
	if !hasRemote1 {
		t.Error("ListNodes() missing remote-1")
	}
	if !hasRemote2 {
		t.Error("ListNodes() missing remote-2")
	}
}

func TestListNodesOnlyLocal(t *testing.T) {
	r := newTestResolver(t)

	nodes := r.ListNodes()
	if len(nodes) != 1 {
		t.Fatalf("ListNodes() returned %d nodes, want 1", len(nodes))
	}
	if nodes[0] != LocalNode {
		t.Errorf("ListNodes()[0] = %q, want %q", nodes[0], LocalNode)
	}
}

func TestRegisterNode(t *testing.T) {
	r := newTestResolver(t)
	remote := &mockProvider{id: "remote-1"}

	err := r.RegisterNode("remote-1", remote)
	if err != nil {
		t.Fatalf("RegisterNode() error = %v", err)
	}

	p, err := r.Resolve("remote-1")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if p != remote {
		t.Error("registered provider not found")
	}
}

func TestRegisterLocalNodeEmptyString(t *testing.T) {
	r := newTestResolver(t)
	remote := &mockProvider{id: ""}

	err := r.RegisterNode("", remote)
	if err == nil {
		t.Error("RegisterNode(\"\") should return error")
	}
}

func TestRegisterLocalNodeExplicit(t *testing.T) {
	r := newTestResolver(t)
	remote := &mockProvider{id: LocalNode}

	err := r.RegisterNode(LocalNode, remote)
	if err == nil {
		t.Error("RegisterNode(LocalNode) should return error")
	}
}

func TestRegisterDuplicateNode(t *testing.T) {
	r := newTestResolver(t)
	remote1 := &mockProvider{id: "remote-1"}
	remote2 := &mockProvider{id: "remote-1"}

	if err := r.RegisterNode("remote-1", remote1); err != nil {
		t.Fatalf("first RegisterNode() error = %v", err)
	}

	err := r.RegisterNode("remote-1", remote2)
	if !errors.Is(err, ErrNodeAlreadyRegistered) {
		t.Errorf("second RegisterNode() error = %v, want ErrNodeAlreadyRegistered", err)
	}
}

func TestUnregisterNode(t *testing.T) {
	r := newTestResolver(t)
	remote := &mockProvider{id: "remote-1"}

	r.RegisterNode("remote-1", remote)

	err := r.UnregisterNode("remote-1")
	if err != nil {
		t.Fatalf("UnregisterNode() error = %v", err)
	}

	_, err = r.Resolve("remote-1")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Error("node should not be resolvable after unregistration")
	}
}

func TestUnregisterUnknownNode(t *testing.T) {
	r := newTestResolver(t)

	err := r.UnregisterNode("unknown")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("UnregisterNode(\"unknown\") error = %v, want ErrNodeNotFound", err)
	}
}

func TestUnregisterLocalNodeEmptyString(t *testing.T) {
	r := newTestResolver(t)

	err := r.UnregisterNode("")
	if err == nil {
		t.Error("UnregisterNode(\"\") should return error")
	}
}

func TestUnregisterLocalNodeExplicit(t *testing.T) {
	r := newTestResolver(t)

	err := r.UnregisterNode(LocalNode)
	if err == nil {
		t.Error("UnregisterNode(LocalNode) should return error")
	}
}

func TestConcurrentOperations(t *testing.T) {
	r := newTestResolver(t)
	const numGoroutines = 100
	const numOperations = 50

	var wg sync.WaitGroup

	// Concurrent registrations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				nodeID := NodeID(string(rune('a'+id)) + string(rune('0'+j)))
				remote := &mockProvider{id: nodeID}
				r.RegisterNode(nodeID, remote)
			}
		}(i % 26) // Use letters a-z
	}
	wg.Wait()

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				r.ListNodes()
				r.Resolve("")
				r.Resolve(LocalNode)
			}
		}()
	}
	wg.Wait()

	// Concurrent mixed operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			nodeID := NodeID("concurrent-" + string(rune('a'+id)))
			remote := &mockProvider{id: nodeID}
			r.RegisterNode(nodeID, remote)
			r.Resolve(nodeID)
			r.ListNodes()
			r.UnregisterNode(nodeID)
		}(i % 26)
	}
	wg.Wait()
}

func TestDefaultResolverImplementsInterface(t *testing.T) {
	var _ ProviderResolver = (*DefaultResolver)(nil)
}

func TestErrNodeNotFound(t *testing.T) {
	if ErrNodeNotFound == nil {
		t.Fatal("ErrNodeNotFound should not be nil")
	}
	if ErrNodeNotFound.Error() != "node not found" {
		t.Errorf("ErrNodeNotFound.Error() = %q, want %q", ErrNodeNotFound.Error(), "node not found")
	}
}

func TestErrNodeAlreadyRegistered(t *testing.T) {
	if ErrNodeAlreadyRegistered == nil {
		t.Fatal("ErrNodeAlreadyRegistered should not be nil")
	}
	if ErrNodeAlreadyRegistered.Error() != "node already registered" {
		t.Errorf("ErrNodeAlreadyRegistered.Error() = %q, want %q", ErrNodeAlreadyRegistered.Error(), "node already registered")
	}
}
