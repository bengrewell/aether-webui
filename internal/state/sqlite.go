package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// ErrNotFound is returned when a requested key or record does not exist.
var ErrNotFound = errors.New("not found")

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite-backed store.
// It creates the data directory and database file if they don't exist,
// and initializes the schema.
func NewSQLiteStore(dataDir string) (*SQLiteStore, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "state.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent read/write performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	store := &SQLiteStore{db: db}

	// Initialize schema
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Run any pending migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

// initSchema creates the database tables if they don't exist.
func (s *SQLiteStore) initSchema() error {
	schema := `
		-- Application key-value state
		CREATE TABLE IF NOT EXISTS app_state (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Cached system information
		CREATE TABLE IF NOT EXISTS system_info_cache (
			info_type TEXT PRIMARY KEY,
			data TEXT NOT NULL,
			collected_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Metrics history (time-series)
		CREATE TABLE IF NOT EXISTS metrics_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			metric_type TEXT NOT NULL,
			data TEXT NOT NULL,
			recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Create index for efficient metrics queries
		CREATE INDEX IF NOT EXISTS idx_metrics_type_time
			ON metrics_history(metric_type, recorded_at);
	`

	_, err := s.db.Exec(schema)
	return err
}

// GetState retrieves a value from the app_state table.
func (s *SQLiteStore) GetState(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx,
		"SELECT value FROM app_state WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("failed to get state: %w", err)
	}
	return value, nil
}

// SetState stores a value in the app_state table.
func (s *SQLiteStore) SetState(ctx context.Context, key string, value string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO app_state (key, value, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = CURRENT_TIMESTAMP
	`, key, value)
	if err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}
	return nil
}

// DeleteState removes a key from the app_state table.
func (s *SQLiteStore) DeleteState(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM app_state WHERE key = ?", key)
	if err != nil {
		return fmt.Errorf("failed to delete state: %w", err)
	}
	return nil
}

// GetWizardStatus retrieves the current wizard completion status.
func (s *SQLiteStore) GetWizardStatus(ctx context.Context) (*WizardStatus, error) {
	status := &WizardStatus{Completed: false}

	// Check if wizard is completed
	completed, err := s.GetState(ctx, KeyWizardCompleted)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if completed == "true" {
		status.Completed = true
	}

	// Get completion time if available
	if status.Completed {
		completedAt, err := s.GetState(ctx, KeyWizardCompletedAt)
		if err == nil {
			t, err := time.Parse(time.RFC3339, completedAt)
			if err == nil {
				status.CompletedAt = &t
			}
		}

		// Get steps if available
		stepsJSON, err := s.GetState(ctx, KeyWizardSteps)
		if err == nil {
			var steps []string
			if json.Unmarshal([]byte(stepsJSON), &steps) == nil {
				status.Steps = steps
			}
		}
	}

	return status, nil
}

// SetWizardComplete marks the wizard as complete with optional step list.
func (s *SQLiteStore) SetWizardComplete(ctx context.Context, steps []string) error {
	now := time.Now().UTC()

	if err := s.SetState(ctx, KeyWizardCompleted, "true"); err != nil {
		return err
	}

	if err := s.SetState(ctx, KeyWizardCompletedAt, now.Format(time.RFC3339)); err != nil {
		return err
	}

	if len(steps) > 0 {
		stepsJSON, err := json.Marshal(steps)
		if err != nil {
			return fmt.Errorf("failed to marshal steps: %w", err)
		}
		if err := s.SetState(ctx, KeyWizardSteps, string(stepsJSON)); err != nil {
			return err
		}
	}

	return nil
}

// ClearWizardStatus resets the wizard status by removing all wizard-related state.
func (s *SQLiteStore) ClearWizardStatus(ctx context.Context) error {
	for _, key := range []string{KeyWizardCompleted, KeyWizardCompletedAt, KeyWizardSteps} {
		if err := s.DeleteState(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// GetCachedSystemInfo retrieves cached system information.
func (s *SQLiteStore) GetCachedSystemInfo(ctx context.Context, infoType string) ([]byte, time.Time, error) {
	var data string
	var collectedAt time.Time

	err := s.db.QueryRowContext(ctx,
		"SELECT data, collected_at FROM system_info_cache WHERE info_type = ?",
		infoType).Scan(&data, &collectedAt)
	if err == sql.ErrNoRows {
		return nil, time.Time{}, ErrNotFound
	}
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to get cached info: %w", err)
	}

	return []byte(data), collectedAt, nil
}

// SetCachedSystemInfo stores system information in the cache.
func (s *SQLiteStore) SetCachedSystemInfo(ctx context.Context, infoType string, data []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO system_info_cache (info_type, data, collected_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(info_type) DO UPDATE SET
			data = excluded.data,
			collected_at = CURRENT_TIMESTAMP
	`, infoType, string(data))
	if err != nil {
		return fmt.Errorf("failed to set cached info: %w", err)
	}
	return nil
}

