#!/bin/bash
#
# Aether WebUI Uninstallation Script
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/bengrewell/aether-webui/main/scripts/uninstall.sh | sudo bash
#
# Options:
#   --purge  - Also remove user and configuration directory
#
set -euo pipefail

# Configuration
BINARY_NAME="aether-webd"
SERVICE_NAME="aether-webd"
SERVICE_USER="aether-webd"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/aether-webd"
SYSTEMD_DIR="/etc/systemd/system"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Flags
PURGE=false

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --purge)
                PURGE=true
                shift
                ;;
            *)
                log_warn "Unknown option: $1"
                shift
                ;;
        esac
    done
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Stop the service if running
stop_service() {
    if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
        log_info "Stopping $SERVICE_NAME service..."
        systemctl stop "$SERVICE_NAME"
    else
        log_info "Service $SERVICE_NAME is not running"
    fi
}

# Disable the service
disable_service() {
    if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
        log_info "Disabling $SERVICE_NAME service..."
        systemctl disable "$SERVICE_NAME"
    else
        log_info "Service $SERVICE_NAME is not enabled"
    fi
}

# Remove the systemd service file
remove_service_file() {
    local service_file="${SYSTEMD_DIR}/${SERVICE_NAME}.service"

    if [[ -f "$service_file" ]]; then
        log_info "Removing service file: $service_file"
        rm -f "$service_file"
    else
        log_info "Service file not found: $service_file"
    fi
}

# Remove the binary
remove_binary() {
    local binary_path="${INSTALL_DIR}/${BINARY_NAME}"

    if [[ -f "$binary_path" ]]; then
        log_info "Removing binary: $binary_path"
        rm -f "$binary_path"
    else
        log_info "Binary not found: $binary_path"
    fi
}

# Remove the system user (only with --purge)
remove_user() {
    if id "$SERVICE_USER" &>/dev/null; then
        log_info "Removing system user: $SERVICE_USER"
        userdel "$SERVICE_USER" 2>/dev/null || log_warn "Failed to remove user $SERVICE_USER"
    else
        log_info "User $SERVICE_USER does not exist"
    fi
}

# Remove the configuration directory (only with --purge)
remove_config() {
    if [[ -d "$CONFIG_DIR" ]]; then
        log_info "Removing config directory: $CONFIG_DIR"
        rm -rf "$CONFIG_DIR"
    else
        log_info "Config directory not found: $CONFIG_DIR"
    fi
}

# Reload systemd daemon
reload_systemd() {
    log_info "Reloading systemd daemon..."
    systemctl daemon-reload
}

# Print uninstallation summary
print_summary() {
    echo ""
    echo "=============================================="
    echo -e "${GREEN}Uninstallation Complete!${NC}"
    echo "=============================================="
    echo ""

    if [[ "$PURGE" == "true" ]]; then
        echo "Removed:"
        echo "  - Binary: ${INSTALL_DIR}/${BINARY_NAME}"
        echo "  - Service file: ${SYSTEMD_DIR}/${SERVICE_NAME}.service"
        echo "  - Config directory: ${CONFIG_DIR}/"
        echo "  - System user: ${SERVICE_USER}"
    else
        echo "Removed:"
        echo "  - Binary: ${INSTALL_DIR}/${BINARY_NAME}"
        echo "  - Service file: ${SYSTEMD_DIR}/${SERVICE_NAME}.service"
        echo ""
        echo "Preserved (use --purge to remove):"
        echo "  - Config directory: ${CONFIG_DIR}/"
        echo "  - System user: ${SERVICE_USER}"
    fi
    echo ""
}

# Main uninstallation flow
main() {
    parse_args "$@"

    echo ""
    echo "=============================================="
    echo "  Aether WebUI Uninstaller"
    echo "=============================================="
    echo ""

    if [[ "$PURGE" == "true" ]]; then
        log_warn "Running with --purge: will remove user and config"
    fi

    check_root
    stop_service
    disable_service
    remove_service_file
    remove_binary

    if [[ "$PURGE" == "true" ]]; then
        remove_config
        remove_user
    fi

    reload_systemd
    print_summary
}

main "$@"
