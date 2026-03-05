#!/bin/bash
#
# Aether Preflight Setup Script
#
# Configures a host for Aether deployment by installing required packages,
# enabling SSH password authentication, and creating the aether service user.
#
# Usage:
#   sudo bash scripts/preflight-setup.sh [--yes]
#
#   --yes   Skip the interactive confirmation prompt and apply all changes
#           automatically. Required when running non-interactively (e.g. via
#           curl | sudo bash).
#
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Track what was done for the summary.
SUMMARY_PACKAGES="skipped"
SUMMARY_SSH="skipped"
SUMMARY_USER="skipped"

# Set to true when --yes flag is passed.
AUTO_CONFIRM=false

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Check if running on Linux
check_os() {
    if [[ "$(uname -s)" != "Linux" ]]; then
        log_error "This script only supports Linux"
        exit 1
    fi
    log_info "Operating system: Linux"
}

# Detect package manager. Sets PM_NAME and PM_PATH.
detect_package_manager() {
    for pm in apt-get dnf yum; do
        local pm_path
        pm_path=$(command -v "$pm" 2>/dev/null) || continue
        PM_NAME="$pm"
        PM_PATH="$pm_path"
        return 0
    done
    return 1
}

# Install required packages (make, ansible) if missing.
install_packages() {
    local missing=()

    for bin in git make ansible-playbook; do
        if ! command -v "$bin" &>/dev/null; then
            missing+=("$bin")
        fi
    done

    if [[ ${#missing[@]} -eq 0 ]]; then
        log_info "Required packages already installed (git, make, ansible)"
        SUMMARY_PACKAGES="already installed (git, make, ansible)"
        return
    fi

    log_info "Missing binaries: ${missing[*]}"

    PM_NAME=""
    PM_PATH=""
    if ! detect_package_manager; then
        log_error "No supported package manager found (tried apt-get, dnf, yum)"
        exit 1
    fi
    log_info "Using package manager: ${PM_NAME} (${PM_PATH})"

    # On Debian/Ubuntu, ansible requires the PPA.
    local needs_ansible=false
    for m in "${missing[@]}"; do
        if [[ "$m" == "ansible-playbook" ]]; then
            needs_ansible=true
            break
        fi
    done

    if [[ "$PM_NAME" == "apt-get" ]] && [[ "$needs_ansible" == "true" ]]; then
        # PPAs are Ubuntu-specific; on Debian, install ansible from default repos.
        local distro_id
        # shellcheck disable=SC1091 # /etc/os-release is a system file, not a script input
        distro_id=$(. /etc/os-release 2>/dev/null && echo "${ID:-}") || true
        if [[ "$distro_id" == "ubuntu" ]]; then
            log_info "Adding Ansible PPA for Ubuntu..."
            "$PM_PATH" install -y software-properties-common
            add-apt-repository --yes --update ppa:ansible/ansible
        else
            log_info "Installing ansible from default repos (non-Ubuntu apt system)"
        fi
    fi

    # Map binary names to package names.
    local pkgs=()
    for m in "${missing[@]}"; do
        case "$m" in
            git)
                pkgs+=("git")
                ;;
            make)
                pkgs+=("make")
                ;;
            ansible-playbook)
                pkgs+=("ansible")
                ;;
        esac
    done

    log_info "Installing packages: ${pkgs[*]}"
    "$PM_PATH" install -y "${pkgs[@]}"

    SUMMARY_PACKAGES="installed ${pkgs[*]} via ${PM_NAME}"
    log_info "Packages installed successfully"
}

