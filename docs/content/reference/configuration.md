---
sidebar_position: 7
title: "OnRamp Configuration"
---

# OnRamp Configuration

The OnRamp deployment toolchain is configured through a YAML file at `vars/main.yml` inside the aether-onramp directory. This file controls Kubernetes settings, 5G core parameters, RAN simulator configurations, and all other deployment options.

## File Location

```
{onramp-dir}/vars/main.yml
```

The default OnRamp directory is `{data-dir}/aether-onramp` (typically `/var/lib/aether-webd/aether-onramp`). This can be overridden with the `--onramp-dir` [CLI flag](./cli.md).

## Top-Level Sections

The configuration file contains up to 9 top-level sections. Each section is optional -- omitted sections use OnRamp's built-in defaults.

| Section | Description |
|---------|-------------|
| `k8s` | Kubernetes (RKE2) cluster settings |
| `core` | 5G core network (SD-Core) settings |
| `gnbsim` | gNBSim simulated RAN settings |
| `amp` | Aether Management Platform settings |
| `sdran` | SD-RAN controller settings |
| `ueransim` | UERANSIM simulator settings |
| `oai` | OpenAirInterface RAN settings |
| `srsran` | srsRAN Project settings |
| `n3iwf` | Non-3GPP Interworking Function settings |

## K8s Section

Controls the RKE2 Kubernetes distribution and Helm settings.

```yaml
k8s:
  rke2:
    version: "v1.28.2+rke2r1"
    config:
      token: "my-cluster-token"
      port: 9345
      params_file:
        master: "config/server.yaml"
        worker: "config/agent.yaml"
  helm:
    version: "v3.14.0"
```

| Field | Type | Description |
|-------|------|-------------|
| `k8s.rke2.version` | string | RKE2 release version |
| `k8s.rke2.config.token` | string | Shared secret for joining nodes to the cluster |
| `k8s.rke2.config.port` | int | Supervisor port for node registration |
| `k8s.rke2.config.params_file.master` | string | Path to server (control-plane) config file |
| `k8s.rke2.config.params_file.worker` | string | Path to agent (worker) config file |
| `k8s.helm.version` | string | Helm version to install |

## Core Section

Controls the SD-Core 5G core network deployment.

```yaml
core:
  standalone: true
  data_iface: "eth0"
  values_file: "sd-core-5g-values.yaml"
  ran_subnet: "192.168.70.0/24"
  helm:
    chart_ref: "aether/sd-core"
    chart_version: "0.12.8"
  upf:
    access_subnet: "192.168.252.0/24"
    core_subnet: "192.168.250.0/24"
    mode: "af_packet"
    default_upf:
      ip:
        access: "192.168.252.3/24"
        core: "192.168.250.3/24"
      ue_ip_pool: "172.250.0.0/16"
  amf:
    ip: "10.42.0.100"
```

| Field | Type | Description |
|-------|------|-------------|
| `core.standalone` | bool | Whether the core runs as a standalone (single-node) deployment |
| `core.data_iface` | string | Host network interface for data-plane traffic |
| `core.values_file` | string | Helm values file for SD-Core |
| `core.ran_subnet` | string | Subnet for RAN-side traffic |
| `core.upf.access_subnet` | string | UPF access network subnet |
| `core.upf.core_subnet` | string | UPF core network subnet |
| `core.upf.mode` | string | UPF data-plane mode (e.g., `af_packet`) |
| `core.upf.default_upf` | object | Default UPF instance configuration |
| `core.upf.additional_upfs` | map | Additional UPF instances (keyed by name) |
| `core.amf.ip` | string | AMF IP address |

## GNBSim Section

Controls the gNBSim simulated RAN.

```yaml
gnbsim:
  docker:
    container:
      image: "omecproject/5gc-gnbsim:main-latest"
      prefix: "gnbsim"
      count: 1
    network:
      macvlan:
        name: "gnbnet"
  router:
    data_iface: "eth0"
    macvlan:
      subnet_prefix: "192.168.251"
  servers:
    1: ["server1"]
```

## UERANSIM Section

Controls the UERANSIM UE and gNB simulator.

