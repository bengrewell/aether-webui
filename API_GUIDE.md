# Aether WebUI — API Guide for Frontend Development

This document is the definitive reference for building a frontend against the **aether-webd** backend API. It is written for AI coding tools (Bolt, Claude, Copilot) and human developers alike. Use it alongside the machine-readable OpenAPI 3.1 spec at [`api/openapi.json`](api/openapi.json) (also served at runtime at `/openapi.json`).

---

## Overview

**aether-webd** is the backend service for the Aether WebUI. It manages [Aether OnRamp](https://github.com/opennetworkinglab/aether-onramp) 5G network deployments, including:

- SD-Core (4G/5G core network)
- gNBs (srsRAN, OAI, UERANSIM, gNBSim)
- Kubernetes (RKE2) cluster lifecycle
- Host system monitoring and metrics
- Ansible inventory and node management

The backend is written in Go using [Huma](https://huma.rocks/) (on Chi) and exposes a REST/JSON API.

---

## Quick Reference

| Item | Value |
|------|-------|
| **Base URL** | `http://<host>:8186/api/v1/` (default) or `https://<host>:8443/api/v1/` with TLS |
| **Auth** | `Authorization: Bearer <token>` (when `--api-token` or `AETHER_API_TOKEN` is set) |
| **Content-Type** | `application/json` for all request and response bodies |
| **Error format** | [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457) Problem Details (`{ "title", "status", "detail" }`) |
| **OpenAPI spec** | `./api/openapi.json` (static, committed) or `/openapi.json` (runtime) |
| **Interactive docs** | `/docs` (served at runtime) |
| **Health check** | `GET /healthz` (always public, returns `"healthy"`) |

### Public paths (no auth required)

`/healthz`, `/openapi.json`, `/docs`, and frontend static files.

---

## Providers Overview

The API is organized into **providers** — modular units that each register a set of endpoints:

| Provider | Path Prefix | Purpose | Endpoints |
|----------|-------------|---------|-----------|
| **meta** | `/api/v1/meta/` | Server introspection — version, build, runtime, config, providers, store health | 6 |
| **system** | `/api/v1/system/` | Host system info — CPU, memory, disk, OS, network, metrics | 8 |
| **nodes** | `/api/v1/nodes` | Cluster node CRUD — manage nodes with role assignments and credentials | 5 |
| **onramp** | `/api/v1/onramp/` | Aether OnRamp operations — components, tasks, config, profiles, inventory, deployment tracking | 18 |

**Total: 37 endpoints**

---

## Endpoint Reference

### Meta Provider

All meta endpoints are `GET` requests with no parameters or request body.

#### GET `/api/v1/meta/version`

Get build and version information.

**Response:**
```json
{
  "version": "1.2.0",
  "build_date": "2026-02-18T12:00:00Z",
  "branch": "main",
  "commit_hash": "a1b2c3d"
}
```

**JavaScript:**
```js
const res = await fetch(`${BASE_URL}/meta/version`, { headers });
const version = await res.json();
```

---

#### GET `/api/v1/meta/build`

Get Go toolchain and target platform info.

**Response:**
```json
{
  "go_version": "go1.25.0",
  "os": "linux",
  "arch": "amd64"
}
```

---

#### GET `/api/v1/meta/runtime`

Get process runtime details.

**Response:**
```json
{
  "pid": 4821,
  "user": { "uid": "1000", "name": "aether" },
  "group": { "gid": "1000", "name": "aether" },
  "binary_path": "/usr/local/bin/aether-webd",
  "start_time": "2026-02-18T12:00:00Z",
  "uptime": "2h30m0s"
}
```

---

#### GET `/api/v1/meta/config`

Get active (non-secret) application configuration.

**Response:**
```json
{
  "listen_address": "127.0.0.1:8186",
  "debug_enabled": false,
  "security": {
    "tls_enabled": false,
    "tls_auto_generated": false,
    "mtls_enabled": false,
    "token_auth_enabled": true,
    "rbac_enabled": false
  },
  "frontend": {
    "enabled": true,
    "source": "embedded",
    "dir": ""
  },
  "storage": {
    "data_dir": "/var/lib/aether-webd"
  },
  "metrics": {
    "interval": "10s",
    "retention": "24h"
  },
  "schema_version": 4
}
```

---

#### GET `/api/v1/meta/providers`

Get registered provider statuses.

**Response:**
```json
{
  "providers": [
    {
      "name": "system",
      "enabled": true,
      "running": true,
      "endpoint_count": 8
    },
    {
      "name": "nodes",
      "enabled": true,
      "running": true,
      "endpoint_count": 5
    },
    {
      "name": "onramp",
      "enabled": true,
      "running": true,
      "endpoint_count": 18
    },
    {
      "name": "meta",
      "enabled": true,
      "running": true,
      "endpoint_count": 6
    }
  ]
}
```

---

#### GET `/api/v1/meta/store`

Get store health and metadata, including live diagnostics.

**Response:**
```json
{
  "engine": "sqlite",
  "path": "/var/lib/aether-webd/app.db",
  "file_size_bytes": 524288,
  "schema_version": 4,
  "status": "healthy",
  "diagnostics": [
    { "name": "ping", "passed": true, "latency": "0.1ms" },
    { "name": "write", "passed": true, "latency": "1.2ms" },
    { "name": "read", "passed": true, "latency": "0.3ms" },
    { "name": "delete", "passed": true, "latency": "0.4ms" }
  ]
}
```

The `status` field is one of: `"healthy"`, `"degraded"`, `"unhealthy"`.

---

### System Provider

System endpoints return live host information. All are `GET` with no parameters unless noted.

#### GET `/api/v1/system/cpu`

**Response:**
```json
{
  "model": "Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz",
  "physical_cores": 8,
  "logical_cores": 16,
  "frequency_mhz": 3800,
  "cache_size_kb": 16384,
  "flags": ["sse4_2", "avx2", "aes"]
}
```

---

#### GET `/api/v1/system/memory`

**Response:**
```json
{
  "total_bytes": 34359738368,
  "available_bytes": 17179869184,
  "used_bytes": 17179869184,
  "usage_percent": 50.0,
  "swap_total_bytes": 8589934592,
  "swap_used_bytes": 0,
  "swap_percent": 0.0
}
```

---

#### GET `/api/v1/system/disks`

**Response:**
```json
{
  "partitions": [
    {
      "device": "/dev/sda1",
      "mountpoint": "/",
      "fs_type": "ext4",
      "total_bytes": 512110190592,
      "used_bytes": 128027547648,
      "free_bytes": 384082642944,
      "usage_percent": 25.0
    }
  ]
}
```

---

#### GET `/api/v1/system/os`

**Response:**
```json
{
  "hostname": "aether-node-01",
  "os": "linux",
  "platform": "ubuntu",
  "platform_version": "22.04",
  "kernel_version": "6.8.0-100-generic",
  "kernel_arch": "x86_64",
  "uptime_seconds": 86400
}
```

---

#### GET `/api/v1/system/network/interfaces`

**Response:**
```json
[
  {
    "name": "eth0",
    "mac": "00:1a:2b:3c:4d:5e",
    "mtu": 1500,
    "flags": ["up", "broadcast", "multicast"],
    "addresses": ["192.168.1.100/24", "fe80::1/64"]
  }
]
```

---

#### GET `/api/v1/system/network/config`

**Response:**
```json
{
  "dns_servers": ["8.8.8.8", "8.8.4.4"],
  "search_domains": ["example.com"]
}
```

---

#### GET `/api/v1/system/network/ports`

**Response:**
```json
[
  {
    "protocol": "tcp",
    "local_addr": "0.0.0.0",
    "local_port": 8186,
    "pid": 1234,
    "process_name": "aether-webd",
    "state": "LISTEN"
  }
]
```

---

#### GET `/api/v1/system/metrics`

Query time-series system metrics.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `metric` | string | **yes** | Metric name (e.g. `system.cpu.usage_percent`) |
| `from` | string | no | Start time, RFC 3339. Default: 1 hour ago |
| `to` | string | no | End time, RFC 3339. Default: now |
| `labels` | string | no | Comma-separated `key=val` label filters (e.g. `cpu=total`) |
| `aggregation` | string | no | Time bucket: `raw`, `10s`, `1m`, `5m`, `1h`. Default: `raw` |

**Example request:**
```
GET /api/v1/system/metrics?metric=system.cpu.usage_percent&from=2026-02-18T21:00:00Z&to=2026-02-18T21:30:00Z&aggregation=1m
```

**JavaScript:**
```js
const params = new URLSearchParams({
  metric: 'system.cpu.usage_percent',
  from: new Date(Date.now() - 3600000).toISOString(),
  to: new Date().toISOString(),
  aggregation: '1m',
});
const res = await fetch(`${BASE_URL}/system/metrics?${params}`, { headers });
const data = await res.json();
```

**Response:**
```json
{
  "series": [
    {
      "metric": "system.cpu.usage_percent",
      "labels": { "cpu": "total" },
      "points": [
        { "timestamp": "2026-02-18T21:00:00Z", "value": 23.5 },
        { "timestamp": "2026-02-18T21:01:00Z", "value": 25.1 }
      ]
    }
  ]
}
```

---

### Nodes Provider

CRUD operations for managed cluster nodes. Nodes represent hosts in the Ansible inventory with role assignments.

#### GET `/api/v1/nodes`

List all managed nodes.

**Response:**
```json
[
  {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "name": "node1",
    "ansible_host": "192.168.1.10",
    "ansible_user": "ubuntu",
    "has_password": true,
    "has_sudo_password": true,
    "has_ssh_key": false,
    "roles": ["master", "worker"],
    "created_at": "2026-02-18T12:00:00Z",
    "updated_at": "2026-02-18T12:00:00Z"
  }
]
```

> **Note:** Secrets (password, sudo_password, ssh_key) are never returned. Only boolean `has_*` flags indicate whether they are set.

---

#### GET `/api/v1/nodes/{id}`

Get a single node by ID.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Node ID |

**Response:** Same shape as a single item in the list response.

**Errors:**
- `404` — Node not found

---

#### POST `/api/v1/nodes`

Create a new node.

**Request Body:**
```json
{
  "name": "node1",
  "ansible_host": "192.168.1.10",
  "ansible_user": "ubuntu",
  "password": "secret",
  "sudo_password": "secret",
  "ssh_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n...",
  "roles": ["master", "worker"]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **yes** | Unique node name (used as Ansible inventory hostname) |
| `ansible_host` | string | **yes** | IP or hostname for SSH |
| `ansible_user` | string | no | SSH username |
| `password` | string | no | SSH password (encrypted at rest) |
| `sudo_password` | string | no | Sudo password (encrypted at rest) |
| `ssh_key` | string | no | SSH private key (encrypted at rest) |
| `roles` | string[] | no | Role assignments |

**Valid roles:** `master`, `worker`, `gnbsim`, `oai`, `ueransim`, `srsran`, `oscric`, `n3iwf`

**JavaScript:**
```js
const res = await fetch(`${BASE_URL}/nodes`, {
  method: 'POST',
  headers: { ...headers, 'Content-Type': 'application/json' },
  body: JSON.stringify({
    name: 'node1',
    ansible_host: '192.168.1.10',
    ansible_user: 'ubuntu',
    password: 'secret',
    sudo_password: 'secret',
    roles: ['master', 'worker'],
  }),
});
const node = await res.json();
```

**Response:** `200` with the created `ManagedNode` (same shape as GET).

**Errors:**
- `422` — Missing required field (`name` or `ansible_host`) or invalid role

---

#### PUT `/api/v1/nodes/{id}`

Partial update — merges non-null fields. Roles replaces entire set when provided.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Node ID |

**Request Body:** (all fields optional)
```json
{
  "name": "updated-name",
  "ansible_host": "10.0.0.5",
  "ansible_user": "admin",
  "password": "new-secret",
  "sudo_password": "new-secret",
  "ssh_key": "",
  "roles": ["master"]
}
```

> Set `password`, `sudo_password`, or `ssh_key` to an empty string `""` to clear the credential.

**Response:** `200` with the updated `ManagedNode`.

**Errors:**
- `404` — Node not found
- `422` — Invalid role

---

#### DELETE `/api/v1/nodes/{id}`

Delete a node and its role assignments.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Node ID |

**Response:**
```json
{
  "message": "node a1b2c3d4-e5f6-7890-abcd-ef1234567890 deleted"
}
```

---

### OnRamp Provider

Manages the Aether OnRamp deployment toolchain. Endpoints are organized into sub-groups: repo, components, tasks, config, profiles, and inventory.

#### Repo

##### GET `/api/v1/onramp/repo`

Get the OnRamp repository status.

**Response:**
```json
{
  "cloned": true,
  "dir": "/var/lib/aether-webd/aether-onramp",
  "repo_url": "https://github.com/opennetworkinglab/aether-onramp.git",
  "version": "main",
  "commit": "abc123def456",
  "branch": "main",
  "tag": "v2.2.0",
  "dirty": false
}
```

---

##### POST `/api/v1/onramp/repo/refresh`

Clone the repo if missing, check out the pinned version, and validate.

**Response:** Same shape as `GET /api/v1/onramp/repo`. If an error occurs, the `error` field is populated but the response is still `200`.

---

#### Components

##### GET `/api/v1/onramp/components`

List all available OnRamp components and their actions.

**Response:**
```json
[
  {
    "name": "k8s",
    "description": "Kubernetes (RKE2) cluster lifecycle",
    "actions": [
      { "name": "install", "description": "Deploy Kubernetes (RKE2)", "target": "aether-k8s-install" },
      { "name": "uninstall", "description": "Remove Kubernetes (RKE2)", "target": "aether-k8s-uninstall" }
    ]
  },
  {
    "name": "5gc",
    "description": "5G core network (SD-Core)",
    "actions": [
      { "name": "install", "description": "Deploy 5G core", "target": "aether-5gc-install" },
      { "name": "uninstall", "description": "Remove 5G core", "target": "aether-5gc-uninstall" },
      { "name": "reset", "description": "Reset 5G core state", "target": "aether-5gc-reset" }
    ]
  }
]
```

There are **12 components**: `k8s`, `5gc`, `4gc`, `gnbsim`, `amp`, `sdran`, `ueransim`, `oai`, `srsran`, `oscric`, `n3iwf`, `cluster`.

---

##### GET `/api/v1/onramp/components/{component}`

Get a single component by name.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `component` | string | Component name (e.g. `k8s`, `cluster`) |

**Response:** Single `Component` object (same shape as list items).

**Errors:**
- `404` — Unknown component name

---

##### POST `/api/v1/onramp/components/{component}/{action}`

Execute a component action (runs a make target).

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `component` | string | Component name |
| `action` | string | Action name (e.g. `install`, `uninstall`, `pingall`) |

**JavaScript:**
```js
const res = await fetch(`${BASE_URL}/onramp/components/cluster/pingall`, {
  method: 'POST',
  headers,
});
const task = await res.json();
// task.id can be used to poll for status
```

**Response:** `200` with an `OnRampTask`:
```json
{
  "id": "task-uuid-here",
  "component": "cluster",
  "action": "pingall",
  "target": "aether-pingall",
  "status": "running",
  "started_at": "2026-02-18T12:00:00Z",
  "exit_code": 0,
  "output": "",
  "output_offset": 0
}
```

**Errors:**
- `404` — Component or action not found
- `409` — A task is already running (only one concurrent task allowed)

---

#### Tasks

##### GET `/api/v1/onramp/tasks`

List all recent task executions.

**Response:**
```json
[
  {
    "id": "task-uuid",
    "component": "cluster",
    "action": "pingall",
    "target": "aether-pingall",
    "status": "succeeded",
    "started_at": "2026-02-18T12:00:00Z",
    "finished_at": "2026-02-18T12:00:05Z",
    "exit_code": 0,
    "output": "PLAY [all] ...\nok: [localhost]\n",
    "output_offset": 512
  }
]
```

---

##### GET `/api/v1/onramp/tasks/{id}`

Get a specific task with output.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Task ID |

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `offset` | int | `0` | Byte offset for incremental output reads |

The `offset` parameter enables efficient polling: pass the `output_offset` value from the previous response to only receive new output since the last read.

**JavaScript (polling pattern):**
```js
let offset = 0;
const poll = async (taskId) => {
  const res = await fetch(
    `${BASE_URL}/onramp/tasks/${taskId}?offset=${offset}`,
    { headers }
  );
  const task = await res.json();
  if (task.output) {
    appendToTerminal(task.output);  // only new output since last poll
  }
  offset = task.output_offset;
  return task;
};
```

**Errors:**
- `404` — Task not found

---

#### Action History

Persistent record of all executed actions, stored in SQLite. Enables audit trails and deployment dashboards.

##### GET `/api/v1/onramp/actions`

List action history with optional filtering and pagination.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `component` | string | | Filter by component name (exact match) |
| `action` | string | | Filter by action name (exact match) |
| `status` | string | | Filter by status (exact match) |
| `limit` | int | `50` | Max results |
| `offset` | int | `0` | Pagination offset |

Filters combine with AND when multiple are provided. Results are ordered by `started_at` descending (newest first).

**Example request:**
```
GET /api/v1/onramp/actions?component=k8s&status=succeeded&limit=10
```

**JavaScript:**
```js
const params = new URLSearchParams({ component: 'k8s', limit: '10' });
const res = await fetch(`${BASE_URL}/onramp/actions?${params}`, { headers });
const actions = await res.json();
```

**Response:**
```json
[
  {
    "id": "action-uuid",
    "component": "k8s",
    "action": "install",
    "target": "aether-k8s-install",
    "status": "succeeded",
    "exit_code": 0,
    "labels": { "profile": "gnbsim" },
    "tags": ["automated"],
    "started_at": 1708257600,
    "finished_at": 1708257660
  }
]
```

> **Note:** `started_at` and `finished_at` are Unix epoch seconds (int64), not ISO 8601 strings. `finished_at` is omitted when the action is still running. `error`, `labels`, and `tags` are omitted when empty.

---

##### GET `/api/v1/onramp/actions/{id}`

Get a single action record by ID.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Action ID |

**Response:** Single `ActionHistoryItem` (same shape as list items).

**Errors:**
- `404` — Action not found

---

#### Component State

Tracks the current deployment state of each component. States are automatically updated when actions complete.

##### GET `/api/v1/onramp/state`

List deployment state for all components. Always returns one entry per known component (currently 12). Components with no deployment history appear as `"not_installed"`.

**Response:**
```json
[
  {
    "component": "k8s",
    "status": "installed",
    "last_action": "install",
    "action_id": "action-uuid",
    "updated_at": 1708257660
  },
  {
    "component": "5gc",
    "status": "not_installed"
  }
]
```

> **Note:** `updated_at` is Unix epoch seconds. `last_action`, `action_id`, and `updated_at` are omitted for components with no recorded history.

**State values:** `not_installed`, `installing`, `installed`, `uninstalling`, `failed`

**State transitions:**
- Install action triggered → `installing`
- Install action succeeds → `installed`
- Install action fails → `failed`
- Uninstall action triggered → `uninstalling`
- Uninstall action succeeds → `not_installed`
- Uninstall action fails → `failed`

The transitional states (`installing`, `uninstalling`) are set when an action is submitted and can be used to show in-progress indicators in the frontend. Final states are set automatically when the action completes.

---

##### GET `/api/v1/onramp/state/{component}`

Get deployment state for a single component.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `component` | string | Component name (e.g. `k8s`, `5gc`) |

**Response:** Single `ComponentStateItem` (same shape as list items).

**Errors:**
- `404` — Unknown component name

---

#### Config

##### GET `/api/v1/onramp/config`

Read the current OnRamp configuration (`vars/main.yml`).

**Response:** An `OnRampConfig` object with optional sections:
```json
{
  "k8s": {
    "rke2": {
      "version": "v1.32.4+rke2r1",
      "config": {
        "token": "my-secret-token",
        "port": 9345,
        "params_file": { "master": "...", "worker": "..." }
      }
    }
  },
  "core": {
    "standalone": true,
    "data_iface": "ens18",
    "values_file": "sd-core-5g-values.yaml"
  }
}
```

All top-level keys are optional: `k8s`, `core`, `gnbsim`, `amp`, `sdran`, `ueransim`, `oai`, `srsran`, `n3iwf`.

---

##### PATCH `/api/v1/onramp/config`

Merge fields into the active configuration. Only provided top-level sections are replaced; untouched sections are preserved.

**Request Body:** Same shape as `OnRampConfig`, with only the sections you want to change:
```json
{
  "core": {
    "data_iface": "ens20"
  }
}
```

**Response:** `200` with the full updated `OnRampConfig`.

---

#### Profiles

##### GET `/api/v1/onramp/config/profiles`

List available config profiles (files matching `vars/main-*.yml`).

**Response:**
```json
["gnbsim", "oai", "srsran", "ueransim"]
```

---

##### GET `/api/v1/onramp/config/profiles/{name}`

Read a specific profile's configuration.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `name` | string | Profile name (e.g. `gnbsim`) |

**Response:** `OnRampConfig` (same shape as `GET /api/v1/onramp/config`).

**Errors:**
- `404` — Profile not found

---

##### POST `/api/v1/onramp/config/profiles/{name}/activate`

Copy the named profile to `vars/main.yml`, making it the active configuration.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `name` | string | Profile name |

**Response:**
```json
{
  "message": "profile \"gnbsim\" activated"
}
```

**Errors:**
- `404` — Profile not found

---

#### Inventory

##### GET `/api/v1/onramp/inventory`

Parse the current `hosts.ini` and return structured inventory data.

**Response:**
```json
{
  "nodes": [
    {
      "name": "localhost",
      "ansible_host": "127.0.0.1",
      "ansible_user": "ubuntu",
      "roles": ["master", "worker", "gnbsim"]
    }
  ]
}
```

---

##### POST `/api/v1/onramp/inventory/sync`

Generate `hosts.ini` from managed nodes in the database and write it to disk.

**Response:**
```json
{
  "message": "hosts.ini written with 1 nodes",
  "path": "/var/lib/aether-webd/aether-onramp/hosts.ini"
}
```

---

## Workflows

### 1. Deploying a Component

Multi-step flow: trigger an action, receive a task ID, poll for completion.

```
POST /api/v1/onramp/components/{component}/{action}
  → { "id": "task-123", "status": "running", ... }

GET /api/v1/onramp/tasks/task-123?offset=0
  → { "status": "running", "output": "...", "output_offset": 256 }

GET /api/v1/onramp/tasks/task-123?offset=256
  → { "status": "running", "output": "...", "output_offset": 512 }

GET /api/v1/onramp/tasks/task-123?offset=512
  → { "status": "succeeded", "exit_code": 0, "output": "...", "output_offset": 600 }
```

**Frontend implementation:**
```js
async function deployComponent(component, action) {
  // 1. Start the action
  const startRes = await fetch(`${BASE_URL}/onramp/components/${component}/${action}`, {
    method: 'POST',
    headers,
  });
  if (!startRes.ok) {
    const err = await startRes.json();
    throw new Error(err.detail || err.title);
  }
  const task = await startRes.json();

  // 2. Poll for completion
  let offset = 0;
  while (true) {
    await new Promise(r => setTimeout(r, 1000)); // 1s interval

    const pollRes = await fetch(
      `${BASE_URL}/onramp/tasks/${task.id}?offset=${offset}`,
      { headers }
    );
    const updated = await pollRes.json();

    if (updated.output) {
      appendToTerminal(updated.output);
    }
    offset = updated.output_offset;

    if (updated.status !== 'running' && updated.status !== 'pending') {
      return updated; // succeeded, failed, or canceled
    }
  }
}
```

### 2. Managing Nodes

CRUD workflow for cluster nodes, followed by inventory sync.

```
1. POST /api/v1/nodes                          ← Create node with roles + credentials
2. GET  /api/v1/nodes                           ← List all nodes
3. PUT  /api/v1/nodes/{id}                      ← Update node (partial merge)
4. POST /api/v1/onramp/inventory/sync           ← Write hosts.ini from DB
5. GET  /api/v1/onramp/inventory                ← Verify inventory state
6. POST /api/v1/onramp/components/cluster/pingall ← Test connectivity
```

After creating/updating nodes, always call inventory sync before running OnRamp actions so that `hosts.ini` reflects the current state.

### 3. Monitoring System Metrics

Poll metrics endpoints for live dashboard data.

```js
// Fetch CPU usage over the last 30 minutes, aggregated to 1-minute buckets
const from = new Date(Date.now() - 30 * 60000).toISOString();
const to = new Date().toISOString();
const params = new URLSearchParams({
  metric: 'system.cpu.usage_percent',
  from,
  to,
  aggregation: '1m',
});

const res = await fetch(`${BASE_URL}/system/metrics?${params}`, { headers });
const { series } = await res.json();
```

For a live dashboard, poll the metrics endpoint every 10–30 seconds. Use the `aggregation` parameter to control data density:
- `raw` — every collected sample (default 10s collection interval)
- `10s`, `1m`, `5m`, `1h` — pre-aggregated time buckets

Static system info (CPU model, total memory, disk layout, OS info, network interfaces) changes infrequently and can be fetched once on page load.

### 4. Deployment State Dashboard

Component state and action history enable a deployment dashboard without client-side tracking.

```
1. GET /api/v1/onramp/state                    ← Load current state for all 12 components
2. GET /api/v1/onramp/actions?limit=20         ← Load recent action history
```

After triggering a deploy, the state updates automatically:

```js
// 1. Start an install
const task = await fetch(`${BASE_URL}/onramp/components/k8s/install`, {
  method: 'POST', headers,
}).then(r => r.json());

// 2. Poll task until done (see "Deploying a Component" workflow)
// ...

// 3. Refresh state — k8s is now "installed" (or "failed")
const states = await fetch(`${BASE_URL}/onramp/state`, { headers }).then(r => r.json());

// 4. Optionally fetch history for a specific component
const history = await fetch(
  `${BASE_URL}/onramp/actions?component=k8s&limit=10`,
  { headers }
).then(r => r.json());
```

> **Note:** Timestamps in action history and component state are Unix epoch seconds (int64), not ISO 8601 strings. Convert on the frontend: `new Date(record.started_at * 1000)`.

### 5. Configuration Management

```
1. GET  /api/v1/onramp/config                    ← Read current config
2. PATCH /api/v1/onramp/config                   ← Update specific sections
3. GET  /api/v1/onramp/config/profiles           ← List available profiles
4. GET  /api/v1/onramp/config/profiles/{name}    ← Preview a profile
5. POST /api/v1/onramp/config/profiles/{name}/activate ← Switch to profile
```

Profiles are preset configurations (e.g., `gnbsim`, `oai`, `srsran`). Activating a profile overwrites `vars/main.yml` entirely.

### 6. Repository Management

```
1. GET  /api/v1/onramp/repo         ← Check clone status
2. POST /api/v1/onramp/repo/refresh ← Clone if missing, checkout pinned version
3. GET  /api/v1/onramp/repo         ← Verify state
```

The OnRamp repo is automatically cloned on server start. Use refresh to recover from corruption or switch versions.

---

## Data Models

### ManagedNode

Represents a cluster host in the node inventory.

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | UUID, server-generated |
| `name` | string | Ansible inventory hostname |
| `ansible_host` | string | IP or hostname for SSH |
| `ansible_user` | string | SSH username |
| `has_password` | bool | Whether an SSH password is stored |
| `has_sudo_password` | bool | Whether a sudo password is stored |
| `has_ssh_key` | bool | Whether an SSH key is stored |
| `roles` | string[] | Role assignments |
| `created_at` | string | ISO 8601 timestamp |
| `updated_at` | string | ISO 8601 timestamp |

> Credentials are encrypted at rest (AES-256-GCM). They are accepted in create/update requests but never returned in responses.

**Valid roles:** `master`, `worker`, `gnbsim`, `oai`, `ueransim`, `srsran`, `oscric`, `n3iwf`

### OnRampTask

Represents a make target execution.

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | UUID |
| `component` | string | Component name (e.g. `k8s`) |
| `action` | string | Action name (e.g. `install`) |
| `target` | string | Make target (e.g. `aether-k8s-install`) |
| `status` | string | Task state |
| `started_at` | string | ISO 8601 timestamp |
| `finished_at` | string | ISO 8601 timestamp (empty if running) |
| `exit_code` | int | Process exit code (0 = success) |
| `output` | string | Command stdout/stderr (or slice from offset) |
| `output_offset` | int | Byte offset for next incremental read |

**Task status values:** `pending` → `running` → `succeeded` | `failed` | `canceled`

### Component

Describes a deployable OnRamp component.

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Component identifier |
| `description` | string | Human-readable description |
| `actions` | Action[] | Available operations |

### Action

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Action identifier |
| `description` | string | Human-readable description |
| `target` | string | Makefile target |

### RepoStatus

| Field | Type | Description |
|-------|------|-------------|
| `cloned` | bool | Whether the repo directory exists |
| `dir` | string | Repo directory path |
| `repo_url` | string | Git remote URL |
| `version` | string | Pinned version (tag/branch/commit) |
| `commit` | string | Current HEAD commit hash |
| `branch` | string | Current branch name |
| `tag` | string | Tag at HEAD, if any |
| `dirty` | bool | Whether there are uncommitted changes |
| `error` | string | Error message, if repo setup failed |

### OnRampConfig

Top-level config object representing the OnRamp `vars/main.yml`. All sections are optional pointers:

| Section | Type | Description |
|---------|------|-------------|
| `k8s` | object | Kubernetes (RKE2) settings |
| `core` | object | 5G core network settings |
| `gnbsim` | object | gNBSim simulator settings |
| `amp` | object | Aether Management Platform settings |
| `sdran` | object | SD-RAN controller settings |
| `ueransim` | object | UERANSIM simulator settings |
| `oai` | object | OpenAirInterface RAN settings |
| `srsran` | object | srsRAN Project settings |
| `n3iwf` | object | Non-3GPP Interworking Function settings |

See the full schema in `api/openapi.json` under `#/components/schemas/OnRampConfig`.

### ActionHistoryItem

Represents one recorded action execution.

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | UUID |
| `component` | string | Component name (e.g. `k8s`) |
| `action` | string | Action name (e.g. `install`) |
| `target` | string | Make target (e.g. `aether-k8s-install`) |
| `status` | string | Execution result: `succeeded`, `failed`, `running` |
| `exit_code` | int | Process exit code (0 = success) |
| `error` | string | Error message (omitted when empty) |
| `labels` | object | Arbitrary key-value metadata (omitted when empty) |
| `tags` | string[] | Additional tags (omitted when empty) |
| `started_at` | int64 | Unix epoch seconds |
| `finished_at` | int64 | Unix epoch seconds (omitted while running) |

### ComponentStateItem

Represents the current deployment state of a component.

| Field | Type | Description |
|-------|------|-------------|
| `component` | string | Component name |
| `status` | string | Current state: `not_installed`, `installing`, `installed`, `uninstalling`, `failed` |
| `last_action` | string | Most recent action executed (omitted if no history) |
| `action_id` | string | ID of the action that set this state (omitted if no history) |
| `updated_at` | int64 | Unix epoch seconds of last state change (omitted if no history) |

### Metrics Series

| Field | Type | Description |
|-------|------|-------------|
| `metric` | string | Metric name |
| `labels` | object | Key-value label pairs identifying the series |
| `points` | PointResult[] | Time-ordered data points |

Each `PointResult` has `timestamp` (RFC 3339) and `value` (float64).

---

## Error Responses

All errors follow [RFC 9457 Problem Details](https://www.rfc-editor.org/rfc/rfc9457):

```json
{
  "title": "Not Found",
  "status": 404,
  "detail": "node not found"
}
```

Common status codes:

| Status | When |
|--------|------|
| `401` | Missing or invalid `Authorization` header (when auth is enabled) |
| `404` | Resource not found (node, task, component, action, profile) |
| `409` | Conflict — a task is already running (max 1 concurrent) |
| `422` | Validation error (missing required field, invalid role) |
| `500` | Internal server error (store failure, IO error) |

---

## Frontend Integration Notes

### Task Polling Pattern

The backend has no WebSocket endpoints — all communication is REST with polling. For task output streaming:

1. Start the action → receive a task ID
2. Poll `GET /tasks/{id}?offset=N` every 1 second
3. Append `output` to the UI terminal
4. Use `output_offset` as the next `offset` value
5. Stop polling when `status` is not `running` or `pending`

> **Future consideration:** WebSocket streaming may be added. Architect the frontend with an abstraction layer (e.g., `TaskStream` class) so the polling implementation can later be swapped for WebSocket without changing UI components.

### Concurrency Limit

Only **one OnRamp task** can run at a time. Attempting to start another returns `409 Conflict`. The frontend should disable action buttons while a task is running.

### CORS

By default, the backend does not set CORS headers. When developing the frontend on a different origin (e.g., `localhost:5173`), either:
- Use a dev proxy (Vite's `server.proxy`)
- Or run the frontend via the backend's built-in static file serving (`--serve-frontend`)

### Authentication Token

The auth token comes from server configuration (`--api-token` flag or `AETHER_API_TOKEN` env var). The frontend should read the token from its own configuration/environment, never hardcode it.

```js
// Example: read from environment or config
const API_TOKEN = import.meta.env.VITE_API_TOKEN;

const headers = API_TOKEN
  ? { Authorization: `Bearer ${API_TOKEN}` }
  : {};
```

### Theme and UI

The backend is UI-agnostic. Theme, layout, and design system choices are entirely the frontend's domain.
