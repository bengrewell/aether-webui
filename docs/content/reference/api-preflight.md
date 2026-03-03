---
sidebar_position: 6
title: "API: Preflight"
---

# Preflight API Reference

Pre-deployment system checks with optional automated fixes. Each check verifies a specific prerequisite and reports whether it passed, along with details and an optional fix.

## Endpoints

| Method | Path | Operation ID | Description |
|--------|------|--------------|-------------|
| `GET` | `/api/v1/preflight` | `preflight-list` | Run all checks, return aggregate summary |
| `GET` | `/api/v1/preflight/{id}` | `preflight-get` | Run a single check by ID |
| `POST` | `/api/v1/preflight/{id}/fix` | `preflight-fix` | Apply the automated fix for a check |

## GET `/api/v1/preflight`

Runs all registered preflight checks in parallel and returns a summary.

**Response:**

```json
{
  "passed": 2,
  "failed": 2,
  "total": 4,
  "results": [
    {
      "id": "required-packages",
      "name": "Required Packages",
      "description": "Checks that required build and deployment tools (make, ansible) are installed.",
      "severity": "required",
      "category": "tooling",
      "passed": true,
      "message": "all required packages found: make (/usr/bin/make), ansible-playbook (/usr/bin/ansible-playbook)",
      "can_fix": true,
      "fix_warning": "This will install system packages using the detected package manager (apt-get, dnf, or yum)."
    }
  ]
}
```

## GET `/api/v1/preflight/{id}`

Runs a single check and returns its result.

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `id` | string | path | Check ID |

**Response:** A single `CheckResult` object.

**Errors:**
- `404` -- Unknown check ID

## POST `/api/v1/preflight/{id}/fix`

Executes the automated fix for a check. Fixes run synchronously and return immediately.

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `id` | string | path | Check ID |

**Response:**

```json
{
  "id": "ssh-configured",
  "applied": true,
  "message": "wrote /etc/ssh/sshd_config.d/99-aether-password-auth.conf and restarted sshd",
  "warning": "Enabling SSH password authentication allows any user to log in with a password."
}
```

**Errors:**
- `404` -- Unknown check ID
- `422` -- Check has no automated fix

## Available Checks

| ID | Category | Severity | Fix | Description |
|----|----------|----------|-----|-------------|
| `required-packages` | tooling | required | Yes | `make` and `ansible-playbook` installed (distro-aware install via apt-get/dnf/yum) |
| `ssh-configured` | access | required | Yes | sshd PasswordAuthentication enabled |
| `aether-user-configured` | access | required | Yes | `aether` user with NOPASSWD sudo |
| `node-ssh-reachable` | network | info | No | TCP port 22 reachable on managed nodes |

## Data Types

### CheckResult

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Check identifier |
| `name` | string | Human-readable name |
| `description` | string | What the check verifies |
| `severity` | string | `required`, `warning`, or `info` |
| `category` | string | `tooling`, `access`, or `network` |
| `passed` | bool | Whether the check passed |
| `message` | string | Summary of the result |
| `details` | string | Additional context (omitted when empty) |
| `can_fix` | bool | Whether an automated fix exists |
| `fix_warning` | string | Security warning for the fix (omitted when empty) |
| `error` | string | Error running the check (omitted when empty) |

### FixResult

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Check identifier |
| `applied` | bool | Whether the fix succeeded |
| `message` | string | Summary of what happened |
| `warning` | string | Security warning (omitted when empty) |
| `error` | string | Error message (omitted when empty) |

### PreflightSummary

| Field | Type | Description |
|-------|------|-------------|
| `passed` | int | Checks that passed |
| `failed` | int | Checks that failed |
| `total` | int | Total checks run |
| `results` | CheckResult[] | Individual results |
