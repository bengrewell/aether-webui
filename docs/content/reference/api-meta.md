---
sidebar_position: 3
title: "Meta Endpoints"
---

# Meta Endpoints

The meta provider exposes 6 read-only endpoints for application introspection. All endpoints accept `GET` requests with no parameters.

| Endpoint | Description |
|----------|-------------|
| [`GET /api/v1/meta/version`](#get-version) | Build-time version metadata |
| [`GET /api/v1/meta/build`](#get-build-info) | Go toolchain and platform details |
| [`GET /api/v1/meta/runtime`](#get-runtime-info) | Process runtime details |
| [`GET /api/v1/meta/config`](#get-config) | Active application configuration |
| [`GET /api/v1/meta/providers`](#get-providers) | Registered provider statuses |
| [`GET /api/v1/meta/store`](#get-store-diagnostics) | Store health and diagnostics |

---

## GET Version

```
GET /api/v1/meta/version
```

Returns the server's version string, build date, git branch, and commit hash.

### Response Schema

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Semantic version string (e.g., `1.2.0`) |
| `build_date` | string | Build timestamp (ISO 8601) |
| `branch` | string | Git branch at build time |
| `commit_hash` | string | Short git commit hash |

### Example

```bash
curl http://localhost:8186/api/v1/meta/version
```

```json
{
  "version": "1.2.0",
  "build_date": "2026-02-18T12:00:00Z",
  "branch": "main",
  "commit_hash": "a1b2c3d"
}
```

---

## GET Build Info

```
GET /api/v1/meta/build
```

Returns the Go toolchain version, target OS, and architecture used to compile the server binary.

### Response Schema

| Field | Type | Description |
|-------|------|-------------|
| `go_version` | string | Go toolchain version (e.g., `go1.25.0`) |
| `os` | string | Target operating system |
| `arch` | string | Target architecture |

### Example

```bash
curl http://localhost:8186/api/v1/meta/build
```

```json
{
  "go_version": "go1.25.0",
  "os": "linux",
  "arch": "amd64"
}
```

---

## GET Runtime Info

```
GET /api/v1/meta/runtime
```

Returns the server's process ID, running user/group, binary path, start time, and uptime.

### Response Schema

| Field | Type | Description |
|-------|------|-------------|
| `pid` | int | Process ID |
| `user` | object | Process owner (`uid`, `name`) |
| `group` | object | Process primary group (`gid`, `name`) |
| `binary_path` | string | Path to the running binary |
| `start_time` | string | Server start time (RFC 3339) |
| `uptime` | string | Time since server start (e.g., `2h30m0s`) |

### Example

```bash
curl http://localhost:8186/api/v1/meta/runtime
```

```json
{
  "pid": 4821,
  "user": {
    "uid": "1000",
    "name": "aether"
  },
  "group": {
    "gid": "1000",
    "name": "aether"
  },
  "binary_path": "/usr/local/bin/aether-webd",
  "start_time": "2026-02-18T12:00:00Z",
  "uptime": "2h30m0s"
}
```

---

## GET Config

```
GET /api/v1/meta/config
```

Returns the active (non-secret) application configuration including listen address, debug state, security settings, frontend config, storage paths, metrics settings, and the current database schema version.

### Response Schema

| Field | Type | Description |
|-------|------|-------------|
| `listen_address` | string | Server listen address |
| `debug_enabled` | bool | Whether debug mode is enabled |
| `security` | object | Security configuration (see below) |
| `frontend` | object | Frontend serving configuration (see below) |
| `storage` | object | Storage configuration (see below) |
| `metrics` | object | Metrics collection configuration (see below) |
| `schema_version` | int | Current database schema version |

**`security` object:**

| Field | Type | Description |
|-------|------|-------------|
| `tls_enabled` | bool | Whether TLS is enabled |
| `tls_auto_generated` | bool | Whether the TLS certificate was auto-generated |
| `mtls_enabled` | bool | Whether mutual TLS is enabled |
| `token_auth_enabled` | bool | Whether bearer token authentication is enabled |
| `rbac_enabled` | bool | Whether RBAC is enabled |
| `cors_origins` | string[] | Allowed CORS origins (omitted when CORS is disabled) |

**`frontend` object:**

| Field | Type | Description |
|-------|------|-------------|
| `enabled` | bool | Whether frontend serving is enabled |
| `source` | string | Frontend source (`embedded` or `directory`) |
| `dir` | string | Custom frontend directory path, if any |

**`storage` object:**

| Field | Type | Description |
|-------|------|-------------|
| `data_dir` | string | Directory for persistent state database |

**`metrics` object:**

| Field | Type | Description |
|-------|------|-------------|
| `interval` | string | Metrics collection interval |
| `retention` | string | Metrics retention period |

### Example

```bash
curl http://localhost:8186/api/v1/meta/config
```

```json
{
  "listen_address": "127.0.0.1:8186",
  "debug_enabled": false,
  "security": {
    "tls_enabled": false,
    "tls_auto_generated": false,
    "mtls_enabled": false,
    "token_auth_enabled": true,
    "rbac_enabled": false,
    "cors_origins": ["http://localhost:5173"]
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

## GET Providers

```
GET /api/v1/meta/providers
```

Returns the name, enabled/running state, and endpoint count for each registered provider.

### Response Schema

The response body contains a `providers` array. Each element:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Provider name |
| `enabled` | bool | Whether the provider is enabled |
| `running` | bool | Whether the provider is currently running |
| `endpoint_count` | int | Number of registered endpoints |

### Example

```bash
curl http://localhost:8186/api/v1/meta/providers
```

```json
{
  "providers": [
    {
      "name": "meta",
      "enabled": true,
      "running": true,
      "endpoint_count": 6
    },
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
    }
  ]
}
```

---

## GET Store Diagnostics

```
GET /api/v1/meta/store
```

Returns store engine metadata, file size, schema version, and live diagnostic results. The diagnostics run ping, write, read, and delete checks against the database.

### Response Schema

| Field | Type | Description |
|-------|------|-------------|
| `engine` | string | Storage engine type (e.g., `sqlite`) |
| `path` | string | Database file path |
| `file_size_bytes` | int64 | Database file size in bytes |
| `schema_version` | int | Current schema migration version |
| `status` | string | Overall store health: `healthy`, `degraded`, or `unhealthy` |
| `diagnostics` | array | Individual diagnostic check results (see below) |

**`diagnostics` array elements:**

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Diagnostic check name (e.g., `ping`, `write`, `read`, `delete`) |
| `passed` | bool | Whether the check passed |
| `latency` | string | Check execution time (e.g., `1.2ms`) |
| `error` | string | Error message if the check failed (omitted on success) |

### Example

```bash
curl http://localhost:8186/api/v1/meta/store
```

```json
{
  "engine": "sqlite",
  "path": "/var/lib/aether-webd/app.db",
  "file_size_bytes": 524288,
  "schema_version": 4,
  "status": "healthy",
  "diagnostics": [
    {
      "name": "ping",
      "passed": true,
      "latency": "0.3ms"
    },
    {
      "name": "write",
      "passed": true,
      "latency": "1.1ms"
    },
    {
      "name": "read",
      "passed": true,
      "latency": "0.2ms"
    },
    {
      "name": "delete",
      "passed": true,
      "latency": "0.8ms"
    }
  ]
}
```
