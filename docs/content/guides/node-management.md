---
sidebar_position: 1
title: "Managing Nodes"
---

# Managing Nodes

Nodes represent the machines in your Aether cluster. Each node has an Ansible host address, SSH credentials, and one or more roles that determine which components are deployed to it.

## Create a node

```bash
curl -X POST http://localhost:8186/api/v1/nodes \
  -H "Content-Type: application/json" \
  -d '{
    "name": "worker-1",
    "ansible_host": "192.168.1.10",
    "ansible_user": "ubuntu",
    "password": "s3cret",
    "roles": ["worker"]
  }'
```

The response includes the generated node `id`, which is required for subsequent update and delete operations.

### Valid roles

| Role | Purpose |
|------|---------|
| `master` | Kubernetes control-plane node |
| `worker` | Kubernetes worker node |
| `gnbsim` | gNBSim simulated RAN host |
| `oai` | OpenAirInterface RAN host |
| `ueransim` | UERANSIM simulator host |
| `srsran` | srsRAN Project host |
| `oscric` | O-RAN SC near-RT RIC host |
| `n3iwf` | Non-3GPP Interworking Function host |

A node can have multiple roles. Roles control which Ansible inventory groups the node appears in after an inventory sync.

## List nodes

```bash
curl http://localhost:8186/api/v1/nodes
```

Returns an array of all registered nodes with their current roles and connection details.

## Update a node

```bash
curl -X PUT http://localhost:8186/api/v1/nodes/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "ansible_host": "192.168.1.20",
    "roles": ["master", "worker"]
  }'
```

Updates use partial merge semantics: only the fields present in the request body are changed. There are two exceptions:

- **roles** -- the provided array replaces the entire role set rather than merging with existing roles.
- **password** -- sending an empty string `""` clears the stored credential.

Omitted fields are left unchanged.

## Delete a node

```bash
curl -X DELETE http://localhost:8186/api/v1/nodes/{id}
```

Deleting a node also removes all associated role assignments.

## Sync the Ansible inventory

After creating, updating, or deleting nodes, sync the inventory so that changes take effect in subsequent Ansible runs:

```bash
curl -X POST http://localhost:8186/api/v1/onramp/inventory/sync
```

This generates a `hosts.ini` file from the current node database and writes it into the OnRamp repository directory. All component actions (`install`, `uninstall`, etc.) read from this inventory file.

Note: Always sync inventory after node changes before running any OnRamp action. Forgetting this step causes Ansible to target stale host lists.

## Test connectivity

Verify that all nodes are reachable via SSH:

```bash
curl -X POST http://localhost:8186/api/v1/onramp/components/cluster/pingall
```

This returns a task ID. Poll the task to see results:

```bash
curl http://localhost:8186/api/v1/onramp/tasks/{task_id}
```

A successful pingall confirms that SSH credentials are correct and all nodes are network-reachable. See [Deploying Components](deploying-components) for details on task polling.

## Typical workflow

1. Create nodes with `POST /api/v1/nodes`
2. Sync inventory with `POST /api/v1/onramp/inventory/sync`
3. Test connectivity with `POST /api/v1/onramp/components/cluster/pingall`
4. Proceed to [component deployment](deploying-components) once all nodes pass
