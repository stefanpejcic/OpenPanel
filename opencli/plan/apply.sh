#!/bin/bash
################################################################################
# Script Name: plan/apply.sh
# Description: Change plan for a user and apply new plan limits.
# Usage: opencli plan-apply <USERNAME> <NEW_PLAN_ID>
# Author: Petar Ćurić
# Created: 17.11.2023
# Last Modified: 06.04.2026
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

limit_text() {
    local value=$1
    local unit=$2
    local description=$3

    if [[ "$value" == "0" ]]; then
        echo "$description limit removed (unlimited)."
    else
        echo "$description limit changed to ${value}${unit}."
    fi
}

# Usage
ram_text=$(limit_text "${ram//[!0-9]/}" "GB" "total")
cpu_text=$(limit_text "$cpu" " core(s)" "total")
disk_text=$(limit_text "$storage_in_blocks" " blocks" "total")
inodes_text=$(limit_text "$inodes_limit" " inodes" "total")
hourly_email_text=$(limit_text "$max_hourly_email" "" "max hourly emails for all domains")

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
    user_id=$(id -u "$username")
    # RAM
    if ! $partial || $doram; then
        sed -i "s/^TOTAL_RAM=\"[^\"]*\"/TOTAL_RAM=\"${ram}\"/" "/home/$context/.env" # legacy
        if [[ "$ram" -eq 0 ]]; then
            systemctl set-property "user-${user_id}.slice" MemoryMax=infinity
        else
            [[ "$ram" != *G ]] && ram="${ram}G"
            systemctl set-property "user-${user_id}.slice" MemoryMax="$ram"
        fi
        echo "- Memory:     [OK]   $ram_text"
    fi
    
    # CPU
    if ! $partial || $docpu; then
        sed -i "s/^TOTAL_CPU=\"[^\"]*\"/TOTAL_CPU=\"${cpu}\"/" "/home/$context/.env" # legacy
        if [[ "$cpu" -eq 0 ]]; then
            systemctl set-property "user-${user_id}.slice" CPUQuota=infinity
        else
            cpu_percent=$(echo "$cpu * 100" | bc)
            systemctl set-property "user-${user_id}.slice" CPUQuota="${cpu_percent}%"
        fi
        echo "- CPU:        [OK]   $cpu_text"
    fi

    # Disk and Inodes
    if ! $partial || $dodsk; then
        setquota -u "$context" "$storage_in_blocks" "$storage_in_blocks" "$inodes_limit" "$inodes_limit" /
        echo "- Disk        [OK]   $disk_text"
        echo "- Inodes:     [OK]   $inodes_text"

    fi

    # Emails
    if ! $partial || $doemail; then
        if [[ $counter -lt $totalc ]]; then
            opencli email-ratelimit --username="$username" --skip-reload >/dev/null 2>&1
        else
            opencli email-ratelimit --username="$username" >/dev/null 2>&1
        fi
        echo "- Emails:     [OK]   $hourly_email_text"
        # TODO: support optional update of max_email_quota for all accounts
    fi

    # Network (bandwidth)
    if ! $partial || $donet; then
        # TODO
        echo "- Port Speed: [WARN] Not implemented yet ($bandwidth)"
    fi
done

echo "+=============================================================================+"
echo "Completed!"

# 6. refresh quotas file if disk limits were updated
if ! $partial || $dodsk; then
    nohup opencli user-quota >/dev/null 2>&1 &
    disown
fi

# 7. Cleanup logs older than 1d
find /tmp -name 'opencli_plan_apply_*' -type f -mtime +1 -exec rm {} \; >/dev/null 2>&1
