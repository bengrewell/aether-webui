---
sidebar_position: 4
title: Preflight Provider
---

# Preflight Provider

The preflight provider runs pre-deployment system checks and offers automated fixes for common configuration issues. It verifies that a fresh host has the tools, user accounts, and network connectivity required before Aether OnRamp deployment begins.

## Design

Checks are plain structs with function fields for dependency injection, making them fully testable without touching the real filesystem or executing commands. A `CheckDeps` struct bundles all external dependencies (file I/O, command execution, user lookup, network dial) so tests can inject stubs.

The check registry is a package-level slice. Adding a new check requires defining a `Check` value and appending it to the registry — no interface implementation or registration boilerplate needed.

## Checks

| ID | Category | Severity | Fix | Description |
|----|----------|----------|-----|-------------|
| `required-packages` | tooling | required | Yes | Verifies `make` and `ansible-playbook` are installed; fix detects distro and installs via apt-get/dnf/yum |
| `ssh-configured` | access | required | Yes | Parses sshd_config (including drop-in files) for PasswordAuthentication |
| `aether-user-configured` | access | required | Yes | Checks for `aether` user with `/etc/sudoers.d/aether` |
| `node-ssh-reachable` | network | info | No | TCP dials port 22 on all managed nodes |

### sshd_config parsing

The SSH password auth check parses `/etc/ssh/sshd_config` and all `*.conf` files in `/etc/ssh/sshd_config.d/`, applying last-directive-wins semantics to match sshd behavior. If no `PasswordAuthentication` directive is found, the OpenSSH default (`yes`) is assumed.

### Fixes

Fixes that modify system state (installing packages, creating users, writing config files, restarting services) run via `sudo` commands. Each fix includes a security warning so the API consumer can display it before the user confirms.

- **required-packages**: Detects the package manager (apt-get, dnf, yum) and installs missing packages. Fails gracefully if no supported package manager is found.
- **ssh-configured**: Writes a drop-in file at `/etc/ssh/sshd_config.d/99-aether-password-auth.conf` and restarts sshd.
- **aether-user-configured**: Creates the `aether` user with a default password, writes a NOPASSWD sudoers file.

## Endpoints

| Method | Path | Operation ID | Description |
|--------|------|--------------|-------------|
| `GET` | `/api/v1/preflight` | `preflight-list` | Run all checks in parallel, return aggregate results |
| `GET` | `/api/v1/preflight/{id}` | `preflight-get` | Run a single check by ID |
| `POST` | `/api/v1/preflight/{id}/fix` | `preflight-fix` | Execute fix for a check (422 if no fix available) |

## Adding a new check

1. Define a function in `checks.go` that returns a `Check` value.
2. Add it to the `registry` slice.
3. The check index, endpoint count, and API responses update automatically.
