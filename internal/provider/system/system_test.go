package system

import (
	"testing"
	"time"

	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/store"
)

func TestNewProvider_EndpointCount(t *testing.T) {
	p := NewProvider(Config{CollectInterval: 10 * time.Second, Retention: 24 * time.Hour})

	descs := p.Base.Descriptors()
	if len(descs) != 8 {
		t.Errorf("registered %d endpoints, want 8", len(descs))
	}
}

func TestNewProvider_EndpointPaths(t *testing.T) {
	p := NewProvider(Config{CollectInterval: 10 * time.Second, Retention: 24 * time.Hour})

	wantOps := map[string]string{
		"system-cpu":                "/api/v1/system/cpu",
		"system-memory":             "/api/v1/system/memory",
		"system-disks":              "/api/v1/system/disks",
		"system-os":                 "/api/v1/system/os",
		"system-network-interfaces": "/api/v1/system/network/interfaces",
		"system-network-config":     "/api/v1/system/network/config",
		"system-network-ports":      "/api/v1/system/network/ports",
		"system-metrics":            "/api/v1/system/metrics",
	}

	descs := p.Base.Descriptors()
	for _, d := range descs {
		want, ok := wantOps[d.OperationID]
		if !ok {
			t.Errorf("unexpected operation %q", d.OperationID)
			continue
		}
		if d.HTTP.Path != want {
			t.Errorf("operation %q path = %q, want %q", d.OperationID, d.HTTP.Path, want)
		}
		delete(wantOps, d.OperationID)
	}
	for op := range wantOps {
		t.Errorf("missing operation %q", op)
	}
}

func TestNewProvider_ImplementsInterface(t *testing.T) {
	var _ provider.Provider = NewProvider(Config{CollectInterval: 10 * time.Second})
}

func TestHandleCPU(t *testing.T) {
	p := NewProvider(Config{CollectInterval: 10 * time.Second})
	out, err := p.handleCPU(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleCPU: %v", err)
	}
	if out.Body.LogicalCores <= 0 {
		t.Error("expected at least 1 logical core")
	}
}

func TestHandleMemory(t *testing.T) {
	p := NewProvider(Config{CollectInterval: 10 * time.Second})
	out, err := p.handleMemory(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleMemory: %v", err)
	}
	if out.Body.TotalBytes == 0 {
		t.Error("expected non-zero total bytes")
	}
}

func TestHandleDisks(t *testing.T) {
	p := NewProvider(Config{CollectInterval: 10 * time.Second})
	out, err := p.handleDisks(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleDisks: %v", err)
	}
	if len(out.Body.Partitions) == 0 {
		t.Error("expected at least one partition")
	}
}

func TestHandleOS(t *testing.T) {
	p := NewProvider(Config{CollectInterval: 10 * time.Second})
	out, err := p.handleOS(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleOS: %v", err)
	}
	if out.Body.Hostname == "" {
		t.Error("expected non-empty hostname")
	}
	if out.Body.OS == "" {
		t.Error("expected non-empty os")
	}
}

func TestHandleNetworkInterfaces(t *testing.T) {
	p := NewProvider(Config{CollectInterval: 10 * time.Second})
	out, err := p.handleNetworkInterfaces(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleNetworkInterfaces: %v", err)
	}
	if len(out.Body) == 0 {
		t.Error("expected at least one network interface")
	}
}

func TestHandleNetworkConfig(t *testing.T) {
	p := NewProvider(Config{CollectInterval: 10 * time.Second})
	out, err := p.handleNetworkConfig(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleNetworkConfig: %v", err)
	}
	// Just verify the shape: slices should be non-nil.
	if out.Body.DNSServers == nil {
		t.Error("expected non-nil DNSServers slice")
	}
	if out.Body.SearchDomains == nil {
		t.Error("expected non-nil SearchDomains slice")
	}
}

