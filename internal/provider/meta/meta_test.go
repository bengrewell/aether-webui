package meta

import (
	"errors"
	"os"
	"runtime"
	"testing"
	"time"
)

func newTestProvider(opts ...func(*Meta)) *Meta {
	m := NewProvider(
		VersionInfo{
			Version:    "1.2.3",
			BuildDate:  "2025-01-01",
			Branch:     "main",
			CommitHash: "abc1234",
		},
		AppConfig{
			ListenAddress:    "127.0.0.1:8680",
			DataDir:          "/tmp/test",
			TLSEnabled:       true,
			MTLSEnabled:      false,
			RBACEnabled:      true,
			DebugEnabled:     false,
			FrontendServing:  true,
			FrontendDir:      "/tmp/frontend",
			OnRampDir:        "/tmp/onramp",
			MetricsInterval:  "10s",
			MetricsRetention: "24h0m0s",
		},
		nil, // schemaVer
		nil, // providersFn
	)
	for _, fn := range opts {
		fn(m)
	}
	return m
}

func TestHandleVersion(t *testing.T) {
	m := newTestProvider()
	out, err := m.handleVersion(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Body.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", out.Body.Version, "1.2.3")
	}
	if out.Body.BuildDate != "2025-01-01" {
		t.Errorf("BuildDate = %q, want %q", out.Body.BuildDate, "2025-01-01")
	}
	if out.Body.Branch != "main" {
		t.Errorf("Branch = %q, want %q", out.Body.Branch, "main")
	}
	if out.Body.CommitHash != "abc1234" {
		t.Errorf("CommitHash = %q, want %q", out.Body.CommitHash, "abc1234")
	}
}

