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
