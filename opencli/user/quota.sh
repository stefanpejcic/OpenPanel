#!/bin/bash
################################################################################
# Script Name: user/quota.sh
# Description: Enforce and recalculate disk and inodes for a user.
# Usage: opencli user-quota <username|--all>
# Author: Stefan Pejcic
# Created: 16.11.2023
# Last Modified: 13.07.2025
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

# Constants
readonly REPQUOTA_PATH="/etc/openpanel/openpanel/core/users/repquota"
readonly DB_CONFIG="/usr/local/opencli/db.sh"
readonly GB_TO_BLOCKS=1024000

# Source database configuration
if [[ ! -f "$DB_CONFIG" ]]; then
    echo "Error: Database configuration file not found: $DB_CONFIG" >&2
    exit 1
fi

# shellcheck source=/usr/local/opencli/db.sh
source "$DB_CONFIG"

# Global variables
declare -g mysql_database config_file

# Function to display usage information
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

# Function to log messages with timestamp
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

# Function to log errors
log_error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $*" >&2
}

# Function to validate database configuration
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

# Function to get plan limits for a user
get_plan_limits() {
    local username="$1"
    local query file_limit disk_limit
    
    # Validate username parameter
    if [[ -z "$username" ]]; then
        log_error "Username parameter is required"
        return 1
    fi
    
    # SQL query to fetch plan limits
    query="SELECT p.inodes_limit, p.disk_limit 
           FROM users u
           JOIN plans p ON u.plan_id = p.id
           WHERE u.username = '$username'"
    
    # Execute query and capture results
    local result
    result=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e "$query" 2>/dev/null)
    
    if [[ -z "$result" ]]; then
        log_error "No plan found for user: $username"
        return 1
    fi
    
    # Parse results
    read -r file_limit disk_limit <<< "$result"
    
    # Validate results
    if [[ -z "$file_limit" ]] || [[ -z "$disk_limit" ]]; then
        log_error "Invalid plan limits retrieved for user: $username"
        return 1
    fi
    
    # Export variables for use in calling function
    export PLAN_FILE_LIMIT="$file_limit"
    export PLAN_DISK_LIMIT="$disk_limit"
    
    return 0
}

# Function to validate user exists in system
validate_user_exists() {
    local username="$1"
    
    if ! id "$username" &>/dev/null; then
        log_error "User does not exist: $username"
        return 1
    fi
    
    return 0
}

# Function to set user quota
set_user_quota() {
    local username="$1"
    local block_limit file_limit
    
    # Calculate block limit (remove " GB" suffix and convert to blocks)
    block_limit="${PLAN_DISK_LIMIT// GB/}"
    block_limit=$((block_limit * GB_TO_BLOCKS))
    file_limit="$PLAN_FILE_LIMIT"
    
    # Set the user's disk quota
    if sudo setquota -u "$username" "$block_limit" "$block_limit" "$file_limit" "$file_limit" /; then
        log "Quota set for user $username: $block_limit blocks (${PLAN_DISK_LIMIT}) and $file_limit inodes"
        return 0
    else
        log_error "Failed to set quota for user: $username"
        return 1
    fi
}

# Function to process a single user
process_user() {
    local username="$1"
    local success=true
    
    log "Processing user: $username"
    
    # Get plan limits
    if ! get_plan_limits "$username"; then
        return 1
    fi
    
    # Validate user exists
    if ! validate_user_exists "$username"; then
        return 1
    fi
    
    # Set quota
    if ! set_user_quota "$username"; then
        return 1
    fi
    
    return 0
}

# Function to get list of active users
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

# Function to process all users
process_all_users() {
    local users user_count current_index=1
    local failed_users=()
    
    log "Fetching list of active users..."
    
    if ! users=$(get_active_users); then
        return 1
    fi
    
    user_count=$(echo "$users" | wc -w)
    log "Found $user_count active users to process"
    
    # Process each user
    for user in $users; do
        echo "Processing user: $user ($current_index/$user_count)"
        
        if ! process_user "$user"; then
            failed_users+=("$user")
        fi
        
        echo "------------------------------"
        ((current_index++))
    done
    
    # Report results
    if [[ ${#failed_users[@]} -eq 0 ]]; then
        log "Successfully processed all $user_count users"
    else
        log_error "Failed to process ${#failed_users[@]} users: ${failed_users[*]}"
        return 1
    fi
    
    return 0
}

# Function to update repquota file
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

# Main function
main() {
    local exit_code=0
    
    # Validate database configuration
    if ! validate_db_config; then
        exit 1
    fi
    
    # Parse command line arguments
    case "${1:-}" in
        "")
            usage
            exit 1
            ;;
        "--all")
            if [[ $# -ne 1 ]]; then
                log_error "Invalid number of arguments"
                usage
                exit 1
            fi
            
            if process_all_users; then
                log "DONE: All users processed successfully"
            else
                exit_code=1
            fi
            ;;
        *)
            if [[ $# -ne 1 ]]; then
                log_error "Invalid number of arguments"
                usage
                exit 1
            fi
            
            if ! process_user "$1"; then
                exit_code=1
            fi
            ;;
    esac
    
    # Update repquota file
    if ! update_repquota; then
        exit_code=1
    fi
    
    exit $exit_code
}

# Run main function with all arguments
main "$@"
