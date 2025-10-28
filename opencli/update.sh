#!/bin/bash
################################################################################
# Script Name: update.sh
# Description: Check if update is available, install updates.
# Usage: opencli update [--check | --force]
# Author: Stefan Pejcic
# Created: 10.10.2023
# Last Modified: 27.10.2025
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

# ---------------------- CONSTANTS ---------------------- #
readonly IMAGE_NAME="openpanel/openpanel-ui"
readonly LOG_FILE="/var/log/openpanel/admin/notifications.log"
readonly CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
readonly SKIP_VERSIONS_FILE="/etc/openpanel/upgrade/skip_versions"
readonly DOCKER_CONTEXT="default"
readonly KEEP_KERNELS=2
readonly UPDATE_TIMEOUT=300

# ---------------------- COLOR CODES ---------------------- #
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # no color/restart

# ---------------------- LOGGING FUNCTIONS ---------------------- #
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

print_header() {
    local title="$1"
    echo -e "${BLUE}===== ${title} =====${NC}"
}

# ---------------------- USAGE ---------------------- #
usage() {
    cat << EOF
Usage: opencli update [OPTION]

Options:
    --check             Check if update is available
    --force             Force update even when autopatch/autoupdate is disabled
    (no argument)       Run update if autopatch/autoupdate is enabled
    -h, --help          Show this help message

Examples:
    opencli update                 # Update if auto-update is enabled
    opencli update --check         # Check for available updates
    opencli update --force         # Force update regardless of settings

EOF
    exit 1
}

# ---------------------- HELPER ---------------------- #
command_exists() {
    command -v "$1" &> /dev/null
}

# ---------------------- INSTALL PACKAGE ---------------------- #
install_package() {
    local package="$1"
    local quiet="${2:-false}"
    
    if [[ "$quiet" == "true" ]]; then
        local output_redirect="> /dev/null 2>&1"
    else
        local output_redirect=""
        log_info "Installing $package"
    fi
    
    if command_exists apt-get; then
        eval "apt-get update $output_redirect && apt-get install -y $package $output_redirect"
    elif command_exists dnf; then
        eval "dnf install -y $package $output_redirect"
    elif command_exists yum; then
        eval "yum install -y $package $output_redirect"
    else
        log_error "No supported package manager found (apt/dnf/yum)"
        return 1
    fi
}

# ---------------------- HELPER ---------------------- #
ensure_jq_installed() {
    if ! command_exists jq; then
        log_info "Installing jq"
        if ! install_package jq true; then
            log_error "Failed to install jq. Please install it manually."
            exit 1
        fi
        if ! command_exists jq; then
            log_error "jq installation failed. Please install jq manually."
            exit 1
        fi
    fi
}

# ---------------------- HELPERS ---------------------- #
get_last_message_content() {
    tail -n 1 "$LOG_FILE" 2>/dev/null || echo ""
}

is_unread_message_present() {
    local unread_message_content="$1"
    grep -q "UNREAD.*$unread_message_content" "$LOG_FILE" 2>/dev/null
}

write_notification() {
    local title="$1"
    local message="$2"
    local current_message="$(date '+%Y-%m-%d %H:%M:%S') UNREAD $title MESSAGE: $message"
    mkdir -p "$(dirname "$LOG_FILE")"
    echo "$current_message" >> "$LOG_FILE"
}

write_notification_for_update_check() {
    local title="$1"
    local message="$2"
    local last_message_content
    last_message_content=$(get_last_message_content)
    if [[ "$message" != "$last_message_content" ]] && ! is_unread_message_present "$title"; then
        write_notification "$title" "$message"
    fi
}

remove_notifications_by_pattern() {
    local pattern="$1"
    if [[ -f "$LOG_FILE" ]]; then
        sed -i "/$pattern/d" "$LOG_FILE"
    fi
}

# ---------------------- GET CURRENT ---------------------- #
get_local_version() {
    opencli version 2>/dev/null || echo "0.0.0"
}

