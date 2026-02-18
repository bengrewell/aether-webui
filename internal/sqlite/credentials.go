package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/bengrewell/aether-webui/internal/store"
)

func (s *Store) UpsertCredential(ctx context.Context, cred store.Credential) error {
	if cred.ID == "" || cred.Provider == "" {
		return store.ErrInvalidArgument
	}
	if cred.UpdatedAt.IsZero() {
		cred.UpdatedAt = s.now()
	}

	labelsJSON, err := json.Marshal(cred.Labels)
	if err != nil {
		return err
	}

	ct, err := s.crypter.Encrypt(cred.Secret)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO credentials(id, provider, labels_json, secret_ciphertext, updated_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			provider = excluded.provider,
			labels_json = excluded.labels_json,
			secret_ciphertext = excluded.secret_ciphertext,
			updated_at = excluded.updated_at
	`, cred.ID, cred.Provider, string(labelsJSON), ct, cred.UpdatedAt.Unix())

	return err
}

func (s *Store) GetCredential(ctx context.Context, id string) (store.Credential, bool, error) {
	if id == "" {
		return store.Credential{}, false, store.ErrInvalidArgument
	}

	var provider string
	var labelsJSON sql.NullString
	var ct []byte
	var updatedAtUnix int64

	err := s.db.QueryRowContext(ctx, `
		SELECT provider, labels_json, secret_ciphertext, updated_at
		FROM credentials
		WHERE id = ?
	`, id).Scan(&provider, &labelsJSON, &ct, &updatedAtUnix)

	if err == sql.ErrNoRows {
		return store.Credential{}, false, nil
	}
	if err != nil {
		return store.Credential{}, false, err
	}

	var labels map[string]string
	if labelsJSON.Valid && labelsJSON.String != "" {
		_ = json.Unmarshal([]byte(labelsJSON.String), &labels)
	}

	pt, err := s.crypter.Decrypt(ct)
	if err != nil {
		return store.Credential{}, false, err
	}

	return store.Credential{
		ID:        id,
		Provider:  provider,
		Labels:    labels,
		Secret:    pt,
		UpdatedAt: time.Unix(updatedAtUnix, 0),
	}, true, nil
}

func (s *Store) DeleteCredential(ctx context.Context, id string) error {
	if id == "" {
		return store.ErrInvalidArgument
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM credentials WHERE id = ?`, id)
	return err
}

func (s *Store) ListCredentials(ctx context.Context) ([]store.CredentialInfo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, provider, labels_json, updated_at
		FROM credentials
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]store.CredentialInfo, 0, 64)
	for rows.Next() {
		var id, provider string
		var labelsJSON sql.NullString
		var updatedAtUnix int64

		if err := rows.Scan(&id, &provider, &labelsJSON, &updatedAtUnix); err != nil {
			return nil, err
		}

		var labels map[string]string
		if labelsJSON.Valid && labelsJSON.String != "" {
			_ = json.Unmarshal([]byte(labelsJSON.String), &labels)
		}

		out = append(out, store.CredentialInfo{
			ID:        id,
			Provider:  provider,
			Labels:    labels,
			UpdatedAt: time.Unix(updatedAtUnix, 0),
		})
	}
	return out, rows.Err()
}
