package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/bengrewell/aether-webui/internal/store"
)

func (s *Store) Save(ctx context.Context, key store.Key, payload []byte, opts ...store.SaveOption) (store.Meta, error) {
	if key.Namespace == "" || key.ID == "" {
		return store.Meta{}, store.ErrInvalidArgument
	}

	cfg := store.SaveOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	now := s.now()
	var expiresAt *time.Time
	if cfg.TTL > 0 {
		t := now.Add(cfg.TTL)
		expiresAt = &t
	}

	// We implement optimistic concurrency via version.
	// Save increments version on update, sets version=1 on insert.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return store.Meta{}, err
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
		return store.Meta{}, err
	}

	if err == sql.ErrNoRows {
		if cfg.ExpectedVersion != 0 {
			return store.Meta{}, store.ErrConflict
		}
		// Insert
		if cfg.CreateOnly == false || cfg.CreateOnly == true {
			// CreateOnly just means "fail if exists" which we already know is false.
		}

		var exp any = nil
		if expiresAt != nil {
			exp = expiresAt.Unix()
		}

		_, err := tx.ExecContext(ctx, `
			INSERT INTO objects(namespace, id, version, payload, created_at, updated_at, expires_at)
			VALUES(?, ?, 1, ?, ?, ?, ?)
		`, key.Namespace, key.ID, payload, now.Unix(), now.Unix(), exp)
		if err != nil {
			return store.Meta{}, err
		}

		if err := tx.Commit(); err != nil {
			return store.Meta{}, err
		}

		return store.Meta{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
			ExpiresAt: expiresAt,
		}, nil
	}

	// Exists
	if cfg.CreateOnly {
		return store.Meta{}, store.ErrConflict
	}
	if cfg.ExpectedVersion != 0 && cfg.ExpectedVersion != curVersion {
		return store.Meta{}, store.ErrConflict
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
		return store.Meta{}, err
	}

	if err := tx.Commit(); err != nil {
		return store.Meta{}, err
	}

	var existingExpires *time.Time
	if exp != nil {
		t := time.Unix(exp.(int64), 0)
		existingExpires = &t
	}

	return store.Meta{
		CreatedAt: time.Unix(createdAtUnix, 0),
		UpdatedAt: now,
		Version:   newVersion,
		ExpiresAt: existingExpires,
	}, nil
}

func (s *Store) Load(ctx context.Context, key store.Key, opts ...store.LoadOption) (store.ItemBytes, bool, error) {
	if key.Namespace == "" || key.ID == "" {
		return store.ItemBytes{}, false, store.ErrInvalidArgument
	}

	cfg := store.LoadOptions{}
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

	err := s.db.QueryRowContext(ctx, `
		SELECT payload, version, created_at, updated_at, expires_at
		FROM objects
		WHERE namespace = ? AND id = ?
	`, key.Namespace, key.ID).Scan(&payload, &version, &createdAtUnix, &updatedAtUnix, &expiresUnix)

	if err == sql.ErrNoRows {
		return store.ItemBytes{}, false, nil
	}
	if err != nil {
		return store.ItemBytes{}, false, err
	}

	var expiresAt *time.Time
	if expiresUnix.Valid {
		t := time.Unix(expiresUnix.Int64, 0)
		expiresAt = &t
		if cfg.RequireFresh && s.now().After(t) {
			return store.ItemBytes{}, false, store.ErrExpired
		}
	}

	return store.ItemBytes{
		Key: key,
		Meta: store.Meta{
			CreatedAt: time.Unix(createdAtUnix, 0),
			UpdatedAt: time.Unix(updatedAtUnix, 0),
			Version:   version,
			ExpiresAt: expiresAt,
		},
		Data: payload,
	}, true, nil
}

func (s *Store) Delete(ctx context.Context, key store.Key) error {
	if key.Namespace == "" || key.ID == "" {
		return store.ErrInvalidArgument
	}
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM objects WHERE namespace = ? AND id = ?
	`, key.Namespace, key.ID)
	return err
}

func (s *Store) List(ctx context.Context, namespace string, opts ...store.ListOption) ([]store.Key, error) {
	if namespace == "" {
		return nil, store.ErrInvalidArgument
	}

	cfg := store.ListOptions{Limit: 1000}
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
		rows, err = s.db.QueryContext(ctx, `
			SELECT id FROM objects
			WHERE namespace = ? AND id LIKE ?
			ORDER BY id
			LIMIT ?
		`, namespace, cfg.Prefix+"%", cfg.Limit)
	} else {
		rows, err = s.db.QueryContext(ctx, `
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

	out := make([]store.Key, 0, 64)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, store.Key{Namespace: namespace, ID: id})
	}
	return out, rows.Err()
}
