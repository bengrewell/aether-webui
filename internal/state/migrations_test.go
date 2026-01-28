package state

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestEnsureMigrationsTable(t *testing.T) {
	db := openTestDB(t)

	if err := ensureMigrationsTable(db); err != nil {
		t.Fatalf("ensureMigrationsTable failed: %v", err)
	}

	// Calling it again should be idempotent
	if err := ensureMigrationsTable(db); err != nil {
		t.Fatalf("ensureMigrationsTable (second call) failed: %v", err)
	}

	// Verify the table exists by querying it
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("failed to query schema_migrations: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows, got %d", count)
	}
}

func TestGetCurrentVersionEmpty(t *testing.T) {
	db := openTestDB(t)

	if err := ensureMigrationsTable(db); err != nil {
		t.Fatalf("ensureMigrationsTable failed: %v", err)
	}

	version, err := getCurrentVersion(db)
	if err != nil {
		t.Fatalf("getCurrentVersion failed: %v", err)
	}
	if version != 0 {
		t.Errorf("expected version 0, got %d", version)
	}
}

func TestApplyMigration(t *testing.T) {
	db := openTestDB(t)

	if err := ensureMigrationsTable(db); err != nil {
		t.Fatalf("ensureMigrationsTable failed: %v", err)
	}

	m := Migration{
		Version:     1,
		Description: "create test_table",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY)")
			return err
		},
	}

	if err := applyMigration(db, m); err != nil {
		t.Fatalf("applyMigration failed: %v", err)
	}

	// Verify version was recorded
	version, err := getCurrentVersion(db)
	if err != nil {
		t.Fatalf("getCurrentVersion failed: %v", err)
	}
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}

	// Verify the table was created
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count); err != nil {
		t.Fatalf("test_table should exist: %v", err)
	}
}

func TestApplyMigrationRollbackOnFailure(t *testing.T) {
	db := openTestDB(t)

	if err := ensureMigrationsTable(db); err != nil {
		t.Fatalf("ensureMigrationsTable failed: %v", err)
	}

	m := Migration{
		Version:     1,
		Description: "should fail",
		Up: func(tx *sql.Tx) error {
			// Create a table, then cause an error
			if _, err := tx.Exec("CREATE TABLE partial_table (id INTEGER PRIMARY KEY)"); err != nil {
				return err
			}
			// This will fail — referencing a nonexistent table
			_, err := tx.Exec("INSERT INTO nonexistent_table VALUES (1)")
			return err
		},
	}

	err := applyMigration(db, m)
	if err == nil {
		t.Fatal("expected applyMigration to fail")
	}

	// Verify the table was NOT created (rolled back)
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='partial_table'").Scan(&name)
	if err != sql.ErrNoRows {
		t.Errorf("expected partial_table to not exist after rollback, got name=%q err=%v", name, err)
	}

	// Verify version was NOT recorded
	version, err := getCurrentVersion(db)
	if err != nil {
		t.Fatalf("getCurrentVersion failed: %v", err)
	}
	if version != 0 {
		t.Errorf("expected version 0 after failed migration, got %d", version)
	}
}

func TestRunMigrationsSkipsApplied(t *testing.T) {
	db := openTestDB(t)

	callCount := 0
	origMigrations := migrations
	migrations = []Migration{
		{
			Version:     1,
			Description: "first",
			Up: func(tx *sql.Tx) error {
				callCount++
				_, err := tx.Exec("CREATE TABLE m1 (id INTEGER PRIMARY KEY)")
				return err
			},
		},
		{
			Version:     2,
			Description: "second",
			Up: func(tx *sql.Tx) error {
				callCount++
				_, err := tx.Exec("CREATE TABLE m2 (id INTEGER PRIMARY KEY)")
				return err
			},
		},
	}
	defer func() { migrations = origMigrations }()

	// First run: both should apply
	if err := runMigrations(db); err != nil {
		t.Fatalf("runMigrations (first) failed: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 migrations applied, got %d", callCount)
	}

	version, err := getCurrentVersion(db)
	if err != nil {
		t.Fatalf("getCurrentVersion failed: %v", err)
	}
	if version != 2 {
		t.Errorf("expected version 2, got %d", version)
	}

	// Second run: neither should apply
	callCount = 0
	if err := runMigrations(db); err != nil {
		t.Fatalf("runMigrations (second) failed: %v", err)
	}
	if callCount != 0 {
		t.Errorf("expected 0 migrations on re-run, got %d", callCount)
	}
}

func TestRunMigrationsPartialApply(t *testing.T) {
	db := openTestDB(t)

	origMigrations := migrations
	migrations = []Migration{
		{
			Version:     1,
			Description: "first",
			Up: func(tx *sql.Tx) error {
				_, err := tx.Exec("CREATE TABLE p1 (id INTEGER PRIMARY KEY)")
				return err
			},
		},
	}
	defer func() { migrations = origMigrations }()

	// Apply first migration
	if err := runMigrations(db); err != nil {
		t.Fatalf("runMigrations failed: %v", err)
	}

	// Add a second migration and run again
	callCount := 0
	migrations = append(migrations, Migration{
		Version:     2,
		Description: "second added later",
		Up: func(tx *sql.Tx) error {
			callCount++
			_, err := tx.Exec("CREATE TABLE p2 (id INTEGER PRIMARY KEY)")
			return err
		},
	})

	if err := runMigrations(db); err != nil {
		t.Fatalf("runMigrations (with new migration) failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected only 1 new migration applied, got %d", callCount)
	}

	version, err := getCurrentVersion(db)
	if err != nil {
		t.Fatalf("getCurrentVersion failed: %v", err)
	}
	if version != 2 {
		t.Errorf("expected version 2, got %d", version)
	}
}

func TestColumnExists(t *testing.T) {
	db := openTestDB(t)

	_, err := db.Exec("CREATE TABLE col_test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback()

	exists, err := columnExists(tx, "col_test", "name")
	if err != nil {
		t.Fatalf("columnExists failed: %v", err)
	}
	if !exists {
		t.Error("expected column 'name' to exist")
	}

	exists, err = columnExists(tx, "col_test", "nonexistent")
	if err != nil {
		t.Fatalf("columnExists failed: %v", err)
	}
	if exists {
		t.Error("expected column 'nonexistent' to not exist")
	}
}

func TestGetSchemaVersionViaStore(t *testing.T) {
	store, err := NewSQLiteStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	version, err := store.GetSchemaVersion()
	if err != nil {
		t.Fatalf("GetSchemaVersion failed: %v", err)
	}
	if version != 0 {
		t.Errorf("expected version 0 for fresh db, got %d", version)
	}
}
