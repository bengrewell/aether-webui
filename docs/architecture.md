# Architecture Overview

## Project Structure

```
cmd/aether-webd/       Entry point — flag parsing, wiring, server startup
internal/
  api/rest/            REST transport (Chi router + Huma API)
  auth/                Authentication middleware (bearer token)
  frontend/            Embedded SPA frontend serving
  logging/             Structured logging (tint) + request middleware
  provider/            Provider framework (base, registration, options)
    meta/              Introspection provider (version, config, health)
  security/            TLS/mTLS configuration and self-signed cert generation
  store/               Persistence layer (SQLite)
deploy/                Deployment configurations
web/                   Frontend source and Caddy configs
docs/                  This documentation
```

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
provider.Register(p.Base, provider.Endpoint[*struct{}, *VersionOutput]{
    Descriptor: provider.Descriptor{
        OperationID: "meta-version",
        Path:        "/api/v1/meta/version",
        Operation:   provider.Read,
        Summary:     "Get version info",
    },
    Handler: p.handleVersion,
})
```

### Adding a Provider

1. Create a package under `internal/provider/<name>/`
2. Define a struct embedding `provider.Base`
3. Register endpoints in the constructor using `provider.Register`
4. Wire the provider in `main.go` using `transport.ProviderOpts("<name>")`

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
