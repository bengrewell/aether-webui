---
sidebar_position: 3
title: First Deployment
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# First Deployment

This page walks through deploying a Kubernetes cluster and the 5G Core network on a single node using the Aether WebUI. By the end you will have both components installed and running.

## Step 1: Check the OnRamp repository status

Before deploying anything, verify that the Aether OnRamp repository has been cloned and is ready:

<Tabs>
  <TabItem value="ui" label="Web UI" default>

Open the Web UI. The dashboard shows the OnRamp repository status, including the clone state, current branch, and commit hash. Verify it shows as **Cloned** and is on the expected version.

If the repository has not been cloned yet, click **Refresh Repository** to trigger a clone.

  </TabItem>
  <TabItem value="api" label="API">

```bash
curl http://localhost:8186/api/v1/onramp/repo
```

The response should include `"cloned": true`:

```json
{
  "cloned": true,
  "dir": "/opt/aether-onramp",
  "repo_url": "https://github.com/opennetworkinglab/aether-onramp.git",
  "version": "main",
  "commit": "abc1234...",
  "branch": "main",
  "dirty": false
}
```

If `cloned` is `false`, trigger a repo refresh and wait for it to complete:

```bash
curl -X POST http://localhost:8186/api/v1/onramp/repo/refresh
```

  </TabItem>
</Tabs>

## Step 2: Deploy Kubernetes

Deploy the Kubernetes (RKE2) cluster:

<Tabs>
  <TabItem value="ui" label="Web UI" default>

Navigate to **Components**. Find **Kubernetes (k8s)** in the list and click **Install**. The task output panel opens automatically, showing real-time Ansible output as the deployment progresses.

  </TabItem>
  <TabItem value="api" label="API">

Send a POST request to start the install:

```bash
curl -X POST http://localhost:8186/api/v1/onramp/components/k8s/install
```

The response contains a task object with an `id` field. Save this ID -- you will use it to track progress:

```json
{
  "id": "d290f1ee-6c54-4b01-90e6-d701748f0851",
  "component": "k8s",
  "action": "install",
  "target": "aether-k8s-install",
  "status": "pending",
  "started_at": "2026-01-15T10:30:00Z"
}
```

  </TabItem>
</Tabs>

## Step 3: Poll the task for output

Tasks run asynchronously. Wait for the Kubernetes install to finish before continuing.

<Tabs>
  <TabItem value="ui" label="Web UI" default>

The task output panel streams output in real time. Wait for the task status to show **Succeeded** before proceeding to the next step.

If the task fails, the panel displays the error output and exit code. Review the output to diagnose the issue.

  </TabItem>
  <TabItem value="api" label="API">

Poll the task endpoint to watch progress and wait for completion:

```bash
curl "http://localhost:8186/api/v1/onramp/tasks/d290f1ee-6c54-4b01-90e6-d701748f0851"
```

The response includes a `status` field and the task output. For long-running tasks, use the `offset` query parameter for incremental output to avoid re-fetching content you have already seen:

```bash
curl "http://localhost:8186/api/v1/onramp/tasks/d290f1ee-6c54-4b01-90e6-d701748f0851?offset=4096"
```

The `output_offset` field in the response tells you the byte position to use as `offset` on your next request.

  </TabItem>
</Tabs>

**Task status values:**

| Status | Meaning |
|--------|---------|
| `pending` | Task is queued but has not started executing |
| `running` | Task is actively executing |
| `succeeded` | Task completed successfully (exit code 0) |
| `failed` | Task completed with a non-zero exit code |
| `canceled` | Task was canceled before completion |

Keep polling until `status` is `succeeded`. The Kubernetes install typically takes several minutes.

## Step 4: Deploy the 5G Core

Once Kubernetes is running, deploy the 5G Core network (SD-Core):

<Tabs>
  <TabItem value="ui" label="Web UI" default>

Return to **Components**. Find **5G Core (5gc)** in the list and click **Install**. The task output panel opens again with live output from the deployment.

  </TabItem>
  <TabItem value="api" label="API">

```bash
curl -X POST http://localhost:8186/api/v1/onramp/components/5gc/install
```

Save the task ID from the response.

  </TabItem>
</Tabs>

## Step 5: Poll for completion

Wait for the 5G Core deployment to finish:

<Tabs>
  <TabItem value="ui" label="Web UI" default>

Monitor the task output panel until the task shows **Succeeded**. The 5G Core deployment typically takes a few minutes as it installs Helm charts and waits for pods to become ready.

  </TabItem>
  <TabItem value="api" label="API">

Poll the new task until it reaches `succeeded`:

```bash
curl "http://localhost:8186/api/v1/onramp/tasks/<task-id>"
```

The 5G Core deployment typically takes a few minutes as it installs Helm charts and waits for pods to become ready.

  </TabItem>
</Tabs>

## Understanding the single-task constraint

Aether WebUI enforces a single-task execution model. Only one deployment task can run at a time. If you attempt to start a second action while a task is already running, the API returns a `409 Conflict` response:

```json
{
  "status": 409,
  "title": "Conflict",
  "detail": "a task is already running"
}
```

Wait for the current task to complete before starting the next one.

## Next step

With Kubernetes and the 5G Core deployed, proceed to [Verifying Your Deployment](verifying) to confirm everything is working correctly.
