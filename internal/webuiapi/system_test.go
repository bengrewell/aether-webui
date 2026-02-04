package webuiapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bengrewell/aether-webui/internal/operator/host"
)

func TestGetCPUInfoSuccess(t *testing.T) {
	hostOp := &mockHostOperator{
		cpuInfo: &host.CPUInfo{
			Model:        "Intel Xeon",
			Vendor:       "Intel",
			Cores:        8,
			Threads:      16,
			FrequencyMHz: 3200,
			CacheKB:      16384,
		},
	}
	router := newSystemTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/cpu", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp host.CPUInfo
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.Model != "Intel Xeon" {
		t.Errorf("Model = %q, want %q", resp.Model, "Intel Xeon")
	}
	if resp.Cores != 8 {
		t.Errorf("Cores = %d, want %d", resp.Cores, 8)
	}
}

func TestGetCPUInfoError(t *testing.T) {
	hostOp := &mockHostOperator{
		cpuInfoErr: errors.New("hardware error"),
	}
	router := newSystemTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/cpu", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetMemoryInfoSuccess(t *testing.T) {
	hostOp := &mockHostOperator{
		memoryInfo: &host.MemoryInfo{
			TotalBytes: 34359738368,
			Type:       "DDR4",
			SpeedMHz:   3200,
			Slots:      4,
			SlotsUsed:  2,
		},
	}
	router := newSystemTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/memory", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp host.MemoryInfo
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.TotalBytes != 34359738368 {
		t.Errorf("TotalBytes = %d, want %d", resp.TotalBytes, 34359738368)
	}
}

func TestGetMemoryInfoError(t *testing.T) {
	hostOp := &mockHostOperator{
		memoryInfoErr: errors.New("memory read error"),
	}
	router := newSystemTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/memory", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetDiskInfoSuccess(t *testing.T) {
	hostOp := &mockHostOperator{
		diskInfo: &host.DiskInfo{
			Disks: []host.Disk{
				{Device: "/dev/sda", Model: "Samsung SSD", SizeBytes: 512000000000, Type: "ssd"},
			},
		},
	}
	router := newSystemTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/disk", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp host.DiskInfo
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Disks) != 1 {
		t.Fatalf("expected 1 disk, got %d", len(resp.Disks))
	}
	if resp.Disks[0].Device != "/dev/sda" {
		t.Errorf("Device = %q, want %q", resp.Disks[0].Device, "/dev/sda")
	}
}

func TestGetDiskInfoError(t *testing.T) {
	hostOp := &mockHostOperator{
		diskInfoErr: errors.New("disk read error"),
	}
	router := newSystemTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/disk", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetNICInfoSuccess(t *testing.T) {
	hostOp := &mockHostOperator{
		nicInfo: &host.NICInfo{
			Interfaces: []host.NetworkInterface{
				{Name: "eth0", MACAddress: "00:11:22:33:44:55", SpeedMbps: 1000, MTU: 1500},
			},
		},
	}
	router := newSystemTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/nic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp host.NICInfo
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(resp.Interfaces))
	}
	if resp.Interfaces[0].Name != "eth0" {
		t.Errorf("Name = %q, want %q", resp.Interfaces[0].Name, "eth0")
	}
}

func TestGetNICInfoError(t *testing.T) {
	hostOp := &mockHostOperator{
		nicInfoErr: errors.New("network error"),
	}
	router := newSystemTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/nic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetOSInfoSuccess(t *testing.T) {
	hostOp := &mockHostOperator{
		osInfo: &host.OSInfo{
			Name:         "Ubuntu",
			Version:      "22.04",
			Kernel:       "5.15.0-generic",
			Architecture: "x86_64",
			Hostname:     "testhost",
			Uptime:       86400,
		},
	}
	router := newSystemTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/os", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp host.OSInfo
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.Name != "Ubuntu" {
		t.Errorf("Name = %q, want %q", resp.Name, "Ubuntu")
	}
	if resp.Hostname != "testhost" {
		t.Errorf("Hostname = %q, want %q", resp.Hostname, "testhost")
	}
}

func TestGetOSInfoError(t *testing.T) {
	hostOp := &mockHostOperator{
		osInfoErr: errors.New("os read error"),
	}
	router := newSystemTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/os", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestSystemEndpointsWithNodeParameter(t *testing.T) {
	hostOp := &mockHostOperator{
		cpuInfo: &host.CPUInfo{Model: "Test CPU"},
	}
	router := newSystemTestRouter(t, hostOp)

	// Test with explicit "local" node parameter
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/cpu?node=local", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with node=local, got %d", w.Code)
	}
}

func TestSystemEndpointsInvalidNode(t *testing.T) {
	hostOp := &mockHostOperator{
		cpuInfo: &host.CPUInfo{Model: "Test CPU"},
	}
	router := newSystemTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/cpu?node=unknown-node", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 with invalid node, got %d", w.Code)
	}
}

func TestSystemEndpointsOperatorUnavailable(t *testing.T) {
	router := newSystemTestRouterNoOperator(t)

	endpoints := []string{
		"/api/v1/system/cpu",
		"/api/v1/system/memory",
		"/api/v1/system/disk",
		"/api/v1/system/nic",
		"/api/v1/system/os",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// When operator is not available, getHostOperator returns 503
			// but the handler wraps it as 400 (invalid node)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestSystemEndpointsContentType(t *testing.T) {
	hostOp := &mockHostOperator{
		cpuInfo: &host.CPUInfo{Model: "Test CPU"},
	}
	router := newSystemTestRouter(t, hostOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/cpu", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

func TestSystemEndpointsMethodNotAllowed(t *testing.T) {
	hostOp := &mockHostOperator{
		cpuInfo: &host.CPUInfo{Model: "Test CPU"},
	}
	router := newSystemTestRouter(t, hostOp)

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	endpoints := []string{
		"/api/v1/system/cpu",
		"/api/v1/system/memory",
		"/api/v1/system/disk",
		"/api/v1/system/nic",
		"/api/v1/system/os",
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
