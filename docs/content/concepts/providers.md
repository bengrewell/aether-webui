---
sidebar_position: 2
title: Providers
---

# Providers

Providers are the modular building blocks of the Aether WebUI API. Each provider owns a set of related endpoints under a dedicated URL path prefix, and each focuses on a single domain of functionality.

## What providers are

A provider is a self-contained unit that:

- Registers one or more HTTP endpoints under a common path prefix
- Has its own scoped logger for traceable log output
- Has access to the shared store for persistence
- Tracks its own health status (enabled and running)

Providers are independent of each other. The **system** provider can collect metrics without the **onramp** provider being active, and the **meta** provider reports on all other providers without depending on their internals.

## Built-in providers

Aether WebUI ships with five providers:

### meta -- Server introspection

**Path prefix:** `/api/v1/meta/`
**Endpoints:** 6

The meta provider exposes information about the server itself. It is always enabled and cannot be disabled, because other parts of the system depend on it for the provider listing.

What it provides:

- **Version and build info** -- the release version, git commit, Go toolchain version, and target platform. Useful for verifying which release is deployed.
- **Runtime info** -- process ID, running user, binary path, server start time, and uptime.
- **Configuration** -- the active server configuration with secrets redacted. Shows listen address, TLS state, data directory, and feature flags.
- **Provider listing** -- every registered provider with its enabled/running status and endpoint count.
- **Store diagnostics** -- database engine, file size, schema version, and a live health check that measures read/write latency.

For the full endpoint reference, see [API Reference: Meta](../reference/api-meta).

### system -- Host system info and metrics

**Path prefix:** `/api/v1/system/`
**Endpoints:** 8

The system provider collects hardware and operating system information from the host machine. It serves two purposes: static inventory (what hardware is present) and time-series metrics (how the system is performing over time).

What it provides:

- **System overview** -- hostname, OS, kernel version, CPU model, core count, total memory, disk capacity.
- **CPU metrics** -- per-core and aggregate utilization over configurable time windows.
- **Memory metrics** -- used, available, cached, and swap usage over time.
- **Disk metrics** -- read/write throughput and IOPS per device.
- **Network metrics** -- bytes and packets in/out per interface.

Metrics are sampled at a configurable interval and stored in the database with configurable retention. See [Monitoring](../guides/monitoring) for usage details.

### nodes -- Cluster node inventory

**Path prefix:** `/api/v1/nodes`
**Endpoints:** 5

The nodes provider manages the inventory of machines that make up the cluster. Aether OnRamp can deploy across multiple nodes, and this provider tracks which nodes exist and what roles they serve.

What it provides:

- **CRUD operations** -- add, list, get, update, and remove nodes from the inventory.
- **Role assignments** -- each node can have one or more roles (e.g., control plane, worker) that determine what components are deployed to it.
- **Connectivity metadata** -- hostname, IP address, and SSH connection details for remote nodes.

For the full endpoint reference, see [API Reference: Nodes](../reference/api-nodes).

### onramp -- Deployment lifecycle

**Path prefix:** `/api/v1/onramp/`
**Endpoints:** 18

The onramp provider is the largest and most complex. It wraps the Aether OnRamp Make/Ansible toolchain and exposes it as a REST API. This is where component installation, task tracking, configuration management, and deployment state all live.

What it provides:

- **Repository management** -- checks the status of the OnRamp git repository (cloned, current version, dirty state) and can refresh it.
- **Component registry** -- lists all deployable components (k8s, 5gc, gnbsim, etc.) and their available actions (install, uninstall, etc.).
- **Task execution** -- triggers async deployment actions and tracks their progress. See [Tasks and Async Execution](tasks).
- **Configuration** -- reads and patches the `vars/main.yml` configuration file that controls OnRamp behavior.
- **Profiles** -- lists and activates named configuration profiles (e.g., `gnbsim`, `5g`) that swap in pre-built configuration sets.
- **Deployment state** -- tracks which components are installed, failed, or in progress. See [Deployment State](deployment-state).
- **Action history** -- a persistent log of every action ever executed, with timestamps, exit codes, and metadata.
- **Inventory** -- reads and writes the Ansible inventory file that defines target hosts.

For the full endpoint reference, see [API Reference: OnRamp](../reference/api-onramp).

### preflight -- Pre-deployment checks

**Path prefix:** `/api/v1/preflight`
**Endpoints:** 3

The preflight provider automates the verification of system prerequisites that a fresh Aether deployment requires. It runs checks against the local host and managed nodes, reporting which prerequisites are met and which are missing.

What it provides:

- **Prerequisite checks** -- verifies that required tools (make, ansible), system configuration (SSH password auth, aether user with sudo), and network connectivity (SSH to managed nodes) are in place.
- **Aggregate status** -- runs all checks in parallel and returns a summary with pass/fail counts, enabling a UI checklist view.
- **Automated fixes** -- some checks offer a one-click fix (e.g., enabling SSH password auth, creating the aether user). Fixes include security warnings so the operator can make an informed decision.

Each check has a severity level (`required`, `warning`, `info`) and category (`tooling`, `access`, `network`) to help the UI prioritize what to display.

For the full endpoint reference, see [API Reference: Preflight](../reference/api-preflight).

## Provider health and status

Every provider has two status flags:

| Flag | Meaning |
|------|---------|
| **enabled** | The provider was registered at startup and is configured to run. |
| **running** | The provider initialized successfully and is actively serving requests. |

A provider can be enabled but not running if it encountered an error during initialization. For example, the onramp provider starts in a degraded state if the OnRamp git repository is missing or corrupt -- it still responds to requests, but operations that need the repository return errors until it is repaired.

The meta provider reports the status of all providers at `GET /api/v1/meta/providers`, giving operators a single view of what the server is capable of at any moment.

## How providers are organized

Each provider is isolated behind its own path prefix, so there is no ambiguity about which provider handles a given request:

```
/api/v1/meta/...       →  meta provider
/api/v1/system/...     →  system provider
/api/v1/nodes/...      →  nodes provider
/api/v1/onramp/...     →  onramp provider
/api/v1/preflight/...  →  preflight provider
```

This isolation means providers can be developed, tested, and reasoned about independently. A bug in the system metrics collection does not affect the onramp deployment workflow.

## Extensibility

The provider framework is designed to be extensible. New providers can be added to the system by implementing the provider interface and registering with the controller. Each new provider gets the same infrastructure -- scoped logging, store access, automatic OpenAPI documentation -- without modifying existing providers.