# ---------------------- GET AVAILABLE ---------------------- #
get_remote_version() {
    local tags
    tags=$(curl -s "https://hub.docker.com/v2/repositories/${IMAGE_NAME}/tags" | jq -r '.results[].name' 2>/dev/null)
    
    if [[ -z "$tags" ]]; then
        return 1
    fi
    
    echo "$tags" | grep -v '^latest$' | sort -V | tail -n 1
}

# ---------------------- COMPARE, DUH! ---------------------- #
compare_versions() {
    local version1="$1"
    local version2="$2"
    
    if [[ "$version1" == "$version2" ]]; then
        echo 0
    elif [[ "$version1" > "$version2" ]]; then
        echo 1
    else
        echo -1
    fi
}

# ---------------------- USERS CAN SET VERSION TO BE SKIPPED ---------------------- #
is_version_skipped() {
    local version="$1"
    [[ -f "$SKIP_VERSIONS_FILE" ]] && grep -q "$version" "$SKIP_VERSIONS_FILE"
}

# ---------------------- READ AUTOPATCH AND AUTOUPDATE VALUES ---------------------- #
get_config_value() {
    local key="$1"
    local default_value="$2"
    
    if [[ -f "$CONFIG_FILE" ]]; then
        awk -F= "/^$key=/{print \$2}" "$CONFIG_FILE" | tr -d '"' | head -1
    else
        echo "$default_value"
    fi
}



# ---------------------- CHECK IF UPDATE IS AVAILABLE ---------------------- #
update_check() {
    local local_version
    local remote_version
    
    local_version=$(get_local_version)
    remote_version=$(get_remote_version)
    
    if [[ -z "$remote_version" ]]; then
        write_notification_for_update_check "Update check failed" "Failed connecting to Docker Hub"
        echo '{"error": "Error fetching remote version"}' >&2
        exit 1
    fi
    
    local comparison
    comparison=$(compare_versions "$local_version" "$remote_version")
    
    case $comparison in
        0)
            echo '{"status": "Up to date", "installed_version": "'"$local_version"'"}'
            ;;
        1)
            echo '{"status": "Local version is greater", "installed_version": "'"$local_version"'", "latest_version": "'"$remote_version"'"}'
            ;;
        -1)
            if is_version_skipped "$remote_version"; then
                echo '{"status": "Skipped version", "installed_version": "'"$local_version"'", "latest_version": "'"$remote_version"'"}'
            else
                write_notification_for_update_check "New OpenPanel update is available" "Installed version: $local_version | Available version: $remote_version"
                echo '{"status": "Update available", "installed_version": "'"$local_version"'", "latest_version": "'"$remote_version"'"}'
            fi
            ;;
    esac
}

# ---------------------- HELPERS USED IN MAJOR UPDATES ONLY ---------------------- #
detect_system() {
    if command_exists apt; then
        echo "debian"
    elif command_exists dnf; then
        echo "dnf"
    elif command_exists yum; then
        echo "yum"
    else
        log_error "[✘] Unsupported Linux distribution"
        exit 1
    fi
}

# ---------------------- RUNS ONLY ON MAJOR ---------------------- #
install_required_tools() {
    local distro
    distro=$(detect_system)
    log_info "Installing required system tools"
    case $distro in
        debian)
            if ! dpkg -s deborphan &>/dev/null; then
                install_package deborphan true
            fi
            ;;
        dnf|yum)
            if ! command_exists package-cleanup || ! command_exists needs-restarting; then
                install_package yum-utils true
            fi
            ;;
    esac
}

