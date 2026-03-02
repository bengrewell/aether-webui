---
sidebar_position: 6
title: "Running with Docker"
---

# Running with Docker

The `ghcr.io/bengrewell/aether-webd:latest` Docker image packages aether-webd for container-based deployments.

## Build the image

To build from source:

```bash
make docker-build
```

## Run the container

### Basic startup

```bash
docker run -d \
  --name aether-webd \
  -p 8186:8186 \
  ghcr.io/bengrewell/aether-webd:latest \
  --listen 0.0.0.0:8186
```

The `--listen 0.0.0.0:8186` flag is required because the default listen address (`127.0.0.1:8186`) is only reachable from inside the container. Binding to `0.0.0.0` makes the API accessible via the published port.

### Pass server flags

Append flags after the image name:

```bash
docker run -d \
  --name aether-webd \
  -p 8186:8186 \
  ghcr.io/bengrewell/aether-webd:latest \
  --listen 0.0.0.0:8186 \
  --metrics-interval 30s \
  --metrics-retention 48h
```

## Persistent data

Mount a volume to preserve the SQLite database and generated certificates across container restarts:

```bash
docker run -d \
  --name aether-webd \
  -p 8186:8186 \
  -v aether-data:/var/lib/aether-webd \
  ghcr.io/bengrewell/aether-webd:latest \
  --listen 0.0.0.0:8186
```

## TLS with mounted certificates

Mount your certificate files read-only and reference them with server flags:

```bash
docker run -d \
  --name aether-webd \
  -p 8186:8186 \
  -v /path/to/certs:/certs:ro \
  -v aether-data:/var/lib/aether-webd \
  ghcr.io/bengrewell/aether-webd:latest \
  --listen 0.0.0.0:8186 \
  --tls-cert /certs/cert.pem \
  --tls-key /certs/key.pem
```

For auto-generated TLS certificates, the persistent data volume is sufficient -- certificates are written to `/var/lib/aether-webd/certs/` automatically:

```bash
docker run -d \
  --name aether-webd \
  -p 8186:8186 \
  -v aether-data:/var/lib/aether-webd \
  ghcr.io/bengrewell/aether-webd:latest \
  --listen 0.0.0.0:8186 \
  --tls
```

## Environment variables

Set the API token via an environment variable instead of a CLI flag:

```bash
docker run -d \
  --name aether-webd \
  -p 8186:8186 \
  -e AETHER_API_TOKEN=mysecrettoken \
  -v aether-data:/var/lib/aether-webd \
  ghcr.io/bengrewell/aether-webd:latest \
  --listen 0.0.0.0:8186
```

This keeps the token out of the process argument list. See the [Security guide](security) for details on token authentication.

## Health check

Verify the container is healthy:

```bash
docker exec aether-webd curl -sf http://localhost:8186/healthz
```

For Docker's built-in health check mechanism, add a `HEALTHCHECK` instruction or pass `--health-cmd`:

```bash
docker run -d \
  --name aether-webd \
  -p 8186:8186 \
  --health-cmd "curl -sf http://localhost:8186/healthz || exit 1" \
  --health-interval 30s \
  --health-timeout 5s \
  --health-retries 3 \
  ghcr.io/bengrewell/aether-webd:latest \
  --listen 0.0.0.0:8186
```

## Full production example

```bash
docker run -d \
  --name aether-webd \
  --restart unless-stopped \
  -p 8186:8186 \
  -e AETHER_API_TOKEN="$(openssl rand -hex 32)" \
  -v aether-data:/var/lib/aether-webd \
  -v /path/to/certs:/certs:ro \
  --health-cmd "curl -sf http://localhost:8186/healthz || exit 1" \
  --health-interval 30s \
  ghcr.io/bengrewell/aether-webd:latest \
  --listen 0.0.0.0:8186 \
  --tls-cert /certs/cert.pem \
  --tls-key /certs/key.pem
```
