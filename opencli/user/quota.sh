#!/bin/bash
################################################################################
# Script Name: user/quota.sh
# Description: Enforce and recalculate disk and inodes for a user.
# Usage: opencli user-quota <username|--all>
# Author: Stefan Pejcic
# Created: 16.11.2023
# Last Modified: 03.02.2026
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

set -euo pipefail

# ======================================================================
# Constants and Variables
readonly REPQUOTA_PATH="/etc/openpanel/openpanel/core/users/repquota"
readonly GB_TO_BLOCKS=1024000
declare -g mysql_database config_file


# ======================================================================
# Helpers

usage() {
    cat << EOF
Usage: opencli user-quota <username> OR opencli user-quota --all

Arguments:
    username    Set quota for specific user
    --all       Set quota for all active users

Description:
    This script enforces and recalculates disk and inode quotas for users
    based on their plan limits stored in the database.

Examples:
    opencli user-quota stefan
    opencli user-quota --all
EOF
}

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

log_error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $*" >&2
}



# ======================================================================
# Functions

validate_db_config() {
    if [[ -z "${mysql_database:-}" ]]; then
        log_error "Database name not configured"
        return 1
    fi
    
    if [[ -z "${config_file:-}" ]]; then
        log_error "Database config file not specified"
        return 1
    fi
    
    if [[ ! -f "$config_file" ]]; then
        log_error "Database config file not found: $config_file"
        return 1
    fi
    
    return 0
}


get_plan_limits() {
    local username="$1"
    local query file_limit disk_limit
    
    # 1. get username
    if [[ -z "$username" ]]; then
        log_error "Username parameter is required"
        return 1
    fi
    
    # 2. fetch plan limits
    query="SELECT p.inodes_limit, p.disk_limit 
           FROM users u
           JOIN plans p ON u.plan_id = p.id
           WHERE u.username = '$username'"
    
    local result
    result=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e "$query" 2>/dev/null)
    
    if [[ -z "$result" ]]; then
        log_error "No plan found for user: $username"
        return 1
    fi
    
    # 3. parse and validate results
    read -r file_limit disk_limit <<< "$result"
    
    if [[ -z "$file_limit" ]] || [[ -z "$disk_limit" ]]; then
        log_error "Invalid plan limits retrieved for user: $username"
        return 1
    fi
    
    export PLAN_FILE_LIMIT="$file_limit"
    export PLAN_DISK_LIMIT="$disk_limit"

    return 0
}


validate_user_exists() {
    local username="$1"

    if ! id "$username" &>/dev/null; then
        log_error "User does not exist: $username"
        return 1
    fi

    return 0
}


set_user_quota() {
    local username="$1"
    local block_limit file_limit

    # 1. remove " GB" suffix and convert to blocks
    block_limit="${PLAN_DISK_LIMIT// GB/}"
    block_limit=$((block_limit * GB_TO_BLOCKS))
    file_limit="$PLAN_FILE_LIMIT"
    
    # 2. setquota for user
    if sudo setquota -u "$username" "$block_limit" "$block_limit" "$file_limit" "$file_limit" /; then
        log "Quota set for user $username: $block_limit blocks (${PLAN_DISK_LIMIT}) and $file_limit inodes"
        return 0
    else
        log_error "Failed to set quota for user: $username"
        return 1
    fi
}


process_user() {
    local username="$1"
    local success=true

    # 1. process single user
    log "Processing user: $username"

    # 2. Get plan limits
    if ! get_plan_limits "$username"; then
        return 1
    fi

    # 3. Check if user exists
    if ! validate_user_exists "$username"; then
        return 1
    fi

    # 4. Set plan quota
    if ! set_user_quota "$username"; then
        return 1
    fi

    return 0
}


get_active_users() {
    local users
    users=$(opencli user-list --json 2>/dev/null | grep -v 'SUSPENDED' | awk -F'"' '/username/ {print $4}')

    if [[ -z "$users" ]] || [[ "$users" == "No users." ]]; then
        log_error "No active users found in the database"
        return 1
    fi

    echo "$users"
    return 0
}


process_all_users() {
    local users user_count current_index=1
    local failed_users=()

    # 1. get all active users from database
    log "Fetching list of active users..."
    
    if ! users=$(get_active_users); then
        return 1
    fi

    user_count=$(echo "$users" | wc -w)
    log "Found $user_count active users to process"
    
    # 2. Process users 1 by 1
    for user in $users; do
        echo "Processing user: $user ($current_index/$user_count)"
        
        if ! process_user "$user"; then
            failed_users+=("$user")
        fi
        
        echo "------------------------------"
        ((current_index++))
    done
    
    # 3. Report results
    if [[ ${#failed_users[@]} -eq 0 ]]; then
        log "Successfully processed all $user_count users"
    else
        log_error "Failed to process ${#failed_users[@]} users: ${failed_users[*]}"
        return 1
    fi
    
    return 0
}

update_repquota() {
    log "Updating repquota file..."
    if repquota -u / > "$REPQUOTA_PATH" 2>/dev/null; then
        log "Repquota file updated successfully: $REPQUOTA_PATH"
    else
        log_error "Failed to update repquota file"
        return 1
    fi
    return 0
}




# ======================================================================
# Main
main() {
    local exit_code=0

    source "/usr/local/opencli/db.sh"

    # 1. test db data
    if ! validate_db_config; then
        exit 1
    fi
    
    # 2. process a single OR all users
    case "${1:-}" in
        ""|"help")
            usage
            exit 1
            ;;
        "--all")
            (( $# == 1 )) || { log_error "Invalid number of arguments"; usage; exit 1; }
            process_all_users && log "DONE: All users processed successfully" || exit_code=1
            ;;
        *)
            (( $# == 1 )) || { log_error "Invalid number of arguments"; usage; exit 1; }
            process_user "$1" || exit_code=1
            ;;
    esac
    
    # 3. Update repquota file
    update_repquota || exit_code=1
    
    exit $exit_code
}

main "$@"
