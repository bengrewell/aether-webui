---
sidebar_position: 8
title: "Managing the OnRamp Repository"
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Managing the OnRamp Repository

Aether-webd clones and manages the [aether-onramp](https://github.com/opennetworkinglab/aether-onramp) repository locally. Component actions execute `make` targets from this repository, and configuration files (`vars/main.yml`) live inside it.

## Check repository status

<Tabs>
  <TabItem value="ui" label="Web UI" default>
    The dashboard header shows the OnRamp repository status including the current version, commit SHA, branch name, and dirty state. A green indicator confirms a clean, cloned repository; a yellow indicator signals uncommitted changes or a missing clone.
  </TabItem>
  <TabItem value="api" label="API">

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

  </TabItem>
</Tabs>

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

<Tabs>
  <TabItem value="ui" label="Web UI" default>
    Click **Refresh Repository** in the repository status panel. Progress is shown inline as the clone or checkout proceeds. The status indicator updates when the operation completes.
  </TabItem>
  <TabItem value="api" label="API">

```bash
curl -X POST http://localhost:8186/api/v1/onramp/repo/refresh
```

  </TabItem>
</Tabs>

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

If the repository enters a bad state (e.g., interrupted clone, manual edits causing conflicts), use the refresh operation to recover.

<Tabs>
  <TabItem value="ui" label="Web UI" default>
    If the repository shows errors, click **Refresh Repository** to re-clone and reset. The status panel updates to reflect the recovery progress. Verify that the status indicator returns to green and the `dirty` flag is cleared.
  </TabItem>
  <TabItem value="api" label="API">

```bash
# Check current status
curl http://localhost:8186/api/v1/onramp/repo

# Re-clone and reset
curl -X POST http://localhost:8186/api/v1/onramp/repo/refresh

# Verify recovery
curl http://localhost:8186/api/v1/onramp/repo
```

  </TabItem>
</Tabs>

The `dirty` field in the status response indicates whether there are uncommitted changes. A dirty repository may produce unexpected behavior during component actions.

## Workflow

1. Start aether-webd with the desired `--onramp-version` (auto-clones on startup).
2. Verify with `GET /api/v1/onramp/repo`.
3. If switching versions, update the `--onramp-version` flag and restart, or call `POST /api/v1/onramp/repo/refresh`.
4. After confirming the repo is clean and on the correct version, proceed to [configuration](configuration) and [deployment](deploying-components).
