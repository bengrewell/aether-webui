---
sidebar_position: 2
title: "Configuring OnRamp"
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Configuring OnRamp

The OnRamp configuration lives in `vars/main.yml` inside the OnRamp repository directory. The API provides endpoints to read, patch, and swap the active configuration using profiles.

## Read the current configuration

<Tabs>
  <TabItem value="ui" label="Web UI" default>
    Navigate to **Configuration**. The current `vars/main.yml` is displayed as an editable form organized by section (k8s, core, gnbsim, etc.). Each section is collapsible and shows all fields with their current values.
  </TabItem>
  <TabItem value="api" label="API">

```bash
curl http://localhost:8186/api/v1/onramp/config
```

Returns the full `vars/main.yml` parsed as a JSON object. Each top-level YAML section becomes a top-level JSON key.

  </TabItem>
</Tabs>

## Patch the configuration

<Tabs>
  <TabItem value="ui" label="Web UI" default>
    Edit the values directly in the configuration form and click **Save**. Only modified sections are sent to the server; unchanged sections are left as-is.
  </TabItem>
  <TabItem value="api" label="API">

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

  </TabItem>
</Tabs>

### Merge behavior

- A section included in the request body **overwrites** the corresponding section in `vars/main.yml`.
- A section **omitted** from the request body is preserved as-is.
- There is no field-level merge within a section. If you send `{"core": {"data_iface": "eth1"}}`, the entire `core` section in the file is replaced with `{"data_iface": "eth1"}`. Include all fields for a section when patching it.

## Work with profiles

Profiles are pre-built configuration variants stored as `vars/main-*.yml` files in the OnRamp repository. Each profile is tuned for a specific RAN simulator or hardware setup.

### List available profiles

<Tabs>
  <TabItem value="ui" label="Web UI" default>
    Open the **Profiles** tab on the Configuration page. Available profiles are listed with their names and a brief description of each variant.
  </TabItem>
  <TabItem value="api" label="API">

```bash
curl http://localhost:8186/api/v1/onramp/config/profiles
```

Example response:

```json
["gnbsim", "oai", "srsran", "ueransim"]
```

Profile names are derived from filenames: `vars/main-gnbsim.yml` becomes profile name `gnbsim`.

  </TabItem>
</Tabs>

### Preview a profile

<Tabs>
  <TabItem value="ui" label="Web UI" default>
    Click a profile name to preview its contents in a read-only view. The preview shows all configuration sections and values that would be applied on activation.
  </TabItem>
  <TabItem value="api" label="API">

Inspect a profile's contents before activating it:

```bash
curl http://localhost:8186/api/v1/onramp/config/profiles/gnbsim
```

Returns the profile's YAML content parsed as JSON, identical in structure to the `GET /api/v1/onramp/config` response.

  </TabItem>
</Tabs>

### Activate a profile

<Tabs>
  <TabItem value="ui" label="Web UI" default>
    Click **Activate** next to the profile name. Confirm in the dialog -- this overwrites the current configuration with the profile's contents. The configuration form reloads with the new values.
  </TabItem>
  <TabItem value="api" label="API">

```bash
curl -X POST http://localhost:8186/api/v1/onramp/config/profiles/gnbsim/activate
```

  </TabItem>
</Tabs>

Warning: Activating a profile **overwrites the entire active configuration** (`vars/main.yml`) with the profile's contents. Any manual patches applied since the last profile activation are lost. Preview the profile first and consider backing up the current config with `GET /api/v1/onramp/config` if needed.

## Workflow example

A typical configuration workflow:

1. Activate a base profile: `POST /api/v1/onramp/config/profiles/gnbsim/activate`
2. Patch site-specific values: `PATCH /api/v1/onramp/config` with your overrides
3. Verify the result: `GET /api/v1/onramp/config`
4. Proceed to [deploy components](deploying-components)
