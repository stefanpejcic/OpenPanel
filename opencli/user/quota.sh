#!/bin/bash
################################################################################
# Script Name: user/quota.sh
# Description: Report or set disk and inodes for users.
# Usage: opencli user-quota <username|--all>
# Author: Stefan Pejcic
# Created: 16.11.2023
# Last Modified: 04.04.2026
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
readonly GB_TO_BLOCKS=1024000
declare -g mysql_database config_file


# ======================================================================
# Helpers

usage() {
    cat << EOF
Usage:
    opencli user-quota                     Reloads cached data for UI
    opencli user-quota --update <username> Set quota for a specific user
    opencli user-quota --update --all      Set quota for all active users

Arguments:
    username    Set quota for specific user
    --all       Set quota for all active users

Description:
    This script enforces and recalculates disk and inode quotas for users
    based on their plan limits stored in the database.
    - Running without arguments will reload cached data for UI.
    - Any quota-setting operation must start with the '--update' keyword.

Examples:
    opencli user-quota                      # reload cached information for UI 
    opencli user-quota --update stefan      # sets quota for user 'stefan'
    opencli user-quota --update --all       # sets quota for all active users
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

generate_report() {

    OUTPUT_FILE="/etc/openpanel/openpanel/quota_report.json"
    local TMP_FILE="${OUTPUT_FILE}.tmp"
    local FILTER_USER="${1-}"
    local TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # AWK: reads /etc/passwd + repquota output, builds full JSON
    repquota -u / 2>/dev/null | awk -v ts="$TIMESTAMP" -v filter="$FILTER_USER" '
    BEGIN {
        # Load UID map from /etc/passwd in one pass
        while ((getline line < "/etc/passwd") > 0) {
            n = split(line, f, ":")
            if (f[3]+0 >= 1000) uid_map[f[1]] = f[3]
        }
        close("/etc/passwd")
        printf "{\n  \"timestamp\": \"%s\",\n  \"users\": [\n", ts
        first = 1
    }
    /^[^#]/ && $2 == "--" {
        user = $1
        if (!(user in uid_map)) next
        uid = uid_map[user]
        if (filter != "" && user != filter) next
    
        # NF tells us total fields — inodes are always the last 3
        inodes_used = $(NF-2)
        inodes_soft = $(NF-1)
        inodes_hard = $NF
    
        # disk fields are always 3,4,5
        disk_used = $3
        disk_soft = $4
        disk_hard = $5
    
        if (!first) printf ",\n"
        first = 0
        printf "    {\"username\":\"%s\",\"uid\":%s,\"home_path\":\"/home/%s/\",\"disk_used\":%s,\"disk_soft\":%s,\"disk_hard\":%s,\"inodes_used\":%s,\"inodes_soft\":%s,\"inodes_hard\":%s}",
            user, uid, user, disk_used, disk_soft, disk_hard, inodes_used, inodes_soft, inodes_hard
    }
    END {
        printf "\n  ]\n}\n"
    }
    ' > "$TMP_FILE"
    
    mv -f "$TMP_FILE" "$OUTPUT_FILE"
    
    if [[ -n "$FILTER_USER" ]]; then
        awk -v u="$FILTER_USER" '
            /"username"/ && index($0, "\"" u "\"") {
                # Print from this line until closing brace
                found=1
            }
            found { print }
            found && /}/ { exit }
        ' "$OUTPUT_FILE"
    else
        cat "$OUTPUT_FILE"
    fi
}

# ======================================================================
# Main
main() {
    local exit_code=0

    # REPORT
    if [[ $# -eq 0 ]] || [[ "$1" != "--update" ]]; then
        generate_report
        exit 0
    fi

    # 3. check for 'set' command as first argument
    if [[ "$1" != "--update" ]]; then
        log_error "You must specify '--update' as the first argument to modify quotas"
        usage
        exit 1
    fi

    shift

    source "/usr/local/opencli/db.sh"

    # 4. process a single OR all users
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

    # 5. Update repquota file
    generate_report &>/dev/null
    exit $exit_code
}

main "$@"
