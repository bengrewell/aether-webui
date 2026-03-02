---
sidebar_position: 4
title: Verifying Your Deployment
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Verifying Your Deployment

After deploying Kubernetes and the 5G Core, this page walks through verifying that everything is installed and healthy.

## Check component state

Confirm that both components show as installed:

<Tabs>
  <TabItem value="ui" label="Web UI" default>

Open the **Components** page. Each component shows its current state alongside the last update timestamp. Kubernetes and 5G Core should both display **Installed**.

If either component shows **Failed**, click on it to view the task output and diagnose the issue.

  </TabItem>
  <TabItem value="api" label="API">

The component state endpoint returns the current deployment status of every component:

```bash
curl http://localhost:8186/api/v1/onramp/state
```

After a successful deployment, `k8s` and `5gc` should both show `installed`:

```json
[
  {
    "component": "k8s",
    "status": "installed",
    "last_action": "install",
    "action_id": "d290f1ee-...",
    "updated_at": 1736935700
  },
  {
    "component": "5gc",
    "status": "installed",
    "last_action": "install",
    "action_id": "e3a1b2c4-...",
    "updated_at": 1736936520
  }
]
```

Note: `updated_at` is a Unix epoch timestamp (seconds). `last_action` and `action_id` are omitted for components with no deployment history.

  </TabItem>
</Tabs>

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

Review the log of all actions that have been executed:

<Tabs>
  <TabItem value="ui" label="Web UI" default>

Navigate to **Action History**. The table shows recent actions sorted by time, with columns for status, component, action name, and duration. Successful actions display a green status indicator; failed actions display red.

  </TabItem>
  <TabItem value="api" label="API">

```bash
curl "http://localhost:8186/api/v1/onramp/actions?limit=5"
```

This returns the most recent actions, including their component, action name, status, and timestamps:

```json
[
  {
    "component": "5gc",
    "action": "install",
    "target": "aether-5gc-install",
    "status": "succeeded",
    "exit_code": 0,
    "started_at": 1736936280,
    "finished_at": 1736936520
  },
  {
    "component": "k8s",
    "action": "install",
    "target": "aether-k8s-install",
    "status": "succeeded",
    "exit_code": 0,
    "started_at": 1736935400,
    "finished_at": 1736935700
  }
]
```

Note: `started_at` and `finished_at` are Unix epoch timestamps (seconds).

  </TabItem>
</Tabs>

## Inspect task output

Review the full output of any deployment task to verify details or diagnose failures:

<Tabs>
  <TabItem value="ui" label="Web UI" default>

Click on any action in the Action History to view its full task output. The detail view displays the complete stdout/stderr from the Ansible execution, along with the exit code, start time, and finish time.

  </TabItem>
  <TabItem value="api" label="API">

Fetch a task by its ID:

```bash
curl "http://localhost:8186/api/v1/onramp/tasks/<task-id>"
```

The `output` field contains the combined stdout and stderr from the Make/Ansible execution. The `exit_code` field indicates the process exit code (`0` for success, non-zero for failure).

  </TabItem>
</Tabs>

## View system metrics

With deployments running, verify that the host is healthy by checking system metrics:

<Tabs>
  <TabItem value="ui" label="Web UI" default>

Open the **Monitoring** dashboard. CPU, memory, and disk usage charts display real-time data with 1-minute aggregation. Verify that resource utilization is within acceptable limits after deploying both components.

  </TabItem>
  <TabItem value="api" label="API">

**CPU usage (1-minute average):**

```bash
curl "http://localhost:8186/api/v1/system/metrics?metric=system.cpu.usage_percent&aggregation=1m"
```

**Memory usage (1-minute average):**

```bash
curl "http://localhost:8186/api/v1/system/metrics?metric=system.memory.usage_percent&aggregation=1m"
```

  </TabItem>
</Tabs>

## Check system information

Confirm that the host meets hardware requirements and review system details:

<Tabs>
  <TabItem value="ui" label="Web UI" default>

The **System Info** page shows hardware details including CPU model and core count, total memory, disk partitions with usage, and OS version. Use this to verify the host meets the minimum requirements for running Aether.

  </TabItem>
  <TabItem value="api" label="API">

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

  </TabItem>
</Tabs>

## Verify store health

Confirm that the internal database is operational:

<Tabs>
  <TabItem value="ui" label="Web UI" default>

Navigate to **Settings > Diagnostics**. The store health section shows the database status and current schema version. A green indicator confirms the persistence layer is operational.

  </TabItem>
  <TabItem value="api" label="API">

The store health endpoint reports the status of the internal SQLite database:

```bash
curl http://localhost:8186/api/v1/meta/store
```

A healthy response confirms that the persistence layer is operational and reports the current schema version.

  </TabItem>
</Tabs>

## Next step

Everything is verified and running. Proceed to [Next Steps](next-steps) to learn about additional components, configuration, monitoring, and production hardening.
