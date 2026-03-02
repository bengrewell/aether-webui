---
sidebar_position: 4
title: Verifying Your Deployment
---

# Verifying Your Deployment

After deploying Kubernetes and the 5G Core, this page walks through verifying that everything is installed and healthy.

## Check component state

The component state endpoint returns the current deployment status of every component:

```bash
curl http://localhost:8186/api/v1/onramp/state
```

After a successful deployment, `k8s` and `5gc` should both show `installed`:

```json
[
  {
    "component": "k8s",
    "state": "installed",
    "updated_at": "2026-01-15T10:35:00Z"
  },
  {
    "component": "5gc",
    "state": "installed",
    "updated_at": "2026-01-15T10:42:00Z"
  }
]
```

**Possible state values:**

| State | Meaning |
|-------|---------|
| `not_installed` | Component has never been deployed |
| `installing` | An install action is currently running |
| `installed` | Component is deployed and the install task succeeded |
| `uninstalling` | An uninstall action is currently running |
| `failed` | The most recent install or uninstall action failed |

If a component shows `failed`, check the task output for details (see below).

## Review action history

The action history endpoint shows a log of all actions that have been executed:

```bash
curl "http://localhost:8186/api/v1/onramp/actions?limit=5"
```

This returns the most recent actions, including their component, action name, status, and timestamps:

```json
[
  {
    "component": "5gc",
    "action": "install",
    "status": "succeeded",
    "started_at": "2026-01-15T10:38:00Z",
    "finished_at": "2026-01-15T10:42:00Z"
  },
  {
    "component": "k8s",
    "action": "install",
    "status": "succeeded",
    "started_at": "2026-01-15T10:30:00Z",
    "finished_at": "2026-01-15T10:35:00Z"
  }
]
```

## Inspect task output

If a deployment failed or you want to review the full output of any task, fetch it by ID:

```bash
curl "http://localhost:8186/api/v1/onramp/tasks/<task-id>"
```

The `output` field contains the combined stdout and stderr from the Make/Ansible execution. The `exit_code` field indicates the process exit code (`0` for success, non-zero for failure).

## View system metrics

With deployments running, verify that the host is healthy by checking system metrics.

**CPU usage (1-minute average):**

```bash
curl "http://localhost:8186/api/v1/system/metrics?metric=system.cpu.usage_percent&aggregation=1m"
```

**Memory usage (1-minute average):**

```bash
curl "http://localhost:8186/api/v1/system/metrics?metric=system.memory.usage_percent&aggregation=1m"
```

## Check system information

The system info endpoints return static information about the host:

```bash
# CPU information
curl http://localhost:8186/api/v1/system/cpu

# Memory information
curl http://localhost:8186/api/v1/system/memory

# Disk information
curl http://localhost:8186/api/v1/system/disks

# Operating system details
curl http://localhost:8186/api/v1/system/os
```

These are useful for confirming that the host meets the minimum hardware requirements and for troubleshooting resource-related issues.

## Verify store health

The store health endpoint reports the status of the internal SQLite database:

```bash
curl http://localhost:8186/api/v1/meta/store
```

A healthy response confirms that the persistence layer is operational and reports the current schema version.

## Next step

Everything is verified and running. Proceed to [Next Steps](next-steps) to learn about additional components, configuration, monitoring, and production hardening.
