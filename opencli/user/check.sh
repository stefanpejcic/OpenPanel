#!/bin/bash
################################################################################
# Script Name: user/check.sh
# Description: Performs comprehensive security checks on user files, Docker daemon and containers.
# Usage: opencli user-check <USERNAME>
# Author: Stefan Pejcic
# Created: 26.07.2025
# Last Modified: 17.11.2025
# Company: openpanel.com
# Copyright (c) openpanel.com
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
################################################################################

#set -euo pipefail

if [ "$EUID" -ne 0 ]; then
  echo "This script must be run as root."
  exit 1
fi

# Color codes
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

# Status counters
PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0
INFO_COUNT=0

# ====== Utility Functions ======

# Print colored header
print_header() {
    local title="$1"
    echo -e "${BLUE}===== ${title} =====${NC}"
}

# Print colored subheader
print_subheader() {
    local title="$1"
    echo -e "${CYAN}---- ${title} ----${NC}"
}

# Print result with colors and increment counters
print_result() {
    local status="$1"
    local message="$2"
    local details="${3:-}"

    case "$status" in
        "PASS")
            echo -e "[${GREEN}${status}${NC}] $message"
            ((PASS_COUNT++))
            ;;
        "FAIL")
            echo -e "[${RED}${status}${NC}] $message"
            ((FAIL_COUNT++))
            ;;
        "WARN")
            echo -e "[${YELLOW}${status}${NC}] $message"
            ((WARN_COUNT++))
            ;;
        "INFO")
            echo -e "[${BLUE}${status}${NC}] $message"
            ((INFO_COUNT++))
            ;;
        *)
            echo "[UNKNOWN] $message"
            ;;
    esac

    # Print additional details if provided
    if [[ -n "$details" ]]; then
        echo "    $details"
    fi
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if file exists and is readable
file_readable() {
    [[ -f "$1" && -r "$1" ]]
}

# get context
get_docker_context() {
    local user="$1"
    local query="SELECT id, server FROM users WHERE username = '${user}';"
    user_info=$(mysql -se "$query")
    user_id=$(echo "$user_info" | awk '{print $1}')
    context=$(echo "$user_info" | awk '{print $2}')
    
    if [ -z "$user_id" ]; then
        echo "FATAL ERROR: user $user does not exist."
        exit 1
    fi
    if [ -z "$context" ]; then
        echo "FATAL ERROR: docker context $context does not exist."
        exit 1
    fi
}



# Get container property safely
get_container_property() {
    local container="$1"
    local format="$2"
    docker --context="$context" inspect --format="$format" "$container" 2>/dev/null || echo ""
}

# ====== Docker Daemon Security Checks ======

