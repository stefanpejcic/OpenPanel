#!/bin/bash
# Enhanced system detection for OpenPanel installation

# Detect OS with more detailed information
detect_os_details() {
    if [ -f "/etc/os-release" ]; then
        source /etc/os-release
        
        # Get specific version details
        echo "ID: $ID"
        echo "VERSION_ID: $VERSION_ID"
        echo "PRETTY_NAME: $PRETTY_NAME"
        
        # Get more specific details for certain distributions
        case $ID in
            ubuntu)
                ubuntu_codename=$(lsb_release -cs 2>/dev/null || echo "unknown")
                echo "UBUNTU_CODENAME: $ubuntu_codename"
                ;;
            debian)
                debian_version=$(cat /etc/debian_version 2>/dev/null || echo "unknown")
                echo "DEBIAN_VERSION: $debian_version"
                ;;
            fedora|centos|rocky|almalinux)
                # Get more detailed version info
                if [ -f "/etc/centos-release" ]; then
                    echo "RELEASE_INFO: $(cat /etc/centos-release)"
                elif [ -f "/etc/fedora-release" ]; then
                    echo "RELEASE_INFO: $(cat /etc/fedora-release)"
                elif [ -f "/etc/rocky-release" ]; then
                    echo "RELEASE_INFO: $(cat /etc/rocky-release)"
                elif [ -f "/etc/almalinux-release" ]; then
                    echo "RELEASE_INFO: $(cat /etc/almalinux-release)"
                fi
                ;;
        esac
    else
        echo "ERROR: Could not detect Linux distribution from /etc/os-release"
        exit 1
    fi
}

# Get detailed CPU information
get_cpu_details() {
    echo "CPU_MODEL: $(grep -m1 'model name' /proc/cpuinfo | cut -d':' -f2 | sed 's/^[ \t]*//')"
    echo "CPU_CORES: $(grep -c '^processor' /proc/cpuinfo)"
    echo "CPU_ARCHITECTURE: $(uname -m)"
    
    # Detect virtualization
    if grep -q "^flags.*vmx" /proc/cpuinfo; then
        echo "VIRTUALIZATION_SUPPORT: Intel VT-x"
    elif grep -q "^flags.*svm" /proc/cpuinfo; then
        echo "VIRTUALIZATION_SUPPORT: AMD-V"
    else
        echo "VIRTUALIZATION_SUPPORT: Not detected"
    fi
}

# Get detailed memory information
get_memory_details() {
    local mem_total=$(grep 'MemTotal' /proc/meminfo | awk '{print $2}')
    local mem_available=$(grep 'MemAvailable' /proc/meminfo | awk '{print $2}')
    local swap_total=$(grep 'SwapTotal' /proc/meminfo | awk '{print $2}')
    
    echo "MEMORY_TOTAL_KB: $mem_total"
    echo "MEMORY_TOTAL_GB: $(awk "BEGIN {printf \"%.2f\", $mem_total/1024/1024}")"
    echo "MEMORY_AVAILABLE_KB: $mem_available"
    echo "MEMORY_AVAILABLE_GB: $(awk "BEGIN {printf \"%.2f\", $mem_available/1024/1024}")"
    echo "SWAP_TOTAL_KB: $swap_total"
    echo "SWAP_TOTAL_GB: $(awk "BEGIN {printf \"%.2f\", $swap_total/1024/1024}")"
    
    # Recommend swap based on memory
    if [ $mem_total -lt 4194304 ]; then # Less than 4GB
        echo "SWAP_RECOMMENDATION: 4GB"
    elif [ $mem_total -lt 8388608 ]; then # Less than 8GB
        echo "SWAP_RECOMMENDATION: 2GB"
    else
        echo "SWAP_RECOMMENDATION: Minimal (1GB)"
    fi
}

# Check for containerization
detect_container() {
    if [ -f "/.dockerenv" ]; then
        echo "CONTAINER_TYPE: Docker"
        return 0
    elif grep -q 'docker\|lxc' /proc/1/cgroup 2>/dev/null; then
        echo "CONTAINER_TYPE: Docker/LXC"
        return 0
    elif systemd-detect-virt --container &>/dev/null; then
        echo "CONTAINER_TYPE: $(systemd-detect-virt --container)"
        return 0
    else
        echo "CONTAINER_TYPE: None detected"
        return 1
    fi
}

# Detect if system uses systemd
detect_init_system() {
    if command -v systemctl >/dev/null 2>&1; then
        echo "INIT_SYSTEM: systemd"
    elif [ -f /sbin/init ] && file /sbin/init | grep -q "systemd"; then
        echo "INIT_SYSTEM: systemd"
    elif [ -f /sbin/init ] && file /sbin/init | grep -q "upstart"; then
        echo "INIT_SYSTEM: upstart"
    elif [ -f /sbin/init ] && file /sbin/init | grep -q "sysvinit"; then
        echo "INIT_SYSTEM: sysvinit"
    else
        echo "INIT_SYSTEM: unknown"
    fi
}

# Detect storage capabilities
detect_storage() {
    echo "DISK_SPACE:"
    df -h / | grep -v Filesystem
    
    echo "FILESYSTEM_TYPE: $(df -T / | awk 'NR==2 {print $2}')"
    
    if command -v lsblk >/dev/null 2>&1; then
        echo "BLOCK_DEVICES:"
        lsblk -o NAME,SIZE,TYPE,MOUNTPOINT | grep -v loop
    fi
}

# Main execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "OpenPanel System Discovery"
    echo "=========================="
    echo "OS Details:"
    detect_os_details
    echo
    echo "CPU Information:"
    get_cpu_details
    echo
    echo "Memory Information:"
    get_memory_details
    echo
    echo "Container Detection:"
    detect_container
    echo
    echo "Init System:"
    detect_init_system
    echo
    echo "Storage Information:"
    detect_storage
fi
