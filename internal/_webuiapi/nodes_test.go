package webuiapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bengrewell/aether-webui/internal/state"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

const testEncryptionKey = "01234567890123456789012345678901"

func newNodeTestRouter(t *testing.T) (http.Handler, state.Store) {
	t.Helper()
	store, err := state.NewSQLiteStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	// Ensure local node exists
	if _, err := store.EnsureLocalNode(t.Context()); err != nil {
		t.Fatalf("EnsureLocalNode failed: %v", err)
	}

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterNodeRoutes(api, NodeRoutesDeps{
		Store:         store,
		EncryptionKey: testEncryptionKey,
	})
	RegisterOperationsRoutes(api, store)
	return router, store
}

func TestListNodesReturnsLocal(t *testing.T) {
	router, _ := newNodeTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var nodes []NodeListItem
	if err := json.NewDecoder(w.Body).Decode(&nodes); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(nodes) < 1 {
		t.Fatal("expected at least 1 node (local)")
	}
	if nodes[0].NodeType != "local" {
		t.Errorf("expected first node to be local, got %q", nodes[0].NodeType)
	}
}

func TestCreateAndGetNode(t *testing.T) {
	router, _ := newNodeTestRouter(t)

	body := `{"id":"node-1","name":"Test Node","address":"192.168.1.100","username":"admin","auth_method":"password","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("create: expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var createResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(w.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if createResp.ID != "node-1" {
		t.Errorf("expected ID 'node-1', got %q", createResp.ID)
	}

	// Get the node
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-1", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d; body: %s", getW.Code, getW.Body.String())
	}

	var nodeResp struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Address  string `json:"address"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(getW.Body).Decode(&nodeResp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if nodeResp.Address != "192.168.1.100" {
		t.Errorf("expected address '192.168.1.100', got %q", nodeResp.Address)
	}
	// Password should never appear in response
	if nodeResp.Password != "" {
		t.Error("password should not be in response")
	}
}

func TestCreateNodeDuplicateID(t *testing.T) {
	router, _ := newNodeTestRouter(t)

	body := `{"id":"dup","name":"First","address":"1.2.3.4","username":"u","auth_method":"password"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("first create: expected 200, got %d", w.Code)
	}

	// Second create with same ID should fail
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/nodes", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code == http.StatusOK {
		t.Error("expected error for duplicate ID, got 200")
	}
}

func TestGetNodeNotFound(t *testing.T) {
	router, _ := newNodeTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdateNode(t *testing.T) {
	router, _ := newNodeTestRouter(t)

	// Create a node
	createBody := `{"id":"upd","name":"Before","address":"1.1.1.1","username":"u","auth_method":"password"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/nodes", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	// Update it
	updateBody := `{"name":"After","address":"2.2.2.2"}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/nodes/upd", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateW := httptest.NewRecorder()
	router.ServeHTTP(updateW, updateReq)

	if updateW.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d; body: %s", updateW.Code, updateW.Body.String())
	}

	var resp struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	}
	if err := json.NewDecoder(updateW.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.Name != "After" {
		t.Errorf("expected name 'After', got %q", resp.Name)
	}
	if resp.Address != "2.2.2.2" {
		t.Errorf("expected address '2.2.2.2', got %q", resp.Address)
	}
}

func TestDeleteNode(t *testing.T) {
	router, _ := newNodeTestRouter(t)

	// Create a node
	body := `{"id":"del","name":"ToDelete","address":"1.1.1.1","username":"u","auth_method":"password"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/nodes", strings.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	// Delete it
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/nodes/del", nil)
	delW := httptest.NewRecorder()
	router.ServeHTTP(delW, delReq)

	if delW.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d; body: %s", delW.Code, delW.Body.String())
	}

	// Verify it's gone
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/del", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", getW.Code)
	}
}

func TestDeleteLocalNodeProtected(t *testing.T) {
	router, _ := newNodeTestRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/nodes/local", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestAssignAndRemoveRole(t *testing.T) {
	router, _ := newNodeTestRouter(t)

	// Assign role
	body := `{"role":"sd-core"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/local/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("assign: expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	// Verify role appears on node
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/local", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	var nodeResp struct {
		Roles []string `json:"roles"`
	}
	if err := json.NewDecoder(getW.Body).Decode(&nodeResp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(nodeResp.Roles) != 1 || nodeResp.Roles[0] != "sd-core" {
		t.Errorf("expected roles [sd-core], got %v", nodeResp.Roles)
	}

	// Remove role
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/nodes/local/roles/sd-core", nil)
	delW := httptest.NewRecorder()
	router.ServeHTTP(delW, delReq)

	if delW.Code != http.StatusOK {
		t.Fatalf("remove: expected 200, got %d; body: %s", delW.Code, delW.Body.String())
	}

	// Verify role removed
	getReq2 := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/local", nil)
	getW2 := httptest.NewRecorder()
	router.ServeHTTP(getW2, getReq2)

	var resp2 struct {
		Roles []string `json:"roles"`
	}
	if err := json.NewDecoder(getW2.Body).Decode(&resp2); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp2.Roles) != 0 {
		t.Errorf("expected empty roles, got %v", resp2.Roles)
	}
}

func TestAssignRoleNodeNotFound(t *testing.T) {
	router, _ := newNodeTestRouter(t)

	body := `{"role":"sd-core"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/nonexistent/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestOperationsLogPopulated(t *testing.T) {
	router, _ := newNodeTestRouter(t)

	// Create a node — generates an operation log entry
	body := `{"id":"ops-node","name":"Ops","address":"1.1.1.1","username":"u","auth_method":"password"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create: got %d", w.Code)
	}

	// Check operations log
	opsReq := httptest.NewRequest(http.MethodGet, "/api/v1/operations", nil)
	opsW := httptest.NewRecorder()
	router.ServeHTTP(opsW, opsReq)

	if opsW.Code != http.StatusOK {
		t.Fatalf("operations: expected 200, got %d", opsW.Code)
	}

	var opsResp struct {
		Operations []OperationLogItem `json:"operations"`
		Total      int                `json:"total"`
	}
	if err := json.NewDecoder(opsW.Body).Decode(&opsResp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if opsResp.Total < 1 {
		t.Error("expected at least 1 operation after creating a node")
	}
	found := false
	for _, op := range opsResp.Operations {
		if op.Operation == "create_node" && op.NodeID == "ops-node" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected create_node operation for ops-node in operations log")
	}
}

func TestOperationsLogFilterByNode(t *testing.T) {
	router, _ := newNodeTestRouter(t)

	// Create two nodes
	for _, id := range []string{"n1", "n2"} {
		body := `{"id":"` + id + `","name":"` + id + `","address":"1.1.1.1","username":"u","auth_method":"password"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Filter by n1
	opsReq := httptest.NewRequest(http.MethodGet, "/api/v1/operations?node_id=n1", nil)
	opsW := httptest.NewRecorder()
	router.ServeHTTP(opsW, opsReq)

	var opsResp struct {
		Operations []OperationLogItem `json:"operations"`
		Total      int                `json:"total"`
	}
	if err := json.NewDecoder(opsW.Body).Decode(&opsResp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	for _, op := range opsResp.Operations {
		if op.NodeID != "n1" {
			t.Errorf("expected all operations to be for n1, got node_id=%q", op.NodeID)
		}
	}
}

func TestNodeEndpointsContentType(t *testing.T) {
	router, _ := newNodeTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}
