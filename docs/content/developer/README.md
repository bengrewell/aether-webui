---
sidebar_position: 1
title: Developer Docs
---

# Aether WebUI Developer Documentation

These docs cover internal implementation details for contributors. For user-facing documentation (installation, API reference, deployment guides), see the [main documentation site](/docs/getting-started/introduction).

## Architecture & Concepts
- [Architecture Overview](architecture.md) — controller lifecycle, provider framework, request flow
- [Security](security.md) — TLS, mTLS, token authentication

## Providers
- [Meta](providers/meta.md) — server introspection and diagnostics
- [OnRamp](providers/onramp.md) — Aether OnRamp lifecycle management
- [Preflight](providers/preflight.md) — pre-deployment system checks and fixes
- [System](providers/system.md) — host system metrics
