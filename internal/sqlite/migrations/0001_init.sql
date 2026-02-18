CREATE TABLE IF NOT EXISTS objects (
                                       namespace   TEXT NOT NULL,
                                       id          TEXT NOT NULL,
                                       version     INTEGER NOT NULL,
                                       payload     BLOB NOT NULL,
                                       created_at  INTEGER NOT NULL,
                                       updated_at  INTEGER NOT NULL,
                                       expires_at  INTEGER,

                                       PRIMARY KEY(namespace, id)
    );

CREATE INDEX IF NOT EXISTS idx_objects_namespace_id
    ON objects(namespace, id);

CREATE TABLE IF NOT EXISTS credentials (
                                           id                TEXT PRIMARY KEY,
                                           provider          TEXT NOT NULL,
                                           labels_json       TEXT,
                                           secret_ciphertext BLOB NOT NULL,
                                           updated_at        INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS metrics_samples (
                                               metric      TEXT NOT NULL,
                                               ts          INTEGER NOT NULL,
                                               value       REAL NOT NULL,
                                               unit        TEXT,
                                               labels_norm TEXT NOT NULL,
                                               labels_hash TEXT NOT NULL,
                                               labels_json TEXT,

    -- No primary key: allow duplicate samples; caller decides semantics.
    -- Add uniqueness later if you want (metric+ts+labels_hash).
    -- PRIMARY KEY(metric, ts, labels_hash)
                                               CHECK(metric <> '')
    );
