CREATE TABLE IF NOT EXISTS deployments (
    id          TEXT PRIMARY KEY,
    status      TEXT NOT NULL DEFAULT 'pending',
    created_at  INTEGER NOT NULL,
    started_at  INTEGER,
    finished_at INTEGER,
    error       TEXT
);

CREATE TABLE IF NOT EXISTS deployment_actions (
    deployment_id TEXT NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    seq           INTEGER NOT NULL,
    action_id     TEXT NOT NULL,
    component     TEXT NOT NULL,
    action        TEXT NOT NULL,
    PRIMARY KEY (deployment_id, seq)
);

CREATE INDEX IF NOT EXISTS idx_da_deployment ON deployment_actions(deployment_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
