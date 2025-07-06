#!/bin/bash
################################################################################
# Script Name: collect_stats.sh
# Description: Collect docker usage information for all users.
# Usage: opencli docker-collect_stats
# Author: Petar Curic, Stefan Pejcic
# Created: 07.10.2023
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

(
flock -n 200 || { echo "Error: Another instance of the script is already running. Exiting."; exit 1; }
output_dir="/etc/openpanel/openpanel/core/users"
current_datetime=$(date +'%Y-%m-%d-%H-%M-%S')

usage() {
    echo "Usage: opencli docker-collect_stats <username> OR opencli docker-collect_stats --all"
    echo ""  
}


# DB
source /usr/local/opencli/db.sh



process_user() {
    local username="$1"
    output_file="$output_dir/$username/docker_usage.txt"  



    get_user_info() {
        local user="$1"
        local query="SELECT id, server FROM users WHERE username = '${user}';"
        
        # Retrieve both id and context
        user_info=$(mysql -se "$query")
        
        # Extract user_id and context from the result
        user_id=$(echo "$user_info" | awk '{print $1}')
        context=$(echo "$user_info" | awk '{print $2}')
        
        echo "$user_id,$context"
    }

    
    result=$(get_user_info "$username")
    user_id=$(echo "$result" | cut -d',' -f1)
    context=$(echo "$result" | cut -d',' -f2)
    
    if [ -z "$user_id" ]; then
        echo "FATAL ERROR: user $username does not exist."
        exit 1
    fi


# Attempt to extract detailed stats
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
  total_mem_precent: (map(.MemPerc | sub("%";"") | tonumber // 0) | add // 0 | .*100 | round / 100),
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



# Check if the first command failed by inspecting `current_usage`
if [[ "$current_usage" == "null" || -z "$current_usage" ]]; then
  #echo "Error in detailed stats extraction, falling back to simple version..."
  current_usage=$(docker --context $context stats --no-stream --format '{{json .}}' | jq -c)
fi


    echo "$current_datetime $current_usage" >> $output_file
    echo ""
    echo $current_usage
    echo ""
}  






# Check if username is provided as an argument
if [ $# -eq 0 ]; then
  usage
  exit 1
elif [[ "$1" == "--all" ]]; then

    sync
    echo 1 > /proc/sys/vm/drop_caches

  # Fetch list of users from opencli user-list --json
  users=$(opencli user-list --json | grep -v 'SUSPENDED' | awk -F'"' '/username/ {print $4}')

  # Check if no sites found
  if [[ -z "$users" || "$users" == "No users." ]]; then
    echo "No users found in the database."
    exit 1
  fi

  # Get total user count
  total_users=$(echo "$users" | wc -w)
  if command -v repquota > /dev/null 2>&1; then
      quotacheck -avm > /dev/null 2>&1
      repquota -u / > /etc/openpanel/openpanel/core/users/repquota
  fi
  # Iterate over each user
  current_user_index=1
  for user in $users; do
    echo "Processing user: $user ($current_user_index/$total_users)"
    process_user "$user"
    echo "------------------------------"
    ((current_user_index++))
  done
  echo "DONE."
    
elif [ $# -eq 1 ]; then
      if command -v repquota > /dev/null 2>&1; then
          quotacheck -avm > /dev/null 2>&1
          repquota -u / > /etc/openpanel/openpanel/core/users/repquota
      fi
  process_user "$1"
else
  usage
  exit 1
fi

)200>/root/openpanel_docker_collect_stats.lock
