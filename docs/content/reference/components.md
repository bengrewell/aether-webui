---
sidebar_position: 8
title: "Components Reference"
---

# Components Reference

Aether WebUI manages 12 deployable components through the OnRamp provider. Each component maps to one or more Make targets in the aether-onramp repository.

## Component Table

| Component | Description | Actions | Make Targets |
|-----------|-------------|---------|-------------|
| `k8s` | Kubernetes (RKE2) cluster lifecycle | `install`, `uninstall` | `aether-k8s-install`, `aether-k8s-uninstall` |
| `5gc` | 5G core network (SD-Core) | `install`, `uninstall`, `reset` | `aether-5gc-install`, `aether-5gc-uninstall`, `aether-5gc-reset` |
| `4gc` | 4G core network | `install`, `uninstall`, `reset` | `aether-4gc-install`, `aether-4gc-uninstall`, `aether-4gc-reset` |
| `gnbsim` | gNBSim simulated RAN | `install`, `uninstall`, `run` | `aether-gnbsim-install`, `aether-gnbsim-uninstall`, `aether-gnbsim-run` |
| `amp` | Aether Management Platform | `install`, `uninstall` | `aether-amp-install`, `aether-amp-uninstall` |
| `sdran` | SD-RAN intelligent RAN controller | `install`, `uninstall` | `aether-sdran-install`, `aether-sdran-uninstall` |
| `ueransim` | UERANSIM UE and gNB simulator | `install`, `uninstall`, `run`, `stop` | `aether-ueransim-install`, `aether-ueransim-uninstall`, `aether-ueransim-run`, `aether-ueransim-stop` |
| `oai` | OpenAirInterface RAN | `gnb-install`, `gnb-uninstall`, `uesim-start`, `uesim-stop` | `aether-oai-gnb-install`, `aether-oai-gnb-uninstall`, `aether-oai-uesim-start`, `aether-oai-uesim-stop` |
| `srsran` | srsRAN Project RAN | `gnb-install`, `gnb-uninstall`, `uesim-start`, `uesim-stop` | `aether-srsran-gnb-install`, `aether-srsran-gnb-uninstall`, `aether-srsran-uesim-start`, `aether-srsran-uesim-stop` |
| `oscric` | O-RAN SC near-RT RIC | `ric-install`, `ric-uninstall` | `aether-oscric-ric-install`, `aether-oscric-ric-uninstall` |
| `n3iwf` | Non-3GPP Interworking Function | `install`, `uninstall` | `aether-n3iwf-install`, `aether-n3iwf-uninstall` |
| `cluster` | Cluster-level operations | `pingall`, `install`, `uninstall`, `add-upfs`, `remove-upfs` | `aether-pingall`, `aether-install`, `aether-uninstall`, `aether-add-upfs`, `aether-remove-upfs` |

## Detailed Action Reference

### k8s -- Kubernetes (RKE2)

| Action | Make Target | Description |
|--------|-------------|-------------|
| `install` | `aether-k8s-install` | Deploy RKE2 cluster across master and worker nodes |
| `uninstall` | `aether-k8s-uninstall` | Remove the RKE2 cluster |

### 5gc -- 5G Core (SD-Core)

| Action | Make Target | Description |
|--------|-------------|-------------|
| `install` | `aether-5gc-install` | Deploy SD-Core with Helm onto the K8s cluster |
| `uninstall` | `aether-5gc-uninstall` | Remove SD-Core |
| `reset` | `aether-5gc-reset` | Reset 5G core state (clear subscriber data) |

### 4gc -- 4G Core

| Action | Make Target | Description |
|--------|-------------|-------------|
| `install` | `aether-4gc-install` | Deploy 4G core |
| `uninstall` | `aether-4gc-uninstall` | Remove 4G core |
| `reset` | `aether-4gc-reset` | Reset 4G core state |

### gnbsim -- gNBSim

| Action | Make Target | Description |
|--------|-------------|-------------|
| `install` | `aether-gnbsim-install` | Deploy gNBSim containers |
| `uninstall` | `aether-gnbsim-uninstall` | Remove gNBSim containers |
| `run` | `aether-gnbsim-run` | Execute gNBSim simulation test |

### amp -- Aether Management Platform

| Action | Make Target | Description |
|--------|-------------|-------------|
| `install` | `aether-amp-install` | Deploy AMP (ROC, monitoring, dashboards) |
| `uninstall` | `aether-amp-uninstall` | Remove AMP |

### sdran -- SD-RAN

| Action | Make Target | Description |
|--------|-------------|-------------|
| `install` | `aether-sdran-install` | Deploy SD-RAN controller (ONOS, xApps) |
| `uninstall` | `aether-sdran-uninstall` | Remove SD-RAN |

### ueransim -- UERANSIM

