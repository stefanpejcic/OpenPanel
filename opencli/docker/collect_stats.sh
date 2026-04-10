#!/bin/bash
################################################################################
# Script Name: collect_stats.sh
# Description: Collect docker usage information for all users.
# Usage: opencli docker-collect_stats
# Author: Petar Curic, Stefan Pejcic
# Created: 07.10.2023
# Last Modified: 09.04.2026
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

OUTPUT_DIR="/etc/openpanel/openpanel/core/users"

resource_usage_retention=$(grep -Eo "resource_usage_retention=[0-9]+" "/etc/openpanel/openpanel/conf/openpanel.config" | cut -d'=' -f2)
if [[ -z $resource_usage_retention ]]; then
  resource_usage_retention=100 #default
fi

# shellcheck source=/usr/local/opencli/db.sh
source /usr/local/opencli/db.sh

(
flock -n 200 || { echo "Error: Script already running."; exit 1; }

process_user() {
    local USER_NAME="$1"
    local UID_NUM=$(id -u $USER_NAME)

    if [ -z "$UID_NUM" ]; then
        echo '{"error": "Context '"$USER_NAME"' not found"}' >&2
        return 1
    fi

    local SLICE="user-${UID_NUM}.slice"
    local CGROUP="/sys/fs/cgroup/user.slice/$SLICE"

    if [ ! -d "$CGROUP" ]; then
        echo '{"error": "Cgroup for '"$USER_NAME"' not found ('"$CGROUP"')"}' >&2
        return 1
    fi

    to_h() {
        local num=$1
        if   [ "$num" -gt 1073741824 ]; then awk "BEGIN {printf \"%.1fG\", $num/1073741824}"
        elif [ "$num" -gt 1048576    ]; then awk "BEGIN {printf \"%.1fM\", $num/1048576}"
        elif [ "$num" -gt 1024       ]; then awk "BEGIN {printf \"%.1fK\", $num/1024}"
        else echo "${num}B"
        fi
    }

    cpu_human() {
        local pct=$1
        awk "BEGIN {printf \"%.1f cores\", $pct/100}"
    }

    # CPU sample 1
    local CPU_STAT1=$(grep '^usage_usec' "$CGROUP/cpu.stat" | awk '{print $2}')
    local T1=$(date +%s%N)

    # Memory
    local MEM_CURRENT=$(cat "$CGROUP/memory.current")
    local ANON=$(grep '^anon ' "$CGROUP/memory.stat" | awk '{print $2}')
    local FILE=$(grep '^file ' "$CGROUP/memory.stat" | awk '{print $2}')
    local KERNEL=$(grep '^kernel ' "$CGROUP/memory.stat" | awk '{print $2}')

    local MEM_MAX=$(systemctl show "$SLICE" -p MemoryMax 2>/dev/null | cut -d= -f2)
    if [[ -z "$MEM_MAX" || "$MEM_MAX" == "max" || "$MEM_MAX" -eq 0 ]] 2>/dev/null; then
        MEM_MAX=$SERVER_MEMORY
    fi

    local USED=$(( ANON + KERNEL ))
    local BUFF_CACHE=$FILE
    local FREE=$(( MEM_MAX - MEM_CURRENT ))
    local AVAILABLE=$(( FREE + BUFF_CACHE ))
    local MEMORY_USAGE_PCT=$(awk "BEGIN {printf \"%d\", ($USED / $MEM_MAX) * 100}")

    # Sleep for CPU delta
    sleep 0.4

    # CPU sample 2
    local CPU_STAT2=$(grep '^usage_usec' "$CGROUP/cpu.stat" | awk '{print $2}')
    local T2=$(date +%s%N)

    local CPU_DELTA=$(( CPU_STAT2 - CPU_STAT1 ))
    local INTERVAL_US=$(( (T2 - T1) / 1000 ))  # ns → μs
    local CPU_TOTAL_SERVER=$(( SERVER_CPUS * 100 ))

    local CPU_MAX_FILE="$CGROUP/cpu.max"
    if [ -f "$CPU_MAX_FILE" ]; then
        read -r QUOTA PERIOD < "$CPU_MAX_FILE"
        QUOTA_NUM=${QUOTA//[^0-9]/}
        PERIOD_NUM=${PERIOD//[^0-9]/}
        if [[ -z "$QUOTA_NUM" || "$QUOTA_NUM" -eq 0 ]] 2>/dev/null; then
            CPU_MAX_PCT=$CPU_TOTAL_SERVER
        else
            CPU_MAX_PCT=$(( QUOTA_NUM * 100 / PERIOD_NUM ))
        fi
    else
        CPU_MAX_PCT=$CPU_TOTAL_SERVER
    fi

    local CPU_USAGE_PCT=$(awk "BEGIN {printf \"%d\", ($CPU_DELTA / $INTERVAL_US) * 100}")
    local CPU_LIMIT_PCT=$(awk "BEGIN {printf \"%d\", ($CPU_USAGE_PCT / $CPU_MAX_PCT) * 100}")

    # Warnings?
    local WARN_MSG=""
    if [ "$MEMORY_USAGE_PCT" -ge 85 ] || [ "$CPU_LIMIT_PCT" -ge 85 ]; then
        WARN_MSG="\""
        [ "$MEMORY_USAGE_PCT" -ge 85 ] && WARN_MSG+="Memory at ${MEMORY_USAGE_PCT}%"
        [ "$MEMORY_USAGE_PCT" -ge 85 ] && [ "$CPU_LIMIT_PCT" -ge 85 ] && WARN_MSG+=", "
        [ "$CPU_LIMIT_PCT" -ge 85 ] && WARN_MSG+="CPU at ${CPU_LIMIT_PCT}%"
        WARN_MSG+=" — above threshold\""
    else
        WARN_MSG="null"
    fi

    local TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local current_usage
    current_usage=$(echo "{\"timestamp\":\"$TIMESTAMP\",\"user\":\"$USER_NAME\",\"uid\":$UID_NUM,\"memory\":{\"total\":{\"bytes\":$MEM_MAX,\"human\":\"$(to_h $MEM_MAX)\"},\"used\":{\"bytes\":$USED,\"human\":\"$(to_h $USED)\"},\"free\":{\"bytes\":$FREE,\"human\":\"$(to_h $FREE)\"},\"buff_cache\":{\"bytes\":$BUFF_CACHE,\"human\":\"$(to_h $BUFF_CACHE)\"},\"available\":{\"bytes\":$AVAILABLE,\"human\":\"$(to_h $AVAILABLE)\"},\"usage_pct\":$MEMORY_USAGE_PCT},\"cpu\":{\"usage\":{\"pct\":$CPU_LIMIT_PCT,\"human\":\"$(cpu_human $CPU_LIMIT_PCT)\"},\"total\":{\"pct\":$CPU_MAX_PCT,\"human\":\"$(cpu_human $CPU_MAX_PCT)\"},\"server\":{\"pct\":$CPU_TOTAL_SERVER,\"human\":\"$(cpu_human $CPU_TOTAL_SERVER)\"}},\"warning\":$WARN_MSG}")

    local usage_file="/home/$USER_NAME/resource_usage.txt"
    echo "$current_usage" >> "$usage_file"
    
    if [ -f "$usage_file" ]; then
        total_lines=$(wc -l < "$usage_file")
        if [ "$resource_usage_retention" -gt 0 ] && [ "$total_lines" -gt "$resource_usage_retention" ]; then
            tail -n "$resource_usage_retention" "$usage_file" > "$usage_file.tmp" && mv "$usage_file.tmp" "$usage_file"
        fi
    fi

    echo "$current_usage"
}

if [ $# -ne 1 ]; then
    echo "Usage: opencli docker-collect_stats <username|--all>"
    exit 1
fi

if [ "$1" == "--all" ]; then
    if command -v repquota &>/dev/null; then
        opencli user-quota #&>/dev/null
    fi

    sync && echo 1 > /proc/sys/vm/drop_caches
    users=($(opencli user-list --json \
      | jq -r '.data[]
               | select(.username | startswith("SUSPENDED_") | not)
               | .context'))
else
    users=("$1")
fi

SERVER_MEMORY=$(grep MemTotal /proc/meminfo | awk '{print $2 * 1024}')  # KB -> bytes
SERVER_CPUS=$(nproc)

for user in "${users[@]}"; do
    process_user "$user"
done

) 200>/root/openpanel_docker_collect_stats.lock
