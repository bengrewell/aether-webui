# Project Status

 > Note: This document is a work in progress and has a lot of unfinished sections or text that is not up to date. Once
 > things like terms and architecture has been finalized this document will be updated to reflect that.

## Recent Updates
<!-- Rolling log of significant changes, most recent first -->
- **Feb 4, 2026**: Refactored exec operator into execution shim (`internal/executor/`) - operators like Aether now receive Executor via dependency injection
- **Feb 4, 2025**: Implemented Host Operator with gopsutil - all static info and dynamic metrics now functional, plus historical metrics with time-series aggregation
- **Week of Feb 3, 2025**: Restructured progress tracking for stakeholder reporting


---

```
Aether WebUI (aether-webui)
├─ Web Console (Frontend)
│  ├─ Application Shell
│  │  ├─ Navigation (tabs/pages)
│  │  ├─ Layout (responsive)
│  │  └─ Theming (light/dark, auto/manual)
│  ├─ Pages / Views
│  │  ├─ Dashboard
│  │  ├─ Nodes
│  │  ├─ Deployments
│  │  ├─ Jobs / Activity
│  │  ├─ Settings
│  │  └─ System / About
│  └─ Client Data Layer
│     ├─ Request/Command Client (REST)
│     ├─ Streaming Client (WebSockets)
│     └─ State & Caching (view-models/hooks)
│
├─ Control Plane (Backend)
│  ├─ API Gateway
│  │  ├─ REST API (routing/handlers)
│  │  ├─ OpenAPI / Swagger (docs)
│  │  └─ API Logging & Request Tracing
│  ├─ Security & Access Layer
│  │  ├─ TLS / mTLS (transport security)
│  │  ├─ Authentication (future-plumbed)
│  │  ├─ Authorization / Policy (future-plumbed)
│  │  └─ Audit Events (who/what/when/result)
│  ├─ Execution Engine
│  │  ├─ Workflow / Plan Runner (Model A)
│  │  ├─ Core Operations (atomic primitives)
│  │  ├─ Domain Workflows (Aether-specific)
│  │  └─ Providers
│  │     ├─ Direct Provider (no shell default)
│  │     ├─ Shell Provider (explicit opt-in)
│  │     ├─ SSH Provider
│  │     ├─ Container Provider (future)
│  │     ├─ Pod Provider (future)
│  │     └─ VM Provider (future)
│  └─ Platform Integrations
│     ├─ Kubernetes Health / Inventory
│     ├─ Deployment Health (core/gNB/etc.)
│     └─ Host Health & Metrics (nodes)
│
└─ Supporting Systems
├─ Installation & Lifecycle Management
│  ├─ Install Script (curl/wget entrypoint)
│  ├─ Uninstall Script
│  └─ Upgrade / Migration (future)
├─ System Services
│  ├─ systemd Units
│  └─ Service Configuration
└─ Validation & Test Framework
├─ E2E Test Definitions
└─ Automated Validation Runs
```

---

## Implementation Progress

Tracks completion of stubbed code that currently returns `ErrNotImplemented`.

### Backend Operators

#### Host Operator (`internal/operator/host`)
System information and metrics collection via gopsutil v4.

- [x] **Static Information** (5-minute in-memory cache)
  - [x] `GetCPUInfo` — CPU model, cores, threads, frequency, cache
  - [x] `GetMemoryInfo` — Total RAM
  - [x] `GetDiskInfo` — Partitions, mount points, disk type detection
  - [x] `GetNICInfo` — Network interfaces, MAC, MTU, IP addresses
  - [x] `GetOSInfo` — Platform, kernel, hostname, uptime, architecture
- [x] **Dynamic Metrics** (real-time)
  - [x] `GetCPUUsage` — Per-core utilization, user/system/idle %, load averages
  - [x] `GetMemoryUsage` — Used/free/available/cached, swap, usage %
  - [x] `GetDiskUsage` — Per-mount usage stats, inodes
  - [x] `GetNICUsage` — Bytes/packets in/out, errors, drops, rates
- [x] **Historical Metrics** (time-series)
  - [x] Background collector with configurable interval/retention
  - [x] SQLite storage with automatic pruning
  - [x] Time-series aggregation (window + granularity)
  - [x] `/history` API endpoints for all metric types

#### Kubernetes Operator (`internal/operator/kube`)
Cluster monitoring and workload visibility.

- [ ] `GetClusterHealth` — Overall cluster status
- [ ] `GetNodes` — Node list with status/resources
- [ ] `GetNamespaces` — Namespace enumeration
- [ ] `GetEvents` — Cluster events (filtered)
- [ ] `GetPods` — Pod list by namespace
- [ ] `GetDeployments` — Deployment list by namespace
- [ ] `GetServices` — Service list by namespace

#### Aether Operator (`internal/operator/aether`)
5G SD-Core and gNB lifecycle management.

- [ ] **SD-Core Management**
  - [ ] `ListCores` — Enumerate core deployments
  - [ ] `GetCore` — Fetch core configuration
  - [ ] `DeployCore` — Deploy new SD-Core
  - [ ] `UpdateCore` — Modify core configuration
  - [ ] `UndeployCore` — Remove core deployment
  - [ ] `GetCoreStatus` — Single core health
  - [ ] `ListCoreStatuses` — All cores health
- [ ] **gNB Management**
  - [ ] `ListGNBs` — Enumerate gNB configs
  - [ ] `GetGNB` — Fetch gNB configuration
  - [ ] `DeployGNB` — Deploy new gNB
  - [ ] `UpdateGNB` — Modify gNB configuration
  - [ ] `UndeployGNB` — Remove gNB deployment
  - [ ] `GetGNBStatus` — Single gNB health
  - [ ] `ListGNBStatuses` — All gNBs health

#### Executor Shim (`internal/executor/`)
Command execution infrastructure (not an operator). Operators use this via dependency injection.

- [x] **Core Interface** — Executor interface with domain-specific methods
- [x] **File Operations** — ReadFile, WriteFile, RenderTemplate, FileExists, MkdirAll
- [x] **Helm Operations** — RunHelmInstall, RunHelmUpgrade, RunHelmUninstall, RunHelmList, RunHelmStatus
- [x] **Kubectl Operations** — RunKubectl, KubectlApply, KubectlDelete, KubectlGet
- [x] **Docker Operations** — RunDockerCommand, DockerRun, DockerStop, DockerRemove
- [x] **Ansible Operations** — RunAnsiblePlaybook
- [x] **Shell/Script** — RunShell, RunScript (explicit opt-in)
- [x] **MockExecutor** — Test double with call tracking for unit testing operators

### Frontend

#### Authentication (`web/frontend/src/auth`)
- [ ] `login` — User authentication flow
- [ ] `logout` — Session termination

---
