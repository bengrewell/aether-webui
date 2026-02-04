# Project Status

 > Note: This document is a work in progress and has a lot of unfinished sections or text that is not up to date. Once
 > things like terms and architecture has been finalized this document will be updated to reflect that.

## Recent Updates
<!-- Rolling log of significant changes, most recent first -->
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
System information and metrics collection.

- [ ] **Static Information**
  - [ ] `GetCPUInfo` — CPU model, cores, cache
  - [ ] `GetMemoryInfo` — Total RAM, type, speed
  - [ ] `GetDiskInfo` — Drives, partitions, filesystem
  - [ ] `GetNICInfo` — Network interfaces, MAC, speed
  - [ ] `GetOSInfo` — Distro, kernel, hostname
- [ ] **Dynamic Metrics**
  - [ ] `GetCPUUsage` — Per-core utilization
  - [ ] `GetMemoryUsage` — Used/free/cached
  - [ ] `GetDiskUsage` — I/O stats, space used
  - [ ] `GetNICUsage` — Bandwidth, packets, errors

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

#### Exec Operator (`internal/operator/exec`)
Command execution and task management.

- [ ] **Core Execution**
  - [ ] `Execute` — Synchronous command execution
  - [ ] `ExecuteAsync` — Async execution with task ID
  - [ ] `GetTaskStatus` — Poll async task status
  - [ ] `CancelTask` — Abort running task
- [ ] **Invocable Operations**
  - [ ] `Shell` — Execute shell commands
  - [ ] `Ansible` — Run Ansible playbooks
  - [ ] `Script` — Execute script files
  - [ ] `Helm` — Run Helm commands
  - [ ] `Kubectl` — Run kubectl commands
  - [ ] `Docker` — Run Docker commands
  - [ ] `TaskStatusOp` — Query task status
  - [ ] `ListTasks` — List all tasks

### Frontend

#### Authentication (`web/frontend/src/auth`)
- [ ] `login` — User authentication flow
- [ ] `logout` — Session termination

---
