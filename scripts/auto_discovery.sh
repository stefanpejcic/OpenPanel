#!/bin/bash
# OpenPanel Auto-Discovery Integration Module
# This script integrates various auto-discovery modules and provides a unified interface

# Call this function before sourcing the file
setup_progress_bar_script
# Then source the file
source "$PROGRESS_BAR_FILE"

# Source required modules
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/network_discovery.sh"
source "$SCRIPT_DIR/system_discovery.sh"

# Add to the installation process
source scripts/auto_discovery.sh
validate_system_compatibility

# Function to validate system compatibility
validate_system_compatibility() {
    # Load system details
    local os_id=$(detect_os_details | grep "^ID:" | cut -d' ' -f2)
    local os_version=$(detect_os_details | grep "^VERSION_ID:" | cut -d' ' -f2 | tr -d '"')
    local cpu_arch=$(get_cpu_details | grep "^CPU_ARCHITECTURE:" | cut -d' ' -f2)
    local mem_gb=$(get_memory_details | grep "^MEMORY_TOTAL_GB:" | cut -d' ' -f2)
    local container_type=$(detect_container | grep "^CONTAINER_TYPE:" | cut -d' ' -f2-)
    local init_system=$(detect_init_system | grep "^INIT_SYSTEM:" | cut -d' ' -f2)
    
    # Check OS compatibility
    local os_supported=false
    case $os_id in
        ubuntu)
            if [[ "$os_version" == "24.04" || "$os_version" == "22.04" ]]; then
                os_supported=true
            fi
            ;;
        debian)
            if [[ "$os_version" == "11" || "$os_version" == "12" ]]; then
                os_supported=true
            fi
            ;;
        fedora)
            if [[ "$os_version" == "38" ]]; then
                os_supported=true
            fi
            ;;
        almalinux|rocky|centos)
            if [[ "$os_version" == "9" || "$os_version" == "9.4" ]]; then
                os_supported=true
            fi
            ;;
    esac
    
    # Display compatibility results
    echo "SYSTEM COMPATIBILITY CHECK:"
    
    if [[ "$os_supported" == "true" ]]; then
        echo "✅ OS: $os_id $os_version is supported"
    else
        echo "❌ OS: $os_id $os_version may not be fully supported"
    fi
    
    if [[ "$cpu_arch" == "x86_64" || "$cpu_arch" == "aarch64" ]]; then
        echo "✅ CPU Architecture: $cpu_arch is supported"
    else
        echo "❌ CPU Architecture: $cpu_arch is not supported"
    fi
    
    if (( $(echo "$mem_gb >= 1.0" | bc -l) )); then
        echo "✅ Memory: $mem_gb GB is sufficient"
    else
        echo "❌ Memory: $mem_gb GB is insufficient (minimum 1 GB recommended)"
    fi
    
    if [[ "$container_type" == "None detected" ]]; then
        echo "✅ Not running in a container"
    else
        echo "❌ Container detected: $container_type (OpenPanel should be installed on the host OS)"
    fi
    
    if [[ "$init_system" == "systemd" ]]; then
        echo "✅ Init System: systemd is supported"
    else
        echo "❌ Init System: $init_system may not be fully supported"
    fi
}

# Function to generate recommended configuration
generate_recommended_config() {
    # Get system specifications
    local mem_gb=$(get_memory_details | grep "^MEMORY_TOTAL_GB:" | cut -d' ' -f2)
    local cpu_cores=$(get_cpu_details | grep "^CPU_CORES:" | cut -d' ' -f2)
    local swap_recommendation=$(get_memory_details | grep "^SWAP_RECOMMENDATION:" | cut -d' ' -f2)
    
    # Calculate recommended resources
    local web_processes=$(( cpu_cores > 2 ? cpu_cores - 1 : 1 ))
    local db_memory_mb=$(awk "BEGIN {printf \"%.0f\", $mem_gb * 1024 * 0.4}")
    local php_memory_limit_mb=$(awk "BEGIN {printf \"%.0f\", $mem_gb * 128}")
    
    # Display recommendations
    echo "RECOMMENDED CONFIGURATION:"
    echo "Web Server Worker Processes: $web_processes"
    echo "Database Memory Allocation: ${db_memory_mb}MB"
    echo "PHP Memory Limit: ${php_memory_limit_mb}MB"
    echo "Swap Size: $swap_recommendation"
    
    # Additional recommendations based on system specifications
    if (( $(echo "$mem_gb >= 4.0" | bc -l) )); then
        echo "Redis Cache: Enabled"
    else
        echo "Redis Cache: Disabled (insufficient memory)"
    fi
    
    if (( $(echo "$mem_gb >= 8.0" | bc -l) )); then
        echo "Page Caching: Enabled"
        echo "Object Caching: Enabled"
    else
        echo "Page Caching: Limited"
        echo "Object Caching: Basic only"
    fi
}

