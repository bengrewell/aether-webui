# OnRamp Provider

The `onramp` provider wraps the [aether-onramp](https://github.com/opennetworkinglab/aether-onramp)
Make/Ansible toolchain and exposes it as a REST API. It manages the lifecycle of
the onramp git repository, dispatches `make` targets asynchronously, and provides
read/write access to the `vars/main.yml` configuration file.

On startup the provider calls `ensureRepo`, which clones the repository if
absent, checks out the configured version, and validates that `Makefile` and
`vars/main.yml` are present. If any step fails the provider logs the error and
continues in degraded mode — endpoints still respond, but operations that require
the repo (e.g. executing actions, reading config) return errors until the
repo is repaired via `POST /api/v1/onramp/repo/refresh`.

## Architecture

`OnRamp` embeds `provider.Base` and implements the `provider.Provider` interface.
All 12 endpoints are registered at construction time via `provider.Register`.

Task execution is managed by a shared `taskrunner.Runner` configured with
`MaxConcurrent: 1`. The runner handles task lifecycle (creation, output
streaming, completion) and enforces the single-task constraint — a second `POST`
to any component action while a task is running returns `409 Conflict`.

Configuration round-trips are handled by `readVarsFile` / `writeVarsFile` (YAML
unmarshal/marshal). The `mergeConfig` helper performs a section-level merge:
non-nil sections in the PATCH body overwrite the corresponding section in
`vars/main.yml`; nil sections are left untouched.

## Endpoints

### Repository

| Method | Path | Operation ID | Description |
|--------|------|--------------|-------------|
| `GET` | `/api/v1/onramp/repo` | `onramp-get-repo-status` | Clone status, current commit, branch, tag, and dirty state |
| `POST` | `/api/v1/onramp/repo/refresh` | `onramp-refresh-repo` | Clone if missing, checkout pinned version, validate |

**Repo status fields:**

| Field | Type | Description |
|-------|------|-------------|
| `cloned` | bool | Whether a valid `.git` directory exists |
| `dir` | string | Filesystem path of the repository |
| `repo_url` | string | Clone URL |
| `version` | string | Configured version pin (tag, branch, or commit) |
| `commit` | string | Full SHA of `HEAD` (omitted when not cloned) |
| `branch` | string | Current branch name, or `HEAD` when detached |
| `tag` | string | Tag pointing at `HEAD`, if any |
| `dirty` | bool | `true` when uncommitted changes exist |
| `error` | string | Error message from the last refresh attempt (omitted on success) |

### Components

| Method | Path | Operation ID | Description |
|--------|------|--------------|-------------|
| `GET` | `/api/v1/onramp/components` | `onramp-list-components` | All components and their actions |
| `GET` | `/api/v1/onramp/components/{component}` | `onramp-get-component` | Single component by name |
| `POST` | `/api/v1/onramp/components/{component}/{action}` | `onramp-execute-action` | Run a make target (async) |

`POST /api/v1/onramp/components/{component}/{action}` returns the newly created
task immediately. Poll `GET /api/v1/onramp/tasks/{id}` to track progress and
retrieve output.

**Component registry** (static, derived from the OnRamp Makefile):

| Component | Description | Actions |
|-----------|-------------|---------|
| `k8s` | Kubernetes (RKE2) cluster lifecycle | `install`, `uninstall` |
| `5gc` | 5G core network (SD-Core) | `install`, `uninstall`, `reset` |
| `4gc` | 4G core network | `install`, `uninstall`, `reset` |
| `gnbsim` | gNBSim simulated RAN | `install`, `uninstall`, `run` |
| `amp` | Aether Management Platform | `install`, `uninstall` |
| `sdran` | SD-RAN intelligent RAN controller | `install`, `uninstall` |
| `ueransim` | UERANSIM UE and gNB simulator | `install`, `uninstall`, `run`, `stop` |
| `oai` | OpenAirInterface RAN | `gnb-install`, `gnb-uninstall`, `uesim-start`, `uesim-stop` |
| `srsran` | srsRAN Project RAN | `gnb-install`, `gnb-uninstall`, `uesim-start`, `uesim-stop` |
| `oscric` | O-RAN SC near-RT RIC | `ric-install`, `ric-uninstall` |
| `n3iwf` | Non-3GPP Interworking Function | `install`, `uninstall` |
| `cluster` | Cluster-level operations | `pingall`, `install`, `uninstall`, `add-upfs`, `remove-upfs` |

### Tasks

| Method | Path | Operation ID | Description |
|--------|------|--------------|-------------|
| `GET` | `/api/v1/onramp/tasks` | `onramp-list-tasks` | All tasks, most recent first |
| `GET` | `/api/v1/onramp/tasks/{id}` | `onramp-get-task` | Single task by UUID (supports incremental output) |

The task detail endpoint accepts an `offset` query parameter (default `0`) for
incremental output reads. The response includes an `output_offset` field
indicating the byte position to use as `offset` on the next request, enabling
efficient polling of long-running tasks without re-fetching earlier output.

**Task fields:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | UUID assigned at creation |
| `component` | string | Component name (e.g. `5gc`) |
| `action` | string | Action name (e.g. `install`) |
| `target` | string | Make target executed (e.g. `aether-5gc-install`) |
| `status` | string | `running`, `succeeded`, or `failed` |
| `started_at` | string | RFC 3339 timestamp |
| `finished_at` | string | RFC 3339 timestamp (omitted while running) |
| `exit_code` | int | Process exit code (omitted while running; `-1` on exec error) |
| `output` | string | Combined stdout/stderr (or incremental chunk when `offset` is used) |
| `output_offset` | int | Byte offset for the next incremental read |

### Configuration

| Method | Path | Operation ID | Description |
|--------|------|--------------|-------------|
| `GET` | `/api/v1/onramp/config` | `onramp-get-config` | Parse and return `vars/main.yml` as JSON |
| `PATCH` | `/api/v1/onramp/config` | `onramp-patch-config` | Merge partial update into `vars/main.yml` |
| `GET` | `/api/v1/onramp/config/profiles` | `onramp-list-profiles` | Names of `vars/main-*.yml` profile files |
| `GET` | `/api/v1/onramp/config/profiles/{name}` | `onramp-get-profile` | Parse a named profile as JSON |
| `POST` | `/api/v1/onramp/config/profiles/{name}/activate` | `onramp-activate-profile` | Copy profile to `vars/main.yml` |

PATCH performs a **section-level** merge: supply only the top-level sections you
want to change. Omitting a section leaves it unchanged. For example, to update
only the data interface used by the core:

```json
{"core": {"data_iface": "eth1"}}
```

Profile names are derived from filenames: `vars/main-gnbsim.yml` → profile name
`gnbsim`. Activating a profile overwrites `vars/main.yml` atomically via
`os.Create` + `io.Copy`.

## Adding a new endpoint

1. Define input/output types in `types.go` (use `struct{}` for empty input).
2. Create an `endpoint.Endpoint[I, O]` with a `Descriptor` and handler method.
3. Call `provider.Register(o.Base, ep)` inside `NewProvider`.
4. Add the handler implementation to `handlers.go`.
