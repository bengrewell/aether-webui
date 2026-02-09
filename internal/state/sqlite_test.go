package state

import (
	"context"
	"errors"
	"testing"
	"time"
)

// newTestStore creates a SQLiteStore backed by a temporary directory for testing.
func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	store, err := NewSQLiteStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

// --- Key-value state ---

func TestGetStateNotFound(t *testing.T) {
	store := newTestStore(t)
	_, err := store.GetState(context.Background(), "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSetAndGetState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.SetState(ctx, "key1", "value1"); err != nil {
		t.Fatalf("SetState failed: %v", err)
	}

	val, err := store.GetState(ctx, "key1")
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got '%s'", val)
	}
}

func TestSetStateOverwrite(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.SetState(ctx, "key1", "first"); err != nil {
		t.Fatalf("SetState (first) failed: %v", err)
	}
	if err := store.SetState(ctx, "key1", "second"); err != nil {
		t.Fatalf("SetState (second) failed: %v", err)
	}

	val, err := store.GetState(ctx, "key1")
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if val != "second" {
		t.Errorf("expected 'second', got '%s'", val)
	}
}

func TestDeleteState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.SetState(ctx, "key1", "value1"); err != nil {
		t.Fatalf("SetState failed: %v", err)
	}
	if err := store.DeleteState(ctx, "key1"); err != nil {
		t.Fatalf("DeleteState failed: %v", err)
	}

	_, err := store.GetState(ctx, "key1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDeleteStateNonExistent(t *testing.T) {
	store := newTestStore(t)
	// Deleting a key that was never set should succeed (no-op).
	if err := store.DeleteState(context.Background(), "nope"); err != nil {
		t.Errorf("expected no error deleting non-existent key, got %v", err)
	}
}

// --- Wizard status ---

func TestGetWizardStatusDefault(t *testing.T) {
	store := newTestStore(t)
	status, err := store.GetWizardStatus(context.Background())
	if err != nil {
		t.Fatalf("GetWizardStatus failed: %v", err)
	}
	if status.Completed {
		t.Error("expected Completed=false on fresh store")
	}
	if status.CompletedAt != nil {
		t.Error("expected CompletedAt=nil on fresh store")
	}
	if len(status.Steps) != 0 {
		t.Errorf("expected no steps on fresh store, got %v", status.Steps)
	}
}

func TestSetWizardCompleteAndGet(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	steps := []string{"network", "storage", "dns"}
	before := time.Now().UTC().Add(-time.Second)

	if err := store.SetWizardComplete(ctx, steps); err != nil {
		t.Fatalf("SetWizardComplete failed: %v", err)
	}

	status, err := store.GetWizardStatus(ctx)
	if err != nil {
		t.Fatalf("GetWizardStatus failed: %v", err)
	}
	if !status.Completed {
		t.Error("expected Completed=true")
	}
	if status.CompletedAt == nil {
		t.Fatal("expected CompletedAt to be set")
	}
	if status.CompletedAt.Before(before) {
		t.Errorf("CompletedAt %v is before test start %v", status.CompletedAt, before)
	}
	if len(status.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(status.Steps))
	}
	for i, want := range steps {
		if status.Steps[i] != want {
			t.Errorf("step[%d]: expected %q, got %q", i, want, status.Steps[i])
		}
	}
}

func TestSetWizardCompleteNoSteps(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.SetWizardComplete(ctx, nil); err != nil {
		t.Fatalf("SetWizardComplete failed: %v", err)
	}

	status, err := store.GetWizardStatus(ctx)
	if err != nil {
		t.Fatalf("GetWizardStatus failed: %v", err)
	}
	if !status.Completed {
		t.Error("expected Completed=true")
	}
	if len(status.Steps) != 0 {
		t.Errorf("expected no steps, got %v", status.Steps)
	}
}