func TestHandleMetricsQuery_EmptyMetric(t *testing.T) {
	p := NewProvider(Config{CollectInterval: 10 * time.Second})
	out, err := p.handleMetricsQuery(t.Context(), &MetricsQueryInput{})
	if err != nil {
		t.Fatalf("handleMetricsQuery: %v", err)
	}
	if len(out.Body.Series) != 0 {
		t.Errorf("expected empty series for empty metric, got %d", len(out.Body.Series))
	}
}

func TestHandleMetricsQuery_WithStore(t *testing.T) {
	ctx := t.Context()
	dbPath := t.TempDir() + "/test.db"
	st, err := store.New(ctx, dbPath)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	defer st.Close()

	now := time.Now()
	err = st.AppendSamples(ctx, []store.Sample{
		{Metric: "test.metric", TS: now, Value: 42.0, Unit: "count"},
	})
	if err != nil {
		t.Fatalf("AppendSamples: %v", err)
	}

	p := NewProvider(Config{CollectInterval: 10 * time.Second}, provider.WithStore(st))
	out, err := p.handleMetricsQuery(ctx, &MetricsQueryInput{
		Metric: "test.metric",
		From:   now.Add(-1 * time.Minute).Format(time.RFC3339),
		To:     now.Add(1 * time.Minute).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("handleMetricsQuery: %v", err)
	}
	if len(out.Body.Series) == 0 {
		t.Fatal("expected at least one series")
	}
	if len(out.Body.Series[0].Points) != 1 {
		t.Errorf("expected 1 point, got %d", len(out.Body.Series[0].Points))
	}
	if out.Body.Series[0].Points[0].Value != 42.0 {
		t.Errorf("expected value 42, got %f", out.Body.Series[0].Points[0].Value)
	}
}

func TestCollector_StartStop(t *testing.T) {
	ctx := t.Context()
	dbPath := t.TempDir() + "/test.db"
	st, err := store.New(ctx, dbPath)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	defer st.Close()

	p := NewProvider(Config{
		CollectInterval: 50 * time.Millisecond,
		Retention:       24 * time.Hour,
	}, provider.WithStore(st))

	if err := p.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Let the collector run a few ticks.
	time.Sleep(200 * time.Millisecond)

	if err := p.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	// Verify some metrics were collected.
	series, err := st.QueryRange(ctx, store.RangeQuery{
		Metric: "system.cpu.usage_percent",
		Range: store.TimeRange{
			From: time.Now().Add(-1 * time.Minute),
			To:   time.Now().Add(1 * time.Minute),
		},
		LabelsExact: map[string]string{"cpu": "total"},
	})
	if err != nil {
		t.Fatalf("QueryRange: %v", err)
	}
	if len(series) == 0 || len(series[0].Points) == 0 {
		t.Error("expected collected CPU metrics, got none")
	}
}

func TestParseLabels(t *testing.T) {
	tests := []struct {
		input string
		want  map[string]string
	}{
		{"", nil},
		{"cpu=total", map[string]string{"cpu": "total"}},
		{"cpu=total, device=sda", map[string]string{"cpu": "total", "device": "sda"}},
		{"bad_format", nil},
	}
	for _, tt := range tests {
		got := parseLabels(tt.input)
		if tt.want == nil && got != nil {
			t.Errorf("parseLabels(%q) = %v, want nil", tt.input, got)
		}
		if tt.want != nil {
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("parseLabels(%q)[%q] = %q, want %q", tt.input, k, got[k], v)
				}
			}
		}
	}
}

func TestParseAggregation(t *testing.T) {
	tests := []struct {
		input string
		want  store.Agg
	}{
		{"", store.AggRaw},
		{"raw", store.AggRaw},
		{"10s", store.Agg10s},
		{"1m", store.Agg1m},
		{"5m", store.Agg5m},
		{"1h", store.Agg1h},
		{"unknown", store.AggRaw},
	}
	for _, tt := range tests {
		got := parseAggregation(tt.input)
		if got != tt.want {
			t.Errorf("parseAggregation(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{9, "9"},
		{10, "10"},
		{42, "42"},
		{100, "100"},
	}
	for _, tt := range tests {
		got := itoa(tt.input)
		if got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