# ---------------------- RUNS ONLY ON MAJOR ---------------------- #
remove_old_kernels() {
    local distro
    distro=$(detect_system)
    case $distro in
        debian)
            log_info "Removing old kernels (Debian/Ubuntu)"
            local current_kernel
            current_kernel=$(uname -r)
            
            local kernels
            mapfile -t kernels < <(dpkg --list | grep linux-image | awk '{print $2}' | grep -v "$current_kernel" | sort -V)
            
            if (( ${#kernels[@]} > KEEP_KERNELS )); then
                local remove_count=$((${#kernels[@]} - KEEP_KERNELS))
                local remove_kernels=("${kernels[@]:0:$remove_count}")
                apt purge -y "${remove_kernels[@]}" >> "$log_file" 2>&1
            fi
            ;;
        dnf|yum)
            log_info "Removing old kernels (RHEL/CentOS)"
            if command_exists package-cleanup; then
                package-cleanup --oldkernels --count=$KEEP_KERNELS -y >> "$log_file" 2>&1
            fi
            ;;
    esac
}

# ---------------------- RUNS ONLY ON MAJOR ---------------------- #
update_system_packages() {
    local distro
    distro=$(detect_system)
    log_info "Updating system packages"
    case $distro in
        debian)
            apt update >> "$log_file" 2>&1
            apt upgrade -y >> "$log_file" 2>&1
            apt full-upgrade -y >> "$log_file" 2>&1
            
            log_info "Cleaning up packages"
            apt autoremove -y >> "$log_file" 2>&1
            apt autoclean -y >> "$log_file" 2>&1
            apt clean >> "$log_file" 2>&1
            
            if command_exists deborphan; then
                deborphan | xargs -r apt purge -y >> "$log_file" 2>&1
            fi
            
            dpkg -l | awk '/^rc/ { print $2 }' | xargs -r apt purge -y >> "$log_file" 2>&1
            ;;
        dnf|yum)
            local pkg_mgr="$distro"
            $pkg_mgr -y update >> "$log_file" 2>&1
            
            log_info "Cleaning up packages"
            $pkg_mgr -y autoremove >> "$log_file" 2>&1
            $pkg_mgr clean all >> "$log_file" 2>&1
            
            if command_exists package-cleanup; then
                package-cleanup --leaves --all --quiet | xargs -r $pkg_mgr remove -y >> "$log_file" 2>&1
            fi
            ;;
    esac
}

# ---------------------- RUNS ONLY ON MAJOR ---------------------- #
check_reboot_required() {
    log_info "Checking if reboot is required"
    local distro
    distro=$(detect_system)
    case $distro in
        debian)
            if [[ -f /var/run/reboot-required ]]; then
                log_warn "[!]  Reboot required!"
                [[ -f /var/run/reboot-required.pkgs ]] && cat /var/run/reboot-required.pkgs >> "$log_file"
            fi
            ;;
        dnf|yum)
            if command_exists needs-restarting && ! needs-restarting -r &>/dev/null; then
                log_warn "[!] Reboot required!"
                needs-restarting >> "$log_file" 2>&1
            fi
            ;;
    esac
}

# ---------------------- HELPER: RUNS AFTER DOWNLOADING NEW IMAGE ---------------------- #
purge_previous_images() {
    log_info "Cleaning up old Docker images"
    local all_images
    all_images=$(docker --context "$DOCKER_CONTEXT" images --format "{{.Repository}} {{.ID}}" | grep "^$IMAGE_NAME" | awk '{print $2}')
    local used_images
    used_images=$(docker --context "$DOCKER_CONTEXT" ps --format "{{.Image}}" | xargs -n1 docker inspect --format '{{.Id}}' 2>/dev/null | sort | uniq)
    for img in $all_images; do
        if echo "$used_images" | grep -q "$img"; then
            log_debug "Skipping in-use image: $img"
        else
            log_info "Removing unused image: $img"
            docker --context "$DOCKER_CONTEXT" rmi "$img" 2>/dev/null || true
        fi
    done
}

