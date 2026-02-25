package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// ---------------------------------------------------------------------------
// Action history
// ---------------------------------------------------------------------------

func (d *db) InsertAction(ctx context.Context, rec ActionRecord) error {
	if rec.ID == "" || rec.Component == "" || rec.Action == "" || rec.Target == "" {
		return ErrInvalidArgument
	}
	labelsJSON, err := marshalJSONField(rec.Labels)
	if err != nil {
		return err
	}
	tagsJSON, err := marshalJSONField(rec.Tags)
	if err != nil {
		return err
	}

	startedAt := rec.StartedAt.Unix()
	var finishedAt *int64
	if !rec.FinishedAt.IsZero() {
		v := rec.FinishedAt.Unix()
		finishedAt = &v
	}

	_, err = d.conn.ExecContext(ctx, `
		INSERT INTO action_history(id, component, action, target, status, exit_code, error, labels_json, tags_json, started_at, finished_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rec.ID, rec.Component, rec.Action, rec.Target, rec.Status, rec.ExitCode,
		nullString(rec.Error), labelsJSON, tagsJSON, startedAt, finishedAt)
	return err
}

func (d *db) UpdateActionResult(ctx context.Context, id string, result ActionResult) error {
	if id == "" {
		return ErrInvalidArgument
	}
	var finishedAt *int64
	if !result.FinishedAt.IsZero() {
		v := result.FinishedAt.Unix()
		finishedAt = &v
	}
	res, err := d.conn.ExecContext(ctx, `
		UPDATE action_history
		SET status = ?, exit_code = ?, error = ?, finished_at = ?
		WHERE id = ?
	`, result.Status, result.ExitCode, nullString(result.Error), finishedAt, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *db) GetAction(ctx context.Context, id string) (ActionRecord, bool, error) {
	if id == "" {
		return ActionRecord{}, false, ErrInvalidArgument
	}
	var rec ActionRecord
	var errStr sql.NullString
	var labelsJSON, tagsJSON sql.NullString
	var startedAt int64
	var finishedAt sql.NullInt64

	err := d.conn.QueryRowContext(ctx, `
		SELECT id, component, action, target, status, exit_code, error, labels_json, tags_json, started_at, finished_at
		FROM action_history WHERE id = ?
	`, id).Scan(&rec.ID, &rec.Component, &rec.Action, &rec.Target, &rec.Status,
		&rec.ExitCode, &errStr, &labelsJSON, &tagsJSON, &startedAt, &finishedAt)

	if err == sql.ErrNoRows {
		return ActionRecord{}, false, nil
	}
	if err != nil {
		return ActionRecord{}, false, err
	}

	rec.Error = errStr.String
	rec.StartedAt = time.Unix(startedAt, 0)
	if finishedAt.Valid {
		rec.FinishedAt = time.Unix(finishedAt.Int64, 0)
	}
	if labelsJSON.Valid && labelsJSON.String != "" {
		if err := json.Unmarshal([]byte(labelsJSON.String), &rec.Labels); err != nil {
			return ActionRecord{}, false, err
		}
	}
	if tagsJSON.Valid && tagsJSON.String != "" {
		if err := json.Unmarshal([]byte(tagsJSON.String), &rec.Tags); err != nil {
			return ActionRecord{}, false, err
		}
	}
	return rec, true, nil
}

func (d *db) ListActions(ctx context.Context, filter ActionFilter) ([]ActionRecord, error) {
	query := `SELECT id, component, action, target, status, exit_code, error, labels_json, tags_json, started_at, finished_at FROM action_history`
	var conditions []string
	var args []any

	if filter.Component != "" {
		conditions = append(conditions, "component = ?")
		args = append(args, filter.Component)
	}
	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for _, c := range conditions[1:] {
			query += " AND " + c
		}
	}

	query += " ORDER BY started_at DESC"

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
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

	out := make([]ActionRecord, 0, limit)
	for rows.Next() {
		var rec ActionRecord
		var errStr sql.NullString
		var labelsJSON, tagsJSON sql.NullString
		var startedAt int64
		var finishedAt sql.NullInt64

		if err := rows.Scan(&rec.ID, &rec.Component, &rec.Action, &rec.Target, &rec.Status,
			&rec.ExitCode, &errStr, &labelsJSON, &tagsJSON, &startedAt, &finishedAt); err != nil {
			return nil, err
		}

		rec.Error = errStr.String
		rec.StartedAt = time.Unix(startedAt, 0)
		if finishedAt.Valid {
			rec.FinishedAt = time.Unix(finishedAt.Int64, 0)
		}
		if labelsJSON.Valid && labelsJSON.String != "" {
			if err := json.Unmarshal([]byte(labelsJSON.String), &rec.Labels); err != nil {
				return nil, err
			}
		}
		if tagsJSON.Valid && tagsJSON.String != "" {
			if err := json.Unmarshal([]byte(tagsJSON.String), &rec.Tags); err != nil {
				return nil, err
			}
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// Component state
// ---------------------------------------------------------------------------

func (d *db) UpsertComponentState(ctx context.Context, cs ComponentState) error {
	if cs.Component == "" || cs.Status == "" {
		return ErrInvalidArgument
	}
	updatedAt := cs.UpdatedAt.Unix()
	if cs.UpdatedAt.IsZero() {
		updatedAt = d.now().Unix()
	}
	_, err := d.conn.ExecContext(ctx, `
		INSERT INTO component_state(component, status, last_action, action_id, updated_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(component) DO UPDATE SET
			status = excluded.status,
			last_action = excluded.last_action,
			action_id = excluded.action_id,
			updated_at = excluded.updated_at
	`, cs.Component, cs.Status, nullString(cs.LastAction), nullString(cs.ActionID), updatedAt)
	return err
}

func (d *db) GetComponentState(ctx context.Context, component string) (ComponentState, bool, error) {
	if component == "" {
		return ComponentState{}, false, ErrInvalidArgument
	}
	var cs ComponentState
	var lastAction, actionID sql.NullString
	var updatedAt int64

	err := d.conn.QueryRowContext(ctx, `
		SELECT component, status, last_action, action_id, updated_at
		FROM component_state WHERE component = ?
	`, component).Scan(&cs.Component, &cs.Status, &lastAction, &actionID, &updatedAt)

	if err == sql.ErrNoRows {
		return ComponentState{}, false, nil
	}
	if err != nil {
		return ComponentState{}, false, err
	}

	cs.LastAction = lastAction.String
	cs.ActionID = actionID.String
	cs.UpdatedAt = time.Unix(updatedAt, 0)
	return cs, true, nil
}

func (d *db) ListComponentStates(ctx context.Context) ([]ComponentState, error) {
	rows, err := d.conn.QueryContext(ctx, `
		SELECT component, status, last_action, action_id, updated_at
		FROM component_state ORDER BY component
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ComponentState, 0, 16)
	for rows.Next() {
		var cs ComponentState
		var lastAction, actionID sql.NullString
		var updatedAt int64

		if err := rows.Scan(&cs.Component, &cs.Status, &lastAction, &actionID, &updatedAt); err != nil {
			return nil, err
		}
		cs.LastAction = lastAction.String
		cs.ActionID = actionID.String
		cs.UpdatedAt = time.Unix(updatedAt, 0)
		out = append(out, cs)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// marshalJSONField marshals v to a JSON string, returning nil if v is nil or empty.
func marshalJSONField(v any) ([]byte, error) {
	switch val := v.(type) {
	case map[string]string:
		if len(val) == 0 {
			return nil, nil
		}
	case []string:
		if len(val) == 0 {
			return nil, nil
		}
	}
	return json.Marshal(v)
}

// nullString returns a sql.NullString; empty strings are stored as NULL.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
