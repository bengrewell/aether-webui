---
sidebar_position: 5
title: "Node Endpoints"
---

# Node Endpoints

The nodes provider exposes 5 CRUD endpoints for managing cluster nodes and their role assignments. Nodes represent hosts in the Ansible inventory used by OnRamp deployments.

| Endpoint | Description |
|----------|-------------|
| [`GET /api/v1/nodes`](#list-nodes) | List all managed nodes |
| [`GET /api/v1/nodes/{id}`](#get-node) | Get a single node |
| [`POST /api/v1/nodes`](#create-node) | Create a new node |
| [`PUT /api/v1/nodes/{id}`](#update-node) | Partial update a node |
| [`DELETE /api/v1/nodes/{id}`](#delete-node) | Delete a node |

## Security

Credentials (password, sudo password, SSH key) are **encrypted at rest** using AES-256-GCM. The API never returns secret values. Instead, boolean presence flags indicate whether each credential is set:

- `has_password` -- SSH password is stored
- `has_sudo_password` -- sudo password is stored
- `has_ssh_key` -- SSH private key is stored

To clear a credential, set it to an empty string (`""`) in an update request.

## ManagedNode Schema

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique node identifier (UUID-like hex string) |
| `name` | string | Unique node name (used as Ansible inventory hostname) |
| `ansible_host` | string | IP address or hostname for SSH connections |
| `ansible_user` | string | SSH username |
| `has_password` | bool | Whether an SSH password is stored |
| `has_sudo_password` | bool | Whether a sudo password is stored |
| `has_ssh_key` | bool | Whether an SSH private key is stored |
| `roles` | string[] | Assigned roles |
| `created_at` | string | Creation timestamp (RFC 3339) |
| `updated_at` | string | Last update timestamp (RFC 3339) |

## Valid Roles

| Role | Description |
|------|-------------|
| `master` | Kubernetes control-plane node |
| `worker` | Kubernetes worker node |
| `gnbsim` | gNBSim simulator host |
| `oai` | OpenAirInterface RAN host |
| `ueransim` | UERANSIM simulator host |
| `srsran` | srsRAN Project host |
| `oscric` | O-RAN SC near-RT RIC host |
| `n3iwf` | N3IWF host |

---

## List Nodes

```
GET /api/v1/nodes
```

Returns all managed cluster nodes with role assignments. The list includes secret-presence flags but never exposes actual credentials.

### Example

```bash
curl http://localhost:8186/api/v1/nodes
```

```json
[
  {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "name": "node-01",
    "ansible_host": "192.168.1.10",
    "ansible_user": "ubuntu",
    "has_password": false,
    "has_sudo_password": true,
    "has_ssh_key": true,
    "roles": ["master"],
    "created_at": "2026-02-18T12:00:00Z",
    "updated_at": "2026-02-18T12:00:00Z"
  },
  {
    "id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
    "name": "node-02",
    "ansible_host": "192.168.1.11",
    "ansible_user": "ubuntu",
    "has_password": false,
    "has_sudo_password": true,
    "has_ssh_key": true,
    "roles": ["worker"],
    "created_at": "2026-02-18T12:05:00Z",
    "updated_at": "2026-02-18T12:05:00Z"
  }
]
```

---

## Get Node

```
GET /api/v1/nodes/{id}
```

Returns a single node by its ID.

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Node ID |

### Example

```bash
curl http://localhost:8186/api/v1/nodes/a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "node-01",
  "ansible_host": "192.168.1.10",
  "ansible_user": "ubuntu",
  "has_password": false,
  "has_sudo_password": true,
  "has_ssh_key": true,
  "roles": ["master"],
  "created_at": "2026-02-18T12:00:00Z",
  "updated_at": "2026-02-18T12:00:00Z"
}
```

### Errors

| Status | When |
|--------|------|
| `404` | No node with the given ID |

---

## Create Node

```
POST /api/v1/nodes
```

Creates a new managed node with optional credentials and role assignments.

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique node name (Ansible inventory hostname) |
| `ansible_host` | string | Yes | IP address or hostname for SSH |
| `ansible_user` | string | No | SSH username |
| `password` | string | No | SSH password (stored encrypted) |
| `sudo_password` | string | No | Sudo password (stored encrypted) |
| `ssh_key` | string | No | SSH private key (stored encrypted) |
| `roles` | string[] | No | Role assignments (validated against [valid roles](#valid-roles)) |

### Example

```bash
curl -X POST http://localhost:8186/api/v1/nodes \
  -H "Content-Type: application/json" \
  -d '{
    "name": "worker-03",
    "ansible_host": "192.168.1.13",
    "ansible_user": "ubuntu",
    "sudo_password": "changeme",
    "ssh_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----",
    "roles": ["worker", "srsran"]
  }'
```

```json
{
  "id": "c3d4e5f6-a7b8-9012-cdef-123456789012",
  "name": "worker-03",
  "ansible_host": "192.168.1.13",
  "ansible_user": "ubuntu",
  "has_password": false,
  "has_sudo_password": true,
  "has_ssh_key": true,
  "roles": ["worker", "srsran"],
  "created_at": "2026-02-18T14:30:00Z",
  "updated_at": "2026-02-18T14:30:00Z"
}
```

### Errors

| Status | When |
|--------|------|
| `422` | `name` or `ansible_host` is missing, or a role is invalid |

---

## Update Node

```
PUT /api/v1/nodes/{id}
```

Partial update -- merges non-null fields into the existing node. Fields omitted from the request body are left unchanged.

**Roles:** When `roles` is provided, it **replaces the entire set**. To add a role, include all existing roles plus the new one.

**Credentials:** To clear a credential, set it to an empty string (`""`). Omitting a credential field leaves it unchanged.

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Node ID |

### Request Body

All fields are optional. Only provided fields are updated.

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique node name |
| `ansible_host` | string | IP or hostname for SSH |
| `ansible_user` | string | SSH username |
| `password` | string | SSH password (empty string clears) |
| `sudo_password` | string | Sudo password (empty string clears) |
| `ssh_key` | string | SSH private key (empty string clears) |
| `roles` | string[] | Role assignments (replaces entire set) |

### Example -- Add a role

```bash
curl -X PUT http://localhost:8186/api/v1/nodes/a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
  -H "Content-Type: application/json" \
  -d '{
    "roles": ["master", "oscric"]
  }'
```

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "node-01",
  "ansible_host": "192.168.1.10",
  "ansible_user": "ubuntu",
  "has_password": false,
  "has_sudo_password": true,
  "has_ssh_key": true,
  "roles": ["master", "oscric"],
  "created_at": "2026-02-18T12:00:00Z",
  "updated_at": "2026-02-18T15:00:00Z"
}
```

### Example -- Clear a credential

```bash
curl -X PUT http://localhost:8186/api/v1/nodes/a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
  -H "Content-Type: application/json" \
  -d '{
    "password": ""
  }'
```

### Errors

| Status | When |
|--------|------|
| `404` | No node with the given ID |
| `422` | An invalid role is provided |

---

## Delete Node

```
DELETE /api/v1/nodes/{id}
```

Deletes a node and all its role assignments. This operation cascades to the `node_roles` table.

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Node ID |

### Example

```bash
curl -X DELETE http://localhost:8186/api/v1/nodes/a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

```json
{
  "message": "node a1b2c3d4-e5f6-7890-abcd-ef1234567890 deleted"
}
```

Note: Deleting a node does not automatically update the Ansible inventory file. Use the [inventory sync](./api-onramp.md#sync-inventory) endpoint to regenerate `hosts.ini` after node changes.