# ---------------------- RUNS ONLY ON MAJOR ---------------------- #
update_docker_compose() {
    log_info "Checking Docker Compose version"
    local dest="$HOME/.docker/cli-plugins/docker-compose"
    local backup="${dest}.bak"
    local local_version="0.0.0"
    
    if command_exists docker-compose; then
        local_version=$(docker-compose version --short 2>/dev/null || echo "0.0.0")
    elif docker compose version &> /dev/null; then
        local_version=$(docker compose version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    fi
    
    log_debug "Local Docker Compose version: $local_version"
    local latest_version
    latest_version=$(curl -s https://api.github.com/repos/docker/compose/releases/latest \
        | jq -r '.tag_name' 2>/dev/null | sed 's/^v//')
    
    if [[ -z "$latest_version" ]]; then
        log_warn "[!] Could not fetch latest Docker Compose version"
        return 1
    fi
    
    log_debug "Latest Docker Compose version: $latest_version"
    
    # Compare versions
    if [[ "$(printf "%s\n%s" "$latest_version" "$local_version" | sort -V | tail -n1)" != "$local_version" ]]; then
        log_info "Updating Docker Compose to $latest_version"
        
        local arch
        arch=$(uname -m)
        local binary
        
        case $arch in
            x86_64)
                binary="docker-compose-linux-x86_64"
                ;;
            aarch64)
                binary="docker-compose-linux-aarch64"
                ;;
            *)
                log_error "Unsupported architecture: $arch"
                return 1
                ;;
        esac
        
        local url="https://github.com/docker/compose/releases/download/v${latest_version}/${binary}"
        
        if [[ -f "$dest" ]]; then
            log_debug "Backing up current Docker Compose binary"
            cp "$dest" "$backup"
        fi
        
        mkdir -p "$(dirname "$dest")"
        if curl -L "$url" -o "$dest" && chmod +x "$dest"; then
            if docker compose version &>/dev/null; then
                log_info "[✔] Docker Compose updated successfully"
                [[ -f "$backup" ]] && rm -f "$backup"
            else
                log_error "[✘] Docker Compose update failed! Reverting"
                if [[ -f "$backup" ]]; then
                    mv "$backup" "$dest"
                    log_info "[!] Reverted to previous version"
                else
                    log_error "[✘] Backup not found. Manual intervention required"
                    return 1
                fi
            fi
        else
            log_error "[!] Failed to download Docker Compose"
            return 1
        fi
    else
        log_info "[✔] Docker Compose is already up to date"
    fi
}

# ---------------------- RUNS AFTER UPDATE ---------------------- #
run_custom_postupdate_script() {
    local script_path="/root/openpanel_run_after_update"
    
    if [[ -f "$script_path" ]]; then
        log_info "[!] Running custom post-update script: $script_path"
        log_info "Documentation: https://dev.openpanel.com/customize.html#After-update"
        bash "$script_path" 2>&1 | tee -a "$log_file"
    fi
}

# ---------------------- CHECKS IF CUSTOM FILE EXISTS AND RUNS IT ---------------------- #
run_version_specific_script() {
    local version="$1"
    local url="https://raw.githubusercontent.com/stefanpejcic/OpenPanel/refs/heads/main/version/$version/UPDATE.sh"
    log_info "Checking for version-specific update script"
    if wget --spider -q "$url" 2>/dev/null; then
        log_info "Downloading and executing version-specific script: $url"
        if timeout "$UPDATE_TIMEOUT" bash -c "wget -q -O - '$url' | bash" &>> "$log_file"; then
            log_info "[✔] Version-specific script executed successfully"
        else
            local exit_code=$?
            if [[ $exit_code -eq 124 ]]; then
                log_error "[!] Version-specific script timed out after ${UPDATE_TIMEOUT} seconds"
            else
                log_error "[!] Version-specific script failed with exit code: $exit_code"
            fi
        fi
    else
        log_info "[✔] No version-specific script"
    fi
}


