#!/bin/bash
#
# Aether WebUI Installation Script
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/bengrewell/aether-webui/main/scripts/install.sh | sudo bash
#
# Options (via environment variables):
#   VERSION=v1.0.0  - Install a specific release version (default: latest)
#   REF=main        - Build from source at a git ref (tag, branch, or commit)
#
set -euo pipefail

# Configuration
GITHUB_REPO="bengrewell/aether-webui"
BINARY_NAME="aether-webd"
SERVICE_NAME="aether-webd"
SERVICE_USER="aether-webd"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/aether-webd"
DATA_DIR="/var/lib/aether-webd"
SYSTEMD_DIR="/etc/systemd/system"

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

# VERSION and REF are mutually exclusive
check_conflict() {
    if [[ -n "${VERSION:-}" && -n "${REF:-}" ]]; then
        log_error "VERSION and REF are mutually exclusive. Set one or neither."
        exit 1
    fi
}

# Verify build dependencies for source builds (git, make, go >= 1.25).
# Note: node/npm may be needed for very old refs where the embedded frontend
# dist is not checked in. For recent refs this is not required.
check_build_deps() {
    local missing=()
    for cmd in git make go; do
        if ! command -v "$cmd" &>/dev/null; then
            missing+=("$cmd")
        fi
    done
    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Source build requires: ${missing[*]}"
        exit 1
    fi
    local go_version go_major go_minor
    go_version=$(go env GOVERSION | sed 's/^go//; s/\([0-9]*\.[0-9]*\).*/\1/')
    IFS='.' read -r go_major go_minor <<< "$go_version"
    if (( go_major < 1 )) || { (( go_major == 1 )) && (( go_minor < 25 )); }; then
        log_error "Go >= 1.25 required (found go${go_version})"
        exit 1
    fi
    log_info "Build dependencies satisfied (go${go_version})"
}

# Clone the repository, checkout the requested ref, build, and install
build_from_source() {
    local tmp_dir
    tmp_dir=$(mktemp -d)
    # shellcheck disable=SC2064 # Intentional: expand tmp_dir now, not at signal time
    trap "rm -rf '$tmp_dir'" EXIT

    log_info "Cloning repository..."
    if ! git clone --depth 50 "https://github.com/${GITHUB_REPO}.git" "${tmp_dir}/src"; then
        log_error "Failed to clone repository"
        exit 1
    fi

    cd "${tmp_dir}/src"

    log_info "Checking out ref: ${REF}"
    if ! git checkout "${REF}" 2>/dev/null; then
        # Ref may be outside shallow depth — unshallow and retry
        log_info "Ref not in shallow clone, fetching full history..."
        git fetch --unshallow
        if ! git checkout "${REF}"; then
            log_error "Failed to checkout ref: ${REF}"
            exit 1
        fi
    fi

    log_info "Building from source (CGO_ENABLED=0)..."
    if [[ "$(id -u)" -eq 0 ]]; then
        log_warn "Running 'make build' as root. Consider building as an unprivileged user and using root only for the install step."
    fi
    CGO_ENABLED=0 make build

    local binary_path="bin/${BINARY_NAME}"
    if [[ ! -f "$binary_path" ]]; then
        log_error "Build did not produce expected binary at ${binary_path}"
        exit 1
    fi

    log_info "Installing binary to ${INSTALL_DIR}/${BINARY_NAME}"
    install -m 755 "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"

    if ! "${INSTALL_DIR}/${BINARY_NAME}" --version &>/dev/null; then
        log_warn "Binary installed but version check failed"
    else
        log_info "Binary verified: $(${INSTALL_DIR}/${BINARY_NAME} --version 2>&1 | head -1)"
    fi

    cd /
}

# Detect system architecture
detect_arch() {
    local arch
    arch=$(uname -m)

    case "$arch" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac

    log_info "Architecture: $ARCH"
}

# Get the latest release version from GitHub
get_latest_version() {
    local latest
    latest=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

    if [[ -z "$latest" ]]; then
        log_error "Failed to fetch latest version from GitHub"
        exit 1
    fi

    echo "$latest"
}

# Determine version to install
determine_version() {
    if [[ -n "${VERSION:-}" ]]; then
        INSTALL_VERSION="$VERSION"
        log_info "Installing specified version: $INSTALL_VERSION"
    else
        INSTALL_VERSION=$(get_latest_version)
        log_info "Installing latest version: $INSTALL_VERSION"
    fi
}

# Download and install the binary
download_binary() {
    local download_url
    local tmp_dir
    local archive_name

    # Strip leading 'v' from version tag for the archive filename
    local version_number="${INSTALL_VERSION#v}"
    archive_name="${BINARY_NAME}_${version_number}_linux_${ARCH}.tar.gz"
    download_url="https://github.com/${GITHUB_REPO}/releases/download/${INSTALL_VERSION}/${archive_name}"

    log_info "Downloading from: $download_url"

    tmp_dir=$(mktemp -d)
    # shellcheck disable=SC2064 # Intentional: expand tmp_dir now, not at signal time
    trap "rm -rf '$tmp_dir'" EXIT

    if ! curl -fsSL -o "${tmp_dir}/${archive_name}" "$download_url"; then
        log_error "Failed to download release archive"
        exit 1
    fi

    log_info "Extracting archive..."
    tar -xzf "${tmp_dir}/${archive_name}" -C "$tmp_dir"

    # Find the binary (it should be in the extracted directory)
    local binary_path="${tmp_dir}/${BINARY_NAME}"
    if [[ ! -f "$binary_path" ]]; then
        log_error "Binary not found in archive"
        exit 1
    fi

    log_info "Installing binary to ${INSTALL_DIR}/${BINARY_NAME}"
    install -m 755 "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"

    # Verify installation
    if ! "${INSTALL_DIR}/${BINARY_NAME}" --version &>/dev/null; then
        log_warn "Binary installed but version check failed"
    else
        log_info "Binary verified: $(${INSTALL_DIR}/${BINARY_NAME} --version 2>&1 | head -1)"
    fi
}