func TestClearWizardStatus(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.SetWizardComplete(ctx, []string{"step1"}); err != nil {
		t.Fatalf("SetWizardComplete failed: %v", err)
	}
	if err := store.ClearWizardStatus(ctx); err != nil {
		t.Fatalf("ClearWizardStatus failed: %v", err)
	}

	status, err := store.GetWizardStatus(ctx)
	if err != nil {
		t.Fatalf("GetWizardStatus failed: %v", err)
	}
	if status.Completed {
		t.Error("expected Completed=false after clear")
	}
	if status.CompletedAt != nil {
		t.Error("expected CompletedAt=nil after clear")
	}
	if len(status.Steps) != 0 {
		t.Errorf("expected no steps after clear, got %v", status.Steps)
	}
}

// --- System info cache ---

func TestGetCachedSystemInfoNotFound(t *testing.T) {
	store := newTestStore(t)
	_, _, err := store.GetCachedSystemInfo(context.Background(), "cpu")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSetAndGetCachedSystemInfo(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	data := []byte(`{"cores":4,"model":"test"}`)
	before := time.Now().UTC().Add(-time.Second)

	if err := store.SetCachedSystemInfo(ctx, "cpu", data); err != nil {
		t.Fatalf("SetCachedSystemInfo failed: %v", err)
	}

	got, collectedAt, err := store.GetCachedSystemInfo(ctx, "cpu")
	if err != nil {
		t.Fatalf("GetCachedSystemInfo failed: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("expected data %q, got %q", data, got)
	}
	if collectedAt.Before(before) {
		t.Errorf("collectedAt %v is before test start %v", collectedAt, before)
	}
}

func TestSetCachedSystemInfoOverwrite(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.SetCachedSystemInfo(ctx, "cpu", []byte(`{"cores":2}`)); err != nil {
		t.Fatalf("SetCachedSystemInfo (first) failed: %v", err)
	}

	updated := []byte(`{"cores":8}`)
	if err := store.SetCachedSystemInfo(ctx, "cpu", updated); err != nil {
		t.Fatalf("SetCachedSystemInfo (second) failed: %v", err)
	}

	got, _, err := store.GetCachedSystemInfo(ctx, "cpu")
	if err != nil {
		t.Fatalf("GetCachedSystemInfo failed: %v", err)
	}
	if string(got) != string(updated) {
		t.Errorf("expected %q, got %q", updated, got)
	}
}

// --- Metrics history ---

func TestRecordAndGetMetrics(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := store.RecordMetrics(ctx, "cpu", []byte(`{"usage":50}`)); err != nil {
			t.Fatalf("RecordMetrics[%d] failed: %v", i, err)
		}
	}

	snapshots, err := store.GetMetricsHistory(ctx, "cpu", 10)
	if err != nil {
		t.Fatalf("GetMetricsHistory failed: %v", err)
	}
	if len(snapshots) != 3 {
		t.Errorf("expected 3 snapshots, got %d", len(snapshots))
	}
}

func TestGetMetricsHistoryOrdering(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Insert with explicit pauses so timestamps differ — but SQLite
	// CURRENT_TIMESTAMP has second resolution, so we just verify the
	// query returns newest-first (descending).
	if err := store.RecordMetrics(ctx, "mem", []byte(`{"val":1}`)); err != nil {
		t.Fatalf("RecordMetrics failed: %v", err)
	}
	if err := store.RecordMetrics(ctx, "mem", []byte(`{"val":2}`)); err != nil {
		t.Fatalf("RecordMetrics failed: %v", err)
	}

	snapshots, err := store.GetMetricsHistory(ctx, "mem", 10)
	if err != nil {
		t.Fatalf("GetMetricsHistory failed: %v", err)
	}
	if len(snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snapshots))
	}
	// Because of autoincrement IDs and descending order, the second insert
	// should appear first. Even if timestamps are identical, ORDER BY
	// recorded_at DESC with same timestamp gives stable ordering via rowid.
	if string(snapshots[0].Data) != `{"val":2}` {
		t.Errorf("expected newest first, got %s", snapshots[0].Data)
	}
}

func TestGetMetricsHistoryLimit(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if err := store.RecordMetrics(ctx, "cpu", []byte(`{}`)); err != nil {
			t.Fatalf("RecordMetrics[%d] failed: %v", i, err)
		}
	}

	snapshots, err := store.GetMetricsHistory(ctx, "cpu", 3)
	if err != nil {
		t.Fatalf("GetMetricsHistory failed: %v", err)
	}
	if len(snapshots) != 3 {
		t.Errorf("expected limit=3 to return 3, got %d", len(snapshots))
	}
}