check_daemon_security() {
    print_subheader "Docker Daemon Security"

    # Check if Docker is installed and running
    if ! command_exists docker; then
        print_result "FAIL" "Docker is not installed or not in PATH"
        return 1
    fi

    if ! docker info >/dev/null 2>&1; then
        print_result "FAIL" "Docker daemon is not running or not accessible"
        return 1
    fi

    current_user="$1"
    print_result "INFO" "Running for user: $current_user"

    # Inter-container communication
    if docker --context="$context" network inspect bridge 2>/dev/null | grep -q '"EnableICC": false'; then
        print_result "PASS" "Inter-container communication on default bridge is restricted"
    else
        print_result "WARN" "Inter-container communication on default bridge is allowed"
    fi

    # Docker daemon logging level
    local log_level
    log_level=$(ps -ef | grep dockerd | grep -oP '(?<=--log-level=)\w+' || echo "")
    if [[ "$log_level" == "info" ]]; then
        print_result "PASS" "Docker logging level is set to 'info'"
    else
        print_result "WARN" "Docker logging level is not set to 'info' (${log_level:-'default'})"
    fi

    # iptables manipulation
    if docker --context="$context" info 2>/dev/null | grep -q "iptables: true"; then
        print_result "FAIL" "Docker is allowed to make changes to iptables"
    else
        print_result "PASS" "Docker is not allowed to modify iptables"
    fi

    # Storage driver check
    if docker info 2>/dev/null | grep -q "Storage Driver: aufs"; then
        print_result "FAIL" "Deprecated aufs storage driver is being used"
    else
        print_result "PASS" "Not using deprecated aufs storage driver"
    fi

    # TLS authentication
    if pgrep -f "dockerd.*--tlsverify" >/dev/null 2>&1; then
        print_result "PASS" "TLS authentication for Docker daemon is configured"
    else
        print_result "WARN" "TLS authentication for Docker daemon is NOT configured"
    fi

    # TCP socket exposure
    if pgrep -f "dockerd.*-H tcp://" >/dev/null 2>&1; then
        print_result "FAIL" "Docker daemon is listening on TCP socket"
    else
        print_result "PASS" "Docker daemon is NOT listening on TCP socket"
    fi


    # Experimental features
    local experimental
    experimental=$(docker version -f '{{.Server.Experimental}}' 2>/dev/null || echo "false")
    if [[ "$experimental" == "true" ]]; then
        print_result "FAIL" "Experimental features are enabled in production"
    else
        print_result "PASS" "Experimental features are NOT enabled"
    fi

    # rootless
    if grep -q '"currentContext": "rootless"' "/home/$context/.docker/config.json"; then
        print_result "PASS" "Rootless context is configured"
    else
        print_result "FAIL" "Rootless context is not configured"
    fi

    # no-new-privileges
    if grep -q '"no-new-privileges": true' "/home/$context/.config/docker/daemon.json"; then
        print_result "PASS" "Containers can not get new privileges"
    else
        print_result "FAIL" "Containers can escalate privileges (--no-new-privileges is not set)"
    fi

    # dns
    if grep -q 'dns' "/home/$context/.config/docker/daemon.json"; then
        print_result "PASS" "Custom DNS resolvers are configured"
    else
        print_result "FAIL" "Custom resolvers are NOT configured - this will cause dns issues"
    fi

}




# ====== Files Checks ======

