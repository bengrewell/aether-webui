# meta provider

The `meta` provider exposes server introspection endpoints — version info,
build metadata, and similar diagnostics. It is always enabled and requires no
external dependencies.

## Architecture

`Meta` embeds `provider.Base` and implements the `provider.Provider` interface.
Endpoints are defined as transport-agnostic `endpoint.Endpoint[I, O]` values and
registered via `provider.Register`, which records the descriptor and optionally
binds the endpoint to a Huma API if one is configured.

## Endpoints

| Operation ID | Semantics | HTTP Path          | Description                              |
|--------------|-----------|--------------------|------------------------------------------|
| `version`    | Read      | `GET /api/v1/version` | Build version, date, branch, commit hash |

## Adding a new endpoint

1. Define input/output types (use `struct{}` for empty input).
2. Create an `endpoint.Endpoint[I, O]` with a `Descriptor` and handler func.
3. Call `provider.Register(m.Base, ep)` inside `NewProvider`.
