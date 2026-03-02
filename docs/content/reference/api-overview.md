---
sidebar_position: 2
title: "API Overview"
---

# API Overview

The Aether WebUI exposes a REST API for managing Aether 5G deployments, querying system information, and monitoring metrics.

## Quick Reference

| Property | Value |
|----------|-------|
| **Base URL** | `http://host:8186/api/v1/` (default) or `https://host:8443/api/v1/` with TLS |
| **Content-Type** | `application/json` |
| **Authentication** | `Authorization: Bearer <token>` (when `--api-token` or `AETHER_API_TOKEN` is set) |
| **Error format** | [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457) Problem Details |
| **OpenAPI spec** | `GET /openapi.json` |
| **Interactive docs** | `GET /docs` (Swagger UI) |

## Public Paths (no authentication required)

These paths bypass bearer-token authentication even when a token is configured:

- `/healthz` -- health check
- `/openapi.json` -- OpenAPI 3.1 specification
- `/docs` -- Swagger UI

## Providers

The API is organized into four providers. Each provider groups related endpoints under a common path prefix.

| Provider | Path Prefix | Endpoints | Description |
|----------|-------------|-----------|-------------|
| [Meta](./api-meta.md) | `/api/v1/meta/` | 6 | Version, build, runtime, config, providers, store diagnostics |
| [System](./api-system.md) | `/api/v1/system/` | 8 | CPU, memory, disk, OS, network, metrics |
| [Nodes](./api-nodes.md) | `/api/v1/nodes` | 5 | Managed cluster node CRUD |
| [OnRamp](./api-onramp.md) | `/api/v1/onramp/` | 18 | Components, tasks, actions, config, profiles, inventory |
| | | **37 total** | |

## Authentication

When bearer-token authentication is enabled (via `--api-token` flag or `AETHER_API_TOKEN` environment variable), all requests to non-public paths must include the token in the `Authorization` header.

```bash
curl -H "Authorization: Bearer my-secret-token" \
     http://localhost:8186/api/v1/meta/version
```

Requests without a valid token receive a `401 Unauthorized` response.

## Error Response Format

All errors follow the [RFC 9457 Problem Details](https://www.rfc-editor.org/rfc/rfc9457) format:

```json
{
  "title": "Not Found",
  "status": 404,
  "detail": "no node with id abc123"
}
```

### Common Error Codes

| Status | When |
|--------|------|
| `401 Unauthorized` | Missing or invalid `Authorization` header |
| `404 Not Found` | Resource not found (node, task, component, action, profile) |
| `409 Conflict` | Task already running (max 1 concurrent task) |
| `422 Unprocessable Entity` | Validation error (missing required field, invalid role, etc.) |
| `500 Internal Server Error` | Unexpected server-side failure |

## Interactive Documentation

When the server is running, visit `/docs` in a browser to access the Swagger UI. This provides an interactive explorer for all endpoints, complete with request/response schemas and a "Try it out" feature.

```
http://localhost:8186/docs
```

The raw OpenAPI 3.1 specification is available at `/openapi.json` and can be imported into tools like Postman, Insomnia, or code generators.
