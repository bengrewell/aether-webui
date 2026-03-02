---
sidebar_position: 6
title: "OnRamp Endpoints"
---

# OnRamp Endpoints

The OnRamp provider exposes 18 endpoints organized into 7 sub-groups for managing the Aether OnRamp deployment toolchain.

## Endpoint Summary

| Sub-group | Endpoint | Description |
|-----------|----------|-------------|
| **Repository** | [`GET /api/v1/onramp/repo`](#get-repo-status) | Repository status |
| | [`POST /api/v1/onramp/repo/refresh`](#refresh-repo) | Clone/checkout/validate repo |
| **Components** | [`GET /api/v1/onramp/components`](#list-components) | List all components |
| | [`GET /api/v1/onramp/components/{component}`](#get-component) | Get single component |
| | [`POST /api/v1/onramp/components/{component}/{action}`](#execute-action) | Execute component action |
| **Tasks** | [`GET /api/v1/onramp/tasks`](#list-tasks) | List tasks |
| | [`GET /api/v1/onramp/tasks/{id}`](#get-task) | Get task with incremental output |
| **Action History** | [`GET /api/v1/onramp/actions`](#list-action-history) | List actions with filters |
| | [`GET /api/v1/onramp/actions/{id}`](#get-action) | Get single action record |
| **Component State** | [`GET /api/v1/onramp/state`](#list-component-states) | All component states |
| | [`GET /api/v1/onramp/state/{component}`](#get-component-state) | Single component state |
| **Config** | [`GET /api/v1/onramp/config`](#get-config) | Read vars/main.yml |
| | [`PATCH /api/v1/onramp/config`](#patch-config) | Section-level merge |
| **Profiles** | [`GET /api/v1/onramp/config/profiles`](#list-profiles) | List profile names |
| | [`GET /api/v1/onramp/config/profiles/{name}`](#get-profile) | Read profile |
| | [`POST /api/v1/onramp/config/profiles/{name}/activate`](#activate-profile) | Copy profile to active config |
| **Inventory** | [`GET /api/v1/onramp/inventory`](#get-inventory) | Parse hosts.ini |
| | [`POST /api/v1/onramp/inventory/sync`](#sync-inventory) | Generate hosts.ini from DB |

---

## Schemas

### OnRampTask

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Task ID (UUID) |
| `component` | string | Component name |
| `action` | string | Action name |
| `target` | string | Make target |
| `status` | string | `pending`, `running`, `succeeded`, or `failed` |
| `started_at` | string | Start time (RFC 3339) |
| `finished_at` | string | Finish time (RFC 3339, omitted if still running) |
| `exit_code` | int | Process exit code (0 = success) |
| `output` | string | Task output (stdout + stderr) |
| `output_offset` | int | Byte offset for incremental reads |

### Component

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Component identifier |
| `description` | string | Human-readable description |
| `actions` | Action[] | Available actions |

### Action

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Action identifier |
| `description` | string | Human-readable description |
| `target` | string | Make target that this action invokes |

### RepoStatus

| Field | Type | Description |
|-------|------|-------------|
| `cloned` | bool | Whether the repo directory exists and contains `.git` |
| `dir` | string | Path to the repo on disk |
| `repo_url` | string | Git clone URL |
| `version` | string | Configured version (tag/branch/commit) |
| `commit` | string | Current HEAD commit hash (omitted if not cloned) |
| `branch` | string | Current branch name (omitted if not cloned) |
| `tag` | string | Tag at HEAD, if any (omitted if none) |
| `dirty` | bool | Whether the working tree has uncommitted changes |
| `error` | string | Error message, if any (omitted on success) |

### ActionHistoryItem

Note: Timestamps in action history are **Unix epoch seconds** (`int64`), not ISO 8601 strings.

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Action record ID (UUID) |
| `component` | string | Component name |
| `action` | string | Action name |
| `target` | string | Make target |
| `status` | string | `running`, `succeeded`, or `failed` |
| `exit_code` | int | Process exit code |
| `error` | string | Error message (omitted on success) |
| `labels` | object | User-supplied key-value labels (omitted if empty) |
| `tags` | string[] | User-supplied tags (omitted if empty) |
| `started_at` | int64 | Start time (Unix epoch seconds) |
| `finished_at` | int64 | Finish time (Unix epoch seconds, omitted if still running) |

### ComponentStateItem

Note: The `updated_at` field is **Unix epoch seconds** (`int64`), not ISO 8601.

| Field | Type | Description |
|-------|------|-------------|
| `component` | string | Component name |
| `status` | string | Deployment status (see [state values](./components.md#deployment-state)) |
| `last_action` | string | Most recent action name (omitted if no history) |
| `action_id` | string | Most recent action record ID (omitted if no history) |
| `updated_at` | int64 | Last state change (Unix epoch seconds, omitted if no history) |

---

## Repository

### Get Repo Status

```
GET /api/v1/onramp/repo
```

Returns the current state of the cloned OnRamp repository on disk.

```bash
curl http://localhost:8186/api/v1/onramp/repo
```

```json
{
  "cloned": true,
  "dir": "/var/lib/aether-webd/aether-onramp",
  "repo_url": "https://github.com/opennetworkinglab/aether-onramp.git",
  "version": "main",
  "commit": "abc123def456789",
  "branch": "main",
  "tag": "",
  "dirty": false
}
```

### Refresh Repo

```
POST /api/v1/onramp/repo/refresh
```

Clones the repo if missing, checks out the pinned version, and validates the directory. Returns the updated repo status.

```bash
curl -X POST http://localhost:8186/api/v1/onramp/repo/refresh
```

```json
{
  "cloned": true,
  "dir": "/var/lib/aether-webd/aether-onramp",
  "repo_url": "https://github.com/opennetworkinglab/aether-onramp.git",
  "version": "main",
  "commit": "abc123def456789",
  "branch": "main",
  "dirty": false
}
```

---

## Components

### List Components

```
GET /api/v1/onramp/components
```

Returns all available OnRamp components and their actions. See the [Components Reference](./components.md) for the full list.

```bash
curl http://localhost:8186/api/v1/onramp/components
```

```json
[
  {
    "name": "k8s",
    "description": "Kubernetes (RKE2) cluster lifecycle",
    "actions": [
      {"name": "install", "description": "Deploy Kubernetes (RKE2)", "target": "aether-k8s-install"},
      {"name": "uninstall", "description": "Remove Kubernetes (RKE2)", "target": "aether-k8s-uninstall"}
    ]
  }
]
```

### Get Component

```
GET /api/v1/onramp/components/{component}
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `component` | string | Component name (e.g., `k8s`, `5gc`, `srsran`) |

```bash
curl http://localhost:8186/api/v1/onramp/components/5gc
```

```json
{
  "name": "5gc",
  "description": "5G core network (SD-Core)",
  "actions": [
    {"name": "install", "description": "Deploy 5G core", "target": "aether-5gc-install"},
    {"name": "uninstall", "description": "Remove 5G core", "target": "aether-5gc-uninstall"},
    {"name": "reset", "description": "Reset 5G core state", "target": "aether-5gc-reset"}
  ]
}
```

#### Errors

| Status | When |
|--------|------|
| `404` | Unknown component name |

### Execute Action

```
POST /api/v1/onramp/components/{component}/{action}
```

Runs the Make target for the specified component and action. The operation is **asynchronous** -- it returns immediately with a task object that can be polled for progress.

Only 1 task can run at a time. Attempting to start a second task returns `409 Conflict`.

| Parameter | Type | Description |
|-----------|------|-------------|
| `component` | string | Component name |
| `action` | string | Action name |

#### Optional Request Body

| Field | Type | Description |
|-------|------|-------------|
| `labels` | object | Key-value labels to attach to the action history record |
| `tags` | string[] | Tags to attach to the action history record |

```bash
curl -X POST http://localhost:8186/api/v1/onramp/components/k8s/install \
  -H "Content-Type: application/json" \
  -d '{"labels": {"env": "staging"}, "tags": ["initial-deploy"]}'
```

```json
{
  "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "component": "k8s",
  "action": "install",
  "target": "aether-k8s-install",
  "status": "running",
  "started_at": "2026-02-18T14:00:00Z",
  "exit_code": 0,
  "output": "",
  "output_offset": 0
}
```

#### Errors

| Status | When |
|--------|------|
| `404` | Unknown component or action name |
| `409` | A task is already running |

---

## Tasks

### List Tasks

```
GET /api/v1/onramp/tasks
```

Returns all recent make target executions and their current status. Includes full output for each task.

```bash
curl http://localhost:8186/api/v1/onramp/tasks
```

```json
[
  {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "component": "k8s",
    "action": "install",
    "target": "aether-k8s-install",
    "status": "succeeded",
    "started_at": "2026-02-18T14:00:00Z",
    "finished_at": "2026-02-18T14:05:32Z",
    "exit_code": 0,
    "output": "PLAY [all] ...\nok: [node-01]\n...",
    "output_offset": 4096
  }
]
```

### Get Task

```
GET /api/v1/onramp/tasks/{id}?offset=N
```

Returns details and output for a specific task. Use the `offset` query parameter for **incremental output streaming** -- pass the `output_offset` from the previous response to receive only new output since the last read.

| Parameter | Type | In | Description |
|-----------|------|-----|-------------|
| `id` | string | path | Task ID |
| `offset` | int | query | Byte offset for incremental output reads (default: `0`) |

#### Incremental Polling Pattern

1. First request: `GET /api/v1/onramp/tasks/{id}` (offset defaults to 0)
2. Note the `output_offset` in the response (e.g., `4096`)
3. Next request: `GET /api/v1/onramp/tasks/{id}?offset=4096`
4. The response contains only output generated since byte 4096
5. Repeat until `status` is `succeeded` or `failed`

```bash
# Initial read
curl "http://localhost:8186/api/v1/onramp/tasks/f47ac10b-58cc-4372-a567-0e02b2c3d479"

# Incremental read (only new output)
curl "http://localhost:8186/api/v1/onramp/tasks/f47ac10b-58cc-4372-a567-0e02b2c3d479?offset=4096"
```

#### Errors

| Status | When |
|--------|------|
| `404` | No task with the given ID |

---

## Action History

Action history provides a persistent record of every component action execution, stored in the database. Unlike tasks (which are in-memory and transient), action history survives server restarts.

### List Action History

```
GET /api/v1/onramp/actions
```

Returns paginated action execution history, filterable by component, action, and status.

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `component` | string | - | Filter by component name |
| `action` | string | - | Filter by action name |
| `status` | string | - | Filter by status (`running`, `succeeded`, `failed`) |
| `limit` | int | `50` | Maximum number of results |
| `offset` | int | `0` | Pagination offset |

```bash
# All actions for the 5gc component
curl "http://localhost:8186/api/v1/onramp/actions?component=5gc"

# Failed actions, page 2
curl "http://localhost:8186/api/v1/onramp/actions?status=failed&limit=10&offset=10"
```

```json
[
  {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "component": "5gc",
    "action": "install",
    "target": "aether-5gc-install",
    "status": "succeeded",
    "exit_code": 0,
    "labels": {"env": "staging"},
    "tags": ["initial-deploy"],
    "started_at": 1708268400,
    "finished_at": 1708268732
  }
]
```

### Get Action

```
GET /api/v1/onramp/actions/{id}
```

Returns a single action execution record by ID.

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Action record ID |

```bash
curl http://localhost:8186/api/v1/onramp/actions/a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "component": "5gc",
  "action": "install",
  "target": "aether-5gc-install",
  "status": "succeeded",
  "exit_code": 0,
  "labels": {"env": "staging"},
  "tags": ["initial-deploy"],
  "started_at": 1708268400,
  "finished_at": 1708268732
}
```

#### Errors

| Status | When |
|--------|------|
| `404` | No action with the given ID |

---

## Component State

### List Component States

```
GET /api/v1/onramp/state
```

Returns the current deployment state of all registered components. Components with no deployment history default to `not_installed`.

```bash
curl http://localhost:8186/api/v1/onramp/state
```

```json
[
  {
    "component": "k8s",
    "status": "installed",
    "last_action": "install",
    "action_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "updated_at": 1708268732
  },
  {
    "component": "5gc",
    "status": "not_installed"
  },
  {
    "component": "4gc",
    "status": "not_installed"
  }
]
```

### Get Component State

```
GET /api/v1/onramp/state/{component}
```

Returns the current deployment state of a single component.

| Parameter | Type | Description |
|-----------|------|-------------|
| `component` | string | Component name |

```bash
curl http://localhost:8186/api/v1/onramp/state/k8s
```

```json
{
  "component": "k8s",
  "status": "installed",
  "last_action": "install",
  "action_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "updated_at": 1708268732
}
```

#### Errors

| Status | When |
|--------|------|
| `404` | Unknown component name |

---

## Config

For a detailed guide on the OnRamp configuration structure, sections, and profiles, see [OnRamp Configuration](./configuration.md).

### Get Config

```
GET /api/v1/onramp/config
```

Reads `vars/main.yml` from the OnRamp directory and returns the parsed configuration.

```bash
curl http://localhost:8186/api/v1/onramp/config
```

```json
{
  "k8s": {
    "rke2": {
      "version": "v1.28.2+rke2r1",
      "config": {
        "token": "my-cluster-token",
        "port": 9345,
        "params_file": {
          "master": "config/server.yaml",
          "worker": "config/agent.yaml"
        }
      }
    }
  },
  "core": {
    "standalone": true,
    "data_iface": "eth0",
    "values_file": "sd-core-5g-values.yaml"
  }
}
```

### Patch Config

```
PATCH /api/v1/onramp/config
```

Merges the provided fields into `vars/main.yml`. This is a **section-level merge**: each top-level key in the request body replaces the corresponding section entirely. Omitting a section leaves it unchanged.

Note: This is **not** a deep merge. If the request body includes a `core` key, the entire `core` section is replaced, not individual fields within it.

```bash
curl -X PATCH http://localhost:8186/api/v1/onramp/config \
  -H "Content-Type: application/json" \
  -d '{
    "core": {
      "standalone": true,
      "data_iface": "ens192",
      "values_file": "sd-core-5g-values.yaml"
    }
  }'
```

The response contains the complete merged configuration (all sections).

---

## Profiles

Profiles are pre-defined configuration files stored as `vars/main-{name}.yml` in the OnRamp directory. Each profile contains a complete `OnRampConfig` tuned for a specific deployment scenario.

Standard profiles: `gnbsim`, `oai`, `srsran`, `ueransim`.

### List Profiles

```
GET /api/v1/onramp/config/profiles
```

Returns an array of available profile names.

```bash
curl http://localhost:8186/api/v1/onramp/config/profiles
```

```json
["gnbsim", "oai", "srsran", "ueransim"]
```

### Get Profile

```
GET /api/v1/onramp/config/profiles/{name}
```

Returns the parsed configuration from the named profile.

| Parameter | Type | Description |
|-----------|------|-------------|
| `name` | string | Profile name (e.g., `srsran`) |

```bash
curl http://localhost:8186/api/v1/onramp/config/profiles/srsran
```

```json
{
  "k8s": {
    "rke2": {
      "version": "v1.28.2+rke2r1"
    }
  },
  "core": {
    "standalone": false,
    "data_iface": "ens192"
  },
  "srsran": {
    "docker": {
      "container": {
        "gnb_image": "softwareradiosystems/srsran-project:latest",
        "ue_image": "softwareradiosystems/srsue:latest"
      }
    },
    "simulation": true
  }
}
```

#### Errors

| Status | When |
|--------|------|
| `404` | No profile with the given name |

### Activate Profile

```
POST /api/v1/onramp/config/profiles/{name}/activate
```

Copies the named profile to `vars/main.yml`, making it the active configuration. The previous `main.yml` is overwritten.

| Parameter | Type | Description |
|-----------|------|-------------|
| `name` | string | Profile name |

```bash
curl -X POST http://localhost:8186/api/v1/onramp/config/profiles/srsran/activate
```

```json
{
  "message": "profile \"srsran\" activated"
}
```

#### Errors

| Status | When |
|--------|------|
| `404` | No profile with the given name |

---

## Inventory

### Get Inventory

```
GET /api/v1/onramp/inventory
```

Parses the current `hosts.ini` file and returns structured inventory data.

```bash
curl http://localhost:8186/api/v1/onramp/inventory
```

```json
{
  "nodes": [
    {
      "name": "node-01",
      "ansible_host": "192.168.1.10",
      "ansible_user": "ubuntu",
      "roles": ["master"]
    },
    {
      "name": "node-02",
      "ansible_host": "192.168.1.11",
      "ansible_user": "ubuntu",
      "roles": ["worker"]
    }
  ]
}
```

### Sync Inventory

```
POST /api/v1/onramp/inventory/sync
```

Generates `hosts.ini` from managed nodes in the database and writes it to disk. This should be called after adding, updating, or removing nodes to keep the Ansible inventory in sync.

```bash
curl -X POST http://localhost:8186/api/v1/onramp/inventory/sync
```

```json
{
  "message": "inventory synced",
  "path": "/var/lib/aether-webd/aether-onramp/hosts.ini"
}
```
