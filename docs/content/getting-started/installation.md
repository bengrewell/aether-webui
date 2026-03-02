---
sidebar_position: 2
title: Installation
---

# Installation

This page walks through installing Aether WebUI and verifying that the service is running.

## Install via the quick-start script

The fastest way to install is the one-line installer. This downloads the latest release binary, installs it to the system path, and creates a systemd service.

```bash
curl -fsSL https://raw.githubusercontent.com/bengrewell/aether-webui/main/scripts/install.sh | sudo bash
```

The script handles downloading the correct binary for your architecture, placing it in `/usr/local/bin`, and enabling the `aether-webd` systemd unit.

## Verify the service is running

After installation, confirm the service started successfully:

```bash
systemctl status aether-webd
```

The output should show `active (running)`. If the service failed to start, check the journal for details:

```bash
journalctl -u aether-webd --no-pager -n 50
```

## Check the health endpoint

The service exposes a health endpoint that returns `"healthy"` when the server is ready to accept requests:

```bash
curl http://localhost:8186/healthz
```

Expected response:

```json
"healthy"
```

## Check the version

Confirm which version is installed by querying the version endpoint:

```bash
curl http://localhost:8186/api/v1/meta/version
```

This returns the build version, commit hash, and build timestamp.

## Default listen address

By default, `aether-webd` listens on `127.0.0.1:8186`. This means the API is only accessible from the local machine. To expose it on all interfaces, restart the service with the `--listen` flag:

```bash
aether-webd --listen 0.0.0.0:8186
```

Note: Exposing the API on all interfaces without authentication is not recommended for production. See the [Security guide](../guides/security) for instructions on enabling TLS and API token authentication.

## Configuration options

The `aether-webd` binary accepts several flags for customizing behavior, including `--tls` for automatic self-signed certificate generation and `--api-token` for bearer token authentication. See the [CLI Reference](../reference/cli) for the full list of options.

## Next step

With the service running, proceed to [First Deployment](first-deployment) to deploy Kubernetes and the 5G Core.
