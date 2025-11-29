#!/bin/bash
################################################################################
# Script Name: user/list.sh
# Description: Display all users: id, username, email, plan, registered date.
# Usage: opencli user-list [--json]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 16.10.2023
# Last Modified: 28.11.2025
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
    ensure_jq_installed


users_data=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "
    SELECT
        users.username,
        users.server,
        IF(users.owner IS NULL OR users.owner = '', 'root', users.owner) AS owner,
        plans.name AS package_name,
        IF(users.owner IS NULL OR users.owner = '', 'root', users.owner) AS package_owner,
        users.email,
        'EN_us' AS locale_code
    FROM users
    INNER JOIN plans ON users.plan_id = plans.id;
" | tail -n +2)

uid_mapped_data=""
while IFS=$'\t' read -r username server owner package_name package_owner email locale_code; do
    uid=$(id -u "$server" 2>/dev/null || echo "null")
    uid_mapped_data+=$uid$'\t'$username$'\t'$server$'\t'$owner$'\t'$package_name$'\t'$package_owner$'\t'$email$'\t'$locale_code$'\n'
done <<< "$users_data"

json_output=$(echo "$uid_mapped_data" | jq -R -s '
    split("\n") | map(select(length > 0)) |
    map(split("\t") | {
        id: (.[0] | if . == "null" then null else tonumber end),
        username: .[1],
        context: .[2],
        owner: .[3],
        package: {
            name: .[4],
            owner: .[5]
        },
        email: .[6],
        locale_code: .[7]
    }) | {
        data: .,
        metadata: {
            result: "ok"
        }
    }
')

echo "$json_output"

    exit 0
else
    users_data=$(mysql --defaults-extra-file=$config_file -D $mysql_database --table -e "SELECT users.id, users.username, users.email, plans.name AS plan_name, users.server, users.owner, users.registered_date FROM users INNER JOIN plans ON users.plan_id = plans.id;")
    if [ -n "$users_data" ]; then
        echo "$users_data"
    else
        echo "No users."
    fi
fi


