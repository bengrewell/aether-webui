# Architecture Overview

## Project Structure

```
cmd/aether-webd/       Entry point — flag parsing, controller configuration
internal/
  controller/          Server lifecycle orchestration (logging, TLS, store, providers, shutdown)
  api/rest/            REST transport (Chi router + Huma API)
  auth/                Authentication middleware (bearer token)
  frontend/            Embedded SPA frontend serving
  logging/             Structured logging (tint) + request middleware
  provider/            Provider framework (base, registration, options)
    meta/              Introspection provider (version, config, health)
    onramp/            Aether OnRamp provider (repo, components, tasks, config, profiles)
    system/            System metrics provider (CPU, memory, disk, NIC)
  security/            TLS/mTLS configuration and self-signed cert generation
  store/               Persistence layer (SQLite)
deploy/                Deployment configurations
web/                   Frontend source and Caddy configs
docs/                  This documentation
```

## Controller

The controller (`internal/controller/`) is the central orchestrator. `main.go` parses flags and configures a `Controller` via functional options, then calls `Run()` which manages the full server lifecycle:

1. Initialize structured logging
2. Resolve API token (flag → env fallback)
3. Build TLS configuration (if enabled)
4. Open the SQLite store
5. Assemble middleware chain (logging + optional token auth)
6. Create the REST transport
7. Initialize registered providers and the meta provider
8. Mount the frontend (embedded or directory)
9. Start the HTTP/HTTPS server
10. Await shutdown (context cancellation or double-Ctrl+C signal pattern)

Providers are registered with the controller via `WithProvider(name, enabled, factory)`. Each factory receives the store client and pre-wired provider options (Huma API, scoped logger, store) from all active transports.

## Provider Framework

Providers are modular units that register HTTP endpoints. Each provider embeds a `provider.Base` which handles:

- Enabled/running state management
- Scoped logger and store client injection
- Endpoint descriptor tracking for introspection

### Registration

Endpoints are registered via the generic `provider.Register[I, O]()` function, which maps semantic operations to HTTP methods:

| Operation | HTTP Method |
|-----------|-------------|
| Read      | GET         |
| Create    | POST        |
| Update    | PUT         |
| Delete    | DELETE      |

```go
provider.Register(p.Base, endpoint.Endpoint[struct{}, VersionOutput]{
    Desc: endpoint.Descriptor{
        OperationID: "meta-version",
        Semantics:   endpoint.Read,
        Summary:     "Get version info",
        Tags:        []string{"meta"},
        HTTP:        endpoint.HTTPHint{Path: "/api/v1/meta/version"},
    },
    Handler: p.handleVersion,
})
```

### Built-in Providers

| Provider | Package | Documentation | Description |
|----------|---------|---------------|-------------|
| `meta` | `internal/provider/meta` | [providers/meta.md](providers/meta.md) | Server introspection — version, build, runtime, config, provider list, store diagnostics |
| `onramp` | `internal/provider/onramp` | [providers/onramp.md](providers/onramp.md) | Aether OnRamp lifecycle — repo management, component deployment, task tracking, config/profile editing |
| `system` | `internal/provider/system` | [providers/system.md](providers/system.md) | Host system metrics — CPU, memory, disk, NIC sampling |

### Adding a Provider

1. Create a package under `internal/provider/<name>/`
2. Define a struct embedding `provider.Base`
3. Register endpoints in the constructor using `provider.Register`
4. Write a `ProviderFactory` and register it with `controller.WithProvider(name, enabled, factory)`

## REST Transport

`rest.Transport` owns the Chi router and Huma API instance. It applies middleware in order during construction:

1. **Request logging** — logs method, path, status, duration for every request
2. **Token authentication** — validates `Authorization: Bearer <token>` on `/api/*` paths (when enabled)

The transport exposes `ProviderOpts(name)` to give each provider pre-wired access to the Huma API, a scoped logger, and the store client.

## Store Layer

The store provides a generic key-value interface backed by SQLite:

- **Generic objects**: `Save[T]()`, `Load[T]()`, `Delete()`, `List()` — JSON-serialized with namespace/ID keys
- **Credentials**: typed CRUD for encrypted credential storage
- **Metrics**: time-series sample append and range queries
- **Schema migrations**: versioned DDL applied on startup

## Request Lifecycle

```
Client Request
  → Controller (owns server, transports, providers)
    → Chi Router
      → Logging Middleware (logs all requests)
      → Token Auth Middleware (rejects unauthorized /api/* requests)
      → Huma API (content negotiation, validation)
        → Provider Handler (business logic)
          → Store (SQLite read/write)
        ← Response
      ← JSON response with status
    ← HTTP response to client
```
