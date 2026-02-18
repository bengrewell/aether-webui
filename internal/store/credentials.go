package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

func (d *db) UpsertCredential(ctx context.Context, cred Credential) error {
	if cred.ID == "" || cred.Provider == "" {
		return ErrInvalidArgument
	}
	if cred.UpdatedAt.IsZero() {
		cred.UpdatedAt = d.now()
	}

	labelsJSON, err := json.Marshal(cred.Labels)
	if err != nil {
		return err
	}

	ct, err := d.crypter.Encrypt(cred.Secret)
	if err != nil {
		return err
	}

	_, err = d.conn.ExecContext(ctx, `
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

func (d *db) GetCredential(ctx context.Context, id string) (Credential, bool, error) {
	if id == "" {
		return Credential{}, false, ErrInvalidArgument
	}

	var provider string
	var labelsJSON sql.NullString
	var ct []byte
	var updatedAtUnix int64

	err := d.conn.QueryRowContext(ctx, `
		SELECT provider, labels_json, secret_ciphertext, updated_at
		FROM credentials
		WHERE id = ?
	`, id).Scan(&provider, &labelsJSON, &ct, &updatedAtUnix)

	if err == sql.ErrNoRows {
		return Credential{}, false, nil
	}
	if err != nil {
		return Credential{}, false, err
	}

	var labels map[string]string
	if labelsJSON.Valid && labelsJSON.String != "" {
		_ = json.Unmarshal([]byte(labelsJSON.String), &labels)
	}

	pt, err := d.crypter.Decrypt(ct)
	if err != nil {
		return Credential{}, false, err
	}

	return Credential{
		ID:        id,
		Provider:  provider,
		Labels:    labels,
		Secret:    pt,
		UpdatedAt: time.Unix(updatedAtUnix, 0),
	}, true, nil
}

func (d *db) DeleteCredential(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidArgument
	}
	_, err := d.conn.ExecContext(ctx, `DELETE FROM credentials WHERE id = ?`, id)
	return err
}

func (d *db) ListCredentials(ctx context.Context) ([]CredentialInfo, error) {
	rows, err := d.conn.QueryContext(ctx, `
		SELECT id, provider, labels_json, updated_at
		FROM credentials
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CredentialInfo, 0, 64)
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

		out = append(out, CredentialInfo{
			ID:        id,
			Provider:  provider,
			Labels:    labels,
			UpdatedAt: time.Unix(updatedAtUnix, 0),
		})
	}
	return out, rows.Err()
}
