package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func applyPragmas(ctx context.Context, db *sql.DB, busyTimeout time.Duration) error {
	stmts := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA foreign_keys=ON;",
		fmt.Sprintf("PRAGMA busy_timeout=%d;", int(busyTimeout.Milliseconds())),
		"PRAGMA temp_store=MEMORY;",
	}

	for _, q := range stmts {
		if _, err := db.ExecContext(ctx, q); err != nil {
			return err
		}
	}
	return nil
}
