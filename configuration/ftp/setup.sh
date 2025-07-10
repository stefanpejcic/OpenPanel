#!/bin/bash

# Configuration
PANEL_CONFIG="/etc/openpanel/openpanel/conf/openpanel.config"
CONTAINER_NAME="openadmin_ftp"

# Exit on any error
set -e

# Utility functions
log() { echo "[*] $1"; }
error() { echo "ERROR: $1" >&2; exit 1; }

# Check if OpenPanel is installed
is_openpanel_installed() {
    [[ -f "$PANEL_CONFIG" ]]
}

# Start FTP container
start_ftp_container() {
    log "Starting FTP container..."
    cd /root && docker --context=default compose up "$CONTAINER_NAME" -d
}

# Check if container is running
is_container_running() {
    docker --context=default ps --filter "name=$CONTAINER_NAME" --filter "status=running" | grep -q "$CONTAINER_NAME"
}

# Enable FTP module in OpenPanel config
enable_ftp_module() {
    if grep -q "^enabled_modules=.*ftp" "$PANEL_CONFIG"; then
        log "FTP module already enabled."
    else
        log "Enabling FTP module..."
        sed -i '/^enabled_modules=/ s/$/,ftp/' "$PANEL_CONFIG"
        docker --context=default restart openpanel
        log "OpenPanel restarted to apply changes."
    fi
}

# Open firewall ports using CSF
open_firewall_ports() {
    if ! command -v csf >/dev/null 2>&1; then
        log "CSF not found. Manually open ports 21 and 21000:21010 in your firewall."
        return
    fi
    
    local csf_conf="/etc/csf/csf.conf"
    local ports_changed=false
    
    # Function to open a port in CSF
    open_port() {
        local port=$1
        if ! grep -q "TCP_IN = .*${port}" "$csf_conf"; then
            sed -i "s/TCP_IN = \"\(.*\)\"/TCP_IN = \"\1,${port}\"/" "$csf_conf"
            log "Opened port $port in CSF."
            ports_changed=true
        fi
    }
    
    open_port 21
    open_port "21000:21010"
    
    $ports_changed && csf -r
}

# Main execution
main() {
    
    # Check OpenPanel installation
    is_openpanel_installed || error "OpenPanel not installed."

    log "Setting up FTP service..."

    # Start container
    start_ftp_container
    
    # Verify container is running
    is_container_running || {
        log "Container failed to start. Cleaning up..."
        docker --context=default stop "$CONTAINER_NAME" 2>/dev/null || true
        docker --context=default rm "$CONTAINER_NAME" 2>/dev/null || true
        error "FTP container failed to start."
    }
    
    # Configure OpenPanel
    enable_ftp_module
    
    # Configure firewall
    open_firewall_ports
    
    log ""
    log "SUCCESS: FTP is now running and enabled for all OpenPanel users."
    log ""
}

# Run main function
main "$@"
