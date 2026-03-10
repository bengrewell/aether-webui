---
sidebar_position: 1
title: "CLI Reference"
---

# CLI Reference

```
aether-webd [options]
```

`aether-webd` is the backend daemon for the Aether WebUI. It serves the REST API, embedded frontend, and manages all deployment operations.

## Flags

### General

| Flag | Env Var | Description | Default |
|------|---------|-------------|---------|
| `-v, --version` | - | Print version information and exit | `false` |
| `-d, --debug` | `AETHER_DEBUG` | Enable debug mode for verbose logging | `false` |
| `-l, --listen` | `AETHER_LISTEN` | Address and port to listen on | `127.0.0.1:8186` |

### Security

| Flag | Env Var | Description | Default |
|------|---------|-------------|---------|
| `--tls` | `AETHER_TLS` | Enable TLS (auto-generates a self-signed cert if `--tls-cert`/`--tls-key` are not provided) | `false` |
| `-t, --tls-cert` | `AETHER_TLS_CERT` | TLS certificate file for HTTPS | - |
| `-k, --tls-key` | `AETHER_TLS_KEY` | TLS private key file for HTTPS | - |
| `-m, --mtls-ca-cert` | `AETHER_MTLS_CA_CERT` | CA certificate for client verification (mTLS) | - |
| `--api-token` | `AETHER_API_TOKEN` | Bearer token for API authentication | - |
| `--encryption-key` | `AETHER_ENCRYPTION_KEY` | 32-byte encryption key for node passwords (auto-generated if not provided) | - |
| `-r, --enable-rbac` | `AETHER_ENABLE_RBAC` | Enable RBAC authentication/authorization | `false` |
| `--cors-origins` | `AETHER_CORS_ORIGINS` | Comma-separated list of allowed CORS origins (e.g., `http://localhost:5173`) | - |

### Execution

| Flag | Env Var | Description | Default |
|------|---------|-------------|---------|
| `-u, --exec-user` | `AETHER_EXEC_USER` | User for command execution | - |
| `-e, --exec-env` | `AETHER_EXEC_ENV` | Environment variables for execution | - |

### OnRamp

| Flag | Env Var | Description | Default |
|------|---------|-------------|---------|
| `--onramp-dir` | `AETHER_ONRAMP_DIR` | Path to the aether-onramp repository on disk | `{data-dir}/aether-onramp` |
| `--onramp-version` | `AETHER_ONRAMP_VERSION` | Tag, branch, or commit to pin aether-onramp to | `main` |

### Frontend

| Flag | Env Var | Description | Default |
|------|---------|-------------|---------|
| `-f, --serve-frontend` | `AETHER_SERVE_FRONTEND` | Enable serving frontend static files | `true` |
| `--frontend-dir` | `AETHER_FRONTEND_DIR` | Override embedded frontend with files from this directory | - |

### Storage

| Flag | Env Var | Description | Default |
|------|---------|-------------|---------|
| `--data-dir` | `AETHER_DATA_DIR` | Directory for persistent state database | `/var/lib/aether-webd` |

### Metrics

| Flag | Env Var | Description | Default |
|------|---------|-------------|---------|
| `--metrics-interval` | `AETHER_METRICS_INTERVAL` | How often to collect system metrics (e.g., `10s`, `30s`, `1m`) | `10s` |
| `--metrics-retention` | `AETHER_METRICS_RETENTION` | How long to retain historical metrics data (e.g., `24h`, `7d`) | `24h` |

## Environment Variables

Every CLI flag (except `--version`) has a corresponding `AETHER_*` environment variable. The precedence order is: **CLI flag > environment variable > hardcoded default**.

| Variable | Description | Corresponding Flag |
|----------|-------------|-------------------|
| `AETHER_LISTEN` | Address and port to listen on | `--listen` |
| `AETHER_DEBUG` | Enable debug logging (`true`, `1`, `yes`) | `--debug` |
| `AETHER_TLS` | Enable TLS (`true`, `1`, `yes`) | `--tls` |
| `AETHER_TLS_CERT` | Path to TLS certificate file | `--tls-cert` |
| `AETHER_TLS_KEY` | Path to TLS private key file | `--tls-key` |
| `AETHER_MTLS_CA_CERT` | Path to CA certificate for mTLS | `--mtls-ca-cert` |
| `AETHER_API_TOKEN` | Bearer token for API authentication | `--api-token` |
| `AETHER_ENCRYPTION_KEY` | 32-byte hex-encoded key for encrypting node passwords at rest (AES-256-GCM). If neither the flag nor the env var is provided, a random key is generated at startup (secrets will not survive restarts). | `--encryption-key` |
| `AETHER_ENABLE_RBAC` | Enable RBAC (`true`, `1`, `yes`) | `--enable-rbac` |
| `AETHER_CORS_ORIGINS` | Comma-separated list of allowed CORS origins | `--cors-origins` |
| `AETHER_DATA_DIR` | Directory for persistent state database | `--data-dir` |
| `AETHER_ONRAMP_DIR` | Path to aether-onramp repository | `--onramp-dir` |
| `AETHER_ONRAMP_VERSION` | Tag, branch, or commit to pin aether-onramp to | `--onramp-version` |
| `AETHER_SERVE_FRONTEND` | Enable frontend serving (`true`, `1`, `yes`) | `--serve-frontend` |
| `AETHER_FRONTEND_DIR` | Override embedded frontend directory | `--frontend-dir` |
| `AETHER_METRICS_INTERVAL` | Metrics collection interval (e.g., `10s`) | `--metrics-interval` |
| `AETHER_METRICS_RETENTION` | Metrics retention duration (e.g., `24h`) | `--metrics-retention` |
| `AETHER_EXEC_USER` | User for command execution | `--exec-user` |
| `AETHER_EXEC_ENV` | Environment variables for execution | `--exec-env` |

## Examples

### Listen on all interfaces

```bash
aether-webd --listen 0.0.0.0:8186
```

### Quick TLS with auto-generated self-signed certificate

```bash
aether-webd --tls
```

The server generates a self-signed certificate at startup and listens on `https://127.0.0.1:8443` by default when TLS is enabled.

### TLS with your own certificate

```bash
aether-webd --tls-cert /etc/aether-webd/cert.pem --tls-key /etc/aether-webd/key.pem
```

### mTLS (mutual TLS) requiring client certificates

```bash
aether-webd \
  --tls-cert /etc/aether-webd/cert.pem \
  --tls-key /etc/aether-webd/key.pem \
  --mtls-ca-cert /etc/aether-webd/ca.pem
```

### Bearer token authentication via environment variable

```bash
export AETHER_API_TOKEN=my-secret-token
aether-webd --listen 0.0.0.0:8186
```

All API requests (except [public paths](./api-overview.md)) require the header `Authorization: Bearer my-secret-token`.

### Allow CORS for a local frontend dev server

```bash
aether-webd --cors-origins http://localhost:5173
```

Multiple origins can be comma-separated. Use `*` to allow all origins (not recommended for production).

### API-only mode (no frontend)

```bash
aether-webd --serve-frontend=false
```

### Custom data directory and metrics retention

```bash
aether-webd \
  --data-dir /opt/aether/data \
  --metrics-interval 30s \
  --metrics-retention 7d
```