# Enable SSH password authentication via sshd drop-in config.
configure_ssh() {
    local main_config="/etc/ssh/sshd_config"
    local drop_in_dir="/etc/ssh/sshd_config.d"
    local drop_in="${drop_in_dir}/99-aether-password-auth.conf"

    # Determine effective PasswordAuthentication value.
    # Last directive wins, matching sshd behavior.
    local enabled="yes" # OpenSSH default
    local source="default"

    if [[ -f "$main_config" ]]; then
        local val
        val=$(grep -i '^[[:space:]]*PasswordAuthentication[[:space:]]' "$main_config" | tail -1 | awk '{print $2}') || true
        if [[ -n "$val" ]]; then
            enabled="$val"
            source="$main_config"
        fi
    fi

    if [[ -d "$drop_in_dir" ]]; then
        for conf in "$drop_in_dir"/*.conf; do
            [[ -f "$conf" ]] || continue
            local dval
            dval=$(grep -i '^[[:space:]]*PasswordAuthentication[[:space:]]' "$conf" | tail -1 | awk '{print $2}') || true
            if [[ -n "$dval" ]]; then
                enabled="$dval"
                source="$conf"
            fi
        done
    fi

    if [[ "${enabled,,}" == "yes" ]]; then
        log_info "SSH PasswordAuthentication already enabled (set in ${source})"
        SUMMARY_SSH="already enabled (set in ${source})"
        return
    fi

    log_warn "WARNING: Enabling SSH password authentication allows any user to log in with a password."
    log_warn "Consider using key-based authentication for production environments."

    mkdir -p "$drop_in_dir"
    echo "PasswordAuthentication yes" > "$drop_in"
    log_info "Wrote ${drop_in}"

    # Restart sshd — RHEL/Fedora use "sshd", Debian/Ubuntu use "ssh".
    local restarted=false
    for unit in sshd ssh; do
        if systemctl restart "$unit" 2>/dev/null; then
            restarted=true
            log_info "Restarted ${unit} service"
            break
        fi
    done

    if [[ "$restarted" == "false" ]]; then
        log_error "Wrote config but failed to restart SSH service (tried sshd and ssh units)"
        exit 1
    fi

    SUMMARY_SSH="enabled PasswordAuthentication via ${drop_in}"
}

# Print the security-critical changes this script will make and require
# explicit confirmation before proceeding.
confirm_changes() {
    echo ""
    echo "=============================================="
    echo -e "${YELLOW}  SECURITY WARNING${NC}"
    echo "=============================================="
    echo ""
    echo "  This script will make the following security-sensitive changes:"
    echo ""
    echo "  1. Enable SSH password authentication"
    echo "     (allows any user to log in with a password)"
    echo "  2. Create the 'aether' OS user with the default"
    echo "     password 'aether' (you must change this after setup)"
    echo "  3. Grant the 'aether' user passwordless sudo (NOPASSWD: ALL)"
    echo ""

    if [[ "$AUTO_CONFIRM" == "true" ]]; then
        log_info "Auto-confirmed via --yes flag"
        return
    fi

    # Guard against non-interactive environments (e.g. curl | sudo bash).
    if [[ ! -t 0 ]]; then
        log_error "stdin is not a terminal. Re-run the script interactively or"
        log_error "pass --yes to confirm the security changes non-interactively:"
        log_error ""
        log_error "  sudo bash preflight-setup.sh --yes"
        exit 1
    fi

    read -r -t 60 -p "  Type 'yes' to confirm and continue: " answer || {
        echo ""
        log_error "Timed out waiting for confirmation. Aborting."
        exit 1
    }
    echo ""
    if [[ "$answer" != "yes" ]]; then
        log_error "Confirmation not given. Aborting."
        exit 1
    fi
}

# Create the aether user with default password and NOPASSWD sudo.
configure_aether_user() {
    local user_existed=false

    if id aether &>/dev/null; then
        user_existed=true
        log_info "User 'aether' already exists"
    else
        log_info "Creating user 'aether'..."
        useradd -m -s /bin/bash aether

        log_warn "WARNING: Setting default password 'aether' for user 'aether'."
        log_warn "Change this password after initial setup."
        echo 'aether:aether' | chpasswd
    fi

    # Write sudoers file via temp file + visudo validation.
    local sudoers_file="/etc/sudoers.d/aether"
    local sudoers_tmp
    sudoers_tmp=$(mktemp)
    echo 'aether ALL=(ALL) NOPASSWD: ALL' > "$sudoers_tmp"

    if visudo -c -f "$sudoers_tmp" >/dev/null 2>&1; then
        install -m 0440 "$sudoers_tmp" "$sudoers_file"
        log_info "Configured NOPASSWD sudo in ${sudoers_file}"
    else
        log_error "Sudoers syntax validation failed; not installing ${sudoers_file}"
        rm -f "$sudoers_tmp"
        exit 1
    fi
    rm -f "$sudoers_tmp"

    if [[ "$user_existed" == "true" ]]; then
        SUMMARY_USER="configured sudo for existing user 'aether'"
    else
        SUMMARY_USER="created user 'aether' with sudo and default password"
    fi
}

# Print summary of actions taken.
print_summary() {
    echo ""
    echo "=============================================="
    echo -e "${GREEN}Preflight Setup Complete${NC}"
    echo "=============================================="
    echo ""
    echo "  Packages:    ${SUMMARY_PACKAGES}"
    echo "  SSH:         ${SUMMARY_SSH}"
    echo "  Aether user: ${SUMMARY_USER}"
    echo ""
}

# Main flow
main() {
    # Parse arguments.
    for arg in "$@"; do
        case "$arg" in
            --yes)
                # Idempotent — setting this flag more than once is harmless.
                AUTO_CONFIRM=true
                ;;
            *)
                log_error "Unknown argument: ${arg}"
                echo "Usage: sudo bash preflight-setup.sh [--yes]"
                exit 1
                ;;
        esac
    done

    echo ""
    echo "=============================================="
    echo "  Aether Preflight Setup"
    echo "=============================================="
    echo ""

    check_root
    check_os
    confirm_changes
    install_packages
    configure_ssh
    configure_aether_user
    print_summary
}

main "$@"