check_files() {
    print_subheader "System and User Files"

    # User privilege check
    local current_uid current_gid
    current_user="$1"
    current_uid=$(id -u "$current_user")
    current_gid=$(id -g "$current_user")

    # custom path
    if grep -q "\"data-root\": \"/home/$context/docker-data\"" "/home/$context/.config/docker/daemon.json"; then
        print_result "PASS" "/home/$context/docker-data is configured for docker data."
    else
        print_result "FAIL" "/home/$context/docker-data is NOT confiugred for docker."
    fi
    
    # system files
    check_file() {
        local filepath=$1
        local filename=$(basename "$filepath")
        local fail_msg=${2:-"$filename missing."}
        local success_msg="$filename file exists."
    
        if file_readable "$filepath"; then
            print_result "PASS" "$success_msg"
        else
            print_result "FAIL" "$fail_msg"
        fi
    }
    
    check_file "/etc/apparmor.d/home.$context.bin.rootlesskit" "AppArmor profile for user does not exist."
    check_file "/home/$context/.env"
    check_file "/home/$context/docker-compose.yml" "compose file does not exist - no containers can be started."
    check_file "/home/$context/backup.env" "backup.env missing - Backups can not be configured via UI"
    check_file "/home/$context/crons.ini" "crons.ini missing - Cron Jobs can not be created via UI"
    check_file "/home/$context/custom.cnf" "custom.cnf missing - MySQL limits can not be configured via UI."
    
    check_file "/home/$context/default.vcl" "default.vcl missing - Varnish Caching can not be configured via UI."
    check_file "/home/$context/httpd.conf" "httpd.conf missing - Apache web server can not be configured via UI."
    check_file "/home/$context/openlitespeed.conf" "openlitespeed.conf missing - OpenLitespeed web server can not be configured via UI."
    check_file "/home/$context/nginx.conf" "nginx.conf missing - Nginx web server can not be configured via UI."
    check_file "/home/$context/openresty.conf" "openresty.conf missing - OpenResty web server can not be configured via UI."
    check_file "/home/$context/pma.php" "pma.php missing - phpMyAdmin autologin from UI is not working."


    # disk and inodes
    quota_output=$(quota -u "$context" 2>/dev/null)
    
    if echo "$quota_output" | grep -q "none"; then
        print_result "WARN" "No quota set for user."
    else
        line=$(echo "$quota_output" | grep '^[[:space:]]*/')
        if [ -z "$line" ]; then
            print_result "WARN" "Quota info not found for user."
        else
            used=$(echo "$line" | awk '{print $2}')
            quota=$(echo "$line" | awk '{print $3}')
            used_inodes=$(echo "$line" | awk '{print $5}')
            quota_inodes=$(echo "$line" | awk '{print $6}')
            #todo: 5 is grace if over quota!!
            
            if [ -z "$used" ] || [ -z "$quota" ] || [ "$quota" -eq 0 ]; then
                print_result "WARN" "Could not parse disk quota info properly for user."
            else
                local usage=$(( used * 100 / quota ))
                local used_gb=$(echo "scale=2; $used / 1024 / 1024" | bc)
                local quota_gb=$(echo "scale=2; $quota / 1024 / 1024" | bc)
                if [ "$usage" -ge 99 ]; then
                    print_result "FAIL" "Disk usage is over quota ($used_gb GB used of $quota_gb GB)."
                elif [ "$usage" -ge 90 ]; then
                    print_result "WARN" "Disk usage is over 90% ($used_gb GB used of $quota_gb GB)."
                elif [ "$usage" -lt 90 ]; then
                    print_result "PASS" "Disk usage is below quota ($used_gb GB used of $quota_gb GB)."
                else
                    print_result "WARN" "Disk usage percentage is unexpected: $usage% ($used_gb GB used of $quota_gb GB)."
                fi
            fi
    
            if [ -z "$used_inodes" ] || [ -z "$quota_inodes" ] || [ "$quota_inodes" -eq 0 ]; then
                print_result "WARN" "Could not parse inode quota info properly for user."
            else
                local inode_usage=$(( used_inodes * 100 / quota_inodes ))
                if [ "$inode_usage" -ge 99 ]; then
                    print_result "FAIL" "Inode usage is over quota ($used_inodes used of $quota_inodes)."
                elif [ "$inode_usage" -ge 90 ]; then
                    print_result "WARN" "Inode usage is over 90% ($used_inodes used of $quota_inodes)."
                elif [ "$inode_usage" -lt 90 ]; then
                    print_result "PASS" "Inode usage is below quota ($used_inodes used of $quota_inodes)."
                else
                    print_result "WARN" "Inode usage percentage is unexpected: $inode_usage ($used_inodes used of $quota_inodes)."
                fi
            fi
        fi
    fi


    print_result "INFO" "Checking files ownership for user: $current_user (UID: $current_uid, GID: $current_gid)"

    # files
    home_dir="/home/$context/docker-data/"
    incorrect_ownership=0

    while IFS= read -r -d $'\0' file; do
        # skip /run/.. and /var/cache/apt/..
        if [[ "$file" == *"/diff/run/"* ]] || [[ "$file" == *"/diff/var/cache/apt"* ]]; then
            continue
        fi
    
        file_uid=$(stat -c '%u' "$file")
        
        if [[ "$file_uid" -ne "$current_uid" ]]; then
            print_result "WARN" "File '$file' is owned by UID:$file_uid instead of UID:$current_uid"
            incorrect_ownership=1
            break  # Stop
        fi
    done < <(find "$home_dir" -mindepth 1 -print0)
    
    if [[ $incorrect_ownership -eq 0 ]]; then
        print_result "PASS" "All docker files are owned by UID:$current_uid"
    else
        print_result "FAIL" "Some docker files in $home_dir are not owned by UID:$current_uid"
    fi
}




