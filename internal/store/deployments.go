package store

import (
	"context"
	"database/sql"
	"time"
)

func (d *db) InsertDeployment(ctx context.Context, dep Deployment) error {
	if dep.ID == "" {
		return ErrInvalidArgument
	}
	if dep.Status == "" {
		dep.Status = "pending"
	}
	if dep.CreatedAt.IsZero() {
		dep.CreatedAt = d.now()
	}

	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	createdAt := dep.CreatedAt.Unix()
	var startedAt, finishedAt *int64
	if !dep.StartedAt.IsZero() {
		v := dep.StartedAt.Unix()
		startedAt = &v
	}
	if !dep.FinishedAt.IsZero() {
		v := dep.FinishedAt.Unix()
		finishedAt = &v
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO deployments(id, status, created_at, started_at, finished_at, error)
		VALUES(?, ?, ?, ?, ?, ?)
	`, dep.ID, dep.Status, createdAt, startedAt, finishedAt, nullString(dep.Error)); err != nil {
		return err
	}

	for _, a := range dep.Actions {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO deployment_actions(deployment_id, seq, action_id, component, action)
			VALUES(?, ?, ?, ?, ?)
		`, dep.ID, a.Seq, a.ActionID, a.Component, a.Action); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *db) UpdateDeploymentStatus(ctx context.Context, id, status, errMsg string, finishedAt time.Time) error {
	if id == "" || status == "" {
		return ErrInvalidArgument
	}
	var fin *int64
	if !finishedAt.IsZero() {
		v := finishedAt.Unix()
		fin = &v
	}
	res, err := d.conn.ExecContext(ctx, `
		UPDATE deployments SET status = ?, error = ?, finished_at = ?
		WHERE id = ?
	`, status, nullString(errMsg), fin, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *db) GetDeployment(ctx context.Context, id string) (Deployment, bool, error) {
	if id == "" {
		return Deployment{}, false, ErrInvalidArgument
	}

	var dep Deployment
	var errStr sql.NullString
	var createdAt int64
	var startedAt, finishedAt sql.NullInt64

	err := d.conn.QueryRowContext(ctx, `
		SELECT id, status, created_at, started_at, finished_at, error
		FROM deployments WHERE id = ?
	`, id).Scan(&dep.ID, &dep.Status, &createdAt, &startedAt, &finishedAt, &errStr)
	if err == sql.ErrNoRows {
		return Deployment{}, false, nil
	}
	if err != nil {
		return Deployment{}, false, err
	}

	dep.CreatedAt = time.Unix(createdAt, 0)
	if startedAt.Valid {
		dep.StartedAt = time.Unix(startedAt.Int64, 0)
	}
	if finishedAt.Valid {
		dep.FinishedAt = time.Unix(finishedAt.Int64, 0)
	}
	dep.Error = errStr.String

	actions, err := d.loadDeploymentActions(ctx, id)
	if err != nil {
		return Deployment{}, false, err
	}
	dep.Actions = actions

	return dep, true, nil
}

func (d *db) ListDeployments(ctx context.Context, filter DeploymentFilter) ([]Deployment, error) {
	query := `SELECT id, status, created_at, started_at, finished_at, error FROM deployments`
	var args []any

	if filter.Status != "" {
		query += " WHERE status = ?"
		args = append(args, filter.Status)
	}

	query += " ORDER BY created_at DESC"

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	query += " LIMIT ?"
	args = append(args, limit)

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := d.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Deployment, 0, limit)
	for rows.Next() {
		var dep Deployment
		var errStr sql.NullString
		var createdAt int64
		var startedAt, finishedAt sql.NullInt64

		if err := rows.Scan(&dep.ID, &dep.Status, &createdAt, &startedAt, &finishedAt, &errStr); err != nil {
			return nil, err
		}

		dep.CreatedAt = time.Unix(createdAt, 0)
		if startedAt.Valid {
			dep.StartedAt = time.Unix(startedAt.Int64, 0)
		}
		if finishedAt.Valid {
			dep.FinishedAt = time.Unix(finishedAt.Int64, 0)
		}
		dep.Error = errStr.String
		out = append(out, dep)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load actions for each deployment.
	for i := range out {
		actions, err := d.loadDeploymentActions(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Actions = actions
	}

	return out, nil
}

func (d *db) loadDeploymentActions(ctx context.Context, deploymentID string) ([]DeploymentAction, error) {
	rows, err := d.conn.QueryContext(ctx, `
		SELECT deployment_id, seq, action_id, component, action
		FROM deployment_actions WHERE deployment_id = ? ORDER BY seq
	`, deploymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []DeploymentAction
	for rows.Next() {
		var a DeploymentAction
		if err := rows.Scan(&a.DeploymentID, &a.Seq, &a.ActionID, &a.Component, &a.Action); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	return actions, rows.Err()
}
