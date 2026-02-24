# Meta Provider

The meta provider exposes server introspection and diagnostic endpoints. It is always registered and cannot be disabled — other providers depend on it for the provider listing.

## Endpoints

| Method | Path | Operation ID | Description |
|--------|------|--------------|-------------|
| `GET` | `/api/v1/meta/version` | `meta-version` | Application version, branch, and commit hash |
| `GET` | `/api/v1/meta/build` | `meta-build` | Go toolchain version, target OS, and architecture |
| `GET` | `/api/v1/meta/runtime` | `meta-runtime` | PID, running user/group, binary path, start time, and uptime |
| `GET` | `/api/v1/meta/config` | `meta-config` | Active non-secret configuration: listen address, storage paths, feature flags, schema version |
| `GET` | `/api/v1/meta/providers` | `meta-providers` | Registered providers with enabled/running state and endpoint counts |
| `GET` | `/api/v1/meta/store` | `meta-store` | Store engine, path, file size, schema version, and live diagnostics (ping, write, read, delete) |

### `/meta/version`

Returns the version string injected at build time via ldflags, along with the git branch and short commit hash. Useful for verifying which release is deployed.

### `/meta/build`

Returns compile-time information: Go version, `GOOS`, `GOARCH`, and compiler. Helps diagnose platform-specific issues.

### `/meta/runtime`

Returns process-level details: PID, the OS user and group the process runs as, the binary's filesystem path, the server start time, and current uptime as a human-readable duration.

### `/meta/config`

Returns the active server configuration. Secrets (API tokens, encryption keys) are redacted. Includes listen address, TLS/mTLS state, frontend serving mode, data directory, metrics interval/retention, and store schema version.

### `/meta/providers`

Lists every registered provider with its name, enabled flag, running flag, and endpoint count. The meta provider uses this to give operators a single view of what the server is capable of.

### `/meta/store`

Returns the store backend type, database file path, file size on disk, and schema version. Also runs a live diagnostic cycle — ping, write, read, delete — and reports per-step latency and any errors. Useful for confirming the database is healthy.
