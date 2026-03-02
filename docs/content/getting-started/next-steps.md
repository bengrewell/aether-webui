---
sidebar_position: 5
title: Next Steps
---

# Next Steps

With Kubernetes and the 5G Core running, there are several directions to explore depending on your goals. This page provides pointers to the most relevant parts of the documentation.

## Guides

Step-by-step instructions for common tasks:

- [Node Management](../guides/node-management) -- add remote nodes, assign roles, and manage multi-node clusters
- [Configuration](../guides/configuration) -- edit OnRamp variables, switch profiles, and tune network parameters
- [Deploying Components](../guides/deploying-components) -- deploy additional components such as gNBSim, srsRAN, UERANSIM, and AMP
- [Monitoring](../guides/monitoring) -- collect and query system metrics for CPU, memory, disk, and network
- [Security](../guides/security) -- enable TLS, configure API token authentication, and lock down production deployments
- [Troubleshooting](../guides/troubleshooting) -- diagnose common issues with deployments, tasks, and connectivity

## Reference

Detailed specifications for the CLI and API:

- [CLI Reference](../reference/cli) -- all command-line flags and environment variables for `aether-webd`
- [API Overview](../reference/api-overview) -- base URL conventions, authentication, error format, and pagination
- [OnRamp API](../reference/api-onramp) -- full endpoint reference for components, tasks, config, and profiles
- [System API](../reference/api-system) -- host metrics and system information endpoints
- [Meta API](../reference/api-meta) -- version, health, config introspection, and store diagnostics
- [Components](../reference/components) -- full list of components and their available actions

## Concepts

Background material on how Aether WebUI works:

- [Architecture](../concepts/architecture) -- controller, providers, transport, and store layers
- [Providers](../concepts/providers) -- how the modular provider framework works
- [Tasks](../concepts/tasks) -- task lifecycle, single-task constraint, and incremental output polling
- [Deployment State](../concepts/deployment-state) -- how component state transitions work and what each state means

## Production hardening

Before exposing Aether WebUI beyond localhost, take the following steps:

- **Enable TLS** -- use `--tls` for auto-generated self-signed certificates, or provide your own via `--tls-cert` and `--tls-key`. See the [Security guide](../guides/security) for details.
- **Set an API token** -- use `--api-token` or the `AETHER_API_TOKEN` environment variable to require bearer token authentication on all API requests.
- **Restrict the listen address** -- the default `127.0.0.1:8186` only accepts local connections. If remote access is needed, bind to a specific interface rather than `0.0.0.0`.

## Upstream documentation

Aether WebUI builds on top of the Aether OnRamp project. For details on the underlying network components, Ansible playbooks, and Makefile targets, refer to the upstream repository:

- [Aether OnRamp on GitHub](https://github.com/opennetworkinglab/aether-onramp)
