# Aether WebUI

[![Build](https://github.com/bengrewell/aether-webui/actions/workflows/build.yaml/badge.svg)](https://github.com/bengrewell/aether-webui/actions/workflows/build.yaml)
[![codecov](https://codecov.io/github/bengrewell/aether-webui/graph/badge.svg?token=J3ZEEWEQT0)](https://codecov.io/github/bengrewell/aether-webui)

Backend API service for the Aether WebUI. This service is responsible for executing deployment tasks, gathering system information, and monitoring the health and metrics of Aether 5G deployments. It manages SD-Core components, gNBs (such as srsRAN and OCUDU), Kubernetes clusters, and host systems.

## Quick Install

Install aether-webd as a systemd service with a single command:

```bash
# Install latest version
curl -fsSL https://raw.githubusercontent.com/bengrewell/aether-webui/main/scripts/install.sh | sudo bash

# Install specific version
curl -fsSL https://raw.githubusercontent.com/bengrewell/aether-webui/main/scripts/install.sh | VERSION=v1.0.0 sudo bash
```

To uninstall:

```bash
# Basic uninstall (keeps config and user)
curl -fsSL https://raw.githubusercontent.com/bengrewell/aether-webui/main/scripts/uninstall.sh | sudo bash

# Full uninstall (removes everything)
curl -fsSL https://raw.githubusercontent.com/bengrewell/aether-webui/main/scripts/uninstall.sh | sudo bash -s -- --purge
```

## Features

- **Provider Framework**: Extensible plugin system for registering API endpoint groups at runtime
- **API Introspection**: Built-in meta provider exposes version, build, runtime, config, provider, and store diagnostics
- **Security**: TLS, mTLS (mutual TLS), and bearer-token authentication out of the box
- **Persistent State**: SQLite-backed store with versioned schema migrations and AES-256-GCM encryption for secrets
- **Embedded Frontend**: React SPA embedded in the Go binary; serve from disk during development
- **OpenAPI Documentation**: Auto-generated OpenAPI 3.1 spec with interactive Swagger UI at `/docs`

### Roadmap

- System monitoring (CPU, memory, disk, NIC metrics)
- Kubernetes cluster management
- Aether 5G lifecycle management (SD-Core, gNB)
- Multi-host distributed deployments

## Building

### Prerequisites

- Go 1.25 or later
- Make
- Node.js 18+ and npm (for frontend)

### Build from Source

```bash
# Initialize the frontend submodule (first time only)
git submodule update --init

# Build frontend and backend with embedded frontend
make all

# Or build backend only (without embedded frontend)
make build

# The binary will be in bin/aether-webd
./bin/aether-webd --version
```

### Other Make Targets

```bash
make all            # Build frontend and backend with embedding
make build          # Build backend only
make frontend       # Build frontend only
make embed-frontend # Build and copy frontend to embed location
make test           # Run tests with coverage
make coverage       # Run tests and display coverage summary
make coverage-html  # Generate HTML coverage report
make clean          # Remove build artifacts
make run            # Build and run
make version        # Display version info that would be injected
make docker-build   # Build Docker image
make docker-push    # Build and push Docker image
```

## Usage

```bash
aether-webd [options]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-v, --version` | Print version information and exit | `false` |
| `-d, --debug` | Enable debug mode for verbose logging | `false` |
| `-l, --listen` | Address and port to listen on | `127.0.0.1:8186` |

### Security Options

| Flag | Description | Default |
|------|-------------|---------|
| `--tls` | Enable TLS (auto-generates a self-signed cert if `--tls-cert`/`--tls-key` are not provided) | `false` |
| `-t, --tls-cert` | TLS certificate file for HTTPS | - |
| `-k, --tls-key` | TLS private key file for HTTPS | - |
| `-m, --mtls-ca-cert` | CA certificate for client verification (mTLS) | - |
| `--api-token` | Bearer token for API authentication (falls back to `AETHER_API_TOKEN` env var) | - |
| `--encryption-key` | 32-byte encryption key for node passwords (falls back to `AETHER_ENCRYPTION_KEY` env var; auto-generated if neither is provided) | - |
| `-r, --enable-rbac` | Enable RBAC authentication/authorization | `false` |

### Execution Options

| Flag | Description | Default |
|------|-------------|---------|
| `-u, --exec-user` | User for command execution | - |
| `-e, --exec-env` | Environment variables for execution | - |

### Frontend Options

| Flag | Description | Default |
|------|-------------|---------|
| `-f, --serve-frontend` | Enable serving frontend static files | `true` |
| `--frontend-dir` | Override embedded frontend with files from this directory | - |

### Storage Options

| Flag | Description | Default |
|------|-------------|---------|
| `--data-dir` | Directory for persistent state database | `/var/lib/aether-webd` |

### Metrics Options

| Flag | Description | Default |
|------|-------------|---------|
| `--metrics-interval` | How often to collect system metrics (e.g., `10s`, `30s`, `1m`) | `10s` |
| `--metrics-retention` | How long to retain historical metrics data (e.g., `24h`, `7d`) | `24h` |

### Examples

```bash
# Run on all interfaces
aether-webd --listen 0.0.0.0:8186

# Quick TLS with auto-generated self-signed certificate
aether-webd --tls

# Run with HTTPS using your own certificate
aether-webd --tls-cert /path/to/cert.pem --tls-key /path/to/key.pem

# Run with mTLS (client certificate required)
aether-webd --tls-cert cert.pem --tls-key key.pem --mtls-ca-cert ca.pem

# Require a bearer token for all API requests
aether-webd --api-token my-secret-token
# Or via environment variable:
AETHER_API_TOKEN=my-secret-token aether-webd

# Run API only (no frontend)
aether-webd --serve-frontend=false

# Run with custom frontend directory (for development)
aether-webd --frontend-dir ./web/frontend/dist
```

## API Documentation

The service provides a REST API built with [Huma](https://huma.rocks/), which auto-generates OpenAPI documentation.

- **Interactive docs**: `http://localhost:8186/docs` — Swagger UI for exploring and testing endpoints
- **OpenAPI spec**: `http://localhost:8186/openapi.json` — Machine-readable OpenAPI 3.1 specification

### Providers

Endpoints are grouped into providers. Additional providers will be added as the roadmap progresses.

| Provider | Endpoints | Description |
|----------|-----------|-------------|
| meta | 6 | Version, build, runtime, config, provider list, store diagnostics |

### Meta Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/meta/version` | Application version, branch, and commit |
| `GET` | `/api/v1/meta/build` | Go version, OS, architecture, compiler |
| `GET` | `/api/v1/meta/runtime` | PID, uptime, start time, running user |
| `GET` | `/api/v1/meta/config` | Active configuration (listen address, security, frontend, storage, metrics) |
| `GET` | `/api/v1/meta/providers` | Registered providers with enabled/running status and endpoint counts |
| `GET` | `/api/v1/meta/store` | Store engine, schema version, file size, and diagnostic checks |

### Built-in Routes

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/healthz` | Health check (returns `{"status":"healthy"}`) |
| `GET` | `/docs` | Interactive Swagger UI |
| `GET` | `/openapi.json` | OpenAPI 3.1 specification |

### Example Requests

```bash
# Health check
curl http://localhost:8186/healthz

# Version info
curl http://localhost:8186/api/v1/meta/version

# Build info
curl http://localhost:8186/api/v1/meta/build

# Running configuration
curl http://localhost:8186/api/v1/meta/config

# Registered providers
curl http://localhost:8186/api/v1/meta/providers

# Store diagnostics
curl http://localhost:8186/api/v1/meta/store

# With bearer token authentication
curl -H "Authorization: Bearer my-secret-token" http://localhost:8186/api/v1/meta/version
```

## Development

### Frontend

The frontend is a React application located in the `web/frontend` submodule (github.com/bengrewell/aether-webui-frontend).

#### Setup

```bash
# Initialize submodule
git submodule update --init

# Install frontend dependencies
cd web/frontend && npm install
```

#### Development Workflow

For frontend development, run the Vite dev server alongside the API:

```bash
# Terminal 1: Run API backend
make build && ./bin/aether-webd --serve-frontend=false

# Terminal 2: Run frontend dev server (with hot reload)
cd web/frontend && npm run dev
```

The Vite dev server proxies API requests to the backend automatically.

#### Production Build

```bash
# Build frontend and embed into Go binary
make all

# The resulting binary includes the frontend - no external files needed
./bin/aether-webd
```

### Versioning

Version information is automatically injected at build time using ldflags:

- **Version**: From `git describe --tags` (e.g., `v0.0.1` or `v0.0.1-3-gabcdef`)
- **Commit**: Short commit hash
- **Branch**: Current git branch
- **Build Date**: UTC timestamp

### Creating a Release

1. Ensure all changes are committed and pushed
2. Create and push a tag:
   ```bash
   git tag v0.0.1
   git push origin v0.0.1
   ```
3. GitHub Actions will automatically:
   - Run tests
   - Build binaries for Linux (amd64, arm64)
   - Create a GitHub release with artifacts

### CI/CD

- **On push/PR to main**: Runs tests and builds (artifact uploaded)
- **On tag push (v*)**: Creates GitHub release with GoReleaser

## Deployment

### Systemd

Install and run as a systemd service:

```bash
# Create service user
sudo useradd -r -s /bin/false aether-webd

# Install binary
sudo cp bin/aether-webd /usr/local/bin/
sudo chmod +x /usr/local/bin/aether-webd

# Install service file
sudo cp deploy/systemd/aether-webd.service /etc/systemd/system/

# Create config directory (optional)
sudo mkdir -p /etc/aether-webd
echo 'AETHER_WEBD_OPTS="--listen 0.0.0.0:8186"' | sudo tee /etc/aether-webd/env

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable aether-webd
sudo systemctl start aether-webd
```

### Docker

Build and run as a container:

```bash
# Build image
make docker-build

# Run container
docker run -d \
  --name aether-webd \
  -p 8186:8186 \
  ghcr.io/bengrewell/aether-webd:latest

# Run with TLS
docker run -d \
  --name aether-webd \
  -p 8186:8186 \
  -v /path/to/certs:/certs:ro \
  ghcr.io/bengrewell/aether-webd:latest \
  --listen 0.0.0.0:8186 \
  --tls-cert /certs/cert.pem \
  --tls-key /certs/key.pem
```

### Kubernetes

Deploy to a Kubernetes cluster:

```bash
# Apply manifests
kubectl apply -f deploy/k8s/

# Check deployment status
kubectl get pods -l app=aether-webd
kubectl get svc aether-webd

# View logs
kubectl logs -l app=aether-webd
```

For TLS, create a secret and uncomment the volume mounts in `deployment.yaml`:

```bash
kubectl create secret tls aether-webd-tls \
  --cert=/path/to/cert.pem \
  --key=/path/to/key.pem
```

## Contributing

### Test Coverage Requirements

We prioritize test coverage to maintain code quality. Please follow these guidelines:

- **Target coverage: 70%** - All new code should aim for at least 70% test coverage
- **No coverage regression** - PRs should not decrease overall coverage
- **Test new features** - All new features must include tests
- **Test bug fixes** - Bug fixes should include regression tests

### Running Tests Locally

```bash
# Run tests with coverage
make test

# View coverage summary in terminal
make coverage

# Generate HTML coverage report (opens coverage.html)
make coverage-html
```

### Writing Tests

- Place tests in `*_test.go` files alongside the code being tested
- Use table-driven tests for multiple test cases
- Mock external dependencies (HTTP clients, databases, etc.)
- Test both success and error paths

## License

See the [LICENSE](LICENSE) file for details.
