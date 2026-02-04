package webuiapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bengrewell/aether-webui/internal/operator/host"
)

func TestGetCPUUsageSuccess(t *testing.T) {
	hostOp := &mockHostOperator{
		cpuUsage: &host.CPUUsage{
			UsagePercent:  45.5,
			UserPercent:   30.0,
			SystemPercent: 15.5,
			IdlePercent:   54.5,
			LoadAverage1:  1.5,
			LoadAverage5:  1.2,
			LoadAverage15: 1.0,
			PerCoreUsage:  []float64{50.0, 40.0, 45.0, 47.0},
		},
	}
	router := newMetricsTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/cpu", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp host.CPUUsage
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.UsagePercent != 45.5 {
		t.Errorf("UsagePercent = %f, want %f", resp.UsagePercent, 45.5)
	}
	if len(resp.PerCoreUsage) != 4 {
		t.Errorf("PerCoreUsage length = %d, want %d", len(resp.PerCoreUsage), 4)
	}
}

func TestGetCPUUsageError(t *testing.T) {
	hostOp := &mockHostOperator{
		cpuUsageErr: errors.New("cpu sampling error"),
	}
	router := newMetricsTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/cpu", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetMemoryUsageSuccess(t *testing.T) {
	hostOp := &mockHostOperator{
		memoryUsage: &host.MemoryUsage{
			UsedBytes:      8589934592,
			FreeBytes:      4294967296,
			AvailableBytes: 6442450944,
			CachedBytes:    2147483648,
			BuffersBytes:   268435456,
			SwapTotalBytes: 8589934592,
			SwapUsedBytes:  0,
			UsagePercent:   50.0,
		},
	}
	router := newMetricsTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/memory", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp host.MemoryUsage
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.UsagePercent != 50.0 {
		t.Errorf("UsagePercent = %f, want %f", resp.UsagePercent, 50.0)
	}
	if resp.UsedBytes != 8589934592 {
		t.Errorf("UsedBytes = %d, want %d", resp.UsedBytes, 8589934592)
	}
}

func TestGetMemoryUsageError(t *testing.T) {
	hostOp := &mockHostOperator{
		memoryUsageErr: errors.New("memory sampling error"),
	}
	router := newMetricsTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/memory", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetDiskUsageSuccess(t *testing.T) {
	hostOp := &mockHostOperator{
		diskUsage: &host.DiskUsage{
			Disks: []host.DiskUsageEntry{
				{
					Device:       "/dev/sda1",
					MountPoint:   "/",
					TotalBytes:   512000000000,
					UsedBytes:    256000000000,
					FreeBytes:    256000000000,
					UsagePercent: 50.0,
					InodesTotal:  32000000,
					InodesUsed:   1000000,
				},
			},
		},
	}
	router := newMetricsTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/disk", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp host.DiskUsage
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Disks) != 1 {
		t.Fatalf("expected 1 disk, got %d", len(resp.Disks))
	}
	if resp.Disks[0].UsagePercent != 50.0 {
		t.Errorf("UsagePercent = %f, want %f", resp.Disks[0].UsagePercent, 50.0)
	}
}

func TestGetDiskUsageError(t *testing.T) {
	hostOp := &mockHostOperator{
		diskUsageErr: errors.New("disk sampling error"),
	}
	router := newMetricsTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/disk", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetNICUsageSuccess(t *testing.T) {
	hostOp := &mockHostOperator{
		nicUsage: &host.NICUsage{
			Interfaces: []host.NICUsageEntry{
				{
					Name:          "eth0",
					RxBytes:       1073741824,
					TxBytes:       536870912,
					RxPackets:     1000000,
					TxPackets:     500000,
					RxErrors:      0,
					TxErrors:      0,
					RxDropped:     0,
					TxDropped:     0,
					RxBytesPerSec: 1048576,
					TxBytesPerSec: 524288,
				},
			},
		},
	}
	router := newMetricsTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/nic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp host.NICUsage
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(resp.Interfaces))
	}
	if resp.Interfaces[0].Name != "eth0" {
		t.Errorf("Name = %q, want %q", resp.Interfaces[0].Name, "eth0")
	}
	if resp.Interfaces[0].RxBytesPerSec != 1048576 {
		t.Errorf("RxBytesPerSec = %f, want %f", resp.Interfaces[0].RxBytesPerSec, 1048576.0)
	}
}

func TestGetNICUsageError(t *testing.T) {
	hostOp := &mockHostOperator{
		nicUsageErr: errors.New("network sampling error"),
	}
	router := newMetricsTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/nic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestMetricsEndpointsWithNodeParameter(t *testing.T) {
	hostOp := &mockHostOperator{
		cpuUsage: &host.CPUUsage{UsagePercent: 50.0},
	}
	router := newMetricsTestRouter(t, hostOp)

	// Test with explicit "local" node parameter
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/cpu?node=local", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with node=local, got %d", w.Code)
	}
}

func TestMetricsEndpointsInvalidNode(t *testing.T) {
	hostOp := &mockHostOperator{
		cpuUsage: &host.CPUUsage{UsagePercent: 50.0},
	}
	router := newMetricsTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/cpu?node=unknown-node", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 with invalid node, got %d", w.Code)
	}
}

func TestMetricsEndpointsOperatorUnavailable(t *testing.T) {
	router := newSystemTestRouterNoOperator(t)

	endpoints := []string{
		"/api/v1/metrics/cpu",
		"/api/v1/metrics/memory",
		"/api/v1/metrics/disk",
		"/api/v1/metrics/nic",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestMetricsEndpointsContentType(t *testing.T) {
	hostOp := &mockHostOperator{
		cpuUsage: &host.CPUUsage{UsagePercent: 50.0},
	}
	router := newMetricsTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/cpu", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

func TestMetricsEndpointsMethodNotAllowed(t *testing.T) {
	hostOp := &mockHostOperator{
		cpuUsage: &host.CPUUsage{UsagePercent: 50.0},
	}
	router := newMetricsTestRouter(t, hostOp)

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	endpoints := []string{
		"/api/v1/metrics/cpu",
		"/api/v1/metrics/memory",
		"/api/v1/metrics/disk",
		"/api/v1/metrics/nic",
	}

	for _, endpoint := range endpoints {
		for _, method := range methods {
			t.Run(method+" "+endpoint, func(t *testing.T) {
				req := httptest.NewRequest(method, endpoint, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != http.StatusMethodNotAllowed {
					t.Fatalf("expected 405, got %d", w.Code)
				}
			})
		}
	}
}
