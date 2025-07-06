#!/bin/bash
################################################################################
# Script Name: user/list.sh
# Description: Display all users: id, username, email, plan, registered date.
# Usage: opencli user-list [--json]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 16.10.2023
# Last Modified: 04.07.2025
# Company: openpanel.co
# Copyright (c) openpanel.co
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

# this should be edited to not print full sh file path..
print_usage() {
    echo "Usage: opencli user-list [--json] [--total]"
    exit 1
}

# Flag variables
json_output=false
total_users=false

report_users_over_quota_only() {
    local repquota_file="/etc/openpanel/openpanel/core/users/repquota"

    if [ ! -f "$repquota_file" ]; then
        quotacheck -avm >/dev/null 2>&1
        repquota -u / > /etc/openpanel/openpanel/core/users/repquota 
    fi
    if grep -q '+' "$repquota_file"; then
        sed -n '3,5p' "$repquota_file"
        grep '+' "$repquota_file"
    else
        No users over quota.
    fi
}


report_users_quotas_only() {
    local repquota_file="/etc/openpanel/openpanel/core/users/repquota"

    if [ ! -f "$repquota_file" ]; then
        quotacheck -avm >/dev/null 2>&1
        repquota -u / > /etc/openpanel/openpanel/core/users/repquota 
    fi
    if grep -q 'root' "$repquota_file"; then
        tail -n +3 $repquota_file
    else
        No users quota.
    fi
}



# Loop through command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --over_quota)
            report_users_over_quota_only
            exit 0
            ;;
        --quota)
            report_users_quotas_only
            exit 0
            ;;
        --json)
            json_output=true
            shift
            ;;
        --total)
            total_users=true
            shift
            ;;
        *)
            print_usage
            ;;
    esac
done


# DB
source /usr/local/opencli/db.sh



# Count total users
if [ "$total_users" = true ]; then
    # Fetch only the count of users
    user_count=$(mysql --defaults-extra-file=$config_file -D $mysql_database -se "SELECT COUNT(*) FROM users")

    if [ "$json_output" = true ]; then
        echo "$user_count"
    else
        echo "Total number of users: $user_count"
    fi
    
exit 0
fi


ensure_jq_installed() {
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        # Detect the package manager and install jq
        if command -v apt-get &> /dev/null; then
            sudo apt-get update > /dev/null 2>&1
            sudo apt-get install -y -qq jq > /dev/null 2>&1
        elif command -v yum &> /dev/null; then
            sudo yum install -y -q jq > /dev/null 2>&1
        elif command -v dnf &> /dev/null; then
            sudo dnf install -y -q jq > /dev/null 2>&1
        else
            echo "Error: No compatible package manager found. Please install jq manually and try again."
            exit 1
        fi

        # Check if installation was successful
        if ! command -v jq &> /dev/null; then
            echo "Error: jq installation failed. Please install jq manually and try again."
            exit 1
        fi
    fi
}


# Get users information
if [ "$json_output" = true ]; then
    # For JSON output without --table option
    ensure_jq_installed
    users_data=$(mysql --defaults-extra-file=$config_file -D $mysql_database -e "SELECT users.id, users.username, users.email, plans.name AS plan_name, users.server, users.owner, users.registered_date FROM users INNER JOIN plans ON users.plan_id = plans.id;" | tail -n +2)
    json_output=$(echo "$users_data" | jq -R 'split("\n") | map(split("\t") | {id: .[0], username: .[1], email: .[2], plan_name: .[3], server: .[4], owner: .[5], registered_date: .[6]})' )
    echo "$json_output"
else
    # For Terminal output with --table option
    users_data=$(mysql --defaults-extra-file=$config_file -D $mysql_database --table -e "SELECT users.id, users.username, users.email, plans.name AS plan_name, users.server, users.owner, users.registered_date FROM users INNER JOIN plans ON users.plan_id = plans.id;")
    # Check if any data is retrieved
    if [ -n "$users_data" ]; then
        # Display data in tabular format
        echo "$users_data"
    else
        echo "No users."
    fi
fi