# ---------------------- MAIN UPDATE FUNCTION - THIS IS WHERE MAGIC HAPPENS ---------------------- #
run_update_immediately() {
    local version="$1"
    local log_dir="/var/log/openpanel/updates"
    
    # ---------------------- 1. PREPARATION
    mkdir -p "$log_dir"
    log_file="$log_dir/$version.log"
    if [[ -f "$log_file" ]]; then
        local timestamp
        timestamp=$(date +"%Y%m%d_%H%M%S")
        log_file="${log_dir}/${version}_${timestamp}.log"
    fi
    
    touch "$log_file"
    
    log() {
        echo "" >> "$log_file"
        echo -e "${CYAN}---- ${1} ----${NC}" | tee -a "$log_file"
    }

    # ---------------------- 2. START PROCESS, WRITE NOTIFICATION
    print_header "Starting update to version $version"
    log_info "Update log: $log_file"
    write_notification "OpenPanel update started" "Started update to version $version - Log file: $log_file"
    
    # ---------------------- 3. DOWNLOAD NEW IMAGE FROM DOCKER HUB
    log "Updating OpenPanel Docker image"
    if ! timeout 60 docker image pull "${IMAGE_NAME}:${version}" 2>&1 | tee -a "$log_file"; then
        log_error "Failed to pull Docker image or command timed out: docker image pull ${IMAGE_NAME}:${version}"
        remove_notifications_by_pattern "OpenPanel update started MESSAGE"
        write_notification "OpenPanel update failed!" "OpenPanel failed to update to version $version - Log file: $log_file"
        log "Update failed!"       
        return 1
    else
        log "[✔] docker image ${IMAGE_NAME}:${version} downloaded successfully"

        log "Updating version in /root/.env"     # ------------------ 3.1 UPDATE TAG
        if [[ -f /root/.env ]]; then
            sed -i "s/^VERSION=.*$/VERSION=\"$version\"/" /root/.env
        fi

        log "Restarting OpenPanel service"       # ------------------ 3.2 RESTART PANEL
        if [[ -f /root/docker-compose.yml ]] || [[ -f /root/compose.yml ]]; then
            cd /root && docker --context "$DOCKER_CONTEXT" compose down openpanel && \
            docker --context "$DOCKER_CONTEXT" compose up -d openpanel 2>&1 | tee -a "$log_file"
        fi

        log "Cleaning up previous docker images" # ------------------ 3.3 DELETE PREVIOUS
        purge_previous_images       
    fi

    # ---------------------- 4. DOWNLOAD BASH SCRIPTS FROM GITHUB
    log "Updating OpenCLI"
    if [[ -d /usr/local/opencli ]]; then
        rm -f /usr/local/opencli/aliases.txt
        cd /usr/local/opencli && git reset --hard origin/main && git pull 2>&1 | tee -a "$log_file"
        latest_commit=$(git rev-parse origin/main)
        current_commit=$(git rev-parse HEAD)
        if [[ "$current_commit" == "$latest_commit" ]]; then
            log "[✔] OpenCLI is up-to-date"
        else
            log_error "OpenCLI is NOT up-to-date - something is blocking update. Run: 'cd /usr/local/opencli && git pull' and check for errors."
        fi       
        log "Generating list of OpenCLI commands for auto-complete"
        opencli commands &>/dev/null || true
    fi
    
    # ---------------------- 5. DOWNLOAD OPENADMIN FILES FROM GITHUB
    log "Updating OpenAdmin"
    if [[ -d /usr/local/admin ]]; then
        cd /usr/local/admin
        current_branch=$(git rev-parse --abbrev-ref HEAD)
        git fetch origin 2>&1 | tee -a "$log_file"
        git reset --hard origin/"$current_branch" 2>&1 | tee -a "$log_file"

        latest_commit=$(git rev-parse origin/"$current_branch")
        current_commit=$(git rev-parse HEAD)
        if [[ "$current_commit" == "$latest_commit" ]]; then
            log "[✔] OpenAdmin is up-to-date"
        else
            log_error "OpenAdmin is NOT up-to-date - something is blocking update. Run: 'cd /usr/local/admin && git pull' and check for errors."
        fi
    fi
    
    # ---------------------- 6. RUN VERSION-SPECIFIC FILE IF EXISTS
    run_version_specific_script "$version"

    # ---------------------- 7. IF MAJOR VERSION, ALSO UPDATE SYSTEM, DOCKER AND KERNEL     
    current_major=$(echo "$local_version" | cut -d. -f1)
    new_major=$(echo "$version" | cut -d. -f1)
    if [[ "$current_major" -lt "$new_major" ]]; then
        log "Updating system packages"
        install_required_tools
        update_system_packages
        remove_old_kernels
        check_reboot_required
        log "Updating Docker Compose"
        update_docker_compose
    else
        log "[✔] Minor update - skipping system and Docker Compose updates"
    fi

    # ---------------------- 8. RUN POST-UPDATE HOOK IF EXISTS        
    log "Checking for custom post-update scripts"
    run_custom_postupdate_script
    
    # ---------------------- 9. CLEANUP     
    remove_notifications_by_pattern "OpenPanel update started MESSAGE"
    remove_notifications_by_pattern "New OpenPanel update is available"
    write_notification "OpenPanel updated successfully!" "OpenPanel updated to version $version - Log file: $log_file"
    log "Update completed successfully!"

    # ---------------------- 10. RESTART ADMIN PANEL
    if systemctl is-active --quiet admin; then
        log "'OpenAdmin' service is running, restarting..."
        systemctl restart admin 2>&1 | tee -a "$log_file" || true
    else
        echo "[!] Service 'admin' is not running, skipping restart." | tee -a "$log_file"
    fi
}