func TestGetMetricsHistoryByType(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.RecordMetrics(ctx, "cpu", []byte(`{"cpu":1}`)); err != nil {
		t.Fatalf("RecordMetrics(cpu) failed: %v", err)
	}
	if err := store.RecordMetrics(ctx, "memory", []byte(`{"mem":1}`)); err != nil {
		t.Fatalf("RecordMetrics(memory) failed: %v", err)
	}
	if err := store.RecordMetrics(ctx, "cpu", []byte(`{"cpu":2}`)); err != nil {
		t.Fatalf("RecordMetrics(cpu) failed: %v", err)
	}

	cpuSnaps, err := store.GetMetricsHistory(ctx, "cpu", 10)
	if err != nil {
		t.Fatalf("GetMetricsHistory(cpu) failed: %v", err)
	}
	if len(cpuSnaps) != 2 {
		t.Errorf("expected 2 cpu snapshots, got %d", len(cpuSnaps))
	}
	for _, snap := range cpuSnaps {
		if snap.MetricType != "cpu" {
			t.Errorf("expected MetricType 'cpu', got %q", snap.MetricType)
		}
	}

	memSnaps, err := store.GetMetricsHistory(ctx, "memory", 10)
	if err != nil {
		t.Fatalf("GetMetricsHistory(memory) failed: %v", err)
	}
	if len(memSnaps) != 1 {
		t.Errorf("expected 1 memory snapshot, got %d", len(memSnaps))
	}
}

func TestPruneOldMetrics(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Insert a metric, then manually backdate it.
	if err := store.RecordMetrics(ctx, "cpu", []byte(`{"old":true}`)); err != nil {
		t.Fatalf("RecordMetrics failed: %v", err)
	}
	_, err := store.db.ExecContext(ctx,
		"UPDATE metrics_history SET recorded_at = datetime('now', '-2 hours')")
	if err != nil {
		t.Fatalf("backdate failed: %v", err)
	}

	// Insert a recent metric.
	if err := store.RecordMetrics(ctx, "cpu", []byte(`{"recent":true}`)); err != nil {
		t.Fatalf("RecordMetrics failed: %v", err)
	}

	// Prune metrics older than 1 hour.
	if err := store.PruneOldMetrics(ctx, time.Hour); err != nil {
		t.Fatalf("PruneOldMetrics failed: %v", err)
	}

	snapshots, err := store.GetMetricsHistory(ctx, "cpu", 10)
	if err != nil {
		t.Fatalf("GetMetricsHistory failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected 1 remaining snapshot, got %d", len(snapshots))
	}
	if string(snapshots[0].Data) != `{"recent":true}` {
		t.Errorf("expected recent snapshot to survive, got %s", snapshots[0].Data)
	}
}

func TestGetMetricsRange(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Insert metrics at different times
	// First, insert a metric and backdate it to 2 hours ago
	if err := store.RecordMetrics(ctx, "cpu", []byte(`{"val":1}`)); err != nil {
		t.Fatalf("RecordMetrics failed: %v", err)
	}
	_, err := store.db.ExecContext(ctx,
		"UPDATE metrics_history SET recorded_at = datetime('now', '-2 hours') WHERE id = 1")
	if err != nil {
		t.Fatalf("backdate failed: %v", err)
	}

	// Insert another at 30 minutes ago
	if err := store.RecordMetrics(ctx, "cpu", []byte(`{"val":2}`)); err != nil {
		t.Fatalf("RecordMetrics failed: %v", err)
	}
	_, err = store.db.ExecContext(ctx,
		"UPDATE metrics_history SET recorded_at = datetime('now', '-30 minutes') WHERE id = 2")
	if err != nil {
		t.Fatalf("backdate failed: %v", err)
	}

	// Insert a recent one
	if err := store.RecordMetrics(ctx, "cpu", []byte(`{"val":3}`)); err != nil {
		t.Fatalf("RecordMetrics failed: %v", err)
	}

	// Query for last hour - should get val:2 and val:3
	end := time.Now().UTC()
	start := end.Add(-time.Hour)

	snapshots, err := store.GetMetricsRange(ctx, "cpu", start, end)
	if err != nil {
		t.Fatalf("GetMetricsRange failed: %v", err)
	}
	if len(snapshots) != 2 {
		t.Fatalf("expected 2 snapshots in last hour, got %d", len(snapshots))
	}
	// Should be sorted ascending by time
	if string(snapshots[0].Data) != `{"val":2}` {
		t.Errorf("expected first snapshot to be val:2, got %s", snapshots[0].Data)
	}
	if string(snapshots[1].Data) != `{"val":3}` {
		t.Errorf("expected second snapshot to be val:3, got %s", snapshots[1].Data)
	}
}

