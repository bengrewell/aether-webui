-- action_history records every action execution.
CREATE TABLE IF NOT EXISTS action_history (
    id          TEXT PRIMARY KEY,
    component   TEXT NOT NULL,
    action      TEXT NOT NULL,
    target      TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'running',
    exit_code   INTEGER NOT NULL DEFAULT -1,
    error       TEXT,
    labels_json TEXT,
    tags_json   TEXT,
    started_at  INTEGER NOT NULL,
    finished_at INTEGER
);

CREATE INDEX IF NOT EXISTS idx_ah_component        ON action_history(component);
CREATE INDEX IF NOT EXISTS idx_ah_component_action ON action_history(component, action);
CREATE INDEX IF NOT EXISTS idx_ah_status           ON action_history(status);
CREATE INDEX IF NOT EXISTS idx_ah_started_at       ON action_history(started_at);

-- component_state tracks the current installed state per component.
CREATE TABLE IF NOT EXISTS component_state (
    component   TEXT PRIMARY KEY,
    status      TEXT NOT NULL,
    last_action TEXT,
    action_id   TEXT,
    updated_at  INTEGER NOT NULL
);
