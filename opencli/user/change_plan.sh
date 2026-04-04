#!/bin/bash
################################################################################
# Script Name: user/change_plan.sh
# Description: Change plan for a user and apply new plan limits.
# Usage: opencli user-change_plan <USERNAME> <NEW_PLAN_NAME>
# Author: Petar Ćurić
# Created: 17.11.2023
# Last Modified: 03.04.2026
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

if [ "$#" -ne 2 ] && [ "$#" -ne 3 ]; then
    echo "Usage: opencli user-change-plan <username> <new_plan_name>"
    exit 1
fi

USERNAME=$1
new_plan_name=$2

debug=false
for arg in "$@"; do
    if [ "$arg" == "--debug" ]; then
        debug=true
        break
    fi
done

source /usr/local/opencli/db.sh

# For old plan: id, name, context
IFS=$'\t' read -r current_plan_id current_plan_name CONTEXT < <(
    mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e "
        SELECT u.plan_id, p.name, u.server
        FROM users u
        JOIN plans p ON p.id = u.plan_id
        WHERE u.username = '$USERNAME'
        LIMIT 1"
)

# if suspended, remove prefix
USERNAME="${USERNAME##*_}" 

if [ -z "$current_plan_id" ]; then
    echo "Error: User '$USERNAME' not found in the database."
    exit 1
fi

# For new plan: id, cpu, ram, disk, inodes, port
safe_plan_name=$(printf "%s" "$new_plan_name" | sed "s/'/''/g")
IFS=$'\t' read -r new_plan_id Ncpu Nram Ndisk_limit Ninodes_limit Nbandwidth < <(
    mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e "
        SELECT id, cpu, ram, disk_limit, inodes_limit, bandwidth
        FROM plans
        WHERE name = '$safe_plan_name'
        LIMIT 1"
)

if [ -z "$new_plan_id" ]; then
    echo "Error: Plan '$new_plan_name' not found in the database."
    exit 1
fi

# convert
numNram=$(echo "$Nram" | tr -d 'g')
numNdisk=$(echo "$Ndisk_limit" | awk '{print $1}')
storage_in_blocks=$((numNdisk * 1024000))

success_count=0
failure_count=0

update_resource() {
    local type="$1"   # cpu or ram
    local value="$2"

    $debug && echo "Updating ${type^^} limit to $value"

    if [[ "$type" == "ram" && "$value" != *G ]]; then
        env_value="${value}G"
    fi
    sed -i "s/^TOTAL_${type^^}=.*/TOTAL_${type^^}=$env_value/" /home/$CONTEXT/.env

    if [[ "$type" == "cpu" ]]; then
        if [[ "$value" -eq 0 ]]; then
            systemctl set-property "user-${user_id}.slice" CPUQuota=infinity
        else
            local cpu_percent
            cpu_percent=$(echo "$value * 100" | bc)
            systemctl set-property "user-${user_id}.slice" CPUQuota="${cpu_percent}%"
        fi
    fi

    if [[ "$type" == "ram" ]]; then
        if [[ "$value" -eq 0 ]]; then
            systemctl set-property "user-${user_id}.slice" MemoryMax=infinity
        else
            [[ "$value" != *G ]] && value="${value}G"
            systemctl set-property "user-${user_id}.slice" MemoryMax="$value"
        fi
    fi

    ((success_count++))
    echo "[✔] Total ${type^^} limit ($value) set successfully for docker context."
}

update_disk_inodes() {
    $debug && echo "Changing disk limit from: $Ndisk_limit ($storage_in_blocks blocks), inodes: $Ninodes_limit"
    if setquota -u "$USERNAME" "$storage_in_blocks" "$storage_in_blocks" "$Ninodes_limit" "$Ninodes_limit" /; then
        ((success_count++))
        echo "[✔] Disk limit ($Ndisk_limit) and inodes ($Ninodes_limit) applied successfully."
    else
        ((failure_count++))
        echo "[✘] Error setting disk/inodes limits."
        echo "    Command: setquota -u $USERNAME $storage_in_blocks $storage_in_blocks $Ninodes_limit $Ninodes_limit /"
    fi
    nohup opencli user-quota >/dev/null 2>&1 &
    disown
}

update_total_tc() {
    # caddy-ratelimit
    echo "Changing port speed to $Nbandwidth is not possible at the moment."
}

change_plan_name_in_db() {
    $debug && echo "Changing plan for user from '$current_plan_name' to '$new_plan_name'"
    if mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e \
        "UPDATE users SET plan_id = $new_plan_id WHERE username = '$USERNAME';"; then
        if (( failure_count > 0 )); then
            echo "Plan changed successfully for user $USERNAME from $current_plan_name to $new_plan_name — ($failure_count warnings)"
        else
            echo "Plan changed successfully for user $USERNAME from $current_plan_name to $new_plan_name"
        fi
    else
        echo "Error: Could not update plan_id in database — is MySQL running?"
    fi
}

drop_redis_cache() {
    nohup docker --context=default exec openpanel_redis bash -c "redis-cli --raw KEYS 'flask_cache_*' | xargs -r redis-cli DEL" >/dev/null 2>&1 &
    disown
}


# Main
user_id=$(id -u $CONTEXT)
update_resource cpu "$Ncpu"
update_resource ram "$numNram"
# update_total_tc   # TODO
update_disk_inodes
change_plan_name_in_db
drop_redis_cache

exit 0
