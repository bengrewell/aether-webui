---
sidebar_position: 9
title: "Troubleshooting"
---

# Troubleshooting

This guide covers common problems and their solutions when operating aether-webd.

## Service will not start

### Check systemd status

```bash
sudo systemctl status aether-webd
sudo journalctl -u aether-webd --no-pager -n 50
```

Look for error messages related to flag parsing, port binding, or database initialization.

### Port conflict

If another process is already using port 8186:

```bash
ss -tlnp | grep 8186
```

Either stop the conflicting process or start aether-webd on a different port:

```bash
aether-webd --listen 0.0.0.0:9090
```

### Permission errors

The data directory (default `/var/lib/aether-webd`) must be writable by the aether-webd process. Check ownership:

```bash
ls -la /var/lib/aether-webd
```

## Task fails

### Check task output

Retrieve the full task output to identify the failure:

```bash
curl "http://localhost:8186/api/v1/onramp/tasks/{task_id}"
```

Ansible playbook errors typically appear near the end of the output. Look for lines containing `fatal:` or `FAILED`.

### Verify repository status

A dirty or mis-versioned repository can cause unexpected failures:

```bash
curl http://localhost:8186/api/v1/onramp/repo
```

If the response shows `"dirty": true` or an unexpected version, refresh the repository:

```bash
curl -X POST http://localhost:8186/api/v1/onramp/repo/refresh
```

See the [Repository guide](repository) for more detail.

### Check connectivity

Verify that all nodes are reachable before retrying:

```bash
curl -X POST http://localhost:8186/api/v1/onramp/components/cluster/pingall
```

Poll the resulting task for results. Connectivity failures indicate SSH credential or network issues. See [Managing Nodes](node-management) for credential setup.

## 409 Conflict: task already running

Only one task can run at a time. A `409` response means a task is already in progress:

```json
{
  "status": 409,
  "title": "Conflict",
  "detail": "a task is already running"
}
```

Check the current task list:

```bash
curl http://localhost:8186/api/v1/onramp/tasks
```

Wait for the running task to complete before starting a new one. See [Deploying Components](deploying-components) for the polling pattern.

## 401 Unauthorized

### Missing or incorrect token

If token authentication is enabled, every request to `/api/*` paths must include the `Authorization` header:

```bash
curl -H "Authorization: Bearer yourtoken" http://localhost:8186/api/v1/meta/version
```

### Check the configured token

Verify which authentication method is active:

```bash
curl http://localhost:8186/api/v1/meta/config
```

The response includes TLS and authentication state (token values are redacted).

### Exempt paths

These paths do not require a token: `/healthz`, `/openapi.json`, `/docs`, and paths not under `/api/`. See the [Security guide](security) for the full list.

## Connectivity issues

### SSH credentials

Ensure node records have correct `ansible_host`, `ansible_user`, and `password` values:

```bash
curl http://localhost:8186/api/v1/nodes
```

Update credentials if needed:

```bash
curl -X PUT http://localhost:8186/api/v1/nodes/{id} \
  -H "Content-Type: application/json" \
  -d '{"password": "newpassword"}'
```

### Inventory sync

After changing nodes, sync the inventory before running actions:

```bash
curl -X POST http://localhost:8186/api/v1/onramp/inventory/sync
```

Forgetting this step causes Ansible to target a stale host list.

### Network reachability

Ensure the aether-webd host can reach all nodes on the SSH port (default 22). Firewalls, security groups, or network segmentation may block access.

## Store diagnostics

The store diagnostic endpoint runs a live health check on the SQLite database:

```bash
curl http://localhost:8186/api/v1/meta/store
```

Example response:

```json
{
  "engine": "sqlite",
  "path": "/var/lib/aether-webd/app.db",
  "file_size_bytes": 49152,
  "schema_version": 4,
  "status": "healthy",
  "diagnostics": [
    { "name": "ping", "passed": true, "latency": "0.1ms" },
    { "name": "write", "passed": true, "latency": "0.5ms" },
    { "name": "read", "passed": true, "latency": "0.1ms" },
    { "name": "delete", "passed": true, "latency": "0.1ms" }
  ]
}
```

The `status` field is one of `healthy`, `degraded`, or `unhealthy`. If any diagnostic check shows `"passed": false`, the `error` field on that check describes the failure. A degraded or unhealthy status may indicate database corruption or a full filesystem.

## Repository issues

### Dirty repository

A dirty repository has uncommitted local changes that may interfere with `make` targets:

```bash
curl http://localhost:8186/api/v1/onramp/repo
# Look for "dirty": true
```

Refresh to reset:

```bash
curl -X POST http://localhost:8186/api/v1/onramp/repo/refresh
```

### Wrong version

If the repository is on the wrong branch or tag, restart aether-webd with the correct `--onramp-version` flag, or call the refresh endpoint after updating the flag.

## Certificate issues

### Regenerate auto-generated certificates

If certificates are expired or corrupted, delete the `certs/` directory and restart:

```bash
rm -rf /var/lib/aether-webd/certs/
sudo systemctl restart aether-webd
```

A new CA and server certificate are generated on startup. After regeneration, re-import the CA into browsers and system trust stores. See the [Security guide](security) for details.

### Certificate verification errors

When using auto-generated certificates, curl requires the CA:

```bash
curl --cacert /var/lib/aether-webd/certs/ca.pem https://localhost:8186/api/v1/meta/version
```

## Metrics not showing

### Check collection interval

Metrics are only collected if the background sampler is running. Verify the configured interval:

```bash
curl http://localhost:8186/api/v1/meta/config
```

Look for the `metrics_interval` field. If it is set to `0`, metric collection is disabled.

### Check retention

Metrics older than the retention window are pruned automatically. If querying for data outside the retention period, no results are returned:

```bash
# Default retention is 24h; this query may return no data
curl "http://localhost:8186/api/v1/system/metrics?metric=system.cpu.usage_percent&from=2026-02-28T00:00:00Z&to=2026-02-28T01:00:00Z"
```

### Verify the time range

Ensure the `from` and `to` parameters are in RFC 3339 format and cover a period where the server was running:

```bash
# Query the last 10 minutes
FROM=$(date -u -d '10 minutes ago' +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8186/api/v1/system/metrics?metric=system.cpu.usage_percent&from=${FROM}"
```

See the [Monitoring guide](monitoring) for query parameter details.
