---
sidebar_position: 2
title: Installation
---

# Installation

This page walks through installing Aether WebUI and verifying that the service is running.

## Prepare the host

Before installing the service, run the preflight setup script to install required packages (`git`, `make`, `ansible`, `openssh-server`, `iptables`), enable SSH password authentication, and create the `aether` service user:

```bash
curl -fsSL https://raw.githubusercontent.com/bengrewell/aether-webui/main/scripts/preflight-setup.sh | sudo bash
```

The script is idempotent — running it again on an already-configured host skips all steps that are already satisfied.

Note: The script enables SSH password authentication and creates a user with a default password. See the [Security guide](../guides/security) for hardening recommendations.

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

The service exposes a health endpoint that returns a JSON object when the server is ready to accept requests:

```bash
curl http://localhost:8186/healthz
```

Expected response:

```json
{"status":"healthy","version":"0.0.9","uptime":"1m39s"}
```

The `version` and `uptime` fields will vary based on the installed release and how long the service has been running.

## Check the version

Confirm which version is installed by querying the version endpoint:

```bash
curl http://localhost:8186/api/v1/meta/version
```

This returns the build version, commit hash, and build timestamp.

## Default listen address

By default, `aether-webd` listens on `127.0.0.1:8186`. This means the API and Web UI are only accessible from the local machine. To expose the service on all interfaces, set the `AETHER_LISTEN` variable in the systemd environment file and restart:

```bash
# Edit the environment file to set the listen address
echo 'AETHER_LISTEN=0.0.0.0:8186' | sudo tee /etc/aether-webd/env

# Restart the service to apply
sudo systemctl restart aether-webd
```

The environment file at `/etc/aether-webd/env` accepts one variable per line. Every CLI flag has a corresponding `AETHER_*` environment variable. For example, to enable TLS and token auth alongside the listen address:

```bash
cat <<'EOF' | sudo tee /etc/aether-webd/env
AETHER_LISTEN=0.0.0.0:8186
AETHER_TLS=true
AETHER_API_TOKEN=my-secret-token
EOF
```

CLI flags override environment variables when both are set. See the [CLI Reference](../reference/cli) for the full mapping.

Note: Exposing the API on all interfaces should be paired with TLS and API token authentication in production. See the [Security guide](../guides/security) for setup instructions.

## Configuration options

The `aether-webd` binary accepts several flags for customizing behavior, including `--tls` for automatic self-signed certificate generation and `--api-token` for bearer token authentication. See the [CLI Reference](../reference/cli) for the full list of options.

## Next step

With the service running, proceed to [First Deployment](first-deployment) to run preflight checks and deploy Kubernetes and the 5G Core.
