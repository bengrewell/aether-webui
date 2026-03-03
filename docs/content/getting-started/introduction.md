---
sidebar_position: 1
title: Introduction
---

# Introduction

Aether WebUI is a REST API backend service that manages [Aether OnRamp](https://github.com/opennetworkinglab/aether-onramp) 5G network deployments. It wraps the OnRamp Make/Ansible toolchain in a JSON API, giving operators a programmatic way to deploy and manage private 5G networks without running shell commands directly on the host.

Through its API, Aether WebUI manages:

- **Kubernetes clusters** -- RKE2 cluster provisioning and lifecycle
- **5G Core (SD-Core)** -- the mobile core network that handles subscriber authentication, session management, and data plane routing
- **Radio Access Network (RAN)** -- simulated and physical gNBs via srsRAN, UERANSIM, OpenAirInterface, and gNBSim
- **Supporting components** -- Aether Management Platform, SD-RAN, O-RAN SC RIC, and Non-3GPP Interworking Function
- **Host system monitoring** -- CPU, memory, disk, and network metrics

All operations are exposed as REST/JSON endpoints at `http://host:8186/api/v1/`.

## Who this guide is for

This tutorial is written for network operators who need to deploy a private 5G network using Aether OnRamp. It assumes basic familiarity with Linux system administration and REST APIs, but no prior experience with Aether or 5G network components.

## What you will accomplish

By the end of this tutorial, you will have:

1. Installed the Aether WebUI service on your host
2. Run preflight checks to verify the host meets all prerequisites
3. Deployed a Kubernetes cluster via the API
4. Deployed the 5G Core network (SD-Core)
5. Verified that all components are running and healthy

## Prerequisites

Before starting, ensure your environment meets the following requirements:

- **Operating system:** Ubuntu 22.04 or later
- **Privileges:** `sudo` access on the target host
- **Networking:** Outbound internet connectivity (the installer downloads binaries and container images)
- **Hardware:** At minimum, 4 CPU cores, 8 GB RAM, and 50 GB disk (a single-node lab deployment)

After installation, the [preflight checks](first-deployment#step-1-run-preflight-checks) automatically verify that the host has the required tools (`make`, `ansible`), SSH configuration, and user accounts. Any missing prerequisites can be fixed with a single API call.

## How the tutorial is structured

Each page builds on the previous one:

1. **Introduction** (this page) -- overview and prerequisites
2. [Installation](installation) -- install the service and verify it is running
3. [First Deployment](first-deployment) -- run preflight checks, then deploy Kubernetes and the 5G Core
4. [Verifying Your Deployment](verifying) -- confirm everything is working
5. [Next Steps](next-steps) -- where to go from here
