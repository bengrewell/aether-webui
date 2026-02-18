package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bengrewell/aether-webui/internal/store"
)

func (s *Store) AppendSample(ctx context.Context, sample store.Sample) error {
	return s.AppendSamples(ctx, []store.Sample{sample})
}

func (s *Store) AppendSamples(ctx context.Context, samples []store.Sample) error {
	if len(samples) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics_samples(metric, ts, value, unit, labels_norm, labels_hash, labels_json)
		VALUES(?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, sm := range samples {
		if sm.Metric == "" || sm.TS.IsZero() {
			return store.ErrInvalidArgument
		}
		labelsNorm, labelsHash := store.NormalizeLabels(sm.Labels)
		labelsJSON, _ := json.Marshal(sm.Labels)

		if _, err := stmt.ExecContext(ctx,
			sm.Metric,
			sm.TS.Unix(),
			sm.Value,
			sm.Unit,
			labelsNorm,
			labelsHash,
			string(labelsJSON),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func bucketSeconds(a store.Agg) (int64, error) {
	switch a {
	case store.AggRaw:
		return 0, nil
	case store.Agg10s:
		return 10, nil
	case store.Agg1m:
		return 60, nil
	case store.Agg5m:
		return 300, nil
	case store.Agg1h:
		return 3600, nil
	default:
		return 0, fmt.Errorf("unknown aggregation")
	}
}

func (s *Store) QueryRange(ctx context.Context, q store.RangeQuery) ([]store.Series, error) {
	if q.Metric == "" {
		return nil, store.ErrInvalidArgument
	}
	if q.Range.To.Before(q.Range.From) {
		return nil, store.ErrInvalidArgument
	}

	labelsNorm, labelsHash := store.NormalizeLabels(q.LabelsExact)

	// Current minimal behavior:
	// - If LabelsExact is provided, match exact labels_hash+labels_norm.
	// - GroupBy ignored.
	// - One series returned.
	sec, err := bucketSeconds(q.Agg)
	if err != nil {
		return nil, err
	}

	limit := q.MaxPoints
	if limit <= 0 {
		limit = 5000
	}

	from := q.Range.From.Unix()
	to := q.Range.To.Unix()

	var rows *sql.Rows

	if q.Agg == store.AggRaw {
		if len(q.LabelsExact) > 0 {
			rows, err = s.db.QueryContext(ctx, `
				SELECT ts, value
				FROM metrics_samples
				WHERE metric = ? AND ts >= ? AND ts <= ?
				  AND labels_hash = ? AND labels_norm = ?
				ORDER BY ts
				LIMIT ?
			`, q.Metric, from, to, labelsHash, labelsNorm, limit)
		} else {
			rows, err = s.db.QueryContext(ctx, `
				SELECT ts, value
				FROM metrics_samples
				WHERE metric = ? AND ts >= ? AND ts <= ?
				ORDER BY ts
				LIMIT ?
			`, q.Metric, from, to, limit)
		}
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		points := make([]store.Point, 0, 256)
		for rows.Next() {
			var ts int64
			var v float64
			if err := rows.Scan(&ts, &v); err != nil {
				return nil, err
			}
			points = append(points, store.Point{TS: time.Unix(ts, 0), Value: v})
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}

		return []store.Series{
			{Metric: q.Metric, Labels: q.LabelsExact, Points: points},
		}, nil
	}

	// Aggregated by time bucket: avg(value) per bucket
	// bucket_ts = (ts / sec) * sec
	if len(q.LabelsExact) > 0 {
		rows, err = s.db.QueryContext(ctx, `
			SELECT (ts / ?) * ? AS bucket_ts, AVG(value) AS avg_value
			FROM metrics_samples
			WHERE metric = ? AND ts >= ? AND ts <= ?
			  AND labels_hash = ? AND labels_norm = ?
			GROUP BY bucket_ts
			ORDER BY bucket_ts
			LIMIT ?
		`, sec, sec, q.Metric, from, to, labelsHash, labelsNorm, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT (ts / ?) * ? AS bucket_ts, AVG(value) AS avg_value
			FROM metrics_samples
			WHERE metric = ? AND ts >= ? AND ts <= ?
			GROUP BY bucket_ts
			ORDER BY bucket_ts
			LIMIT ?
		`, sec, sec, q.Metric, from, to, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]store.Point, 0, 256)
	for rows.Next() {
		var ts int64
		var v float64
		if err := rows.Scan(&ts, &v); err != nil {
			return nil, err
		}
		points = append(points, store.Point{TS: time.Unix(ts, 0), Value: v})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return []store.Series{
		{Metric: q.Metric, Labels: q.LabelsExact, Points: points},
	}, nil
}
