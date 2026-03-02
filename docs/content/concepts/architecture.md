---
sidebar_position: 1
title: Architecture
---

# Architecture

Aether WebUI is a REST API service that sits between operators and the Aether OnRamp toolchain. Rather than running Make and Ansible commands directly on the host, operators interact with a JSON API that manages deployments, tracks state, and collects metrics.

This page describes the major layers of the system and how they fit together.

## System layers

The service is organized into four layers, each with a distinct responsibility:

```
┌─────────────────────────────────────────────────────┐
│                     Controller                       │
│  Startup sequencing, shutdown, lifecycle management  │
├─────────────────────────────────────────────────────┤
│                   REST Transport                     │
│        Chi router, middleware, Huma API framework    │
├──────────┬──────────┬──────────┬────────────────────┤
│   meta   │  system  │  nodes   │      onramp        │
│ provider │ provider │ provider │     provider        │
├──────────┴──────────┴──────────┴────────────────────┤
│                    Store Layer                        │
│         SQLite database, migrations, encryption      │
└─────────────────────────────────────────────────────┘
```

### Controller

The controller is the central orchestrator. It owns the server process and coordinates everything else. At startup, the controller:

1. Initializes structured logging
2. Configures TLS if enabled
3. Opens the SQLite store and runs any pending schema migrations
4. Assembles the middleware chain (request logging and optional token authentication)
5. Creates the REST transport (router and API framework)
6. Initializes each registered provider
7. Mounts the frontend (either embedded in the binary or served from a directory)
8. Starts the HTTP or HTTPS server

On shutdown, the controller drains active connections and closes the store cleanly. It handles OS signals (Ctrl+C) and supports graceful termination so that in-flight requests complete before the process exits.

### Provider framework

Providers are modular units that each own a set of related API endpoints. The service ships with four built-in providers:

| Provider | Responsibility |
|----------|---------------|
| **meta** | Server introspection and diagnostics |
| **system** | Host hardware info and time-series metrics |
| **nodes** | Cluster node inventory management |
| **onramp** | Deployment lifecycle -- components, tasks, configuration, state |

Each provider operates independently under its own URL path prefix. The framework gives every provider access to the store and a scoped logger, but providers do not depend on each other. See [Providers](providers) for details on each one.

### REST transport

The REST transport layer handles HTTP routing, content negotiation, request validation, and response serialization. It is built on [Chi](https://github.com/go-chi/chi) (a lightweight HTTP router) and [Huma](https://huma.rocks/) (an API framework that generates OpenAPI documentation automatically).

Two middleware layers process every request before it reaches a provider handler:

1. **Request logging** -- logs the HTTP method, path, response status, and duration for every request.
2. **Token authentication** -- validates the `Authorization: Bearer <token>` header on all `/api/*` paths. Requests to non-API paths (the frontend, health checks, OpenAPI spec) bypass authentication.

### Store layer

The store provides persistent storage backed by SQLite. It manages several categories of data:

- **Node inventory** -- the set of machines in the cluster, with roles and connectivity information
- **Deployment state** -- which components are currently installed, failed, or in progress
- **Action history** -- a persistent log of every deployment action, with timestamps, exit codes, and metadata
- **Time-series metrics** -- sampled CPU, memory, disk, and network measurements
- **Configuration objects** -- generic JSON key-value storage for settings and internal state
- **Credentials** -- encrypted storage for sensitive values like SSH keys

The store applies versioned schema migrations on startup, so the database schema evolves automatically as the service is upgraded.

## Request flow

Every API request follows the same path through the system:

```
Client
  │
  ▼
Controller (HTTP/HTTPS listener)
  │
  ▼
Chi Router (path matching)
  │
  ▼
Request Logging Middleware (logs method, path, status, duration)
  │
  ▼
Token Auth Middleware (validates bearer token on /api/* paths)
  │
  ▼
Huma API Framework (content negotiation, input validation)
  │
  ▼
Provider Handler (business logic)
  │
  ▼
Store (SQLite read/write)
  │
  ▼
Response (JSON back to client)
```

For async operations like component deployments, the provider handler creates a background task and returns immediately with a task ID. The client then polls for progress. See [Tasks and Async Execution](tasks) for details.

## Security layers

The service supports multiple layers of security that can be enabled independently or combined:

| Layer | What it protects | How it works |
|-------|-----------------|--------------|
| **TLS** | Data in transit | Encrypts all traffic between client and server. Can use auto-generated or user-provided certificates. |
| **Mutual TLS (mTLS)** | Client identity | Requires clients to present a certificate signed by a trusted CA. Unauthenticated clients are rejected at the TLS layer before any HTTP processing. |
| **Token authentication** | API access | Requires a bearer token in the `Authorization` header for all `/api/*` requests. Constant-time comparison prevents timing attacks. |

These layers are applied in order during the request lifecycle. TLS is negotiated first (at the connection level), then token authentication is checked (at the HTTP level). A production deployment typically enables all three.

For configuration details, see the [Security guide](../guides/security).

## What the service does not do

Understanding the boundaries helps set expectations:

- **No direct network configuration.** The service delegates to Aether OnRamp's Make targets and Ansible playbooks for all infrastructure changes.
- **No multi-user access control.** Authentication is a single shared token, not a user/role system.
- **No clustering.** The service runs as a single instance on one host. The SQLite store is local to that host.
- **No real-time streaming.** Task output is retrieved via polling with byte offsets, not via WebSockets or server-sent events.
