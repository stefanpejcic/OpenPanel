#!/bin/bash
################################################################################
# Script Name: backup.sh
# Description: Generates a backup for all users.
# Usage: opencli docker-backup
# Author: Stefan Pejcic
# Created: 22.07.2025
# Last Modified: 07.11.2025
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



run_for_user() {
    local username="$1"
    source /usr/local/opencli/db.sh
    username_query="SELECT server FROM users WHERE username = '$username'"
    context=$(mysql -D "$mysql_database" -e "$username_query" -sN)
    if [ -z "$context" ]; then
        context=$username
    fi
    
    cd /home/$context/ && docker --context=$context compose run --remove-orphans --rm --entrypoint backup backup
}

# Function to process a single user
process_user() {
    local username="$1"
    local success=true
    
    echo "Processing user: $username"

    # get context and run backup
    if ! run_for_user "$username"; then
        return 1
    fi

    return 0
}

# Function to get list of active users
get_active_users() {
    local users
    
    users=$(opencli user-list --json 2>/dev/null | grep -v 'SUSPENDED' | awk -F'"' '/username/ {print $4}')
    
    if [[ -z "$users" ]] || [[ "$users" == "No users." ]]; then
        echo "No active users found in the database"
        return 1
    fi
    
    echo "$users"
    return 0
}

# Function to process all users
process_all_users() {
    local users user_count current_index=1
    local failed_users=()
    
    echo "Fetching list of active users..."
    
    if ! users=$(get_active_users); then
        return 1
    fi
    
    user_count=$(echo "$users" | wc -w)
    echo "Found $user_count active users to process"
    
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
        echo "Successfully processed all $user_count users"
    else
        echo "Failed to process ${#failed_users[@]} users: ${failed_users[*]}"
        return 1
    fi
    
    return 0
}




process_all_users
