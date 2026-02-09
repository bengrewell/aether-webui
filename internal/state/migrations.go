package state

import (
	"database/sql"
	"fmt"
)

// Migration represents a single schema migration.
type Migration struct {
	Version     int
	Description string
	Up          func(tx *sql.Tx) error
}

// migrations is the ordered list of all schema migrations.
// IMPORTANT: Always append new migrations at the end. Never reorder or remove.
var migrations = []Migration{
	{
		Version:     1,
		Description: "add nodes, node_roles, and operations_log tables",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`
				CREATE TABLE IF NOT EXISTS nodes (
					id TEXT PRIMARY KEY,
					name TEXT NOT NULL,
					node_type TEXT NOT NULL DEFAULT 'remote',
					address TEXT,
					ssh_port INTEGER DEFAULT 22,
					username TEXT,
					auth_method TEXT,
					encrypted_password TEXT,
					private_key_path TEXT,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);

				CREATE TABLE IF NOT EXISTS node_roles (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					node_id TEXT NOT NULL,
					role TEXT NOT NULL,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					UNIQUE(node_id, role),
					FOREIGN KEY (node_id) REFERENCES nodes(id) ON DELETE CASCADE
				);

				CREATE TABLE IF NOT EXISTS operations_log (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					operation TEXT NOT NULL,
					node_id TEXT,
					detail TEXT,
					status TEXT NOT NULL,
					error TEXT,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);

				CREATE INDEX IF NOT EXISTS idx_operations_log_created_at
					ON operations_log(created_at);
				CREATE INDEX IF NOT EXISTS idx_operations_log_node_id
					ON operations_log(node_id);
			`)
			return err
		},
	},
}

// runMigrations applies all pending migrations in order.
func runMigrations(db *sql.DB) error {
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	for _, m := range migrations {
		if m.Version <= currentVersion {
			continue
		}
		if err := applyMigration(db, m); err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w",
				m.Version, m.Description, err)
		}
	}
	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			description TEXT NOT NULL
		)
	`)
	return err
}

func getCurrentVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version)
	return version, err
}

func applyMigration(db *sql.DB, m Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := m.Up(tx); err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO schema_migrations (version, description) VALUES (?, ?)`,
		m.Version, m.Description)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// columnExists checks if a column exists in a table (useful for idempotent migrations).
func columnExists(tx *sql.Tx, table, column string) (bool, error) {
	rows, err := tx.Query("SELECT 1 FROM pragma_table_info(?) WHERE name = ?", table, column)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}
