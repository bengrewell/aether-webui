---
sidebar_position: 4
title: Deployment State
---

# Deployment State

Aether WebUI tracks the installation state of every deployable component. This state tells you at a glance which components are installed, which are in progress, and which have failed -- without needing to inspect Kubernetes or Ansible directly.

## What deployment state tracks

Each component in the system has a current state value that reflects whether it is installed on the cluster. The service updates this state automatically as deployment actions complete, so it always reflects the most recent outcome.

The 12 tracked components are:

| Component | Description |
|-----------|-------------|
| `k8s` | Kubernetes cluster (RKE2) |
| `5gc` | 5G Core network (SD-Core) |
| `4gc` | 4G Core network |
| `gnbsim` | gNBSim simulated RAN |
| `amp` | Aether Management Platform |
| `sdran` | SD-RAN intelligent RAN controller |
| `ueransim` | UERANSIM UE and gNB simulator |
| `oai` | OpenAirInterface RAN |
| `srsran` | srsRAN Project RAN |
| `oscric` | O-RAN SC near-RT RIC |
| `n3iwf` | Non-3GPP Interworking Function |
| `cluster` | Cluster-level operations |

## State values

A component is always in one of five states:

| State | Meaning |
|-------|---------|
| `not_installed` | The component has never been installed, or was successfully uninstalled. This is the default state. |
| `installing` | An install action is currently running. |
| `installed` | The most recent install action succeeded. |
| `uninstalling` | An uninstall action is currently running. |
| `failed` | The most recent install or uninstall action failed. |

## State transitions

State changes happen automatically when deployment actions are triggered and when they complete. The following diagram shows all valid transitions:

```
                    install triggered
   not_installed ─────────────────────────► installing
        ▲                                     │
        │                              ┌──────┴──────┐
        │                              │             │
        │                           success       failure
        │                              │             │
        │                              ▼             ▼
        │                          installed       failed
        │                              │           │   │
        │                  uninstall   │           │   │
        │                  triggered   │           │   │
        │                              ▼           │   │
        │                         uninstalling ◄───┘   │
        │                              │       uninstall│
        │                       ┌──────┴──────┐triggered│
        │                       │             │         │
        │                    success       failure      │
        │                       │             │         │
        └───────────────────────┘             ▼         │
                                           failed ◄─────┘
                                             │  install
                                             │  triggered
                                             └──────────► installing
```

Written as transitions:

| From | Trigger | To |
|------|---------|-----|
| `not_installed` | Install triggered | `installing` |
| `installing` | Action succeeds | `installed` |
| `installing` | Action fails | `failed` |
| `installed` | Uninstall triggered | `uninstalling` |
| `uninstalling` | Action succeeds | `not_installed` |
| `uninstalling` | Action fails | `failed` |
| `failed` | Install triggered | `installing` |
| `failed` | Uninstall triggered | `uninstalling` |

Note that `failed` is not a terminal state. You can retry an install or trigger an uninstall from the failed state to attempt recovery.

## How state updates work

State updates are automatic. When you trigger an action via the API:

1. **Action triggered** -- You `POST /api/v1/onramp/components/5gc/install`. The component state immediately changes to `installing`.
2. **Task runs** -- The Make target executes in the background. The state remains `installing` for the duration.
3. **Task completes** -- When the task finishes, the service examines the exit code. If the exit code is 0, the state moves to `installed`. If non-zero, it moves to `failed`.

The same pattern applies to uninstall actions: the state moves to `uninstalling` when triggered, then to `not_installed` on success or `failed` on failure.

You never need to update component state manually. It is derived entirely from the outcomes of deployment actions.

## Relationship between tasks, actions, and state

Three related concepts work together to give a complete picture of deployments:

```
┌─────────────────────────────────────────────────────────┐
│                    Deployment Action                      │
│                                                          │
│  ┌──────────┐    completes    ┌────────────────────┐     │
│  │   Task   │ ──────────────► │  Action History    │     │
│  │ (live)   │                 │  (persistent log)  │     │
│  └──────────┘                 └────────────────────┘     │
│       │                              │                   │
│       │ updates on                   │ derives           │
│       │ trigger + completion         │ current status    │
│       ▼                              ▼                   │
│              ┌──────────────────┐                        │
│              │ Component State   │                        │
│              │ (current status)  │                        │
│              └──────────────────┘                        │
└─────────────────────────────────────────────────────────┘
```

| Concept | What it is | Lifetime | Contains |
|---------|-----------|----------|----------|
| **Task** | A live execution of a Make target | In-memory; exists while the server runs | Real-time output, status, exit code |
| **Action history** | A persistent record of a completed action | Stored in the database; survives restarts | Timestamps, exit code, component, action, labels, tags |
| **Component state** | The current installation status of a component | Stored in the database; survives restarts | One of the five state values |

When a task completes:
- The result is recorded as an entry in the action history.
- The component state is updated based on the outcome.

This means you can always reconstruct how a component reached its current state by examining the action history.

## Querying state

### All components

```
GET /api/v1/onramp/state
```

Returns the state of every tracked component:

```json
[
  { "component": "k8s", "state": "installed" },
  { "component": "5gc", "state": "installed" },
  { "component": "gnbsim", "state": "not_installed" },
  ...
]
```

### Single component

```
GET /api/v1/onramp/state/5gc
```

Returns the state for a specific component:

```json
{
  "component": "5gc",
  "state": "installed"
}
```

## Querying action history

The action history endpoint provides a filterable log of all past actions:

```
GET /api/v1/onramp/actions
```

Filters are available for narrowing results by component, action type, status, and time range. Each history entry includes:

| Field | Description |
|-------|-------------|
| `component` | The component acted on |
| `action` | The action performed (e.g., `install`, `uninstall`) |
| `target` | The Make target that was executed |
| `status` | The outcome: `succeeded` or `failed` |
| `exit_code` | The process exit code |
| `started_at` | When the action began |
| `finished_at` | When the action completed |
| `labels` | Key-value metadata attached to the action |
| `tags` | Freeform tags for categorization |

Action history is append-only. Records are never modified or deleted, providing a complete audit trail of every deployment operation.

For the full API reference, see [API Reference: OnRamp](../reference/api-onramp). For a practical walkthrough of deploying components and observing state changes, see [Deploying Components](../guides/deploying-components).