# ====== Container Security Checks ======

check_container_user() {
    local container="$1"
    local name="$2"
    
    local user
    user=$(get_container_property "$container" '{{.Config.User}}')
    
    if [[ "$user" == "root" || -z "$user" ]]; then
        print_result "PASS" "$name: Container is running as root"
    else
        print_result "FAIL" "$name: Container is running as non-root user ($user)"
    fi
}

check_container_status() {
    local container="$1"
    local name="$2"

    local status
    status=$(get_container_property "$container" '{{.State.Status}}')

    case "$status" in
        running)
            print_result "PASS" "$name: Container is running"
            ;;
        created)
            print_result "INFO" "$name: Container is created but not started"
            ;;
        paused)
            print_result "WARN" "$name: Container is paused"
            ;;
        exited)
            print_result "FAIL" "$name: Container has exited"
            ;;
        dead)
            print_result "FAIL" "$name: Container is in dead state"
            ;;
        *)
            print_result "UNKNOWN" "$name: Container status is unknown ($status)"
            ;;
    esac
}

check_container_image() {
    local container="$1"
    local name="$2"
    
    # Skip specific containers as per original logic
    if [[ "$name" == php-fpm-* ]]; then
        return 0
    fi
    
    local image
    image=$(get_container_property "$container" '{{.Config.Image}}')
    
    if [[ "$image" == *"latest"* ]]; then
        print_result "WARN" "$name: Using 'latest' tag is not recommended for image: $image"
    elif [[ "$image" == *":"* ]]; then
        print_result "PASS" "$name: Using specific tag for image: $image"
    else
        print_result "WARN" "$name: Image tag not specified for image: $image"
    fi
}

check_container_health() {
    local container="$1"
    local name="$2"
    
    local healthcheck
    healthcheck=$(get_container_property "$container" '{{.Config.Healthcheck}}')
    
    if [[ "$healthcheck" == *"Test"* ]]; then
        print_result "PASS" "$name: HEALTHCHECK configured"
    else
        print_result "WARN" "$name: No HEALTHCHECK configured"
    fi
}

check_container_security_options() {
    local container="$1"
    local name="$2"
    
    local security_opt
    security_opt=$(get_container_property "$container" '{{.HostConfig.SecurityOpt}}')
    
    # SELinux check
    if [[ "$security_opt" == *"label"* ]]; then
        print_result "PASS" "$name: SELinux security options set"
    else
        print_result "WARN" "$name: SELinux options not set"
    fi
    
    # No new privileges check
    if [[ "$security_opt" == *"no-new-privileges"* ]]; then
        print_result "PASS" "$name: Container cannot acquire additional privileges"
    else
        print_result "WARN" "$name: --no-new-privileges restriction not set"
    fi
}

check_container_capabilities() {
    local container="$1"
    local name="$2"
    
    local cap_add
    cap_add=$(get_container_property "$container" '{{.HostConfig.CapAdd}}')
    
    if [[ "$cap_add" == "[]" || -z "$cap_add" ]]; then
        print_result "PASS" "$name: No extra Linux capabilities added"
    else
        print_result "WARN" "$name: Additional Linux capabilities may be added" "Capabilities: $cap_add"
    fi
}

check_container_privileged() {
    local container="$1"
    local name="$2"
    
    local privileged
    privileged=$(get_container_property "$container" '{{.HostConfig.Privileged}}')
    
    if [[ "$privileged" == "true" ]]; then
        print_result "FAIL" "$name: Privileged mode is enabled"
    else
        print_result "PASS" "$name: Privileged mode is disabled"
    fi
}

