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

| Flag | Description | Default |
|------|-------------|---------|
| `-v, --version` | Print version information and exit | `false` |
| `-d, --debug` | Enable debug mode for verbose logging | `false` |
| `-l, --listen` | Address and port to listen on | `127.0.0.1:8186` |

### Security

| Flag | Description | Default |
|------|-------------|---------|
| `--tls` | Enable TLS (auto-generates a self-signed cert if `--tls-cert`/`--tls-key` are not provided) | `false` |
| `-t, --tls-cert` | TLS certificate file for HTTPS | - |
| `-k, --tls-key` | TLS private key file for HTTPS | - |
| `-m, --mtls-ca-cert` | CA certificate for client verification (mTLS) | - |
| `--api-token` | Bearer token for API authentication (falls back to `AETHER_API_TOKEN` env var) | - |
| `--encryption-key` | 32-byte encryption key for node passwords (falls back to `AETHER_ENCRYPTION_KEY` env var; auto-generated if neither is provided) | - |
| `-r, --enable-rbac` | Enable RBAC authentication/authorization | `false` |

### Execution

| Flag | Description | Default |
|------|-------------|---------|
| `-u, --exec-user` | User for command execution | - |
| `-e, --exec-env` | Environment variables for execution | - |

### OnRamp

| Flag | Description | Default |
|------|-------------|---------|
| `--onramp-dir` | Path to the aether-onramp repository on disk | `{data-dir}/aether-onramp` |
| `--onramp-version` | Tag, branch, or commit to pin aether-onramp to | `main` |

### Frontend

| Flag | Description | Default |
|------|-------------|---------|
| `-f, --serve-frontend` | Enable serving frontend static files | `true` |
| `--frontend-dir` | Override embedded frontend with files from this directory | - |

### Storage

| Flag | Description | Default |
|------|-------------|---------|
| `--data-dir` | Directory for persistent state database | `/var/lib/aether-webd` |

### Metrics

| Flag | Description | Default |
|------|-------------|---------|
| `--metrics-interval` | How often to collect system metrics (e.g., `10s`, `30s`, `1m`) | `10s` |
| `--metrics-retention` | How long to retain historical metrics data (e.g., `24h`, `7d`) | `24h` |

## Environment Variables

| Variable | Description | Corresponding Flag |
|----------|-------------|-------------------|
| `AETHER_API_TOKEN` | Bearer token for API authentication. Used when `--api-token` is not set on the command line. | `--api-token` |
| `AETHER_ENCRYPTION_KEY` | 32-byte hex-encoded key for encrypting node passwords at rest (AES-256-GCM). Used when `--encryption-key` is not set. If neither the flag nor the env var is provided, a random key is generated at startup (secrets will not survive restarts). | `--encryption-key` |

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
