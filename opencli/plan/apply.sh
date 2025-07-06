#!/bin/bash
################################################################################
# Script Name: plan/apply.sh
# Description: Change plan for a user and apply new plan limits.
# Usage: opencli plan-apply <USERNAME> <NEW_PLAN_ID>
# Author: Petar Ćurić
# Created: 17.11.2023
# Last Modified: 04.07.2025
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

# DB
source /usr/local/opencli/db.sh

# Usage info
usage() {
    echo "Usage: opencli plan-apply <plan_id> <username1> <username2>... [--debug] [--all] [--cpu] [--ram] [--dsk] [--net]"
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

# Parse arguments
for arg in "$@"; do
    case "$arg" in
        --debug)   debug=true ;;
        --all)     bulk=true ;;
        --cpu)     partial=true; docpu=true ;;
        --ram)     partial=true; doram=true ;;
        --dsk)     partial=true; dodsk=true ;;
        --net)     partial=true; donet=true ;;
        --*)       ;; # ignore unknown flags
        *)         usernames+=("$arg") ;;
    esac
done

# Fetch bulk usernames if --all
if $bulk; then
    mapfile -t usernames < <(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -e \
        "SELECT username FROM users WHERE plan_id = '$new_plan_id';")
    $debug && echo "Applying plan changes to users: ${usernames[*]}"
fi

# Helper: DB queries
get_current_plan_id() {
    local user="$1"
    read -r current_plan_id server < <(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e \
        "SELECT plan_id, server FROM users WHERE username = '$user'")
    [ -z "$server" ] || [ "$server" = "default" ] && server="$user"
}

get_plan_limit() {
    local plan_id="$1" resource="$2"
    mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e "SELECT $resource FROM plans WHERE id = '$plan_id'"
}

get_plan_name() {
    local plan_id="$1"
    mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e "SELECT name FROM plans WHERE id = '$plan_id'"
}

reload_user_quotas() {
    quotacheck -avm >/dev/null 2>&1
    repquota -u / > /etc/openpanel/openpanel/core/users/repquota
}

# Main loop
totalc="${#usernames[@]}"
counter=0

for container in "${usernames[@]}"; do
    ((counter++))
    echo "+=============================================================================+"
    echo "Processing user: $container ($counter/$totalc)"
    echo ""

    get_current_plan_id "$container"
    current_plan_name=$(get_plan_name "$current_plan_id")
    new_plan_name=$(get_plan_name "$new_plan_id")

    # Plan limits
    Ncpu=$(get_plan_limit "$new_plan_id" "cpu")
    Ocpu=$(get_plan_limit "$current_plan_id" "cpu")
    Nram=$(get_plan_limit "$new_plan_id" "ram")
    Oram=$(get_plan_limit "$current_plan_id" "ram")
    Ndisk=$(get_plan_limit "$new_plan_id" "disk_limit")
    Ninodes=$(get_plan_limit "$new_plan_id" "inodes_limit")

    # System limits
    maxCPU=$(nproc)
    maxRAM=$(free -g | awk '/^Mem/ {print $2}')
    numOram=${Oram//[!0-9]/}
    numNram=${Nram//[!0-9]/}
    numNdisk=$(echo "$Ndisk" | awk '{print $1}')
    storage_in_blocks=$((numNdisk * 1024000))

    # RAM
    if ! $partial || $doram; then
        if (( numNram > maxRAM )); then
            echo "- Memory: [ERROR] New RAM value exceeds server limit ($numNram > $maxRAM GB)."
        else
            sed -i "s/^TOTAL_RAM=\"[^\"]*\"/TOTAL_RAM=\"${Nram}\"/" "/home/$server/.env"
            echo "- Memory: [OK] total limit changed to ${numNram}GB."
        fi
    fi

    # CPU
    if ! $partial || $docpu; then
        if (( Ncpu > maxCPU )); then
            echo "- CPU: [ERROR] Number of cores exceeds those of server ($Ncpu > $maxCPU)."
        else
            sed -i "s/^TOTAL_CPU=\"[^\"]*\"/TOTAL_CPU=\"${Ncpu}\"/" "/home/$server/.env"
            echo "- CPU: [OK] limit set to ${Ncpu}"
        fi
    fi

    # Disk/Inodes
    if ! $partial || $dodsk; then
        setquota -u "$server" "$storage_in_blocks" "$storage_in_blocks" "$Ninodes" "$Ninodes" /
        echo "- Storage: [OK] limit set to ${storage_in_blocks} blocks and $Ninodes inodes"
        reload_user_quotas
    fi

    # Network (stub)
    if ! $partial || $donet; then
        # TODO
        echo "- Port Speed: [WARN] Not implemented yet"
        :
    fi
done

echo "+=============================================================================+"
echo "Completed!"

# Cleanup
find /tmp -name 'opencli_plan_apply_*' -type f -mtime +1 -exec rm {} \; >/dev/null 2>&1
