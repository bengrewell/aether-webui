---
sidebar_position: 4
title: "System Endpoints"
---

# System Endpoints

The system provider exposes 8 endpoints for querying host hardware, OS, network, and time-series metrics. All endpoints accept `GET` requests.

| Endpoint | Description |
|----------|-------------|
| [`GET /api/v1/system/cpu`](#get-cpu) | CPU model, cores, frequency, cache, flags |
| [`GET /api/v1/system/memory`](#get-memory) | Physical and swap memory usage |
| [`GET /api/v1/system/disks`](#get-disks) | Disk partitions with usage |
| [`GET /api/v1/system/os`](#get-os) | Hostname, OS, platform, kernel, uptime |
| [`GET /api/v1/system/network/interfaces`](#get-network-interfaces) | Network interfaces with addresses |
| [`GET /api/v1/system/network/config`](#get-network-config) | DNS servers and search domains |
| [`GET /api/v1/system/network/ports`](#get-listening-ports) | TCP/UDP ports in LISTEN state |
| [`GET /api/v1/system/metrics`](#get-metrics) | Time-series metrics query |

---

## GET CPU

```
GET /api/v1/system/cpu
```

Returns CPU model, core counts, base frequency, cache size, and feature flags.

### Response Schema

| Field | Type | Description |
|-------|------|-------------|
| `model` | string | CPU model name |
| `physical_cores` | int | Number of physical CPU cores |
| `logical_cores` | int | Number of logical CPU cores (includes hyperthreading) |
| `frequency_mhz` | float64 | Base clock frequency in MHz |
| `cache_size_kb` | int32 | CPU cache size in kilobytes |
| `flags` | string[] | CPU feature flags |

### Example

```bash
curl http://localhost:8186/api/v1/system/cpu
```

```json
{
  "model": "Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz",
  "physical_cores": 8,
  "logical_cores": 16,
  "frequency_mhz": 3800,
  "cache_size_kb": 16384,
  "flags": ["sse4_2", "avx2", "aes", "vmx"]
}
```

---

## GET Memory

```
GET /api/v1/system/memory
```

Returns physical and swap memory usage statistics.

### Response Schema

| Field | Type | Description |
|-------|------|-------------|
| `total_bytes` | uint64 | Total physical memory in bytes |
| `available_bytes` | uint64 | Available physical memory in bytes |
| `used_bytes` | uint64 | Used physical memory in bytes |
| `usage_percent` | float64 | Physical memory usage as a percentage |
| `swap_total_bytes` | uint64 | Total swap space in bytes |
| `swap_used_bytes` | uint64 | Used swap space in bytes |
| `swap_percent` | float64 | Swap usage as a percentage |

### Example

```bash
curl http://localhost:8186/api/v1/system/memory
```

```json
{
  "total_bytes": 34359738368,
  "available_bytes": 17179869184,
  "used_bytes": 17179869184,
  "usage_percent": 50.0,
  "swap_total_bytes": 8589934592,
  "swap_used_bytes": 0,
  "swap_percent": 0.0
}
```

---

## GET Disks

```
GET /api/v1/system/disks
```

Returns a list of disk partitions with device path, mount point, filesystem type, and usage.

### Response Schema

The response body contains a `partitions` array. Each element:

| Field | Type | Description |
|-------|------|-------------|
| `device` | string | Device path (e.g., `/dev/sda1`) |
| `mountpoint` | string | Filesystem mount point |
| `fs_type` | string | Filesystem type (e.g., `ext4`, `xfs`) |
| `total_bytes` | uint64 | Total partition size in bytes |
| `used_bytes` | uint64 | Used space in bytes |
| `free_bytes` | uint64 | Free space in bytes |
| `usage_percent` | float64 | Disk usage as a percentage |

### Example

```bash
curl http://localhost:8186/api/v1/system/disks
```

```json
{
  "partitions": [
    {
      "device": "/dev/sda1",
      "mountpoint": "/",
      "fs_type": "ext4",
      "total_bytes": 512110190592,
      "used_bytes": 128027547648,
      "free_bytes": 384082642944,
      "usage_percent": 25.0
    }
  ]
}
```

---

## GET OS

```
GET /api/v1/system/os
```

Returns hostname, operating system, platform, kernel version, architecture, and uptime.

### Response Schema

| Field | Type | Description |
|-------|------|-------------|
| `hostname` | string | System hostname |
| `os` | string | Operating system name (e.g., `linux`) |
| `platform` | string | OS distribution or platform (e.g., `ubuntu`) |
| `platform_version` | string | Platform version (e.g., `22.04`) |
| `kernel_version` | string | Kernel version string |
| `kernel_arch` | string | Kernel architecture (e.g., `x86_64`) |
| `uptime_seconds` | uint64 | System uptime in seconds |

### Example

```bash
curl http://localhost:8186/api/v1/system/os
```

```json
{
  "hostname": "aether-node-01",
  "os": "linux",
  "platform": "ubuntu",
  "platform_version": "22.04",
  "kernel_version": "6.8.0-100-generic",
  "kernel_arch": "x86_64",
  "uptime_seconds": 86400
}
```

---

## GET Network Interfaces

```
GET /api/v1/system/network/interfaces
```

Returns all network interfaces with hardware address, MTU, flags, and assigned IP addresses.

### Response Schema

The response body is an array. Each element:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Interface name (e.g., `eth0`) |
| `mac` | string | Hardware MAC address |
| `mtu` | int | Maximum transmission unit |
| `flags` | string[] | Interface flags (e.g., `up`, `broadcast`, `multicast`) |
| `addresses` | string[] | Assigned IP addresses with CIDR prefix (e.g., `192.168.1.10/24`) |

### Example

```bash
curl http://localhost:8186/api/v1/system/network/interfaces
```

```json
[
  {
    "name": "eth0",
    "mac": "00:1a:2b:3c:4d:5e",
    "mtu": 1500,
    "flags": ["up", "broadcast", "multicast"],
    "addresses": ["192.168.1.10/24", "fe80::1/64"]
  },
  {
    "name": "lo",
    "mac": "",
    "mtu": 65536,
    "flags": ["up", "loopback"],
    "addresses": ["127.0.0.1/8", "::1/128"]
  }
]
```

---

## GET Network Config

```
GET /api/v1/system/network/config
```

Returns DNS servers and search domains parsed from `/etc/resolv.conf`.

### Response Schema

| Field | Type | Description |
|-------|------|-------------|
| `dns_servers` | string[] | Configured DNS nameservers |
| `search_domains` | string[] | DNS search domains |

### Example

```bash
curl http://localhost:8186/api/v1/system/network/config
```

```json
{
  "dns_servers": ["8.8.8.8", "8.8.4.4"],
  "search_domains": ["example.com"]
}
```

---

## GET Listening Ports

```
GET /api/v1/system/network/ports
```

Returns TCP and UDP sockets in `LISTEN` state with the owning process information.

### Response Schema

The response body is an array. Each element:

| Field | Type | Description |
|-------|------|-------------|
| `protocol` | string | Network protocol (`tcp` or `udp`) |
| `local_addr` | string | Local bind address |
| `local_port` | uint32 | Local port number |
| `pid` | int32 | Process ID of the listener |
| `process_name` | string | Name of the listening process |
| `state` | string | Connection state (always `LISTEN`) |

### Example

```bash
curl http://localhost:8186/api/v1/system/network/ports
```

```json
[
  {
    "protocol": "tcp",
    "local_addr": "0.0.0.0",
    "local_port": 8186,
    "pid": 1234,
    "process_name": "aether-webd",
    "state": "LISTEN"
  },
  {
    "protocol": "tcp",
    "local_addr": "0.0.0.0",
    "local_port": 22,
    "pid": 890,
    "process_name": "sshd",
    "state": "LISTEN"
  }
]
```

---

## GET Metrics

```
GET /api/v1/system/metrics
```

Queries time-series metric data with optional time range, label filtering, and time-bucket aggregation.

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `metric` | string | Yes | - | Metric name to query (e.g., `system.cpu.usage_percent`) |
| `from` | string | No | 1 hour ago | Start time (RFC 3339) |
| `to` | string | No | now | End time (RFC 3339) |
| `labels` | string | No | - | Comma-separated `key=val` label filters (e.g., `cpu=total`) |
| `aggregation` | string | No | `raw` | Time bucket aggregation: `raw`, `10s`, `1m`, `5m`, `1h` |

### Response Schema

The response body contains a `series` array. Each element:

| Field | Type | Description |
|-------|------|-------------|
| `metric` | string | Metric name |
| `labels` | object | Label key-value pairs identifying the series |
| `points` | array | Time-ordered data points |

Each `points` element:

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string | Sample timestamp (RFC 3339) |
| `value` | float64 | Metric value at this timestamp |

### Example Request

```bash
curl "http://localhost:8186/api/v1/system/metrics?metric=system.cpu.usage_percent&from=2026-02-18T21:00:00Z&to=2026-02-18T21:30:00Z&aggregation=1m"
```

### Example Response

```json
{
  "series": [
    {
      "metric": "system.cpu.usage_percent",
      "labels": {
        "cpu": "total"
      },
      "points": [
        {
          "timestamp": "2026-02-18T21:00:00Z",
          "value": 23.5
        },
        {
          "timestamp": "2026-02-18T21:01:00Z",
          "value": 31.2
        },
        {
          "timestamp": "2026-02-18T21:02:00Z",
          "value": 18.7
        }
      ]
    }
  ]
}
```

### Aggregation Modes

| Value | Behavior |
|-------|----------|
| `raw` | Return every recorded data point (no aggregation) |
| `10s` | Average values into 10-second buckets |
| `1m` | Average values into 1-minute buckets |
| `5m` | Average values into 5-minute buckets |
| `1h` | Average values into 1-hour buckets |

Aggregation reduces the number of data points returned for large time ranges, improving response size and rendering performance. For dashboards, `1m` or `5m` provides a good balance between detail and performance.
