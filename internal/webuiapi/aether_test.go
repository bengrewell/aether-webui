package webuiapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bengrewell/aether-webui/internal/operator"
	"github.com/bengrewell/aether-webui/internal/operator/aether"
	"github.com/bengrewell/aether-webui/internal/provider"
)

// --- List Nodes ---

func TestListNodesSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{}
	localProvider := &mockProvider{
		id:      provider.LocalNode,
		isLocal: true,
		operators: map[operator.Domain]operator.Operator{
			operator.DomainAether: aetherOp,
		},
	}
	resolver := newTestResolver(localProvider)
	resolver.RegisterNode("remote-1", &mockProvider{id: "remote-1"})
	resolver.RegisterNode("remote-2", &mockProvider{id: "remote-2"})

	router := newAetherTestRouterWithResolver(t, resolver)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/nodes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Nodes []string `json:"nodes"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(resp.Nodes))
	}
	// Verify local is included
	hasLocal := false
	for _, n := range resp.Nodes {
		if n == "local" {
			hasLocal = true
		}
	}
	if !hasLocal {
		t.Error("expected 'local' node in list")
	}
}

// --- Core Operations ---

func TestListCoresSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		cores: &aether.CoreList{
			Cores: []aether.CoreConfig{
				{ID: "core-1", Name: "Test Core"},
				{ID: "core-2", Name: "Another Core"},
			},
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp aether.CoreList
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Cores) != 2 {
		t.Fatalf("expected 2 cores, got %d", len(resp.Cores))
	}
}

func TestListCoresError(t *testing.T) {
	aetherOp := &mockAetherOperator{
		coresErr: errors.New("failed to list cores"),
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetCoreSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		core: &aether.CoreConfig{
			ID:         "core-1",
			Name:       "Test Core",
			Standalone: true,
			DataIface:  "eth0",
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core/core-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp aether.CoreConfig
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.ID != "core-1" {
		t.Errorf("ID = %q, want %q", resp.ID, "core-1")
	}
}

func TestGetCoreNotFound(t *testing.T) {
	aetherOp := &mockAetherOperator{
		coreErr: errors.New("core not found"),
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core/unknown", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeployCoreSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		deployCore: &aether.DeploymentResponse{
			Success: true,
			Message: "Core deployed successfully",
			ID:      "core-123",
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	body := `{"name":"Test Core","standalone":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/aether/core", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp aether.DeploymentResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.ID != "core-123" {
		t.Errorf("ID = %q, want %q", resp.ID, "core-123")
	}
}