# Function to generate a configuration file for the installer
generate_config_file() {
    local output_file="$1"
    
    # Get key information
    local os_id=$(detect_os_details | grep "^ID:" | cut -d' ' -f2)
    local cpu_arch=$(get_cpu_details | grep "^CPU_ARCHITECTURE:" | cut -d' ' -f2)
    local mem_gb=$(get_memory_details | grep "^MEMORY_TOTAL_GB:" | cut -d' ' -f2)
    local swap_recommendation=$(get_memory_details | grep "^SWAP_RECOMMENDATION:" | cut -d' ' -f2)
    local public_ip=$(get_public_ipv4)
    
    # Write configuration file
    cat > "$output_file" << EOL
# OpenPanel Auto-Generated Configuration
# Generated on $(date)
# System Information
OS_TYPE=$os_id
CPU_ARCHITECTURE=$cpu_arch
MEMORY_GB=$mem_gb
PUBLIC_IP=$public_ip

# Recommended Configuration
SWAP_SIZE=${swap_recommendation%G}
SETUP_SWAP_ANYWAY=true
PACKAGE_MANAGER=$(detect_package_manager)
CSF_SETUP=$([ "$mem_gb" -ge 2 ] && echo "true" || echo "false")
CORAZA=$([ "$mem_gb" -ge 4 ] && echo "true" || echo "false")
EOL
    
    echo "Configuration written to $output_file"
}

# Helper function to detect package manager
detect_package_manager() {
    if [ -f "/etc/os-release" ]; then
        . /etc/os-release
        case $ID in
            ubuntu|debian)
                echo "apt-get"
                ;;
            fedora|rocky|almalinux|centos|rhel|sles)
                if command -v dnf >/dev/null 2>&1; then
                    echo "dnf"
                else
                    echo "yum"
                fi
                ;;
            *)
                echo "unknown"
                ;;
        esac
    else
        echo "unknown"
    fi
}

# Function to set version to install
set_version_to_install(){
    if [ "$CUSTOM_VERSION" = false ]; then
        PANEL_VERSION=$(curl -L --silent --max-time 10 -4 https://openpanel.org/version || echo "1.1.8")
    fi
}

# Main execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "OpenPanel Auto-Discovery Module"
    echo "=============================="
    
    # Parse command-line arguments
    ACTION="all"
    OUTPUT_FILE=""
    while [[ $# -gt 0 ]]; do
        case $1 in
            --check-compatibility)
                ACTION="check"
                shift
                ;;
            --recommend-config)
                ACTION="recommend"
                shift
                ;;
            --generate-config)
                ACTION="generate"
                OUTPUT_FILE="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTION]"
                echo "Options:"
                echo "  --check-compatibility   Check if the system meets OpenPanel requirements"
                echo "  --recommend-config      Generate recommended configuration values"
                echo "  --generate-config FILE  Generate configuration file for the installer"
                echo "  --help                  Display this help message"
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    case $ACTION in
        "check")
            validate_system_compatibility
            ;;
        "recommend")
            generate_recommended_config
            ;;
        "generate")
            if [[ -z "$OUTPUT_FILE" ]]; then
                OUTPUT_FILE="/tmp/openpanel_auto_config.conf"
            fi
            generate_config_file "$OUTPUT_FILE"
            ;;
        "all")
            echo "System Compatibility:"
            validate_system_compatibility
            echo "Recommended Configuration:"
            generate_recommended_config
            ;;
    esac
fi

# Ensure proper if/fi balance in the functions
if [ condition ]; then
    command
fi  # Make sure this fi has a matching if