check_container_mounts() {
    local container="$1"
    local name="$2"
    
    local mounts
    mounts=$(get_container_property "$container" '{{range .Mounts}}{{.Source}}{{"\n"}}{{end}}')
    
    # Check for sensitive directory mounts
    local sensitive_mounts critical_mounts
    sensitive_mounts=$(echo "$mounts" | grep -E '^/etc/openpanel/|^/hostfs/home/|^/home/' || true)
    critical_mounts=$(echo "$mounts" | grep -E '^/etc|^/boot' | grep -vE '^/etc/openpanel/' || true)
    
    if [[ -n "$critical_mounts" ]]; then
        print_result "FAIL" "$name: Critical host directories mounted" "$critical_mounts"
    elif [[ -n "$sensitive_mounts" ]]; then
        print_result "WARN" "$name: Bind mounts under /home/ or /etc/openpanel/"
    else
        print_result "PASS" "$name: No sensitive host directories mounted"
    fi
    
    # Check for Docker socket mount
    if echo "$mounts" | grep -q "/docker.sock"; then
        print_result "FAIL" "$name: Docker socket is mounted inside the container"
    else
        print_result "PASS" "$name: Docker socket is NOT mounted"
    fi
}

check_container_processes() {
    local container="$1"
    local name="$2"
    
    # Check for SSH daemon
    if docker --context="$context" exec "$container" ps aux 2>/dev/null | grep -q "[s]shd"; then
        print_result "FAIL" "$name: sshd is running inside container"
    else
        print_result "PASS" "$name: sshd is NOT running inside container"
    fi
}

check_container_networking() {
    local container="$1"
    local name="$2"
    
    # Network mode check
    local network_mode
    network_mode=$(get_container_property "$container" '{{.HostConfig.NetworkMode}}')
    
    if [[ -z "$network_mode" ]]; then
        print_result "INFO" "$name: No network configured"
    elif [[ "$network_mode" == *"host"* ]]; then
        print_result "FAIL" "$name: Host network namespace is used"
    elif [[ "$network_mode" =~ (_www|_db) ]]; then
        print_result "PASS" "$name: Using OpenPanel network ($network_mode)"
    else
        print_result "WARN" "$name: Using custom network ($network_mode)"
    fi
    
    # Port binding checks
    local ports
    ports=$(docker --context="$context" port "$container" 2>/dev/null || true)
    
    if [[ -n "$ports" ]]; then
        # Check for privileged ports (< 1024)
        if echo "$ports" | grep -qE ':[0-9]{1,3}$'; then
            print_result "WARN" "$name: Container may be exposing privileged ports"
        fi
        
        # Check interface binding
        if echo "$ports" | grep -q '0.0.0.0'; then
            print_result "WARN" "$name: Container traffic bound to all interfaces"
        elif echo "$ports" | awk '{print $3}' | cut -d':' -f1 | grep -v '^127\.0\.0\.1$' >/dev/null; then
            print_result "WARN" "$name: Some ports are not bound to 127.0.0.1"
            local port_details
            port_details=$(echo "$ports" | awk '{print $3}' | paste -sd ' ' -)
            print_result "INFO" "$name: Open container ports: $port_details"
        else
            print_result "PASS" "$name: All ports are bound to 127.0.0.1"
        fi
    else
        print_result "PASS" "$name: No exposed ports"
    fi
}