# Create system user for the service
create_user() {
    if id "$SERVICE_USER" &>/dev/null; then
        log_info "User '$SERVICE_USER' already exists"
    else
        log_info "Creating system user: $SERVICE_USER"
        useradd -r -s /bin/false -M "$SERVICE_USER"
    fi
}

# Install systemd service file
install_service() {
    local service_url
    service_url="https://raw.githubusercontent.com/${GITHUB_REPO}/main/deploy/systemd/${SERVICE_NAME}.service"

    log_info "Downloading systemd service file..."
    if ! curl -fsSL -o "${SYSTEMD_DIR}/${SERVICE_NAME}.service" "$service_url"; then
        log_error "Failed to download service file"
        exit 1
    fi

    chmod 644 "${SYSTEMD_DIR}/${SERVICE_NAME}.service"
    log_info "Service file installed to ${SYSTEMD_DIR}/${SERVICE_NAME}.service"
}

# Create configuration directory
create_config_dir() {
    if [[ -d "$CONFIG_DIR" ]]; then
        log_info "Config directory already exists: $CONFIG_DIR"
    else
        log_info "Creating config directory: $CONFIG_DIR"
        mkdir -p "$CONFIG_DIR"
    fi

    # Create default environment file if it doesn't exist
    if [[ ! -f "${CONFIG_DIR}/env" ]]; then
        log_info "Creating default environment file"
        cat > "${CONFIG_DIR}/env" << 'EOF'
# Aether WebUI daemon configuration
# Uncomment and set values to configure the service.
# CLI flags override these values when both are set.
#
# AETHER_LISTEN=0.0.0.0:8186
# AETHER_TLS=true
# AETHER_API_TOKEN=
# AETHER_ENCRYPTION_KEY=
# AETHER_DATA_DIR=/var/lib/aether-webd
# AETHER_METRICS_INTERVAL=10s
# AETHER_METRICS_RETENTION=24h
EOF
    fi

    chown -R "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR" 2>/dev/null || true
}

# Create data directory for persistent state
create_data_dir() {
    if [[ -d "$DATA_DIR" ]]; then
        log_info "Data directory already exists: $DATA_DIR"
    else
        log_info "Creating data directory: $DATA_DIR"
        mkdir -p "$DATA_DIR"
    fi
    chown -R "$SERVICE_USER:$SERVICE_USER" "$DATA_DIR"
    chmod 750 "$DATA_DIR"
}

# Exclude aether-webd from needrestart auto-restarts.
# Ubuntu 24.04 ships needrestart which automatically restarts services after
# apt upgrades. OnRamp Ansible playbooks install packages (Docker, etc.) that
# trigger needrestart, killing aether-webd mid-task execution.
configure_needrestart() {
    local nr_conf_dir="/etc/needrestart/conf.d"
    if [[ ! -d "$nr_conf_dir" ]]; then
        log_info "needrestart not installed, skipping exclusion"
        return
    fi

    local nr_conf="${nr_conf_dir}/aether-webd.conf"
    if [[ -f "$nr_conf" ]]; then
        log_info "needrestart exclusion already configured"
        return
    fi

    log_info "Configuring needrestart to exclude $SERVICE_NAME"
    cat > "$nr_conf" << 'EOF'
# Exclude aether-webd from automatic restarts.
# The service runs long-lived deployment tasks that must not be interrupted
# by package-triggered restarts.
$nrconf{override_rc}{qr(^aether-webd)} = 0;
EOF
    log_info "Created $nr_conf"
}

# Enable and start the service
enable_service() {
    log_info "Reloading systemd daemon..."
    systemctl daemon-reload

    log_info "Enabling $SERVICE_NAME service..."
    systemctl enable "$SERVICE_NAME"

    log_info "Starting $SERVICE_NAME service..."
    systemctl start "$SERVICE_NAME"

    # Give it a moment to start
    sleep 2

    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log_info "Service is running"
    else
        log_warn "Service may not have started correctly. Check status with: systemctl status $SERVICE_NAME"
    fi
}

# Print installation summary
print_summary() {
    echo ""
    echo "=============================================="
    echo -e "${GREEN}Installation Complete!${NC}"
    echo "=============================================="
    echo ""
    echo "Binary:     ${INSTALL_DIR}/${BINARY_NAME}"
    echo "Service:    ${SYSTEMD_DIR}/${SERVICE_NAME}.service"
    echo "Config:     ${CONFIG_DIR}/"
    echo "Data:       ${DATA_DIR}/"
    echo "User:       ${SERVICE_USER}"
    echo ""
    echo "Useful commands:"
    echo "  systemctl status $SERVICE_NAME   - Check service status"
    echo "  systemctl restart $SERVICE_NAME  - Restart the service"
    echo "  journalctl -u $SERVICE_NAME -f   - View logs"
    echo ""
    echo "Health check:"
    echo "  curl http://localhost:8186/healthz"
    echo ""
}

# Main installation flow
main() {
    echo ""
    echo "=============================================="
    echo "  Aether WebUI Installer"
    echo "=============================================="
    echo ""

    check_root
    check_os
    check_conflict

    if [[ -n "${REF:-}" ]]; then
        check_build_deps
        build_from_source
    else
        detect_arch
        determine_version
        download_binary
    fi

    create_user
    install_service
    create_config_dir
    create_data_dir
    configure_needrestart
    enable_service
    print_summary
}

main "$@"
