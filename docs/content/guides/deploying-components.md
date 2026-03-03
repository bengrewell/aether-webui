---
sidebar_position: 3
title: "Deploying Components"
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Deploying Components

Component actions (install, uninstall, etc.) are executed asynchronously. A POST request starts the action and returns a task ID immediately. Poll the task endpoint to track progress and retrieve output.

## Trigger an action

<Tabs>
  <TabItem value="ui" label="Web UI" default>
    Navigate to **Components**. Click the action button (e.g., **Install**) on the target component row. The task output panel opens automatically and begins streaming Ansible output.
  </TabItem>
  <TabItem value="api" label="API">

```bash
curl -X POST http://localhost:8186/api/v1/onramp/components/k8s/install
```

Response:

```json
{
  "id": "a1b2c3d4-...",
  "component": "k8s",
  "action": "install",
  "target": "aether-k8s-install",
  "status": "running",
  "started_at": "2026-03-02T10:00:00Z",
  "output": "",
  "output_offset": 0
}
```

  </TabItem>
</Tabs>

## Poll for incremental output

<Tabs>
  <TabItem value="ui" label="Web UI" default>
    The task output panel streams Ansible output in real time. The status indicator updates as the task progresses through each play and task. When the action completes, the panel displays the final status and exit code.
  </TabItem>
  <TabItem value="api" label="API">

Use the `offset` query parameter to fetch only new output since your last read:

```bash
# First poll: start from offset 0
curl "http://localhost:8186/api/v1/onramp/tasks/a1b2c3d4-...?offset=0"
```

The response includes an `output_offset` field indicating where the output stream currently ends:

```json
{
  "id": "a1b2c3d4-...",
  "status": "running",
  "output": "PLAY [all] ***\nTASK [Gathering Facts] ***\n...",
  "output_offset": 1284
}
```

Pass the returned `output_offset` as the `offset` on your next request:

```bash
# Subsequent poll: pick up where the last response left off
curl "http://localhost:8186/api/v1/onramp/tasks/a1b2c3d4-...?offset=1284"
```

This returns only the output produced since byte 1284, avoiding redundant data transfer. Continue polling until the task status is no longer `running`.

  </TabItem>
</Tabs>

## Handle 409 Conflict

Only one task can run at a time. If you attempt to start an action while another is in progress, the server returns `409 Conflict`:

```json
{
  "status": 409,
  "title": "Conflict",
  "detail": "a task is already running"
}
```

Wait for the current task to finish or check its status:

```bash
curl http://localhost:8186/api/v1/onramp/tasks
```

## Check final task status

When a task completes, the `status` field is one of:

| Status | Meaning |
|--------|---------|
| `succeeded` | The action completed with exit code 0 |
| `failed` | The action exited with a non-zero code or encountered an error |
| `canceled` | The action was interrupted |

A finished task also includes `finished_at` and `exit_code` fields.

## Check deployment state

<Tabs>
  <TabItem value="ui" label="Web UI" default>
    The **Components** page shows the current state of each component (not_installed, installing, installed, etc.) with color-coded badges. Click a component row to see its full state details including the last action timestamp.
  </TabItem>
  <TabItem value="api" label="API">

View the current installed state of all components:

```bash
curl http://localhost:8186/api/v1/onramp/state
```

Or for a single component:

```bash
curl http://localhost:8186/api/v1/onramp/state/5gc
```

  </TabItem>
</Tabs>

Component state transitions:

```
not_installed → installing → installed
                           → failed
installed → uninstalling → not_installed
                         → failed
```

## View action history

<Tabs>
  <TabItem value="ui" label="Web UI" default>
    Navigate to **Action History**. Filter by component or status using the dropdowns. Click any row to see full details including timestamps, exit code, and output.

    The table supports pagination and defaults to showing the 50 most recent actions.
  </TabItem>
  <TabItem value="api" label="API">

Query past actions with optional filters:

```bash
# Last 10 actions for the k8s component
curl "http://localhost:8186/api/v1/onramp/actions?component=k8s&limit=10"

# All failed actions
curl "http://localhost:8186/api/v1/onramp/actions?status=failed"
```

Supported query parameters: `component`, `action`, `status`, `limit` (default 50), `offset` (default 0).

  </TabItem>
</Tabs>

## Typical deployment order

Run preflight checks first, then install components in dependency order:

0. `GET /api/v1/preflight` -- verify all prerequisites pass (fix any failures before proceeding)
1. `k8s install` -- Kubernetes cluster
2. `5gc install` -- 5G core network
3. `gnbsim install` or `srsran gnb-install` -- RAN simulator

Uninstall in reverse order:

1. `gnbsim uninstall` or `srsran gnb-uninstall`
2. `5gc uninstall`
3. `k8s uninstall`

## Complete deployment example

```bash
# Run preflight checks first
PREFLIGHT=$(curl -s http://localhost:8186/api/v1/preflight)
FAILED=$(echo "$PREFLIGHT" | jq '.failed')
if [ "$FAILED" -gt 0 ]; then
  echo "Preflight checks failed — fix issues before deploying"
  echo "$PREFLIGHT" | jq '.results[] | select(.passed == false)'
  exit 1
fi

# Install Kubernetes
TASK_ID=$(curl -s -X POST http://localhost:8186/api/v1/onramp/components/k8s/install | jq -r '.id')

# Poll until complete
OFFSET=0
while true; do
  RESP=$(curl -s "http://localhost:8186/api/v1/onramp/tasks/${TASK_ID}?offset=${OFFSET}")
  STATUS=$(echo "$RESP" | jq -r '.status')
  OUTPUT=$(echo "$RESP" | jq -r '.output')
  OFFSET=$(echo "$RESP" | jq -r '.output_offset')

  [ -n "$OUTPUT" ] && echo "$OUTPUT"
  [ "$STATUS" != "running" ] && break
  sleep 5
done

echo "Task finished with status: $STATUS"
```