func TestHandleBuild(t *testing.T) {
	m := newTestProvider()
	out, err := m.handleBuild(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Body.GoVersion != runtime.Version() {
		t.Errorf("GoVersion = %q, want %q", out.Body.GoVersion, runtime.Version())
	}
	if out.Body.OS != runtime.GOOS {
		t.Errorf("OS = %q, want %q", out.Body.OS, runtime.GOOS)
	}
	if out.Body.Arch != runtime.GOARCH {
		t.Errorf("Arch = %q, want %q", out.Body.Arch, runtime.GOARCH)
	}
}

func TestHandleRuntime(t *testing.T) {
	m := newTestProvider()
	out, err := m.handleRuntime(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Body.PID != os.Getpid() {
		t.Errorf("PID = %d, want %d", out.Body.PID, os.Getpid())
	}
	if _, err := time.Parse(time.RFC3339, out.Body.StartTime); err != nil {
		t.Errorf("StartTime %q is not valid RFC3339: %v", out.Body.StartTime, err)
	}
	if out.Body.Uptime == "" {
		t.Error("Uptime is empty")
	}
	if out.Body.User.UID == "" {
		t.Error("User.UID is empty")
	}
}

func TestHandleConfig(t *testing.T) {
	schemaVer := func() (int, error) { return 5, nil }
	m := newTestProvider(func(m *Meta) { m.schemaVer = schemaVer })

	out, err := m.handleConfig(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Body.ListenAddress != "127.0.0.1:8680" {
		t.Errorf("ListenAddress = %q, want %q", out.Body.ListenAddress, "127.0.0.1:8680")
	}
	if out.Body.DataDir != "/tmp/test" {
		t.Errorf("DataDir = %q, want %q", out.Body.DataDir, "/tmp/test")
	}
	if !out.Body.TLSEnabled {
		t.Error("TLSEnabled = false, want true")
	}
	if out.Body.MTLSEnabled {
		t.Error("MTLSEnabled = true, want false")
	}
	if !out.Body.RBACEnabled {
		t.Error("RBACEnabled = false, want true")
	}
	if out.Body.DebugEnabled {
		t.Error("DebugEnabled = true, want false")
	}
	if !out.Body.FrontendServing {
		t.Error("FrontendServing = false, want true")
	}
	if out.Body.FrontendDir != "/tmp/frontend" {
		t.Errorf("FrontendDir = %q, want %q", out.Body.FrontendDir, "/tmp/frontend")
	}
	if out.Body.OnRampDir != "/tmp/onramp" {
		t.Errorf("OnRampDir = %q, want %q", out.Body.OnRampDir, "/tmp/onramp")
	}
	if out.Body.MetricsInterval != "10s" {
		t.Errorf("MetricsInterval = %q, want %q", out.Body.MetricsInterval, "10s")
	}
	if out.Body.MetricsRetention != "24h0m0s" {
		t.Errorf("MetricsRetention = %q, want %q", out.Body.MetricsRetention, "24h0m0s")
	}
	if out.Body.SchemaVersion != 5 {
		t.Errorf("SchemaVersion = %d, want %d", out.Body.SchemaVersion, 5)
	}
}

func TestHandleConfigSchemaVersionError(t *testing.T) {
	schemaVer := func() (int, error) { return 0, errors.New("db error") }
	m := newTestProvider(func(m *Meta) { m.schemaVer = schemaVer })

	out, err := m.handleConfig(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Body.SchemaVersion != 0 {
		t.Errorf("SchemaVersion = %d, want 0 on error", out.Body.SchemaVersion)
	}
}

func TestHandleConfigSchemaVersionNil(t *testing.T) {
	m := newTestProvider() // schemaVer is nil

	out, err := m.handleConfig(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Body.SchemaVersion != 0 {
		t.Errorf("SchemaVersion = %d, want 0 with nil callback", out.Body.SchemaVersion)
	}
}

func TestHandleProviders(t *testing.T) {
	providersFn := func() []ProviderStatus {
		return []ProviderStatus{
			{Name: "meta", Enabled: true, Running: false, EndpointCount: 5},
			{Name: "health", Enabled: true, Running: true, EndpointCount: 2},
		}
	}
	m := newTestProvider(func(m *Meta) { m.providersFn = providersFn })

	out, err := m.handleProviders(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Body.Providers) != 2 {
		t.Fatalf("len(Providers) = %d, want 2", len(out.Body.Providers))
	}
	if out.Body.Providers[0].Name != "meta" {
		t.Errorf("Providers[0].Name = %q, want %q", out.Body.Providers[0].Name, "meta")
	}
	if out.Body.Providers[1].EndpointCount != 2 {
		t.Errorf("Providers[1].EndpointCount = %d, want 2", out.Body.Providers[1].EndpointCount)
	}
}

func TestHandleProvidersNilCallback(t *testing.T) {
	m := newTestProvider() // providersFn is nil

	out, err := m.handleProviders(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Body.Providers == nil {
		t.Fatal("Providers is nil, want empty slice")
	}
	if len(out.Body.Providers) != 0 {
		t.Errorf("len(Providers) = %d, want 0", len(out.Body.Providers))
	}
}

func TestNewProviderEndpointCount(t *testing.T) {
	m := newTestProvider()
	descs := m.Base.Descriptors()
	if len(descs) != 5 {
		t.Errorf("registered %d descriptors, want 5", len(descs))
	}
}

func TestNewProviderEndpointPaths(t *testing.T) {
	m := newTestProvider()
	descs := m.Base.Descriptors()

	want := map[string]string{
		"meta-version":   "/api/v1/meta/version",
		"meta-build":     "/api/v1/meta/build",
		"meta-runtime":   "/api/v1/meta/runtime",
		"meta-config":    "/api/v1/meta/config",
		"meta-providers": "/api/v1/meta/providers",
	}

	got := make(map[string]string, len(descs))
	for _, d := range descs {
		got[d.OperationID] = d.HTTP.Path
	}

	for opID, wantPath := range want {
		if gotPath, ok := got[opID]; !ok {
			t.Errorf("missing endpoint %q", opID)
		} else if gotPath != wantPath {
			t.Errorf("endpoint %q path = %q, want %q", opID, gotPath, wantPath)
		}
	}
}