# Main update check and execution
check_update() {
    local force_update=false
    
    if [[ "$1" == "--force" ]]; then
        force_update=true
        log_info "[!] Forcing update, ignoring autopatch and autoupdate settings"
    fi
    
    ensure_jq_installed
    local autopatch autoupdate
    if [[ "$force_update" == "true" ]]; then
        autopatch="on"
        autoupdate="on"
    else
        autopatch=$(get_config_value "autopatch" "off")
        autoupdate=$(get_config_value "autoupdate" "off")
    fi
    
    if [[ "$autopatch" != "on" && "$autoupdate" != "on" && "$force_update" != "true" ]]; then
        log_info "[!] Autopatch and Autoupdate are both disabled. No updates will be installed automatically."
        return 0
    fi
    
    local local_version remote_version
    local_version=$(get_local_version)
    remote_version=$(opencli update --check 2>/dev/null | jq -r '.latest_version' 2>/dev/null)
    
    if [[ -z "$remote_version" || "$remote_version" == "null" ]]; then
        log_info "No update available or unable to check for updates"
        return 0
    fi
    
    if is_version_skipped "$remote_version"; then
        log_info "[!] Version $remote_version is skipped due to skip_versions configuration"
        return 0
    fi
    
    local comparison
    comparison=$(compare_versions "$local_version" "$remote_version")
    
    if [[ $comparison -eq -1 || "$force_update" == "true" ]]; then
        if [[ "$autoupdate" == "off" && "$force_update" == "false" ]]; then
            log_info "Update available, but only autopatch is enabled. Installing patch"
        else
            log_info "Update available and will be automatically installed"
        fi

    (
      flock -n 200 || { echo "[✘] Error: Update process is already running."; echo "Please wait for it to complete before retrying."; exit 1; }
      run_update_immediately "$remote_version"
    ) 200>/var/lock/openpanel_update.lock
        
    else
        log_info "[✔] No update needed"
    fi
}


# ---------------------- RUNS CHECK OR STARTS UPDATE ---------------------- #
main() {
    case "${1:-}" in
        --check)
            update_check
            ;;
        "")
            check_update
            ;;
        --force)
            check_update --force
            ;;
        -h|--help)
            usage
            ;;
        *)
            log_error "[!] Unknown argument: $1"
            usage
            ;;
    esac
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi

exit 0
