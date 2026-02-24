# System Provider

The system provider collects host-level metrics and exposes them through REST endpoints. It provides both point-in-time system information (CPU specs, memory, disks, OS, network) and a time-series metrics query interface backed by the store.

## Background Collector

On `Start()`, the provider spawns a background goroutine that samples system metrics at a configurable interval (default `10s`, set via `--metrics-interval`). Collected samples are written to the store, which enforces a retention window (default `24h`, set via `--metrics-retention`) and prunes expired data automatically.

The collector runs until `Stop()` is called, which cancels the context and waits for the goroutine to exit.

**Metrics collected:**

| Metric | Labels | Unit |
|--------|--------|------|
| `system.cpu.usage_percent` | `core` (total + per-core) | percent |
| `system.memory.used_bytes` | — | bytes |
| `system.memory.available_bytes` | — | bytes |
| `system.memory.usage_percent` | — | percent |
| `system.swap.used_bytes` | — | bytes |
| `system.disk.used_bytes` | `partition` | bytes |
| `system.disk.usage_percent` | `partition` | percent |
| `system.disk.read_bytes` | `device` | bytes |
| `system.disk.write_bytes` | `device` | bytes |
| `system.net.bytes_sent` | `interface` | bytes |
| `system.net.bytes_recv` | `interface` | bytes |
| `system.load.1m` | — | — |
| `system.load.5m` | — | — |
| `system.load.15m` | — | — |

## Endpoints

### System information

| Method | Path | Operation ID | Description |
|--------|------|--------------|-------------|
| `GET` | `/api/v1/system/cpu` | `system-cpu` | CPU model, core counts, frequency, cache sizes, and feature flags |
| `GET` | `/api/v1/system/memory` | `system-memory` | Physical and swap memory usage |
| `GET` | `/api/v1/system/disks` | `system-disks` | Partition list with mount points, filesystem types, and usage |
| `GET` | `/api/v1/system/os` | `system-os` | Hostname, OS, platform, kernel version, architecture, and uptime |

### Network

| Method | Path | Operation ID | Description |
|--------|------|--------------|-------------|
| `GET` | `/api/v1/system/network/interfaces` | `system-network-interfaces` | Network interfaces with addresses, MAC, MTU, and flags |
| `GET` | `/api/v1/system/network/config` | `system-network-config` | DNS servers and search domains from resolv.conf |
| `GET` | `/api/v1/system/network/ports` | `system-network-ports` | TCP/UDP ports in LISTEN state with owning process info |

### Metrics

| Method | Path | Operation ID | Description |
|--------|------|--------------|-------------|
| `GET` | `/api/v1/system/metrics` | `system-metrics` | Query time-series metric data with optional time range, label filtering, and aggregation |

The metrics endpoint accepts query parameters for time range (`start`, `end`), metric name filtering, and label selectors. Refer to `/docs` for the full query parameter schema.
