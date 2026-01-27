# Aether WebUI

[![Build](https://github.com/bengrewell/aether-webui/actions/workflows/build.yaml/badge.svg)](https://github.com/bengrewell/aether-webui/actions/workflows/build.yaml)
[![codecov](https://codecov.io/gh/bengrewell/aether-webui/branch/main/graph/badge.svg)](https://codecov.io/gh/bengrewell/aether-webui)

Backend API service for the Aether WebUI. This service is responsible for executing deployment tasks, gathering system information, and monitoring the health and metrics of Aether 5G deployments. It manages SD-Core components, gNBs (such as srsRAN and OCUDU), Kubernetes clusters, and host systems.

## Building

### Prerequisites

- Go 1.22 or later
- Make

### Build from Source

```bash
# Build the binary with version info from git
make build

# The binary will be in bin/aether-webd
./bin/aether-webd --version
```

### Other Make Targets

```bash
make test           # Run tests with coverage
make coverage       # Run tests and display coverage summary
make coverage-html  # Generate HTML coverage report
make clean          # Remove build artifacts
make run            # Build and run
make version        # Display version info that would be injected
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

### Examples

```bash
# Run on all interfaces
aether-webd --listen 0.0.0.0:8680

# Run with HTTPS
aether-webd --tls-cert /path/to/cert.pem --tls-key /path/to/key.pem

# Run with mTLS (client certificate required)
aether-webd --tls-cert cert.pem --tls-key key.pem --mtls-ca-cert ca.pem
```

## Development

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