// RecordMetrics appends a metrics snapshot to the history.
func (s *SQLiteStore) RecordMetrics(ctx context.Context, metricType string, data []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO metrics_history (metric_type, data, recorded_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`, metricType, string(data))
	if err != nil {
		return fmt.Errorf("failed to record metrics: %w", err)
	}
	return nil
}

// GetMetricsHistory retrieves recent metrics snapshots of a given type.
func (s *SQLiteStore) GetMetricsHistory(ctx context.Context, metricType string, limit int) ([]MetricsSnapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT metric_type, data, recorded_at
		FROM metrics_history
		WHERE metric_type = ?
		ORDER BY recorded_at DESC
		LIMIT ?
	`, metricType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics history: %w", err)
	}
	defer rows.Close()

	var snapshots []MetricsSnapshot
	for rows.Next() {
		var snap MetricsSnapshot
		var data string
		if err := rows.Scan(&snap.MetricType, &data, &snap.RecordedAt); err != nil {
			return nil, fmt.Errorf("failed to scan metrics row: %w", err)
		}
		snap.Data = []byte(data)
		snapshots = append(snapshots, snap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metrics rows: %w", err)
	}

	return snapshots, nil
}

// GetMetricsRange retrieves metrics snapshots within a time range.
func (s *SQLiteStore) GetMetricsRange(ctx context.Context, metricType string, start, end time.Time) ([]MetricsSnapshot, error) {
	// Format times to match SQLite's CURRENT_TIMESTAMP format (UTC)
	startStr := start.UTC().Format("2006-01-02 15:04:05")
	endStr := end.UTC().Format("2006-01-02 15:04:05")

	rows, err := s.db.QueryContext(ctx, `
		SELECT metric_type, data, recorded_at
		FROM metrics_history
		WHERE metric_type = ? AND recorded_at >= ? AND recorded_at <= ?
		ORDER BY recorded_at ASC
	`, metricType, startStr, endStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics range: %w", err)
	}
	defer rows.Close()

	var snapshots []MetricsSnapshot
	for rows.Next() {
		var snap MetricsSnapshot
		var data string
		if err := rows.Scan(&snap.MetricType, &data, &snap.RecordedAt); err != nil {
			return nil, fmt.Errorf("failed to scan metrics row: %w", err)
		}
		snap.Data = []byte(data)
		snapshots = append(snapshots, snap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metrics rows: %w", err)
	}

	return snapshots, nil
}

// PruneOldMetrics removes metrics older than the specified duration.
func (s *SQLiteStore) PruneOldMetrics(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().UTC().Add(-olderThan).Format("2006-01-02 15:04:05")
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM metrics_history WHERE recorded_at < ?", cutoff)
	if err != nil {
		return fmt.Errorf("failed to prune old metrics: %w", err)
	}
	return nil
}

// GetSchemaVersion returns the current schema migration version.
func (s *SQLiteStore) GetSchemaVersion() (int, error) {
	return getCurrentVersion(s.db)
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
