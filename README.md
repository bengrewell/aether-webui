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

- **System Monitoring**: Query hardware and OS information (CPU, memory, disk, NICs) and collect real-time metrics
- **Kubernetes Integration**: Monitor cluster health, nodes, pods, deployments, services, and events
- **Aether 5G Management**: Full lifecycle management of SD-Core and gNB deployments (install, start, stop, restart, uninstall)
- **Multi-host Support**: Manage distributed deployments across multiple hosts

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
| `-d, --debug` | Enable debug mode for verbose logging | `false` |
| `-l, --listen` | Address and port to listen on | `127.0.0.1:8680` |

### Security Options

| Flag | Description | Default |
|------|-------------|---------|
| `-t, --tls-cert` | TLS certificate file for HTTPS | - |
| `-k, --tls-key` | TLS private key file for HTTPS | - |
| `-m, --mtls-ca-cert` | CA certificate for client verification (mTLS) | - |
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

### Examples

```bash
# Run on all interfaces
aether-webd --listen 0.0.0.0:8680

# Run with HTTPS
aether-webd --tls-cert /path/to/cert.pem --tls-key /path/to/key.pem

# Run with mTLS (client certificate required)
aether-webd --tls-cert cert.pem --tls-key key.pem --mtls-ca-cert ca.pem

# Run API only (no frontend)
aether-webd --serve-frontend=false

# Run with custom frontend directory (for development)
aether-webd --frontend-dir ./web/frontend/dist
```

## API Documentation

The service provides a REST API built with [Huma](https://huma.rocks/), which auto-generates OpenAPI documentation.

- **Interactive docs**: `http://localhost:8680/docs` - Swagger UI for exploring and testing endpoints
- **OpenAPI spec**: `http://localhost:8680/openapi.json` - Machine-readable OpenAPI 3.1 specification

### API Endpoints

The API provides 36 endpoints across 6 categories:

| Category | Endpoints | Description |
|----------|-----------|-------------|
| Health | 1 | Service health check |
| Setup | 3 | Setup wizard status and completion |
| System Info | 5 | CPU, memory, disk, NIC, and OS information |
| Metrics | 4 | Real-time CPU, memory, disk, and network usage |
| Kubernetes | 7 | Cluster health, nodes, pods, deployments, services, events |
| Aether 5G | 16 | Host management, SD-Core and gNB lifecycle operations |

### Example Requests

```bash
# Health check
curl http://localhost:8680/healthz

# Setup wizard status
curl http://localhost:8680/api/v1/setup/status
curl -X POST http://localhost:8680/api/v1/setup/complete
curl -X DELETE http://localhost:8680/api/v1/setup/status

# Get system information
curl http://localhost:8680/api/v1/system/cpu
curl http://localhost:8680/api/v1/system/memory
curl http://localhost:8680/api/v1/system/os

# Get real-time metrics
curl http://localhost:8680/api/v1/metrics/cpu
curl http://localhost:8680/api/v1/metrics/memory

# Kubernetes status (requires cluster access)
curl http://localhost:8680/api/v1/kubernetes/health
curl http://localhost:8680/api/v1/kubernetes/pods

# Aether 5G management
curl http://localhost:8680/api/v1/aether/hosts
curl http://localhost:8680/api/v1/aether/sdcore/status
curl -X POST http://localhost:8680/api/v1/aether/sdcore/install
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
echo 'AETHER_WEBD_OPTS="--listen 0.0.0.0:8680"' | sudo tee /etc/aether-webd/env

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
  -p 8680:8680 \
  ghcr.io/bengrewell/aether-webd:latest

# Run with TLS
docker run -d \
  --name aether-webd \
  -p 8680:8680 \
  -v /path/to/certs:/certs:ro \
  ghcr.io/bengrewell/aether-webd:latest \
  --listen 0.0.0.0:8680 \
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

[License information here]