| Action | Make Target | Description |
|--------|-------------|-------------|
| `install` | `aether-ueransim-install` | Deploy UERANSIM gNB and UE containers |
| `uninstall` | `aether-ueransim-uninstall` | Remove UERANSIM |
| `run` | `aether-ueransim-run` | Start UERANSIM simulation |
| `stop` | `aether-ueransim-stop` | Stop UERANSIM simulation |

### oai -- OpenAirInterface

| Action | Make Target | Description |
|--------|-------------|-------------|
| `gnb-install` | `aether-oai-gnb-install` | Deploy OAI gNB |
| `gnb-uninstall` | `aether-oai-gnb-uninstall` | Remove OAI gNB |
| `uesim-start` | `aether-oai-uesim-start` | Start OAI UE simulator |
| `uesim-stop` | `aether-oai-uesim-stop` | Stop OAI UE simulator |

### srsran -- srsRAN Project

| Action | Make Target | Description |
|--------|-------------|-------------|
| `gnb-install` | `aether-srsran-gnb-install` | Deploy srsRAN gNB |
| `gnb-uninstall` | `aether-srsran-gnb-uninstall` | Remove srsRAN gNB |
| `uesim-start` | `aether-srsran-uesim-start` | Start srsRAN UE simulator |
| `uesim-stop` | `aether-srsran-uesim-stop` | Stop srsRAN UE simulator |

### oscric -- O-RAN SC near-RT RIC

| Action | Make Target | Description |
|--------|-------------|-------------|
| `ric-install` | `aether-oscric-ric-install` | Deploy O-RAN SC near-RT RIC |
| `ric-uninstall` | `aether-oscric-ric-uninstall` | Remove O-RAN SC near-RT RIC |

### n3iwf -- Non-3GPP Interworking Function

| Action | Make Target | Description |
|--------|-------------|-------------|
| `install` | `aether-n3iwf-install` | Deploy N3IWF containers |
| `uninstall` | `aether-n3iwf-uninstall` | Remove N3IWF |

### cluster -- Cluster-Level Operations

| Action | Make Target | Description |
|--------|-------------|-------------|
| `pingall` | `aether-pingall` | Ping all cluster nodes to verify connectivity |
| `install` | `aether-install` | Deploy the full Aether stack (K8s + core + RAN) |
| `uninstall` | `aether-uninstall` | Remove the full Aether stack |
| `add-upfs` | `aether-add-upfs` | Add additional UPF instances to the deployment |
| `remove-upfs` | `aether-remove-upfs` | Remove additional UPF instances |

## Deployment State

Each component tracks its current deployment state. The state is updated automatically when install/uninstall actions are executed.

### State Values

| State | Description |
|-------|-------------|
| `not_installed` | No deployment exists (default for all components) |
| `installing` | An install action is currently running |
| `installed` | The most recent install action succeeded |
| `uninstalling` | An uninstall action is currently running |
| `failed` | The most recent install or uninstall action failed |

### State Transitions

```
                    install (start)
  not_installed ────────────────────> installing
       ^                                  │
       │                       ┌──────────┴──────────┐
       │                       │                      │
       │                   succeeded               failed
       │                       │                      │
       │                       v                      v
       │                  installed               failed
       │                       │                      │
       │            uninstall (start)     install (start)
       │                       │                      │
       │                       v                      v
       │                 uninstalling            installing
       │                       │                      │
       │            ┌──────────┴──────────┐    ┌──────┴──────┐
       │            │                      │    │              │
       │        succeeded               failed  succeeded   failed
       │            │                      │    │              │
       └────────────┘                      v    v              v
                                        failed  installed   failed
```

**Transition rules:**

- `install` (start): transitions to `installing` from any state
- `install` (succeeded): transitions to `installed`
- `install` (failed): transitions to `failed`
- `uninstall` (start): transitions to `uninstalling` from any state
- `uninstall` (succeeded): transitions to `not_installed`
- `uninstall` (failed): transitions to `failed`

Note: Actions that do not match install/uninstall patterns (e.g., `run`, `stop`, `reset`, `pingall`, `add-upfs`, `remove-upfs`) do **not** affect the component's deployment state.

### Querying State

**All components:**

```bash
curl http://localhost:8186/api/v1/onramp/state
```

**Single component:**

```bash
curl http://localhost:8186/api/v1/onramp/state/k8s
```

See the [OnRamp Endpoints](./api-onramp.md#component-state) reference for full response schemas.

## Executing Actions

Actions are executed via the [execute action endpoint](./api-onramp.md#execute-action):

```bash
curl -X POST http://localhost:8186/api/v1/onramp/components/{component}/{action}
```

All actions run asynchronously. The response includes a task object with an ID that can be polled for progress:

```bash
# Start an install
curl -X POST http://localhost:8186/api/v1/onramp/components/k8s/install

# Poll for progress
curl "http://localhost:8186/api/v1/onramp/tasks/{task-id}?offset=0"
```

Only one task can run at a time. Attempting to start a second task while one is in progress returns `409 Conflict`.

For a step-by-step deployment walkthrough, see [Deploying Components](../guides/deploying-components).