```yaml
ueransim:
  gnb:
    ip: "192.168.70.132"
  servers:
    1:
      gnb: "config/gnb.yaml"
      ue: "config/ue.yaml"
```

## OAI Section

Controls the OpenAirInterface RAN deployment.

```yaml
oai:
  docker:
    container:
      gnb_image: "oaisoftwarealliance/oai-gnb:develop"
      ue_image: "oaisoftwarealliance/oai-nr-ue:develop"
    network:
      data_iface: "eth0"
      name: "oai-net"
      subnet: "192.168.72.0/24"
  simulation: true
  servers:
    1:
      gnb_conf: "config/gnb.sa.band78.106prb.conf"
      gnb_ip: "192.168.72.10"
      ue_conf: "config/ue.conf"
```

## srsRAN Section

Controls the srsRAN Project RAN deployment.

```yaml
srsran:
  docker:
    container:
      gnb_image: "softwareradiosystems/srsran-project:latest"
      ue_image: "softwareradiosystems/srsue:latest"
    network:
      name: "srsran-net"
  simulation: true
  servers:
    1:
      gnb_ip: "192.168.72.20"
      gnb_conf: "config/gnb_zmq.yaml"
      ue_conf: "config/ue_zmq.conf"
```

## N3IWF Section

Controls the Non-3GPP Interworking Function.

```yaml
n3iwf:
  docker:
    image: "free5gc/n3iwf:v3.3.0"
    network:
      name: "n3iwf-net"
  servers:
    1:
      conf_file: "config/n3iwf.yaml"
      n3iwf_ip: "192.168.73.10"
      n2_ip: "10.42.0.110"
      n3_ip: "10.42.0.111"
      nwu_ip: "192.168.73.11"
```

## Profiles

Profiles are pre-defined configuration files stored alongside the active config. Each profile is a complete `vars/main.yml` tuned for a specific deployment scenario.

### File Naming

Profiles follow the naming convention:

```
vars/main-{name}.yml
```

### Standard Profiles

| Profile | File | Description |
|---------|------|-------------|
| `gnbsim` | `vars/main-gnbsim.yml` | Single-node deployment with gNBSim |
| `oai` | `vars/main-oai.yml` | Deployment with OpenAirInterface RAN |
| `srsran` | `vars/main-srsran.yml` | Deployment with srsRAN Project |
| `ueransim` | `vars/main-ueransim.yml` | Deployment with UERANSIM simulator |

### Managing Profiles via API

**List available profiles:**

```bash
curl http://localhost:8186/api/v1/onramp/config/profiles
```

**Read a profile without activating it:**

```bash
curl http://localhost:8186/api/v1/onramp/config/profiles/srsran
```

**Activate a profile** (copies it to `vars/main.yml`):

```bash
curl -X POST http://localhost:8186/api/v1/onramp/config/profiles/srsran/activate
```

**Read the current active config:**

```bash
curl http://localhost:8186/api/v1/onramp/config
```

**Patch a section of the active config** (section-level merge):

```bash
curl -X PATCH http://localhost:8186/api/v1/onramp/config \
  -H "Content-Type: application/json" \
  -d '{
    "core": {
      "standalone": false,
      "data_iface": "ens192",
      "values_file": "sd-core-5g-values.yaml"
    }
  }'
```

### Merge Behavior

The `PATCH /api/v1/onramp/config` endpoint performs a **section-level merge**:

- Each top-level key in the request body (`k8s`, `core`, `gnbsim`, etc.) **replaces** the corresponding section entirely.
- Sections **not included** in the request body are left unchanged.
- This is not a deep merge. Submitting a partial `core` section replaces the entire `core` section, not just the fields specified.

**Example:** If the current config has both `k8s` and `core` sections, and the PATCH request only includes `core`, the `k8s` section remains untouched while `core` is fully replaced.

### Workflow

A typical configuration workflow:

1. **Activate a profile** as a starting point: `POST /api/v1/onramp/config/profiles/srsran/activate`
2. **Read the active config** to review: `GET /api/v1/onramp/config`
3. **Patch specific sections** as needed: `PATCH /api/v1/onramp/config`
4. **Deploy components** using the active config: `POST /api/v1/onramp/components/k8s/install`

For a step-by-step guide, see [Deploying Components](../guides/deploying-components).