func TestGetMetricsRangeEmpty(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Query for non-existent metric type
	end := time.Now().UTC()
	start := end.Add(-time.Hour)

	snapshots, err := store.GetMetricsRange(ctx, "nonexistent", start, end)
	if err != nil {
		t.Fatalf("GetMetricsRange failed: %v", err)
	}
	if len(snapshots) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snapshots))
	}
}

// --- Node management ---

func TestCreateAndGetNode(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	node := &Node{
		ID:       "node-1",
		Name:     "Test Node",
		NodeType: NodeTypeRemote,
		Address:  "192.168.1.100",
		SSHPort:  22,
		Username: "admin",
	}
	if err := store.CreateNode(ctx, node); err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}

	got, err := store.GetNode(ctx, "node-1")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if got.Name != "Test Node" {
		t.Errorf("expected name 'Test Node', got %q", got.Name)
	}
	if got.NodeType != NodeTypeRemote {
		t.Errorf("expected node_type 'remote', got %q", got.NodeType)
	}
	if got.Address != "192.168.1.100" {
		t.Errorf("expected address '192.168.1.100', got %q", got.Address)
	}
}

func TestCreateNodeDuplicateID(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	node := &Node{ID: "dup", Name: "First", NodeType: NodeTypeRemote, Address: "1.2.3.4"}
	if err := store.CreateNode(ctx, node); err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}
	node2 := &Node{ID: "dup", Name: "Second", NodeType: NodeTypeRemote, Address: "5.6.7.8"}
	if err := store.CreateNode(ctx, node2); err == nil {
		t.Error("expected error for duplicate ID")
	}
}

func TestGetNodeNotFound(t *testing.T) {
	store := newTestStore(t)
	_, err := store.GetNode(context.Background(), "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListNodesOrdering(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Ensure local node
	if _, err := store.EnsureLocalNode(ctx); err != nil {
		t.Fatalf("EnsureLocalNode failed: %v", err)
	}

	// Create remote nodes in reverse alphabetical order
	for _, name := range []string{"Zulu", "Alpha", "Mike"} {
		n := &Node{ID: "node-" + name, Name: name, NodeType: NodeTypeRemote, Address: "1.2.3.4"}
		if err := store.CreateNode(ctx, n); err != nil {
			t.Fatalf("CreateNode(%s) failed: %v", name, err)
		}
	}

	nodes, err := store.ListNodes(ctx)
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}
	if len(nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(nodes))
	}
	// Local should be first
	if nodes[0].NodeType != NodeTypeLocal {
		t.Errorf("expected first node to be local, got %q", nodes[0].NodeType)
	}
	// Remote nodes should be alphabetical
	if nodes[1].Name != "Alpha" {
		t.Errorf("expected second node to be 'Alpha', got %q", nodes[1].Name)
	}
	if nodes[2].Name != "Mike" {
		t.Errorf("expected third node to be 'Mike', got %q", nodes[2].Name)
	}
	if nodes[3].Name != "Zulu" {
		t.Errorf("expected fourth node to be 'Zulu', got %q", nodes[3].Name)
	}
}