check_container_resources() {
    local container="$1"
    local name="$2"
    
    # Memory limit check
    local mem_limit
    mem_limit=$(get_container_property "$container" '{{.HostConfig.Memory}}')
    
    if [[ "$mem_limit" -gt 0 ]] 2>/dev/null; then
        local mem_limit_gb
        mem_limit_gb=$(echo "scale=2; $mem_limit / (1024*1024*1024)" | bc 2>/dev/null || echo "N/A")
        print_result "PASS" "$name: Memory usage is limited to ${mem_limit_gb} GB"
    else
        print_result "WARN" "$name: No memory limit set"
    fi
    
    # CPU limit check
    local cpu_limit
    cpu_limit=$(get_container_property "$container" '{{.HostConfig.NanoCpus}}')
    
    if [[ "$cpu_limit" -gt 0 ]] 2>/dev/null; then
        local cpu_limit_cpus
        cpu_limit_cpus=$(echo "scale=2; $cpu_limit / 1000000000" | bc 2>/dev/null || echo "N/A")
        print_result "PASS" "$name: CPU usage is limited to ${cpu_limit_cpus} CPUs"
    else
        print_result "WARN" "$name: No CPU limit set"
    fi
    
    # PID limit check
    local pid_limit
    pid_limit=$(get_container_property "$container" '{{.HostConfig.PidsLimit}}')
    
    if [[ "$pid_limit" -gt 0 ]] 2>/dev/null; then
        print_result "PASS" "$name: PID cgroup limit is set ($pid_limit)"
    else
        print_result "WARN" "$name: PID cgroup limit is not set"
    fi
}

check_container_filesystem() {
    local container="$1"
    local name="$2"
    
    # Read-only root filesystem
    local readonly_rootfs
    readonly_rootfs=$(get_container_property "$container" '{{.HostConfig.ReadonlyRootfs}}')
    
    if [[ "$readonly_rootfs" == "true" ]]; then
        print_result "PASS" "$name: Root filesystem is read-only"
    else
        print_result "WARN" "$name: Root filesystem is NOT read-only"
    fi
}

check_container_devices() {
    local container="$1"
    local name="$2"
    
    local devices
    devices=$(get_container_property "$container" '{{json .HostConfig.Devices}}')
    
    if [[ "$devices" == *"/dev"* ]]; then
        print_result "FAIL" "$name: Host devices are exposed" "Devices: $devices"
    else
        print_result "PASS" "$name: Host devices are NOT exposed"
    fi
}

check_container_ulimits() {
    local container="$1"
    local name="$2"
    
    local ulimits
    ulimits=$(get_container_property "$container" '{{.HostConfig.Ulimits}}')
    
    if [[ "$ulimits" != "[]" && -n "$ulimits" ]]; then
        print_result "INFO" "$name: Custom ulimit configured: $ulimits"
    else
        print_result "WARN" "$name: Default ulimit is used"
    fi
}