func TestDeployCoreWithoutBody(t *testing.T) {
	aetherOp := &mockAetherOperator{
		deployCore: &aether.DeploymentResponse{
			Success: true,
			Message: "Core deployed with defaults",
			ID:      "core-456",
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/aether/core", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestDeployCoreError(t *testing.T) {
	aetherOp := &mockAetherOperator{
		deployCoreErr: errors.New("deployment failed"),
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/aether/core", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateCoreSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		updateCoreErr: nil,
	}
	router := newAetherTestRouter(t, aetherOp)

	body := `{"name":"Updated Core","standalone":false}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/aether/core/core-1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp aether.CoreConfig
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	// ID should be set from path parameter
	if resp.ID != "core-1" {
		t.Errorf("ID = %q, want %q", resp.ID, "core-1")
	}
}

func TestUpdateCoreNotFound(t *testing.T) {
	aetherOp := &mockAetherOperator{
		updateCoreErr: errors.New("core not found"),
	}
	router := newAetherTestRouter(t, aetherOp)

	body := `{"name":"Updated Core"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/aether/core/unknown", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUndeployCoreSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		undeployCore: &aether.DeploymentResponse{
			Success: true,
			Message: "Core undeployed successfully",
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/aether/core/core-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestUndeployCoreNotFound(t *testing.T) {
	aetherOp := &mockAetherOperator{
		undeployCoreErr: errors.New("core not found"),
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/aether/core/unknown", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetCoreStatusSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		coreStatus: &aether.CoreStatus{
			ID:    "core-1",
			Name:  "Test Core",
			Host:  "localhost",
			State: aether.StateDeployed,
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core/core-1/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp aether.CoreStatus
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.State != aether.StateDeployed {
		t.Errorf("State = %q, want %q", resp.State, aether.StateDeployed)
	}
}

func TestGetCoreStatusNotFound(t *testing.T) {
	aetherOp := &mockAetherOperator{
		coreStatusErr: errors.New("core not found"),
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core/unknown/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestListCoreStatusesSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		coreStatuses: &aether.CoreStatusList{
			Cores: []aether.CoreStatus{
				{ID: "core-1", State: aether.StateDeployed},
				{ID: "core-2", State: aether.StateDeploying},
			},
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp aether.CoreStatusList
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Cores) != 2 {
		t.Fatalf("expected 2 core statuses, got %d", len(resp.Cores))
	}
}

func TestListCoreStatusesError(t *testing.T) {
	aetherOp := &mockAetherOperator{
		coreStatusesErr: errors.New("failed to list statuses"),
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// --- gNB Operations ---

func TestListGNBsSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		gnbs: &aether.GNBList{
			GNBs: []aether.GNBConfig{
				{ID: "gnb-1", Name: "Test gNB", Type: "srsran"},
				{ID: "gnb-2", Name: "Another gNB", Type: "ocudu"},
			},
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp aether.GNBList
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.GNBs) != 2 {
		t.Fatalf("expected 2 gNBs, got %d", len(resp.GNBs))
	}
}

func TestListGNBsError(t *testing.T) {
	aetherOp := &mockAetherOperator{
		gnbsErr: errors.New("failed to list gNBs"),
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetGNBSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		gnb: &aether.GNBConfig{
			ID:   "gnb-1",
			Name: "Test gNB",
			Type: "srsran",
			IP:   "192.168.1.100",
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb/gnb-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp aether.GNBConfig
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.ID != "gnb-1" {
		t.Errorf("ID = %q, want %q", resp.ID, "gnb-1")
	}
}

func TestGetGNBNotFound(t *testing.T) {
	aetherOp := &mockAetherOperator{
		gnbErr: errors.New("gNB not found"),
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb/unknown", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeployGNBSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		deployGNB: &aether.DeploymentResponse{
			Success: true,
			Message: "gNB deployed successfully",
			ID:      "gnb-123",
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	body := `{"name":"Test gNB","type":"srsran"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/aether/gnb", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp aether.DeploymentResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestDeployGNBWithoutBody(t *testing.T) {
	aetherOp := &mockAetherOperator{
		deployGNB: &aether.DeploymentResponse{
			Success: true,
			Message: "gNB deployed with defaults",
			ID:      "gnb-456",
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/aether/gnb", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestDeployGNBError(t *testing.T) {
	aetherOp := &mockAetherOperator{
		deployGNBErr: errors.New("deployment failed"),
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/aether/gnb", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateGNBSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		updateGNBErr: nil,
	}
	router := newAetherTestRouter(t, aetherOp)

	body := `{"name":"Updated gNB","type":"ocudu"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/aether/gnb/gnb-1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp aether.GNBConfig
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.ID != "gnb-1" {
		t.Errorf("ID = %q, want %q", resp.ID, "gnb-1")
	}
}

func TestUpdateGNBNotFound(t *testing.T) {
	aetherOp := &mockAetherOperator{
		updateGNBErr: errors.New("gNB not found"),
	}
	router := newAetherTestRouter(t, aetherOp)

	body := `{"name":"Updated gNB"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/aether/gnb/unknown", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUndeployGNBSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		undeployGNB: &aether.DeploymentResponse{
			Success: true,
			Message: "gNB undeployed successfully",
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/aether/gnb/gnb-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestUndeployGNBNotFound(t *testing.T) {
	aetherOp := &mockAetherOperator{
		undeployGNBErr: errors.New("gNB not found"),
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/aether/gnb/unknown", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetGNBStatusSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		gnbStatus: &aether.GNBStatus{
			ID:        "gnb-1",
			Name:      "Test gNB",
			Host:      "localhost",
			Type:      "srsran",
			State:     aether.StateDeployed,
			Connected: true,
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb/gnb-1/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp aether.GNBStatus
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !resp.Connected {
		t.Error("expected Connected=true")
	}
}

func TestGetGNBStatusNotFound(t *testing.T) {
	aetherOp := &mockAetherOperator{
		gnbStatusErr: errors.New("gNB not found"),
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb/unknown/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestListGNBStatusesSuccess(t *testing.T) {
	aetherOp := &mockAetherOperator{
		gnbStatuses: &aether.GNBStatusList{
			GNBs: []aether.GNBStatus{
				{ID: "gnb-1", State: aether.StateDeployed, Connected: true},
				{ID: "gnb-2", State: aether.StateDeploying, Connected: false},
			},
		},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp aether.GNBStatusList
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.GNBs) != 2 {
		t.Fatalf("expected 2 gNB statuses, got %d", len(resp.GNBs))
	}
}

func TestListGNBStatusesError(t *testing.T) {
	aetherOp := &mockAetherOperator{
		gnbStatusesErr: errors.New("failed to list statuses"),
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// --- Common Error Cases ---

func TestAetherEndpointsInvalidNode(t *testing.T) {
	aetherOp := &mockAetherOperator{
		cores: &aether.CoreList{},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core?node=unknown-node", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 with invalid node, got %d", w.Code)
	}
}

func TestAetherEndpointsOperatorUnavailable(t *testing.T) {
	router := newAetherTestRouterNoOperator(t)

	endpoints := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/v1/aether/core", ""},
		{http.MethodGet, "/api/v1/aether/core/test-id", ""},
		{http.MethodPost, "/api/v1/aether/core", ""},
		{http.MethodPut, "/api/v1/aether/core/test-id", `{"name":"test"}`},
		{http.MethodDelete, "/api/v1/aether/core/test-id", ""},
		{http.MethodGet, "/api/v1/aether/core/test-id/status", ""},
		{http.MethodGet, "/api/v1/aether/core/status", ""},
		{http.MethodGet, "/api/v1/aether/gnb", ""},
		{http.MethodGet, "/api/v1/aether/gnb/test-id", ""},
		{http.MethodPost, "/api/v1/aether/gnb", ""},
		{http.MethodPut, "/api/v1/aether/gnb/test-id", `{"name":"test"}`},
		{http.MethodDelete, "/api/v1/aether/gnb/test-id", ""},
		{http.MethodGet, "/api/v1/aether/gnb/test-id/status", ""},
		{http.MethodGet, "/api/v1/aether/gnb/status", ""},
	}

	for _, tc := range endpoints {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestAetherEndpointsContentType(t *testing.T) {
	aetherOp := &mockAetherOperator{
		cores: &aether.CoreList{},
	}
	router := newAetherTestRouter(t, aetherOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}
