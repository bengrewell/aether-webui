# meta provider

The `meta` provider exposes server introspection endpoints — version info,
build metadata, runtime identity, active configuration, and registered provider
statuses. It is always enabled and requires no external dependencies.

## Architecture

`Meta` embeds `provider.Base` and implements the `provider.Provider` interface.
Endpoints are defined as transport-agnostic `endpoint.Endpoint[I, O]` values and
registered via `provider.Register`, which records the descriptor and optionally
binds the endpoint to a Huma API if one is configured.

External data (schema version, provider list) is injected via callback functions
to avoid circular imports with the state and provider-registry packages.

## Endpoints

| Operation ID     | Semantics | HTTP Path                | Description                                        |
|------------------|-----------|--------------------------|----------------------------------------------------|
| `meta-version`   | Read      | `GET /api/v1/meta/version`   | Build version, date, branch, commit hash           |
| `meta-build`     | Read      | `GET /api/v1/meta/build`     | Go toolchain version, target OS, architecture      |
| `meta-runtime`   | Read      | `GET /api/v1/meta/runtime`   | PID, user/group, binary path, start time, uptime   |
| `meta-config`    | Read      | `GET /api/v1/meta/config`    | Non-secret config values and schema version        |
| `meta-providers` | Read      | `GET /api/v1/meta/providers` | Registered providers with enabled/running/endpoint count |

## Adding a new endpoint

1. Define input/output types in `types.go` (use `struct{}` for empty input).
2. Create an `endpoint.Endpoint[I, O]` with a `Descriptor` and handler func.
3. Call `provider.Register(m.Base, ep)` inside `NewProvider`.
