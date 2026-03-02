---
sidebar_position: 8
title: "Managing the OnRamp Repository"
---

# Managing the OnRamp Repository

Aether-webd clones and manages the [aether-onramp](https://github.com/opennetworkinglab/aether-onramp) repository locally. Component actions execute `make` targets from this repository, and configuration files (`vars/main.yml`) live inside it.

## Check repository status

```bash
curl http://localhost:8186/api/v1/onramp/repo
```

Example response:

```json
{
  "cloned": true,
  "dir": "/opt/aether-onramp",
  "repo_url": "https://github.com/opennetworkinglab/aether-onramp.git",
  "version": "main",
  "commit": "abc123def456...",
  "branch": "main",
  "tag": "",
  "dirty": false
}
```

### Status fields

| Field | Description |
|-------|-------------|
| `cloned` | Whether a valid `.git` directory exists |
| `dir` | Filesystem path of the local clone |
| `repo_url` | Git clone URL |
| `version` | Configured version pin (tag, branch, or commit) |
| `commit` | Full SHA of HEAD (omitted when not cloned) |
| `branch` | Current branch name, or `HEAD` when detached |
| `tag` | Tag pointing at HEAD, if any |
| `dirty` | `true` when uncommitted changes exist in the working tree |
| `error` | Error message from the last refresh attempt (omitted on success) |

## Refresh or clone the repository

```bash
curl -X POST http://localhost:8186/api/v1/onramp/repo/refresh
```

This endpoint performs several steps:

1. Clones the repository if it does not exist locally.
2. Checks out the configured version (tag, branch, or commit).
3. Validates that `Makefile` and `vars/main.yml` are present.

If the local clone is corrupted or in a bad state, the refresh endpoint re-clones the repository from scratch.

## Auto-clone on startup

When aether-webd starts, it automatically runs the equivalent of a refresh: if the repository is absent, it is cloned; if it exists, the configured version is checked out and validated. If this initial setup fails, the server starts in degraded mode -- endpoints that depend on the repository return errors until a successful `POST /api/v1/onramp/repo/refresh`.

## Version pinning

Lock the repository to a specific tag, branch, or commit SHA:

```bash
aether-webd --onramp-version v2.1.0
```

Examples:

```bash
# Pin to a release tag
aether-webd --onramp-version v2.1.0

# Pin to a branch
aether-webd --onramp-version main

# Pin to a specific commit
aether-webd --onramp-version abc123def456
```

The pinned version is checked out during startup and on each `POST /api/v1/onramp/repo/refresh`.

## Custom repository directory

By default, the OnRamp repository is cloned to a directory managed by aether-webd. Override this with:

```bash
aether-webd --onramp-dir /opt/aether-onramp
```

This is useful when the repository is already cloned and managed externally, or when specific filesystem permissions are required.

## Recovery from corruption

If the repository enters a bad state (e.g., interrupted clone, manual edits causing conflicts), use the refresh endpoint to recover:

```bash
# Check current status
curl http://localhost:8186/api/v1/onramp/repo

# Re-clone and reset
curl -X POST http://localhost:8186/api/v1/onramp/repo/refresh

# Verify recovery
curl http://localhost:8186/api/v1/onramp/repo
```

The `dirty` field in the status response indicates whether there are uncommitted changes. A dirty repository may produce unexpected behavior during component actions.

## Workflow

1. Start aether-webd with the desired `--onramp-version` (auto-clones on startup).
2. Verify with `GET /api/v1/onramp/repo`.
3. If switching versions, update the `--onramp-version` flag and restart, or call `POST /api/v1/onramp/repo/refresh`.
4. After confirming the repo is clean and on the correct version, proceed to [configuration](configuration) and [deployment](deploying-components).
