package sysinfo

import (
	"testing"
)

func TestNewDefaultNodeResolver(t *testing.T) {
	provider := NewMockProvider()
	resolver := NewDefaultNodeResolver(provider)
	if resolver == nil {
		t.Fatal("NewDefaultNodeResolver returned nil")
	}
}

func TestDefaultNodeResolver_ResolveLocal(t *testing.T) {
	mockProvider := NewMockProvider()
	resolver := NewDefaultNodeResolver(mockProvider)

	tests := []struct {
		name  string
		node  string
	}{
		{"empty string", ""},
		{"local keyword", "local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := resolver.Resolve(tt.node)
			if err != nil {
				t.Fatalf("Resolve(%q) returned error: %v", tt.node, err)
			}
			if provider == nil {
				t.Fatalf("Resolve(%q) returned nil provider", tt.node)
			}
			if provider != mockProvider {
				t.Errorf("Resolve(%q) returned different provider than expected", tt.node)
			}
		})
	}
}

func TestDefaultNodeResolver_ResolveRemote(t *testing.T) {
	mockProvider := NewMockProvider()
	resolver := NewDefaultNodeResolver(mockProvider)

	remoteNodes := []string{
		"remote-node-1",
		"192.168.1.100",
		"node.example.com",
	}

	for _, node := range remoteNodes {
		t.Run(node, func(t *testing.T) {
			provider, err := resolver.Resolve(node)
			if err == nil {
				t.Errorf("Resolve(%q) should return error for remote node", node)
			}
			if provider != nil {
				t.Errorf("Resolve(%q) should return nil provider for remote node", node)
			}
		})
	}
}

func TestDefaultNodeResolverImplementsInterface(t *testing.T) {
	var _ NodeResolver = (*DefaultNodeResolver)(nil)
}