check_container_compose_location() {
    local container="$1"
    local name="$2"
    
    # Get container labels to find compose file location
    local compose_project_dir compose_file
    compose_project_dir=$(get_container_property "$container" '{{index .Config.Labels "com.docker.compose.project.working_dir"}}')
    compose_file=$(get_container_property "$container" '{{index .Config.Labels "com.docker.compose.project.config_files"}}')
    
    # Get current user for expected path
    local expected_path
    expected_path1="/home/$context/docker-compose.yml"
    expected_path2="/hostfs/home/$context/docker-compose.yml"
    
    if [[ -n "$compose_project_dir" && -n "$compose_file" ]]; then
        # Container was started with docker-compose
        local full_compose_path="$compose_project_dir/$compose_file"
        
        # Handle case where compose_file might be absolute path
        if [[ "$compose_file" == /* ]]; then
            full_compose_path="$compose_file"
        fi
        
        if [[ "$full_compose_path" == "$expected_path1" || "$full_compose_path" == "$expected_path2" ]]; then
            print_result "PASS" "$name: Container running from OpenPanel"
        else
            print_result "FAIL" "$name: Container running outside of OpenPanel: $full_compose_path"
        fi
    else
        # Check if container has compose labels at all
        local compose_service
        compose_service=$(get_container_property "$container" '{{index .Config.Labels "com.docker.compose.service"}}')
        
        if [[ -n "$compose_service" ]]; then
            # Has compose labels but missing required info
            print_result "WARN" "$name: Container has compose labels but missing location info"
        else
            # Container started manually (not via docker-compose)
            print_result "FAIL" "$name: Container not started via docker-compose from expected location" "Expected: $expected_path"
        fi
    fi
}



check_container_name_service_match() {
    local container="$1"
    local name="$2"
    
    # Get the compose service name from container labels
    local compose_service compose_project
    compose_service=$(get_container_property "$container" '{{index .Config.Labels "com.docker.compose.service"}}')
    compose_project=$(get_container_property "$container" '{{index .Config.Labels "com.docker.compose.project"}}')
    
    if [[ -n "$compose_service" && -n "$compose_project" ]]; then
        local expected_name="${compose_service}"
        
        if [[ "$name" == "$expected_name" ]]; then
            print_result "PASS" "$name: Container name matches compose service name: $compose_service"
        else
            print_result "FAIL" "$name: Container name does not match compose service name: $compose_service"
        fi
    else
        # No compose labels found
        print_result "WARN" "$name: Container not managed by docker-compose or missing service labels"
    fi
}



check_single_container() {
    local container="$1"
    local name
    name=$(get_container_property "$container" '{{.Name}}' | sed 's|^/||')
    
    print_subheader "Container: $name"
    
    # Run all container security checks
    check_container_compose_location "$container" "$name"
    check_container_name_service_match "$container" "$name"
    check_container_status "$container" "$name"
    check_container_user "$container" "$name"
    check_container_image "$container" "$name"
    check_container_health "$container" "$name"
    check_container_security_options "$container" "$name"
    check_container_capabilities "$container" "$name"
    check_container_privileged "$container" "$name"
    check_container_mounts "$container" "$name"
    check_container_processes "$container" "$name"
    check_container_networking "$container" "$name"
    check_container_resources "$container" "$name"
    check_container_filesystem "$container" "$name"
    check_container_devices "$container" "$name"
    check_container_ulimits "$container" "$name"


}

check_all_containers() {
    print_subheader "Container Security Checks"
    
    local containers
    containers=$(docker --context="$context" ps -q)
    
    if [[ -z "$containers" ]]; then
        print_result "INFO" "No running containers found"
        return 0
    fi
    
    local container_count
    container_count=$(echo "$containers" | wc -l)
    print_result "INFO" "Found $container_count running container(s)"
    
    while IFS= read -r container; do
        [[ -n "$container" ]] && check_single_container "$container"
    done <<< "$containers"
}

# ====== Summary Functions ======

print_summary() {
    print_header "Audit Summary"
    
    local total_checks=$((PASS_COUNT + FAIL_COUNT + WARN_COUNT + INFO_COUNT))
    
    echo -e "Total checks performed: ${BLUE}$total_checks${NC}"
    echo -e "✅ Passed: ${GREEN}$PASS_COUNT${NC}"
    echo -e "❌ Failed: ${RED}$FAIL_COUNT${NC}"
    echo -e "⚠️  Warnings: ${YELLOW}$WARN_COUNT${NC}"
    echo -e "ℹ️  Info: ${BLUE}$INFO_COUNT${NC}"
    
    if [[ $FAIL_COUNT -gt 0 ]]; then
        echo -e "\n${RED}⚠️  Critical issues found! Please review and address failed checks.${NC}"
        exit 1
    elif [[ $WARN_COUNT -gt 0 ]]; then
        echo -e "\n${YELLOW}⚠️  Some improvements recommended. Please review warnings.${NC}"
        exit 0
    else
        echo -e "\n${GREEN}✅ All checks passed!${NC}"
        exit 0
    fi
}

# ====== Main Execution ======

main() {
    local user="$1"

    if [[ -z "$user" ]]; then
        echo "Error: Username not provided."
        echo "Usage: opencli user-check <username>"
        exit 1
    fi

    print_header "Checking user: $user"

    # context
    get_docker_context "$user"
    
    # service
    check_daemon_security "$user"
    echo ""

    # files
    check_files "$user"
    echo ""
    
    # containers
    check_all_containers
    echo ""
    
    # Print summary
    print_summary
}

# Check if script is being sourced or executed
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi

exit 0
