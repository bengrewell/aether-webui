package aether

import (
	"testing"
)

func TestNewDefaultHostResolver(t *testing.T) {
	provider := NewMockProvider("local")
	resolver := NewDefaultHostResolver(provider)
	if resolver == nil {
		t.Fatal("NewDefaultHostResolver returned nil")
	}
}

func TestDefaultHostResolver_ResolveLocal(t *testing.T) {
	mockProvider := NewMockProvider("local")
	resolver := NewDefaultHostResolver(mockProvider)

	tests := []struct {
		name string
		host string
	}{
		{"empty string", ""},
		{"local keyword", "local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := resolver.Resolve(tt.host)
			if err != nil {
				t.Fatalf("Resolve(%q) returned error: %v", tt.host, err)
			}
			if provider == nil {
				t.Fatalf("Resolve(%q) returned nil provider", tt.host)
			}
			if provider != mockProvider {
				t.Errorf("Resolve(%q) returned different provider than expected", tt.host)
			}
		})
	}
}

func TestDefaultHostResolver_ResolveRemote(t *testing.T) {
	localProvider := NewMockProvider("local")
	remoteProvider := NewMockProvider("server-1")
	resolver := NewDefaultHostResolver(localProvider)
	resolver.AddHost("server-1", remoteProvider)

	t.Run("configured remote host", func(t *testing.T) {
		provider, err := resolver.Resolve("server-1")
		if err != nil {
			t.Fatalf("Resolve returned error: %v", err)
		}
		if provider != remoteProvider {
			t.Error("Resolve returned wrong provider")
		}
	})

	t.Run("unconfigured remote host", func(t *testing.T) {
		_, err := resolver.Resolve("unknown-server")
		if err == nil {
			t.Error("Expected error for unconfigured host")
		}
	})
}

func TestDefaultHostResolver_AddHost(t *testing.T) {
	localProvider := NewMockProvider("local")
	resolver := NewDefaultHostResolver(localProvider)

	// Initially should fail for remote host
	_, err := resolver.Resolve("remote-1")
	if err == nil {
		t.Error("Expected error before adding host")
	}

	// Add remote host
	remoteProvider := NewMockProvider("remote-1")
	resolver.AddHost("remote-1", remoteProvider)

	// Now should succeed
	provider, err := resolver.Resolve("remote-1")
	if err != nil {
		t.Fatalf("Resolve returned error after AddHost: %v", err)
	}
	if provider != remoteProvider {
		t.Error("Resolve returned wrong provider after AddHost")
	}
}

func TestDefaultHostResolver_ListHosts(t *testing.T) {
	localProvider := NewMockProvider("local")
	resolver := NewDefaultHostResolver(localProvider)

	t.Run("local only", func(t *testing.T) {
		hosts := resolver.ListHosts()
		if len(hosts) != 1 {
			t.Errorf("Expected 1 host, got %d", len(hosts))
		}
		if hosts[0] != "local" {
			t.Errorf("Expected 'local', got '%s'", hosts[0])
		}
	})

	t.Run("with remote hosts", func(t *testing.T) {
		resolver.AddHost("server-1", NewMockProvider("server-1"))
		resolver.AddHost("server-2", NewMockProvider("server-2"))

		hosts := resolver.ListHosts()
		if len(hosts) != 3 {
			t.Errorf("Expected 3 hosts, got %d", len(hosts))
		}

		// Verify local is in the list
		foundLocal := false
		for _, h := range hosts {
			if h == "local" {
				foundLocal = true
				break
			}
		}
		if !foundLocal {
			t.Error("Expected 'local' in host list")
		}
	})
}

func TestDefaultHostResolverImplementsInterface(t *testing.T) {
	var _ HostResolver = (*DefaultHostResolver)(nil)
}
