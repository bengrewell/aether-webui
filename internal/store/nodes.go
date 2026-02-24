package store

import (
	"context"
	"database/sql"
	"time"
)

func (d *db) UpsertNode(ctx context.Context, node Node) error {
	if node.ID == "" || node.Name == "" || node.AnsibleHost == "" {
		return ErrInvalidArgument
	}

	now := d.now()
	if node.CreatedAt.IsZero() {
		node.CreatedAt = now
	}
	if node.UpdatedAt.IsZero() {
		node.UpdatedAt = now
	}

	passwordCT, err := d.encryptOptional(node.Password)
	if err != nil {
		return err
	}
	sudoPassCT, err := d.encryptOptional(node.SudoPassword)
	if err != nil {
		return err
	}
	sshKeyCT, err := d.encryptOptional(node.SSHKey)
	if err != nil {
		return err
	}

	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO nodes(id, name, ansible_host, ansible_user, password_ct, sudo_pass_ct, ssh_key_ct, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			ansible_host = excluded.ansible_host,
			ansible_user = excluded.ansible_user,
			password_ct = excluded.password_ct,
			sudo_pass_ct = excluded.sudo_pass_ct,
			ssh_key_ct = excluded.ssh_key_ct,
			updated_at = excluded.updated_at
	`, node.ID, node.Name, node.AnsibleHost, node.AnsibleUser,
		passwordCT, sudoPassCT, sshKeyCT,
		node.CreatedAt.Unix(), node.UpdatedAt.Unix())
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM node_roles WHERE node_id = ?`, node.ID); err != nil {
		return err
	}
	for _, role := range node.Roles {
		if _, err := tx.ExecContext(ctx, `INSERT INTO node_roles(node_id, role) VALUES(?, ?)`, node.ID, role); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *db) GetNode(ctx context.Context, id string) (Node, bool, error) {
	if id == "" {
		return Node{}, false, ErrInvalidArgument
	}

	var name, ansibleHost, ansibleUser string
	var passwordCT, sudoPassCT, sshKeyCT []byte
	var createdAtUnix, updatedAtUnix int64

	err := d.conn.QueryRowContext(ctx, `
		SELECT name, ansible_host, ansible_user, password_ct, sudo_pass_ct, ssh_key_ct, created_at, updated_at
		FROM nodes WHERE id = ?
	`, id).Scan(&name, &ansibleHost, &ansibleUser, &passwordCT, &sudoPassCT, &sshKeyCT, &createdAtUnix, &updatedAtUnix)

	if err == sql.ErrNoRows {
		return Node{}, false, nil
	}
	if err != nil {
		return Node{}, false, err
	}

	password, err := d.decryptOptional(passwordCT)
	if err != nil {
		return Node{}, false, err
	}
	sudoPass, err := d.decryptOptional(sudoPassCT)
	if err != nil {
		return Node{}, false, err
	}
	sshKey, err := d.decryptOptional(sshKeyCT)
	if err != nil {
		return Node{}, false, err
	}

	roles, err := d.nodeRoles(ctx, id)
	if err != nil {
		return Node{}, false, err
	}

	return Node{
		ID:           id,
		Name:         name,
		AnsibleHost:  ansibleHost,
		AnsibleUser:  ansibleUser,
		Password:     password,
		SudoPassword: sudoPass,
		SSHKey:       sshKey,
		Roles:        roles,
		CreatedAt:    time.Unix(createdAtUnix, 0),
		UpdatedAt:    time.Unix(updatedAtUnix, 0),
	}, true, nil
}

func (d *db) DeleteNode(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidArgument
	}
	_, err := d.conn.ExecContext(ctx, `DELETE FROM nodes WHERE id = ?`, id)
	return err
}

func (d *db) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	rows, err := d.conn.QueryContext(ctx, `
		SELECT id, name, ansible_host, ansible_user, created_at, updated_at
		FROM nodes ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]NodeInfo, 0, 32)
	for rows.Next() {
		var id, name, ansibleHost, ansibleUser string
		var createdAtUnix, updatedAtUnix int64

		if err := rows.Scan(&id, &name, &ansibleHost, &ansibleUser, &createdAtUnix, &updatedAtUnix); err != nil {
			return nil, err
		}

		roles, err := d.nodeRoles(ctx, id)
		if err != nil {
			return nil, err
		}

		out = append(out, NodeInfo{
			ID:          id,
			Name:        name,
			AnsibleHost: ansibleHost,
			AnsibleUser: ansibleUser,
			Roles:       roles,
			CreatedAt:   time.Unix(createdAtUnix, 0),
			UpdatedAt:   time.Unix(updatedAtUnix, 0),
		})
	}
	return out, rows.Err()
}

// nodeRoles queries the node_roles table for the given node ID.
func (d *db) nodeRoles(ctx context.Context, nodeID string) ([]string, error) {
	rows, err := d.conn.QueryContext(ctx, `SELECT role FROM node_roles WHERE node_id = ? ORDER BY role`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

// encryptOptional encrypts non-nil/non-empty plaintext via the store's Crypter.
func (d *db) encryptOptional(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, nil
	}
	return d.crypter.Encrypt(plaintext)
}

// decryptOptional decrypts non-nil/non-empty ciphertext via the store's Crypter.
func (d *db) decryptOptional(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, nil
	}
	return d.crypter.Decrypt(ciphertext)
}