func TestUpdateNode(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	node := &Node{ID: "upd", Name: "Before", NodeType: NodeTypeRemote, Address: "1.1.1.1", SSHPort: 22}
	if err := store.CreateNode(ctx, node); err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}

	node.Name = "After"
	node.Address = "2.2.2.2"
	if err := store.UpdateNode(ctx, node); err != nil {
		t.Fatalf("UpdateNode failed: %v", err)
	}

	got, err := store.GetNode(ctx, "upd")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if got.Name != "After" {
		t.Errorf("expected name 'After', got %q", got.Name)
	}
	if got.Address != "2.2.2.2" {
		t.Errorf("expected address '2.2.2.2', got %q", got.Address)
	}
}

func TestUpdateNodeNotFound(t *testing.T) {
	store := newTestStore(t)
	node := &Node{ID: "ghost", Name: "Ghost"}
	err := store.UpdateNode(context.Background(), node)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteNode(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	node := &Node{ID: "del", Name: "ToDelete", NodeType: NodeTypeRemote, Address: "1.1.1.1"}
	if err := store.CreateNode(ctx, node); err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}
	if err := store.DeleteNode(ctx, "del"); err != nil {
		t.Fatalf("DeleteNode failed: %v", err)
	}
	_, err := store.GetNode(ctx, "del")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDeleteNodeNotFound(t *testing.T) {
	store := newTestStore(t)
	err := store.DeleteNode(context.Background(), "nope")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteLocalNodeProtected(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if _, err := store.EnsureLocalNode(ctx); err != nil {
		t.Fatalf("EnsureLocalNode failed: %v", err)
	}
	err := store.DeleteNode(ctx, LocalNodeID)
	if !errors.Is(err, ErrLocalNodeDelete) {
		t.Errorf("expected ErrLocalNodeDelete, got %v", err)
	}
}

func TestEnsureLocalNodeIdempotent(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	node1, err := store.EnsureLocalNode(ctx)
	if err != nil {
		t.Fatalf("EnsureLocalNode (first) failed: %v", err)
	}
	node2, err := store.EnsureLocalNode(ctx)
	if err != nil {
		t.Fatalf("EnsureLocalNode (second) failed: %v", err)
	}
	if node1.ID != node2.ID {
		t.Errorf("expected same ID, got %q and %q", node1.ID, node2.ID)
	}
	if node1.NodeType != NodeTypeLocal {
		t.Errorf("expected node_type 'local', got %q", node1.NodeType)
	}
}

// --- Node roles ---

func TestAssignAndGetRoles(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if _, err := store.EnsureLocalNode(ctx); err != nil {
		t.Fatalf("EnsureLocalNode failed: %v", err)
	}

	if err := store.AssignRole(ctx, LocalNodeID, "sd-core"); err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}
	if err := store.AssignRole(ctx, LocalNodeID, "gnb"); err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}

	roles, err := store.GetNodeRoles(ctx, LocalNodeID)
	if err != nil {
		t.Fatalf("GetNodeRoles failed: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
}

func TestAssignRoleDuplicate(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if _, err := store.EnsureLocalNode(ctx); err != nil {
		t.Fatalf("EnsureLocalNode failed: %v", err)
	}

	if err := store.AssignRole(ctx, LocalNodeID, "sd-core"); err != nil {
		t.Fatalf("AssignRole (first) failed: %v", err)
	}
	// Second assignment should be a no-op
	if err := store.AssignRole(ctx, LocalNodeID, "sd-core"); err != nil {
		t.Fatalf("AssignRole (duplicate) failed: %v", err)
	}

	roles, err := store.GetNodeRoles(ctx, LocalNodeID)
	if err != nil {
		t.Fatalf("GetNodeRoles failed: %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("expected 1 role after duplicate assign, got %d", len(roles))
	}
}

func TestRemoveRole(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if _, err := store.EnsureLocalNode(ctx); err != nil {
		t.Fatalf("EnsureLocalNode failed: %v", err)
	}
	if err := store.AssignRole(ctx, LocalNodeID, "sd-core"); err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}
	if err := store.RemoveRole(ctx, LocalNodeID, "sd-core"); err != nil {
		t.Fatalf("RemoveRole failed: %v", err)
	}

	roles, err := store.GetNodeRoles(ctx, LocalNodeID)
	if err != nil {
		t.Fatalf("GetNodeRoles failed: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("expected 0 roles after remove, got %d", len(roles))
	}
}

func TestRolesCascadeOnNodeDelete(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	node := &Node{ID: "cascade", Name: "Cascade", NodeType: NodeTypeRemote, Address: "1.1.1.1"}
	if err := store.CreateNode(ctx, node); err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}
	if err := store.AssignRole(ctx, "cascade", "sd-core"); err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}
	if err := store.DeleteNode(ctx, "cascade"); err != nil {
		t.Fatalf("DeleteNode failed: %v", err)
	}

	roles, err := store.GetNodeRoles(ctx, "cascade")
	if err != nil {
		t.Fatalf("GetNodeRoles failed: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("expected 0 roles after cascade delete, got %d", len(roles))
	}
}

func TestNodeIncludesRoles(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if _, err := store.EnsureLocalNode(ctx); err != nil {
		t.Fatalf("EnsureLocalNode failed: %v", err)
	}
	if err := store.AssignRole(ctx, LocalNodeID, "gnb"); err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}

	node, err := store.GetNode(ctx, LocalNodeID)
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if len(node.Roles) != 1 || node.Roles[0] != "gnb" {
		t.Errorf("expected roles [gnb], got %v", node.Roles)
	}
}

// --- Operations log ---

func TestLogAndGetOperations(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		entry := &OperationLog{
			Operation: OpCreateNode,
			NodeID:    "node-1",
			Detail:    "test detail",
			Status:    OpStatusSuccess,
		}
		if err := store.LogOperation(ctx, entry); err != nil {
			t.Fatalf("LogOperation[%d] failed: %v", i, err)
		}
	}

	entries, total, err := store.GetOperationsLog(ctx, 10, 0)
	if err != nil {
		t.Fatalf("GetOperationsLog failed: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total=5, got %d", total)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}
}

