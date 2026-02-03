#!/bin/bash
################################################################################
# Script Name: collect_stats.sh
# Description: Collect docker usage information for all users.
# Usage: opencli docker-collect_stats
# Author: Petar Curic, Stefan Pejcic
# Created: 07.10.2023
# Last Modified: 02.02.2026
# Company: openpanel.comm
# Copyright (c) openpanel.comm
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

(
flock -n 200 || { echo "Error: Script already running."; exit 1; }

OUTPUT_DIR="/etc/openpanel/openpanel/core/users"
mkdir -p "$OUTPUT_DIR"
CURRENT_DATETIME=$(date +'%Y-%m-%d-%H-%M-%S')

source /usr/local/opencli/db.sh

# docker --context=USERNAME_HERE stats --no-stream
process_user() {
    local username="$1"
    
    read -r user_id context <<< $(mysql -se "SELECT id, server FROM users WHERE username = '$username';" | awk '{print $1, $2}')

    if [ -z "$user_id" ]; then
        echo "FATAL ERROR: user $username does not exist."
        return 1
    fi

current_usage=$(docker --context $context stats --no-stream --format '{{json .}}' | jq -s '{
  total_block_rx: (map(.BlockIO | capture("(?<rx>\\d+\\.?\\d*)([KMGT]?B) / .*") | .rx // "0" | tonumber) | add // 0 | .*100 | round / 100),
  total_block_tx: (map(.BlockIO | capture(".* / (?<tx>\\d+\\.?\\d*)([KMGT]?B)") | .tx // "0" | tonumber) | add // 0 | .*100 | round / 100),
  total_cpu: (map(.CPUPerc | sub("%";"") | tonumber // 0) | add // 0 | .*100 | round / 100),
  total_mem_usage: (map(.MemUsage | capture("(?<value>\\d+\\.?\\d*)(?<unit>[KMGT]?iB) / .*") // {"value": "0", "unit": "MiB"} |
    if .unit == "GiB" then (.value | tonumber * 1024)
    elif .unit == "KiB" then (.value | tonumber / 1024)
    else (.value | tonumber) end) | add // 0 | .*100 | round / 100),
  total_mem_limit: (map(.MemUsage | capture(".+ / (?<value>\\d+\\.?\\d*)(?<unit>[KMGT]?iB)") // {"value": "0", "unit": "MiB"} |
    if .unit == "GiB" then (.value | tonumber * 1024)
    elif .unit == "KiB" then (.value | tonumber / 1024)
    else (.value | tonumber) end) | add // 0 | .*100 | round / 100),


total_mem_precent: (
  (
    map(
      .MemUsage 
      | capture("(?<value>\\d+\\.?\\d*)(?<unit>[KMGT]?iB) / (?<limit>\\d+\\.?\\d*)(?<limit_unit>[KMGT]?iB)") 
      // {"value": "0", "unit": "MiB", "limit": "1", "limit_unit": "MiB"}
      | {
          usage_mib: (
            if .unit == "GiB" then (.value | tonumber * 1024)
            elif .unit == "KiB" then (.value | tonumber / 1024)
            elif .unit == "TiB" then (.value | tonumber * 1024 * 1024)
            else (.value | tonumber)
            end
          ),
          limit_mib: (
            if .limit_unit == "GiB" then (.limit | tonumber * 1024)
            elif .limit_unit == "KiB" then (.limit | tonumber / 1024)
            elif .limit_unit == "TiB" then (.limit | tonumber * 1024 * 1024)
            else (.limit | tonumber)
            end
          )
        }
    )
    | reduce .[] as $item ({"usage":0, "limit":0};
        {
          usage: (.usage + $item.usage_mib),
          limit: (.limit + $item.limit_mib)
        }
    )
    | if .limit > 0 then (.usage / .limit * 100) else 0 end
    | .*100 | round / 100
  )
),


  
  total_pids: (map(.PIDs | tonumber // 0) | add // 0),
  total_net_rx: (map(.NetIO | capture("(?<rx>\\d+\\.?\\d*)([KMGT]?B) / .*") | .rx // "0" | tonumber) | add // 0 | .*100 | round / 100),
  total_containers: (map(.Container) | unique | length),
  total_net_tx: (map(.NetIO | capture(".* / (?<tx>\\d+\\.?\\d*)([KMGT]?B)") | .tx // "0" | tonumber) | add // 0 | .*100 | round / 100)
} | {
  "BlockIO": "\(.total_block_rx)B / \(.total_block_tx)B",
  "CPUPerc": "\(.total_cpu) %",
  "Container": "\(.total_containers)",
  "ID": "",
  "MemPerc": "\(.total_mem_precent) %",
  "MemUsage": "\(.total_mem_usage)MiB / \(.total_mem_limit)MiB",
  "Name": "",
  "NetIO": "\(.total_net_rx) / \(.total_net_tx)",
  "PIDs": .total_pids
  }' | jq -c)
    if [[ -n "$current_usage" && "$current_usage" != "null" ]]; then
        mkdir -p "$OUTPUT_DIR/$username"
        echo "$CURRENT_DATETIME $current_usage" >> "$OUTPUT_DIR/$username/docker_usage.txt"
        echo "$current_usage"
    fi
}




if [ $# -ne 1 ]; then
    echo "Usage: opencli docker-collect_stats <username|--all>"
    exit 1
fi

if command -v repquota &>/dev/null; then
    quotacheck -avm &>/dev/null
    repquota -u / > /etc/openpanel/openpanel/core/users/repquota
fi

if [ "$1" == "--all" ]; then
    sync && echo 1 > /proc/sys/vm/drop_caches
    users=($(opencli user-list --json \
      | jq -r '.data[]
               | select(.username | startswith("SUSPENDED_") | not)
               | .username'))
else
    users=("$1")
fi

for user in "${users[@]}"; do
    process_user "$user"
done

) 200>/root/openpanel_docker_collect_stats.lock
