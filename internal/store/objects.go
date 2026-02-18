package store

import (
	"context"
	"database/sql"
	"time"
)

func (d *db) Save(ctx context.Context, key Key, payload []byte, opts ...SaveOption) (Meta, error) {
	if key.Namespace == "" || key.ID == "" {
		return Meta{}, ErrInvalidArgument
	}

	cfg := SaveOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	now := d.now()
	var expiresAt *time.Time
	if cfg.TTL > 0 {
		t := now.Add(cfg.TTL)
		expiresAt = &t
	}

	// We implement optimistic concurrency via version.
	// Save increments version on update, sets version=1 on insert.
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return Meta{}, err
	}
	defer func() { _ = tx.Rollback() }()

	// Check existing row
	var curVersion int64
	var createdAtUnix int64
	var updatedAtUnix int64
	var expiresUnix sql.NullInt64

	err = tx.QueryRowContext(ctx, `
		SELECT version, created_at, updated_at, expires_at
		FROM objects
		WHERE namespace = ? AND id = ?
	`, key.Namespace, key.ID).Scan(&curVersion, &createdAtUnix, &updatedAtUnix, &expiresUnix)

	if err != nil && err != sql.ErrNoRows {
		return Meta{}, err
	}

	if err == sql.ErrNoRows {
		if cfg.ExpectedVersion != 0 {
			return Meta{}, ErrConflict
		}
		// Insert

		var exp any = nil
		if expiresAt != nil {
			exp = expiresAt.Unix()
		}

		_, err := tx.ExecContext(ctx, `
			INSERT INTO objects(namespace, id, version, payload, created_at, updated_at, expires_at)
			VALUES(?, ?, 1, ?, ?, ?, ?)
		`, key.Namespace, key.ID, payload, now.Unix(), now.Unix(), exp)
		if err != nil {
			return Meta{}, err
		}

		if err := tx.Commit(); err != nil {
			return Meta{}, err
		}

		return Meta{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
			ExpiresAt: expiresAt,
		}, nil
	}

	// Exists
	if cfg.CreateOnly {
		return Meta{}, ErrConflict
	}
	if cfg.ExpectedVersion != 0 && cfg.ExpectedVersion != curVersion {
		return Meta{}, ErrConflict
	}

	newVersion := curVersion + 1

	var exp any = nil
	if expiresAt != nil {
		exp = expiresAt.Unix()
	} else if expiresUnix.Valid {
		// preserve existing expires_at if caller didn't provide TTL
		exp = expiresUnix.Int64
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE objects
		SET version = ?, payload = ?, updated_at = ?, expires_at = ?
		WHERE namespace = ? AND id = ?
	`, newVersion, payload, now.Unix(), exp, key.Namespace, key.ID)
	if err != nil {
		return Meta{}, err
	}

	if err := tx.Commit(); err != nil {
		return Meta{}, err
	}

	var existingExpires *time.Time
	if exp != nil {
		t := time.Unix(exp.(int64), 0)
		existingExpires = &t
	}

	return Meta{
		CreatedAt: time.Unix(createdAtUnix, 0),
		UpdatedAt: now,
		Version:   newVersion,
		ExpiresAt: existingExpires,
	}, nil
}

func (d *db) Load(ctx context.Context, key Key, opts ...LoadOption) (ItemBytes, bool, error) {
	if key.Namespace == "" || key.ID == "" {
		return ItemBytes{}, false, ErrInvalidArgument
	}

	cfg := LoadOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	var payload []byte
	var version int64
	var createdAtUnix int64
	var updatedAtUnix int64
	var expiresUnix sql.NullInt64

	err := d.conn.QueryRowContext(ctx, `
		SELECT payload, version, created_at, updated_at, expires_at
		FROM objects
		WHERE namespace = ? AND id = ?
	`, key.Namespace, key.ID).Scan(&payload, &version, &createdAtUnix, &updatedAtUnix, &expiresUnix)

	if err == sql.ErrNoRows {
		return ItemBytes{}, false, nil
	}
	if err != nil {
		return ItemBytes{}, false, err
	}

	var expiresAt *time.Time
	if expiresUnix.Valid {
		t := time.Unix(expiresUnix.Int64, 0)
		expiresAt = &t
		if cfg.RequireFresh && d.now().After(t) {
			return ItemBytes{}, false, ErrExpired
		}
	}

	return ItemBytes{
		Key: key,
		Meta: Meta{
			CreatedAt: time.Unix(createdAtUnix, 0),
			UpdatedAt: time.Unix(updatedAtUnix, 0),
			Version:   version,
			ExpiresAt: expiresAt,
		},
		Data: payload,
	}, true, nil
}

func (d *db) Delete(ctx context.Context, key Key) error {
	if key.Namespace == "" || key.ID == "" {
		return ErrInvalidArgument
	}
	_, err := d.conn.ExecContext(ctx, `
		DELETE FROM objects WHERE namespace = ? AND id = ?
	`, key.Namespace, key.ID)
	return err
}

func (d *db) List(ctx context.Context, namespace string, opts ...ListOption) ([]Key, error) {
	if namespace == "" {
		return nil, ErrInvalidArgument
	}

	cfg := ListOptions{Limit: 1000}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 1000
	}

	var rows *sql.Rows
	var err error

	if cfg.Prefix != "" {
		rows, err = d.conn.QueryContext(ctx, `
			SELECT id FROM objects
			WHERE namespace = ? AND id LIKE ?
			ORDER BY id
			LIMIT ?
		`, namespace, cfg.Prefix+"%", cfg.Limit)
	} else {
		rows, err = d.conn.QueryContext(ctx, `
			SELECT id FROM objects
			WHERE namespace = ?
			ORDER BY id
			LIMIT ?
		`, namespace, cfg.Limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Key, 0, 64)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, Key{Namespace: namespace, ID: id})
	}
	return out, rows.Err()
}
