#!/bin/bash
# Enhanced network auto-discovery for OpenPanel
# This script provides comprehensive network information detection

# Function to detect all network interfaces
detect_interfaces() {
    echo "Detecting network interfaces..."
    ip -o link show | awk -F': ' '{print $2}' | grep -v "lo"
}

# Function to get primary interface
get_primary_interface() {
    ip route get 1 | awk '{print $5;exit}'
}

# Function to get all IPv4 addresses with better validation
get_all_ipv4_addresses() {
    local interfaces=$(detect_interfaces)
    for interface in $interfaces; do
        local ip=$(ip -4 addr show $interface | grep -oP '(?<=inet\s)\d+(\.\d+){3}')
        if [ -n "$ip" ]; then
            echo "$interface: $ip"
        fi
    done
}

# Enhanced public IP detection with more fallback services
get_public_ipv4() {
    # List of IP servers for checks with timeouts
    IP_SERVICES=(
        "https://ip.openpanel.com"
        "https://ipv4.openpanel.com"
        "https://ifconfig.me"
        "https://api.ipify.org"
        "https://icanhazip.com"
        "https://checkip.amazonaws.com"
    )
    
    for service in "${IP_SERVICES[@]}"; do
        public_ip=$(curl --silent --max-time 3 -4 $service)
        if [[ -n "$public_ip" && $public_ip =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            echo "$public_ip"
            return 0
        fi
    done
    
    # If all online services fail, attempt to determine from routing
    echo "Warning: Couldn't reliably determine public IP from external services" >&2
    ip route get 1.1.1.1 | awk '{print $7; exit}'
}

# Detect available ports
detect_available_ports() {
    local start_port=$1
    local end_port=$2
    local num_ports=$3
    
    echo "Finding $num_ports available ports between $start_port and $end_port..."
    
    local available_ports=()
    for port in $(seq $start_port $end_port); do
        if ! ss -tuln | grep -q ":$port "; then
            available_ports+=($port)
            if [ ${#available_ports[@]} -eq $num_ports ]; then
                break
            fi
        fi
    done
    
    echo "${available_ports[@]}"
}

# Main execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "OpenPanel Network Discovery"
    echo "=========================="
    echo "Primary Interface: $(get_primary_interface)"
    echo "Public IPv4: $(get_public_ipv4)"
    echo "Available interfaces and IPs:"
    get_all_ipv4_addresses
    echo "Recommended available ports:"
    detect_available_ports 8000 9000 5
fi
