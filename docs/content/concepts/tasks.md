---
sidebar_position: 3
title: Tasks and Async Execution
---

# Tasks and Async Execution

Deploying network components takes time -- installing a Kubernetes cluster or a 5G core network can run for several minutes. Rather than blocking the HTTP request for the entire duration, Aether WebUI executes deployment actions asynchronously using tasks.

## Why tasks are async

When you trigger an action like installing the 5G core, the server needs to run a Make target that in turn invokes Ansible playbooks. These playbooks download container images, configure services, and wait for pods to become healthy. This can take anywhere from 30 seconds to 10 minutes or more.

Holding an HTTP connection open for that long is fragile -- proxies may time out, clients may disconnect, and there is no way to show incremental progress. Instead, the API returns immediately with a task ID, and the work continues in the background.

## Task lifecycle

A task moves through a predictable set of states:

```
  POST /onramp/components/{component}/{action}
                    │
                    ▼
              ┌──────────┐
              │  pending  │   Task created, waiting to start
              └─────┬─────┘
                    │
                    ▼
              ┌──────────┐
              │  running  │   Make target executing, output accumulating
              └─────┬─────┘
                    │
            ┌───────┴───────┐
            ▼               ▼
     ┌───────────┐   ┌──────────┐
     │ succeeded  │   │  failed  │
     └───────────┘   └──────────┘
```

- **pending** -- The task has been created but execution has not started yet.
- **running** -- The Make target is executing. Output (combined stdout and stderr) is being captured.
- **succeeded** -- The Make target exited with code 0. The deployment action completed successfully.
- **failed** -- The Make target exited with a non-zero code, or could not be started at all.

A task can also reach the **canceled** status if it is terminated before completing.

## Starting a task

To start a deployment action, send a POST request to the component action endpoint:

```
POST /api/v1/onramp/components/5gc/install
```

The response returns immediately with the new task:

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "component": "5gc",
  "action": "install",
  "target": "aether-5gc-install",
  "status": "running",
  "started_at": "2026-03-02T14:30:00Z",
  "output": "",
  "output_offset": 0
}
```

## Polling for progress

Once you have a task ID, poll the task endpoint to check status and retrieve output:

```
GET /api/v1/onramp/tasks/{id}
```

### Incremental output with offsets

Task output can be large (thousands of lines of Ansible output). To avoid re-fetching the entire output on every poll, the API supports offset-based incremental reads.

Each response includes an `output_offset` field that tells you where to start reading next:

```
1.  GET /api/v1/onramp/tasks/{id}
    → output: "PLAY [all] ***\nTASK [Gathering Facts]...",  output_offset: 1482

2.  GET /api/v1/onramp/tasks/{id}?offset=1482
    → output: "TASK [Install SD-Core]...",  output_offset: 3271

3.  GET /api/v1/onramp/tasks/{id}?offset=3271
    → output: "PLAY RECAP ***\nlocalhost: ok=42...",  output_offset: 3890
```

Pass the `output_offset` from the previous response as the `offset` query parameter on the next request. This returns only the new output since your last read.

### When to stop polling

Stop polling when the task status is no longer `running` or `pending`. At that point, the task is complete and no more output will be produced. Check the final status and exit code to determine whether the action succeeded:

| Status | Exit code | Meaning |
|--------|-----------|---------|
| `succeeded` | `0` | Action completed successfully |
| `failed` | Non-zero | Action failed; check the output for error details |
| `failed` | `-1` | The Make target could not be started at all (e.g., binary not found) |
| `canceled` | -- | The task was terminated before completing |

### Recommended polling pattern

A typical polling loop looks like:

1. `POST /api/v1/onramp/components/{component}/{action}` -- start the action, save the returned task ID.
2. Wait a short interval (1-3 seconds).
3. `GET /api/v1/onramp/tasks/{id}?offset={last_offset}` -- fetch new output.
4. Display or log the new output.
5. If `status` is `running` or `pending`, go to step 2.
6. If `status` is `succeeded`, `failed`, or `canceled`, the task is done.

## Single-task constraint

Only one OnRamp task can run at a time. If you attempt to start a new action while another task is still running, the server responds with `409 Conflict`:

```json
{
  "status": 409,
  "title": "Conflict",
  "detail": "a task is already running"
}
```

This constraint exists because OnRamp deployment actions modify shared cluster state -- running two installations concurrently could leave the system in an inconsistent state. Wait for the current task to finish before starting another.

## Task fields

Each task includes the following fields:

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | UUID assigned when the task is created |
| `component` | string | The component being acted on (e.g., `5gc`, `k8s`) |
| `action` | string | The action being performed (e.g., `install`, `uninstall`) |
| `target` | string | The Make target being executed (e.g., `aether-5gc-install`) |
| `status` | string | Current status: `pending`, `running`, `succeeded`, `failed`, or `canceled` |
| `started_at` | string | RFC 3339 timestamp when the task started |
| `finished_at` | string | RFC 3339 timestamp when the task completed (absent while running) |
| `exit_code` | int | Process exit code (absent while running; `-1` if the process could not start) |
| `output` | string | Combined stdout/stderr, or an incremental chunk when `offset` is used |
| `output_offset` | int | Byte position for the next incremental read |

## Tasks vs. action history

Tasks and action history serve different purposes:

| | Tasks | Action history |
|---|-------|----------------|
| **Storage** | In-memory | Persistent in the database |
| **Lifetime** | Exist while the server is running; lost on restart | Survive server restarts |
| **Purpose** | Track live execution progress and output | Provide a permanent audit log of all actions |
| **Content** | Full output stream with incremental reads | Timestamps, exit codes, labels, and tags (no output) |

When a task completes, its result is recorded in the action history automatically. The task itself remains accessible for the duration of the server process, but it is not persisted across restarts.

For information about how completed actions affect component state, see [Deployment State](deployment-state).

For the full task and action API reference, see [API Reference: OnRamp](../reference/api-onramp).
