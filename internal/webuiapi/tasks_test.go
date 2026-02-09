package webuiapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bengrewell/aether-webui/internal/onramp"
	"github.com/bengrewell/aether-webui/internal/state"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

func newTaskTestStore(t *testing.T) *state.SQLiteStore {
	t.Helper()
	store, err := state.NewSQLiteStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func newTaskTestRouter(t *testing.T, store *state.SQLiteStore) http.Handler {
	t.Helper()
	runner := onramp.NewRunner(t.TempDir())
	mgr := onramp.NewManager(onramp.Config{WorkDir: t.TempDir()}, store)
	taskMgr := onramp.NewTaskManager(store, runner, mgr)

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterTaskRoutes(api, taskMgr, store)
	return router
}

func TestListTasksEmpty(t *testing.T) {
	store := newTaskTestStore(t)
	handler := newTaskTestRouter(t, store)

	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Tasks []state.DeploymentTask `json:"tasks"`
		Total int                    `json:"total"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 0 {
		t.Errorf("total = %d, want 0", resp.Total)
	}
	if len(resp.Tasks) != 0 {
		t.Errorf("len(tasks) = %d, want 0", len(resp.Tasks))
	}
}

func TestListTasksWithData(t *testing.T) {
	store := newTaskTestStore(t)
	ctx := t.Context()

	store.CreateTask(ctx, &state.DeploymentTask{ID: "t1", Operation: "op1", Status: state.TaskStatusPending})
	store.CreateTask(ctx, &state.DeploymentTask{ID: "t2", Operation: "op2", Status: state.TaskStatusRunning})

	handler := newTaskTestRouter(t, store)
	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Tasks []state.DeploymentTask `json:"tasks"`
		Total int                    `json:"total"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}
}

func TestGetTask(t *testing.T) {
	store := newTaskTestStore(t)
	ctx := t.Context()

	store.CreateTask(ctx, &state.DeploymentTask{ID: "t1", Operation: "deploy_core", Status: state.TaskStatusRunning})

	handler := newTaskTestRouter(t, store)
	req := httptest.NewRequest("GET", "/api/v1/tasks/t1", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var task state.DeploymentTask
	json.NewDecoder(w.Body).Decode(&task)
	if task.Operation != "deploy_core" {
		t.Errorf("operation = %q, want %q", task.Operation, "deploy_core")
	}
}

func TestGetTaskNotFound(t *testing.T) {
	store := newTaskTestStore(t)
	handler := newTaskTestRouter(t, store)

	req := httptest.NewRequest("GET", "/api/v1/tasks/nonexistent", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestListDeployments(t *testing.T) {
	store := newTaskTestStore(t)
	ctx := t.Context()

	store.SetDeploymentState(ctx, "5gc", state.DeployStateDeployed, "t1")
	store.SetDeploymentState(ctx, "srsran-gnb", state.DeployStateDeploying, "t2")

	handler := newTaskTestRouter(t, store)
	req := httptest.NewRequest("GET", "/api/v1/deployments", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Deployments []state.ComponentDeploymentState `json:"deployments"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Deployments) != 2 {
		t.Errorf("len(deployments) = %d, want 2", len(resp.Deployments))
	}
}

func TestGetDeployment(t *testing.T) {
	store := newTaskTestStore(t)
	ctx := t.Context()

	store.SetDeploymentState(ctx, "5gc", state.DeployStateDeployed, "t1")

	handler := newTaskTestRouter(t, store)
	req := httptest.NewRequest("GET", "/api/v1/deployments/5gc", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var ds state.ComponentDeploymentState
	json.NewDecoder(w.Body).Decode(&ds)
	if ds.Status != state.DeployStateDeployed {
		t.Errorf("status = %q, want %q", ds.Status, state.DeployStateDeployed)
	}
}

func TestGetDeploymentNotFound(t *testing.T) {
	store := newTaskTestStore(t)
	handler := newTaskTestRouter(t, store)

	req := httptest.NewRequest("GET", "/api/v1/deployments/nonexistent", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