func TestOperationsLogPagination(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		entry := &OperationLog{
			Operation: OpCreateNode,
			Status:    OpStatusSuccess,
		}
		if err := store.LogOperation(ctx, entry); err != nil {
			t.Fatalf("LogOperation[%d] failed: %v", i, err)
		}
	}

	entries, total, err := store.GetOperationsLog(ctx, 3, 2)
	if err != nil {
		t.Fatalf("GetOperationsLog failed: %v", err)
	}
	if total != 10 {
		t.Errorf("expected total=10, got %d", total)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries with limit=3, got %d", len(entries))
	}
}

func TestOperationsLogByNode(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_ = store.LogOperation(ctx, &OperationLog{Operation: OpCreateNode, NodeID: "a", Status: OpStatusSuccess})
	_ = store.LogOperation(ctx, &OperationLog{Operation: OpCreateNode, NodeID: "b", Status: OpStatusSuccess})
	_ = store.LogOperation(ctx, &OperationLog{Operation: OpUpdateNode, NodeID: "a", Status: OpStatusSuccess})

	entries, total, err := store.GetOperationsLogByNode(ctx, "a", 10, 0)
	if err != nil {
		t.Fatalf("GetOperationsLogByNode failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total=2 for node 'a', got %d", total)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for node 'a', got %d", len(entries))
	}
}

func TestOperationsLogRecordsError(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	entry := &OperationLog{
		Operation: OpDeleteNode,
		NodeID:    "node-1",
		Status:    OpStatusFailure,
		Error:     "cannot delete local node",
	}
	if err := store.LogOperation(ctx, entry); err != nil {
		t.Fatalf("LogOperation failed: %v", err)
	}

	entries, _, err := store.GetOperationsLog(ctx, 10, 0)
	if err != nil {
		t.Fatalf("GetOperationsLog failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Status != OpStatusFailure {
		t.Errorf("expected status 'failure', got %q", entries[0].Status)
	}
	if entries[0].Error != "cannot delete local node" {
		t.Errorf("expected error message, got %q", entries[0].Error)
	}
}

// --- Lifecycle ---

func TestClose(t *testing.T) {
	store, err := NewSQLiteStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// After closing, operations should fail.
	_, err = store.GetState(context.Background(), "key")
	if err == nil {
		t.Error("expected error after Close, got nil")
	}
}
