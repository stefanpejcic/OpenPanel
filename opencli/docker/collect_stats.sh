#!/bin/bash
################################################################################
# Script Name: collect_stats.sh
# Description: Collect docker usage information for all users.
# Usage: opencli docker-collect_stats
# Author: Petar Curic, Stefan Pejcic
# Created: 07.10.2023
# Last Modified: 08.01.2026
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

    current_usage=$(docker --context "$context" stats --no-stream --format '{{json .}}' | jq -s -c '
    def to_mib(val; unit):
        if unit == "GiB" then (val * 1024)
        elif unit == "KiB" then (val / 1024)
        elif unit == "TiB" then (val * 1024 * 1024)
        else val end;

    map(
        (.MemUsage | capture("(?<u_v>[0-9.]+)(?<u_u>[KMGT]?iB) / (?<l_v>[0-9.]+)(?<l_u>[KMGT]?iB)")) as $mem |
        (.NetIO | capture("(?<n_rx>[0-9.]+)(?<n_rx_u>[KMGT]?B) / (?<n_tx>[0-9.]+)(?<n_tx_u>[KMGT]?B)")) as $net |
        (.BlockIO | capture("(?<b_rx>[0-9.]+)(?<b_rx_u>[KMGT]?B) / (?<b_tx>[0-9.]+)(?<b_tx_u>[KMGT]?B)")) as $blk |
        {
            mem_u: to_mib($mem.u_v|tonumber; $mem.u_u),
            mem_l: to_mib($mem.l_v|tonumber; $mem.l_u),
            cpu: (.CPUPerc | sub("%";"") | tonumber),
            pids: (.PIDs | tonumber),
            net_rx: ($net.n_rx|tonumber),
            net_tx: ($net.n_tx|tonumber),
            blk_rx: ($blk.b_rx|tonumber),
            blk_tx: ($blk.b_tx|tonumber)
        }
    ) as $all | 
    {
        total_mem_u: ($all | map(.mem_u) | add),
        total_mem_l: ($all | map(.mem_l) | add),
        total_cpu: ($all | map(.cpu) | add),
        total_net_rx: ($all | map(.net_rx) | add),
        total_net_tx: ($all | map(.net_tx) | add),
        total_blk_rx: ($all | map(.blk_rx) | add),
        total_blk_tx: ($all | map(.blk_tx) | add),
        total_pids: ($all | map(.pids) | add),
        count: ($all | length)
    } | {
        BlockIO: "\(.total_blk_rx|round)B / \(.total_blk_tx|round)B",
        CPUPerc: "\(.total_cpu|.*100|round/100) %",
        Container: "\(.count)",
        ID: "",
        MemPerc: "\((if .total_mem_l > 0 then (.total_mem_u / .total_mem_l * 100) else 0 end | .*100 | round / 100)) %",
        MemUsage: "\(.total_mem_u|round)MiB / \(.total_mem_l|round)MiB",
        Name: "",
        NetIO: "\(.total_net_rx|round) / \(.total_net_tx|round)",
        PIDs: .total_pids
    }')

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
    users=($(opencli user-list --json | jq -r '.[] | select(.status != "SUSPENDED") | .username'))
else
    users=("$1")
fi

for user in "${users[@]}"; do
    process_user "$user"
done

) 200>/root/openpanel_docker_collect_stats.lock
