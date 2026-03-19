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

## Preflight checks failing

### Run all checks

The preflight endpoint verifies system prerequisites and reports which are passing or failing:

```bash
curl http://localhost:8186/api/v1/preflight
```

Review the `results` array. Each item has `passed`, `message`, and `can_fix` fields.

### Apply automated fixes

For checks with `"can_fix": true`, apply the fix and re-run:

```bash
# Fix a specific check
curl -X POST http://localhost:8186/api/v1/preflight/required-packages/fix

# Re-run to verify
curl http://localhost:8186/api/v1/preflight/required-packages
```

Fixes that modify system state (installing packages, creating users, writing config files) run via `sudo`. Review the `fix_warning` field before applying.

### Common preflight failures

| Check | Common cause | Resolution |
|-------|-------------|------------|
| `required-packages` | Fresh host without build tools | Apply the automated fix, or manually install `make` and the `ansible` package (which provides `ansible-playbook`) |
| `ssh-configured` | Cloud images disable password auth by default | Apply the automated fix, which writes an sshd drop-in config |
| `aether-user-configured` | User not yet created | Apply the automated fix to create the `aether` user with sudo |
| `node-ssh-reachable` | Firewall blocking port 22, incorrect `ansible_host` | Check node configuration and network connectivity |

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

## Task interrupted by service restart

### Symptom

A deployment or task that was in progress suddenly shows `status: "failed"` with the error `"service restarted while task was running"`. The service journal shows a SIGTERM during task execution.

### Cause

Something restarted the `aether-webd` service while a task was running. The most common trigger is Ubuntu's `needrestart` tool, which automatically restarts services after `apt` installs or upgrades packages. Since Aether OnRamp playbooks install packages (Docker, Python modules, etc.) via Ansible, `needrestart` detects that `aether-webd` is using updated libraries and restarts it mid-task.

### Diagnosis

Check the service journal for the restart event:

```bash
journalctl -u aether-webd --no-pager -b | grep -E 'SIGTERM|Stop|Start'
```

Check if `needrestart` is installed and whether `aether-webd` is excluded:

```bash
dpkg -l needrestart
cat /etc/needrestart/conf.d/aether-webd.conf
```

### Fix

Create a `needrestart` exclusion so it never restarts `aether-webd`:

```bash
sudo mkdir -p /etc/needrestart/conf.d
cat <<'EOF' | sudo tee /etc/needrestart/conf.d/aether-webd.conf
$nrconf{override_rc}{qr(^aether-webd)} = 0;
EOF
```

Note: The install script creates this file automatically. This step is only needed for manual or source-built installations.

### Recovery

On startup, `aether-webd` automatically detects actions and deployments that were interrupted by the previous shutdown and marks them as failed. Re-submit the deployment to retry.

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

## CORS errors in the browser

### Symptom

The browser console shows errors like:

```
Access to fetch at 'http://localhost:8186/api/v1/...' from origin 'http://localhost:5173'
has been blocked by CORS policy: No 'Access-Control-Allow-Origin' header is present
on the requested resource.
```

### Cause

The frontend is served from a different origin than the API (e.g., a Vite dev server on `:5173` and the API on `:8186`), and CORS is not configured.

### Fix

Set `AETHER_CORS_ORIGINS` to the frontend origin:

```bash
echo 'AETHER_CORS_ORIGINS=http://localhost:5173' | sudo tee -a /etc/aether-webd/env
sudo systemctl restart aether-webd
```

Or pass it directly:

```bash
aether-webd --cors-origins http://localhost:5173
```

### Verify

Confirm that the origin appears in the active configuration:

```bash
curl http://localhost:8186/api/v1/meta/config | jq '.security.cors_origins'
```

See the [Security guide](security#cors-cross-origin-resource-sharing) for details on allowed methods, headers, and production recommendations.

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

Ensure the aether-webd host can reach all nodes on the SSH port (default 22). The preflight `node-ssh-reachable` check can verify this:

```bash
curl http://localhost:8186/api/v1/preflight/node-ssh-reachable
```

Firewalls, security groups, or network segmentation may block access.

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
