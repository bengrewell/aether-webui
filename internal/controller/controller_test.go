package controller

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/bengrewell/aether-webui/internal/endpoint"
	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/provider/meta"
	"github.com/bengrewell/aether-webui/internal/store"
)

func TestNew_Defaults(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if c.listenAddr != "127.0.0.1:8186" {
		t.Errorf("listenAddr = %q, want %q", c.listenAddr, "127.0.0.1:8186")
	}
	if c.dataDir != "/var/lib/aether-webd" {
		t.Errorf("dataDir = %q, want %q", c.dataDir, "/var/lib/aether-webd")
	}
	if !c.frontendEnabled {
		t.Error("frontendEnabled = false, want true")
	}
	if c.metricsInterval != "10s" {
		t.Errorf("metricsInterval = %q, want %q", c.metricsInterval, "10s")
	}
	if c.metricsRetention != "24h" {
		t.Errorf("metricsRetention = %q, want %q", c.metricsRetention, "24h")
	}
	if c.debug {
		t.Error("debug = true, want false")
	}
}

func TestNew_WithOptions(t *testing.T) {
	c, err := New(
		WithListenAddr("0.0.0.0:9999"),
		WithDebug(true),
		WithDataDir("/tmp/test"),
		WithVersion(meta.VersionInfo{Version: "1.0.0", Branch: "main"}),
		WithTLS(true, "cert.pem", "key.pem", "ca.pem"),
		WithAPIToken("secret"),
		WithRBAC(true),
		WithFrontend(false, "/custom"),
		WithMetrics("30s", "7d"),
		WithEncryptionKey("0123456789abcdef0123456789abcdef"),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if c.listenAddr != "0.0.0.0:9999" {
		t.Errorf("listenAddr = %q, want %q", c.listenAddr, "0.0.0.0:9999")
	}
	if !c.debug {
		t.Error("debug = false, want true")
	}
	if c.dataDir != "/tmp/test" {
		t.Errorf("dataDir = %q, want %q", c.dataDir, "/tmp/test")
	}
	if c.versionInfo.Version != "1.0.0" {
		t.Errorf("version = %q, want %q", c.versionInfo.Version, "1.0.0")
	}
	if !c.tlsAuto {
		t.Error("tlsAuto = false, want true")
	}
	if c.tlsCert != "cert.pem" {
		t.Errorf("tlsCert = %q, want %q", c.tlsCert, "cert.pem")
	}
	if c.tlsKey != "key.pem" {
		t.Errorf("tlsKey = %q, want %q", c.tlsKey, "key.pem")
	}
	if c.tlsMTLSCA != "ca.pem" {
		t.Errorf("tlsMTLSCA = %q, want %q", c.tlsMTLSCA, "ca.pem")
	}
	if c.apiToken != "secret" {
		t.Errorf("apiToken = %q, want %q", c.apiToken, "secret")
	}
	if !c.rbacEnabled {
		t.Error("rbacEnabled = false, want true")
	}
	if c.frontendEnabled {
		t.Error("frontendEnabled = true, want false")
	}
	if c.frontendDir != "/custom" {
		t.Errorf("frontendDir = %q, want %q", c.frontendDir, "/custom")
	}
	if c.metricsInterval != "30s" {
		t.Errorf("metricsInterval = %q, want %q", c.metricsInterval, "30s")
	}
	if c.metricsRetention != "7d" {
		t.Errorf("metricsRetention = %q, want %q", c.metricsRetention, "7d")
	}
	if c.encryptionKey != "0123456789abcdef0123456789abcdef" {
		t.Error("encryptionKey not set correctly")
	}
}

func waitForServer(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("server at %s did not start within 5 seconds", addr)
}

func ephemeralAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := l.Addr().String()
	l.Close()
	return addr
}

func TestRun_StartsAndStops(t *testing.T) {
	addr := ephemeralAddr(t)

	ctrl, err := New(
		WithListenAddr(addr),
		WithDataDir(t.TempDir()),
		WithFrontend(false, ""),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- ctrl.Run(ctx) }()

	waitForServer(t, addr)

	// Verify the meta version endpoint responds (registered by the meta provider).
	resp, err := http.Get("http://" + addr + "/api/v1/meta/version")
	if err != nil {
		t.Fatalf("GET /api/v1/meta/version error: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/v1/meta/version status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not exit within 5 seconds")
	}
}

// mockProvider satisfies provider.Provider for testing.
type mockProvider struct {
	*provider.Base
	started  bool
	stopped  bool
	disabled bool
}

func (m *mockProvider) Endpoints() []endpoint.AnyEndpoint { return nil }
func (m *mockProvider) Start() error                      { m.started = true; return nil }
func (m *mockProvider) Stop() error                       { m.stopped = true; return nil }
func (m *mockProvider) Disable()                          { m.disabled = true; m.Base.Disable() }

func TestRun_WithProvider(t *testing.T) {
	addr := ephemeralAddr(t)
	createdCh := make(chan *mockProvider, 1)

	factory := func(_ context.Context, _ store.Client, opts []provider.Option) (provider.Provider, error) {
		m := &mockProvider{Base: provider.New("test", opts...)}
		createdCh <- m
		return m, nil
	}

	ctrl, err := New(
		WithListenAddr(addr),
		WithDataDir(t.TempDir()),
		WithFrontend(false, ""),
		WithProvider("test", true, factory),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- ctrl.Run(ctx) }()

	// Wait for factory to be called (happens-before edge via channel).
	var created *mockProvider
	select {
	case created = <-createdCh:
	case <-time.After(5 * time.Second):
		t.Fatal("provider factory was not called within 5 seconds")
	}

	// Wait for server to start, then shut down.
	waitForServer(t, addr)
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not exit within 5 seconds")
	}

	// Assert after Run returns — no concurrent writes possible.
	if !created.started {
		t.Error("provider was not started")
	}
}

func TestRun_DisabledProvider(t *testing.T) {
	addr := ephemeralAddr(t)
	createdCh := make(chan *mockProvider, 1)

	factory := func(_ context.Context, _ store.Client, opts []provider.Option) (provider.Provider, error) {
		m := &mockProvider{Base: provider.New("disabled-test", opts...)}
		createdCh <- m
		return m, nil
	}

	ctrl, err := New(
		WithListenAddr(addr),
		WithDataDir(t.TempDir()),
		WithFrontend(false, ""),
		WithProvider("disabled-test", false, factory),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- ctrl.Run(ctx) }()

	// Wait for factory to be called (happens-before edge via channel).
	var created *mockProvider
	select {
	case created = <-createdCh:
	case <-time.After(5 * time.Second):
		t.Fatal("provider factory was not called within 5 seconds")
	}

	// Wait for server to start, then shut down.
	waitForServer(t, addr)
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not exit within 5 seconds")
	}

	// Assert after Run returns — no concurrent writes possible.
	if !created.disabled {
		t.Error("provider Disable() was not called")
	}
	if created.started {
		t.Error("disabled provider should not have been started")
	}
}

func TestRun_ProviderFactoryError(t *testing.T) {
	factory := func(_ context.Context, _ store.Client, _ []provider.Option) (provider.Provider, error) {
		return nil, fmt.Errorf("factory failed")
	}

	ctrl, err := New(
		WithListenAddr(ephemeralAddr(t)),
		WithDataDir(t.TempDir()),
		WithFrontend(false, ""),
		WithProvider("broken", true, factory),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	err = ctrl.Run(t.Context())
	if err == nil {
		t.Fatal("expected error from broken factory")
	}
	if want := `providers: provider "broken": factory failed`; err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}
