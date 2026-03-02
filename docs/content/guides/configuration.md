---
sidebar_position: 2
title: "Configuring OnRamp"
---

# Configuring OnRamp

The OnRamp configuration lives in `vars/main.yml` inside the OnRamp repository directory. The API provides endpoints to read, patch, and swap the active configuration using profiles.

## Read the current configuration

```bash
curl http://localhost:8186/api/v1/onramp/config
```

Returns the full `vars/main.yml` parsed as a JSON object. Each top-level YAML section becomes a top-level JSON key.

## Patch the configuration

The PATCH endpoint performs a **section-level merge**: supply only the top-level sections to change. Omitted sections are left untouched.

```bash
curl -X PATCH http://localhost:8186/api/v1/onramp/config \
  -H "Content-Type: application/json" \
  -d '{
    "core": {
      "data_iface": "eth1"
    }
  }'
```

This updates only the `core` section's `data_iface` field. All other sections and fields remain unchanged.

### Merge behavior

- A section included in the request body **overwrites** the corresponding section in `vars/main.yml`.
- A section **omitted** from the request body is preserved as-is.
- There is no field-level merge within a section. If you send `{"core": {"data_iface": "eth1"}}`, the entire `core` section in the file is replaced with `{"data_iface": "eth1"}`. Include all fields for a section when patching it.

## Work with profiles

Profiles are pre-built configuration variants stored as `vars/main-*.yml` files in the OnRamp repository. Each profile is tuned for a specific RAN simulator or hardware setup.

### List available profiles

```bash
curl http://localhost:8186/api/v1/onramp/config/profiles
```

Example response:

```json
["gnbsim", "oai", "srsran", "ueransim"]
```

Profile names are derived from filenames: `vars/main-gnbsim.yml` becomes profile name `gnbsim`.

### Preview a profile

Inspect a profile's contents before activating it:

```bash
curl http://localhost:8186/api/v1/onramp/config/profiles/gnbsim
```

Returns the profile's YAML content parsed as JSON, identical in structure to the `GET /api/v1/onramp/config` response.

### Activate a profile

```bash
curl -X POST http://localhost:8186/api/v1/onramp/config/profiles/gnbsim/activate
```

Warning: Activating a profile **overwrites the entire active configuration** (`vars/main.yml`) with the profile's contents. Any manual patches applied since the last profile activation are lost. Preview the profile first and consider backing up the current config with `GET /api/v1/onramp/config` if needed.

## Workflow example

A typical configuration workflow:

1. Activate a base profile: `POST /api/v1/onramp/config/profiles/gnbsim/activate`
2. Patch site-specific values: `PATCH /api/v1/onramp/config` with your overrides
3. Verify the result: `GET /api/v1/onramp/config`
4. Proceed to [deploy components](deploying-components)
