---
sidebar_position: 3
title: First Deployment
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# First Deployment

This page walks through deploying a Kubernetes cluster and the 5G Core network on a single node using the Aether WebUI. By the end you will have both components installed and running.

## Step 1: Run preflight checks

Before deploying anything, run the preflight checks to verify that the host meets all prerequisites. These checks verify that required tools are installed, SSH password authentication is enabled, and the `aether` service user exists with the correct permissions.

If you already ran the preflight setup script during [Installation](installation#prepare-the-host), all fixable checks should already pass. You can skip ahead to verifying the results or proceed directly to Step 2.

<Tabs>
  <TabItem value="script" label="Setup Script" default>

The preflight setup script fixes all three fixable checks in one pass. If you did not run it during installation, run it now:

```bash
curl -fsSL https://raw.githubusercontent.com/bengrewell/aether-webui/main/scripts/preflight-setup.sh | sudo bash
```

The script installs required packages (`make`, `ansible`), enables SSH password authentication, and creates the `aether` user with passwordless sudo. It is idempotent — running it again skips anything already configured.

After running the script, verify all checks pass via the API or Web UI:

```bash
curl http://localhost:8186/api/v1/preflight
```

  </TabItem>
  <TabItem value="ui" label="Web UI">

Open the Web UI. The **Preflight** page shows a checklist of system prerequisites. Each check displays a pass/fail status and, where applicable, a **Fix** button to automatically resolve the issue.

Review all checks. If any required checks are failing, click **Fix** to apply the automated fix, then re-run the checks to confirm they pass.

  </TabItem>
  <TabItem value="api" label="API">

```bash
curl http://localhost:8186/api/v1/preflight
```

The response includes a summary and individual results for each check:

```json
{
  "passed": 2,
  "failed": 2,
  "total": 4,
  "results": [
    {
      "id": "required-packages",
      "name": "Required Packages",
      "severity": "required",
      "passed": false,
      "message": "missing required packages: make, ansible-playbook",
      "can_fix": true,
      "fix_warning": "This will install system packages using the detected package manager (apt-get, dnf, or yum)."
    },
    ...
  ]
}
```

For any failed check that has `"can_fix": true`, apply the fix:

```bash
curl -X POST http://localhost:8186/api/v1/preflight/required-packages/fix
```

After fixing, re-run the checks to confirm everything passes:

```bash
curl http://localhost:8186/api/v1/preflight
```

  </TabItem>
</Tabs>

**Preflight checks:**

| Check | What it verifies | Auto-fix |
|-------|------------------|----------|
| Required Packages | `make` and `ansible-playbook` are installed | Yes -- installs via apt-get, dnf, or yum |
| SSH Configuration | SSH password authentication is enabled | Yes -- writes sshd drop-in config and restarts sshd |
| Aether User | `aether` user exists with passwordless sudo | Yes -- creates user with sudo access |
| Node SSH Reachability | All managed nodes are reachable on port 22 | No |

All required checks must pass before proceeding. The node reachability check is informational and only relevant for multi-node deployments.

## Step 2: Check the OnRamp repository status

Verify that the Aether OnRamp repository has been cloned and is ready:

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

## Step 3: Deploy Kubernetes

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

## Step 4: Poll the task for output

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

## Step 5: Deploy the 5G Core

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

## Step 6: Poll for completion

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
