---
sidebar_position: 3
title: First Deployment
---

# First Deployment

This page walks through deploying a Kubernetes cluster and the 5G Core network on a single node using the Aether WebUI API. By the end you will have both components installed and running.

## Step 1: Check the OnRamp repository status

Before deploying anything, verify that the Aether OnRamp repository has been cloned and is ready:

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

## Step 2: Deploy Kubernetes

Deploy the Kubernetes (RKE2) cluster by sending a POST request:

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

## Step 3: Poll the task for output

Tasks run asynchronously. Poll the task endpoint to watch progress and wait for completion:

```bash
curl "http://localhost:8186/api/v1/onramp/tasks/d290f1ee-6c54-4b01-90e6-d701748f0851"
```

The response includes a `status` field and the task output. For long-running tasks, use the `offset` query parameter for incremental output to avoid re-fetching content you have already seen:

```bash
curl "http://localhost:8186/api/v1/onramp/tasks/d290f1ee-6c54-4b01-90e6-d701748f0851?offset=4096"
```

The `output_offset` field in the response tells you the byte position to use as `offset` on your next request.

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

```bash
curl -X POST http://localhost:8186/api/v1/onramp/components/5gc/install
```

Save the task ID from the response.

## Step 5: Poll for completion

Poll the new task until it reaches `succeeded`:

```bash
curl "http://localhost:8186/api/v1/onramp/tasks/<task-id>"
```

The 5G Core deployment typically takes a few minutes as it installs Helm charts and waits for pods to become ready.

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
