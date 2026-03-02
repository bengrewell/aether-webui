---
sidebar_position: 4
title: "Monitoring"
---

# Monitoring

The system provider exposes both point-in-time system information and time-series metrics. Use the static endpoints to understand the host hardware and the metrics endpoint to track resource usage over time.

## Static system information

These endpoints return current system details and do not change frequently. Fetch them once at startup or on demand.

### CPU

```bash
curl http://localhost:8186/api/v1/system/cpu
```

Returns CPU model, core counts, frequency, cache sizes, and feature flags.

### Memory

```bash
curl http://localhost:8186/api/v1/system/memory
```

Returns physical and swap memory usage.

### Disks

```bash
curl http://localhost:8186/api/v1/system/disks
```

Returns partition list with mount points, filesystem types, and usage.

### Operating system

```bash
curl http://localhost:8186/api/v1/system/os
```

Returns hostname, OS, platform, kernel version, architecture, and uptime.

### Network interfaces

```bash
curl http://localhost:8186/api/v1/system/network/interfaces
```

Returns network interfaces with addresses, MAC, MTU, and flags.

### Network configuration

```bash
curl http://localhost:8186/api/v1/system/network/config
```

Returns DNS servers and search domains from resolv.conf.

### Listening ports

```bash
curl http://localhost:8186/api/v1/system/network/ports
```

Returns TCP/UDP ports in LISTEN state with owning process info.

## Time-series metrics

The metrics endpoint returns historical data points collected by the background metric sampler.

```bash
curl "http://localhost:8186/api/v1/system/metrics?metric=system.cpu.usage_percent&from=2026-03-02T09:00:00Z&to=2026-03-02T10:00:00Z&aggregation=1m"
```

### Query parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `metric` | Yes | Metric name (e.g., `system.cpu.usage_percent`) |
| `from` | No | Start time in RFC 3339 format. Defaults to 1 hour ago. |
| `to` | No | End time in RFC 3339 format. Defaults to now. |
| `labels` | No | Comma-separated `key=val` label filters (e.g., `core=total`) |
| `aggregation` | No | Time bucket size: `raw`, `10s`, `1m`, `5m`, `1h` |

### Available metrics

| Metric | Labels | Unit |
|--------|--------|------|
| `system.cpu.usage_percent` | `core` (total + per-core) | percent |
| `system.memory.used_bytes` | -- | bytes |
| `system.memory.available_bytes` | -- | bytes |
| `system.memory.usage_percent` | -- | percent |
| `system.swap.used_bytes` | -- | bytes |
| `system.disk.used_bytes` | `partition` | bytes |
| `system.disk.usage_percent` | `partition` | percent |
| `system.disk.read_bytes` | `device` | bytes |
| `system.disk.write_bytes` | `device` | bytes |
| `system.net.bytes_sent` | `interface` | bytes |
| `system.net.bytes_recv` | `interface` | bytes |
| `system.load.1m` | -- | -- |
| `system.load.5m` | -- | -- |
| `system.load.15m` | -- | -- |

### Examples

Total CPU usage over the last hour, aggregated to 1-minute buckets:

```bash
curl "http://localhost:8186/api/v1/system/metrics?metric=system.cpu.usage_percent&labels=core%3Dtotal&aggregation=1m"
```

Memory usage percentage over a specific time range:

```bash
curl "http://localhost:8186/api/v1/system/metrics?metric=system.memory.usage_percent&from=2026-03-02T08:00:00Z&to=2026-03-02T09:00:00Z"
```

Network bytes received on eth0, raw resolution:

```bash
curl "http://localhost:8186/api/v1/system/metrics?metric=system.net.bytes_recv&labels=interface%3Deth0&aggregation=raw"
```

## Polling for live dashboards

For a live dashboard, poll the metrics endpoint at a 10--30 second interval. Use the `from` parameter set to your last poll time to avoid fetching duplicate data:

```bash
# Poll every 15 seconds for the latest CPU usage
while true; do
  FROM=$(date -u -d '30 seconds ago' +%Y-%m-%dT%H:%M:%SZ)
  curl -s "http://localhost:8186/api/v1/system/metrics?metric=system.cpu.usage_percent&labels=core%3Dtotal&from=${FROM}&aggregation=10s"
  sleep 15
done
```

## Configuration

The background metric collector is configured with server flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--metrics-interval` | `10s` | How often metrics are sampled |
| `--metrics-retention` | `24h` | How long samples are kept before pruning |

Lower intervals increase storage usage; higher retention extends the queryable time window. See the [CLI reference](../reference/cli) for all server flags.
