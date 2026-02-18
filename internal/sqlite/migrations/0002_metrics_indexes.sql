CREATE INDEX IF NOT EXISTS idx_metrics_samples_metric_ts
    ON metrics_samples(metric, ts);

CREATE INDEX IF NOT EXISTS idx_metrics_samples_metric_labels_ts
    ON metrics_samples(metric, labels_hash, ts);
