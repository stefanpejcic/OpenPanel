#!/bin/bash
################################################################################
# Script Name: plan/apply.sh
# Description: Change plan for a user and apply new plan limits.
# Usage: opencli plan-apply <USERNAME> <NEW_PLAN_ID>
# Author: Petar Ćurić
# Created: 17.11.2023
# Last Modified: 24.03.2026
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

# Usage info
usage() {
    echo "Usage: opencli plan-apply <plan_id> <username1> <username2>... [--debug] [--all] [--cpu] [--ram] [--dsk] [--net] [--email]"
    exit 1
}

# Ensure minimum params
if [ "$#" -lt 2 ]; then
    usage
fi

new_plan_id="$1"
shift

# Flags
usernames=()
partial=false
debug=false
bulk=false
docpu=false
doram=false
dodsk=false
donet=false
doemail=false


# Parse arguments
for arg in "$@"; do
    case "$arg" in
        --debug)   debug=true ;;
        --all)     bulk=true ;;
        --cpu)     partial=true; docpu=true ;;
        --ram)     partial=true; doram=true ;;
        --dsk)     partial=true; dodsk=true ;;
        --net)     partial=true; donet=true ;;
        --email)   partial=true; doemail=true ;;
        --*)       ;; # ignore unknown flags
        *)         usernames+=("$arg") ;;
    esac
done


# 1. get plan limits
source /usr/local/opencli/db.sh

IFS=$'\t' read -r cpu ram disk_limit inodes_limit max_hourly_email bandwidth < <(
    mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e "
        SELECT cpu, ram, disk_limit, inodes_limit, max_hourly_email, bandwidth
        FROM plans
        WHERE id = '$new_plan_id'
        LIMIT 1;
    "
)

numNdisk=$(echo "$disk_limit" | awk '{print $1}')
storage_in_blocks=$((numNdisk * 1024000))

# 2. fetch all users if --all
if $bulk; then
    mapfile -t usernames < <(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -e \
        "SELECT username FROM users WHERE plan_id = '$new_plan_id';")
    $debug && echo "Applying plan changes to users: ${usernames[*]}"
fi

# 3. main loop
totalc="${#usernames[@]}"
counter=0

for username in "${usernames[@]}"; do
    ((counter++))
    echo "+=============================================================================+"
    echo "Processing user: $username ($counter/$totalc)"
    echo ""

    # 4. get docker context
    read -r current_plan_id context < <(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e \
        "SELECT plan_id, server FROM users WHERE username = '$username'")

    # 5. update limits
    
    # RAM
    if ! $partial || $doram; then
        sed -i "s/^TOTAL_RAM=\"[^\"]*\"/TOTAL_RAM=\"${ram}\"/" "/home/$context/.env"
        echo "- Memory: [OK] total limit changed to ${ram//[!0-9]/}GB."
    fi
    
    # CPU
    if ! $partial || $docpu; then
        sed -i "s/^TOTAL_CPU=\"[^\"]*\"/TOTAL_CPU=\"${cpu}\"/" "/home/$context/.env"
        echo "- CPU: [OK] limit set to ${cpu}"
    fi

    # Disk and Inodes
    if ! $partial || $dodsk; then
        setquota -u "$context" "$storage_in_blocks" "$storage_in_blocks" "$inodes_limit" "$inodes_limit" /
        echo "- Storage: [OK] limit set to ${storage_in_blocks} blocks and $inodes_limit inodes"
    fi

    # Emails
    if ! $partial || $doemail; then
        opencli email-server ratelimit --username=$username
        # TODO: support optional update of max_email_quota for all accounts
        echo "- Emails: [OK] Max hourly emails limit set to $max_hourly_email"
    fi

    # Network (bandwidth)
    if ! $partial || $donet; then
        # TODO
        echo "- Port Speed: [WARN] Not implemented yet ($bandwidth)"
    fi
done

echo "+=============================================================================+"
echo "Completed!"

# 6. quotacheck if disk limits were updated
if ! $partial || $dodsk; then
    nohup bash -c 'quotacheck -avm >/dev/null 2>&1; repquota -u / > /etc/openpanel/openpanel/core/users/repquota' >/dev/null 2>&1 &
    disown
fi

# 7. Cleanup logs older than 1d
find /tmp -name 'opencli_plan_apply_*' -type f -mtime +1 -exec rm {} \; >/dev/null 2>&1
