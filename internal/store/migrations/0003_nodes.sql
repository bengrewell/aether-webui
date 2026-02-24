CREATE TABLE IF NOT EXISTS nodes (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL UNIQUE,
    ansible_host TEXT NOT NULL,
    ansible_user TEXT NOT NULL DEFAULT '',
    password_ct  BLOB,
    sudo_pass_ct BLOB,
    ssh_key_ct   BLOB,
    created_at   INTEGER NOT NULL,
    updated_at   INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS node_roles (
    node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    role    TEXT NOT NULL,
    UNIQUE(node_id, role)
);

CREATE INDEX IF NOT EXISTS idx_node_roles_node_id ON node_roles(node_id);
CREATE INDEX IF NOT EXISTS idx_node_roles_role ON node_roles(role);
