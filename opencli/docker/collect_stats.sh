#!/bin/bash
################################################################################
# Script Name: collect_stats.sh
# Description: Collect resource usage information for user(s).
# Usage: opencli docker-collect_stats
# Author: Stefan Pejcic
# Created: 22.07.2025
# Last Modified: 10.06.2026
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

resource_usage_retention=$(grep -Eo "resource_usage_retention=[0-9]+" "/etc/openpanel/openpanel/conf/openpanel.config" | cut -d'=' -f2)
if [[ -z $resource_usage_retention ]]; then
  resource_usage_retention=100
fi

# shellcheck source=/usr/local/opencli/db.sh
source /usr/local/opencli/db.sh

(
flock -n 200 || { echo "Error: Script already running."; exit 1; }

declare -A DOCKERD_PIDS

build_dockerd_pid_map() {
    while IFS= read -r pid; do
        local uid_line
        uid_line=$(awk '/^Uid:/{print $2; exit}' "/proc/$pid/status" 2>/dev/null)
        [[ -z "$uid_line" ]] && continue
        local uname
        uname=$(getent passwd "$uid_line" 2>/dev/null | cut -d: -f1)
        [[ -z "$uname" ]] && continue
        [[ -z "${DOCKERD_PIDS[$uname]}" ]] && DOCKERD_PIDS[$uname]=$pid
    done < <(pgrep -x dockerd 2>/dev/null)
}

SEMAPHORE_DIR=$(mktemp -d /run/openpanel_stats_XXXXXX)
SEMAPHORE_FIFO="$SEMAPHORE_DIR/sem"
mkfifo "$SEMAPHORE_FIFO"

exec 3<>"$SEMAPHORE_FIFO"

semaphore_init() {
    local slots=$1
    for (( i=0; i<slots; i++ )); do
        printf 'x' >&3
    done
}

semaphore_acquire() {
    read -r -n1 -u3
}

semaphore_release() {
    printf 'x' >&3
}

semaphore_cleanup() {
    exec 3>&-
    rm -rf "$SEMAPHORE_DIR"
}
trap semaphore_cleanup EXIT

process_user() {
    local USER_NAME="$1"
    local SAMPLE_DELAY="${2:-0.1}"

    semaphore_acquire

    {
        local UID_NUM
        if ! UID_NUM=$(id -u "$USER_NAME" 2>/dev/null); then
            echo '{"error": "Context '"$USER_NAME"' not found"}' >&2
            semaphore_release
            return 1
        fi

        local SLICE="user-${UID_NUM}.slice"
        local CGROUP="/sys/fs/cgroup/user.slice/$SLICE"

        if [ ! -d "$CGROUP" ]; then
            echo '{"error": "Cgroup for '"$USER_NAME"' not found ('"$CGROUP"')"}' >&2
            semaphore_release
            return 1
        fi

        # CPU sample 1
        local CPU_STAT1
        CPU_STAT1=$(awk '/^usage_usec/{print $2; exit}' "$CGROUP/cpu.stat")
        local T1
        T1=$(date +%s%N)

        local MEM_CURRENT MEM_MAX MEM_STAT
        MEM_CURRENT=$(< "$CGROUP/memory.current")
        MEM_MAX=$(< "$CGROUP/memory.max")   # contains "max" or a byte count
        if [[ -z "$MEM_MAX" || "$MEM_MAX" == "max" || "$MEM_MAX" -eq 0 ]] 2>/dev/null; then
            MEM_MAX=$SERVER_MEMORY
        fi
        MEM_STAT=$(< "$CGROUP/memory.stat")

        local BW_LIMIT_BITS=0 BW_USED_BYTES=0 BW_USAGE_PCT=0
        local DOCKERD_PID="${DOCKERD_PIDS[$USER_NAME]:-}"
        if [[ -n "$DOCKERD_PID" ]]; then
            local TC_OUTPUT
            TC_OUTPUT=$(nsenter -t "$DOCKERD_PID" -n tc -s class show dev ifb0 2>/dev/null)
            BW_USED_BYTES=$(awk '/class htb 1:10/{found=1} found && /Sent/{print $2; exit}' <<< "$TC_OUTPUT")
            BW_LIMIT_BITS=$(awk '/class htb 1:10/{
                for(i=1;i<=NF;i++){
                    if($i=="ceil"){
                        val=$(i+1)
                        if(val~/Gbit/) { gsub(/Gbit/,"",val); val=val*1000000000 }
                        else if(val~/Mbit/) { gsub(/Mbit/,"",val); val=val*1000000 }
                        else if(val~/Kbit/) { gsub(/Kbit/,"",val); val=val*1000 }
                        printf "%d", val; exit
                    }
                }
            }' <<< "$TC_OUTPUT")
            [[ -z "$BW_USED_BYTES"  ]] && BW_USED_BYTES=0
            [[ -z "$BW_LIMIT_BITS"  ]] && BW_LIMIT_BITS=0
            local BW_LIMIT_BYTES=$(( BW_LIMIT_BITS / 8 ))
            if [[ "$BW_LIMIT_BYTES" -gt 0 ]]; then
                BW_USAGE_PCT=$(awk "BEGIN {printf \"%d\", ($BW_USED_BYTES / $BW_LIMIT_BYTES) * 100}")
            fi
        fi

        sleep "$SAMPLE_DELAY"

        local CPU_STAT2
        CPU_STAT2=$(awk '/^usage_usec/{print $2; exit}' "$CGROUP/cpu.stat")
        local T2
        T2=$(date +%s%N)

        local CPU_TOTAL_SERVER=$(( SERVER_CPUS * 100 ))

        local QUOTA PERIOD CPU_MAX_PCT
        if [ -f "$CGROUP/cpu.max" ]; then
            read -r QUOTA PERIOD < "$CGROUP/cpu.max"
            local QUOTA_NUM=${QUOTA//[^0-9]/}
            local PERIOD_NUM=${PERIOD//[^0-9]/}
            if [[ -z "$QUOTA_NUM" || "$QUOTA_NUM" -eq 0 ]] 2>/dev/null; then
                CPU_MAX_PCT=$CPU_TOTAL_SERVER
            else
                CPU_MAX_PCT=$(( QUOTA_NUM * 100 / PERIOD_NUM ))
            fi
        else
            CPU_MAX_PCT=$CPU_TOTAL_SERVER
        fi

        local TIMESTAMP
        TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

        local current_usage
        current_usage=$(awk \
            -v mem_stat="$MEM_STAT" \
            -v mem_current="$MEM_CURRENT" \
            -v mem_max="$MEM_MAX" \
            -v cpu_stat1="$CPU_STAT1" \
            -v cpu_stat2="$CPU_STAT2" \
            -v t1="$T1" \
            -v t2="$T2" \
            -v cpu_max_pct="$CPU_MAX_PCT" \
            -v cpu_total_srv="$CPU_TOTAL_SERVER" \
            -v bw_limit_bits="$BW_LIMIT_BITS" \
            -v bw_used_bytes="$BW_USED_BYTES" \
            -v bw_usage_pct="$BW_USAGE_PCT" \
            -v user="$USER_NAME" \
            -v uid="$UID_NUM" \
            -v ts="$TIMESTAMP" \
            -v srv_mem="$SERVER_MEMORY" \
        'function to_h(n,   s) {
            if      (n>1073741824) s=sprintf("%.1fG",n/1073741824)
            else if (n>1048576)   s=sprintf("%.1fM",n/1048576)
            else if (n>1024)      s=sprintf("%.1fK",n/1024)
            else                  s=n"B"
            return s
        }
        function to_h_bits(n,   s) {
            if      (n>1000000000) s=sprintf("%.1fGbit",n/1000000000)
            else if (n>1000000)   s=sprintf("%.1fMbit",n/1000000)
            else if (n>1000)      s=sprintf("%.1fKbit",n/1000)
            else                  s=n"bit"
            return s
        }
        function cpu_h(p) { return sprintf("%.1f cores",p/100) }
        BEGIN {
            # Memory
            n = split(mem_stat, lines, "\n")
            for (i=1; i<=n; i++) {
                split(lines[i], f, " ")
                if (f[1]=="anon")   anon=f[2]+0
                if (f[1]=="file")   file=f[2]+0
                if (f[1]=="kernel") kernel=f[2]+0
            }

            used      = anon + kernel
            buff      = file
            free      = mem_max - mem_current
            avail     = free + buff
            mem_pct   = int((used / mem_max) * 100)

            # CPU
            cpu_delta   = cpu_stat2 - cpu_stat1
            interval_us = int((t2 - t1) / 1000)
            cpu_usage   = (interval_us > 0) ? int((cpu_delta / interval_us) * 100) : 0
            cpu_limit   = (cpu_max_pct > 0) ? int((cpu_usage / cpu_max_pct) * 100) : 0

            # Warnings
            warn = "null"
            wmsg = ""
            if (mem_pct >= 85) {
                wmsg = "Memory at " mem_pct "%"
            }
            if (cpu_limit >= 85) {
                if (wmsg != "") wmsg = wmsg ", "
                wmsg = wmsg "CPU at " cpu_limit "%"
            }
            if (bw_usage_pct >= 90) {
                if (wmsg != "") wmsg = wmsg ", "
                wmsg = wmsg "Bandwidth at " bw_usage_pct "%"
            }
            if (wmsg != "") warn = "\"" wmsg " \xe2\x80\x94 above threshold\""

            printf "{\
\"timestamp\":\"%s\",\
\"user\":\"%s\",\
\"uid\":%s,\
\"memory\":{\
\"total\":{\"bytes\":%d,\"human\":\"%s\"},\
\"used\":{\"bytes\":%d,\"human\":\"%s\"},\
\"free\":{\"bytes\":%d,\"human\":\"%s\"},\
\"buff_cache\":{\"bytes\":%d,\"human\":\"%s\"},\
\"available\":{\"bytes\":%d,\"human\":\"%s\"},\
\"usage_pct\":%d},\
\"cpu\":{\
\"usage\":{\"pct\":%d,\"human\":\"%s\"},\
\"total\":{\"pct\":%d,\"human\":\"%s\"},\
\"server\":{\"pct\":%d,\"human\":\"%s\"}},\
\"bandwidth\":{\
\"limit\":{\"bits\":%d,\"human\":\"%s\"},\
\"total_sent\":{\"bytes\":%d,\"human\":\"%s\"},\
\"usage_pct\":%d},\
\"warning\":%s}\n",
                ts, user, uid,
                mem_max,  to_h(mem_max),
                used,     to_h(used),
                free,     to_h(free),
                buff,     to_h(buff),
                avail,    to_h(avail),
                mem_pct,
                cpu_limit, cpu_h(cpu_limit),
                cpu_max_pct, cpu_h(cpu_max_pct),
                cpu_total_srv, cpu_h(cpu_total_srv),
                bw_limit_bits, to_h_bits(bw_limit_bits),
                bw_used_bytes, to_h_bits(bw_used_bytes*8),
                bw_usage_pct,
                warn
        }')

        local usage_file="/home/$USER_NAME/resource_usage.txt"
        echo "$current_usage" >> "$usage_file"

        local total_lines
        total_lines=$(wc -l < "$usage_file")
        if [ "$resource_usage_retention" -gt 0 ] && [ "$total_lines" -gt "$resource_usage_retention" ]; then
            tail -n "$resource_usage_retention" "$usage_file" > "$usage_file.tmp" && mv "$usage_file.tmp" "$usage_file"
        fi

        echo "$current_usage"
        semaphore_release
    } &
}

# Main
if [ $# -ne 1 ]; then
    echo "Usage: opencli docker-collect_stats <username|--all>"
    exit 1
fi

SERVER_MEMORY=$(awk '/MemTotal/{print $2 * 1024}' /proc/meminfo)
SERVER_CPUS=$(nproc)

if [ "$1" == "--all" ]; then
    if command -v repquota &>/dev/null; then
        opencli user-quota &>/dev/null
    fi

    mapfile -t users < <(opencli user-list --json | jq -r '.data[] | select(.username | startswith("SUSPENDED_") | not) | .context')

    total=${#users[@]}
    if [[ $total -eq 0 ]]; then
        exit 0
    fi

    build_dockerd_pid_map

    MAX_JOBS=$(( SERVER_CPUS * 2 ))
    [[ $MAX_JOBS -gt 8 ]] && MAX_JOBS=8
    [[ $MAX_JOBS -lt 2 ]] && MAX_JOBS=2

    semaphore_init "$MAX_JOBS"

    JITTER_WINDOW=0.5
    idx=0
    for user in "${users[@]}"; do
        jitter=$(awk "BEGIN {d=($idx/$total)*$JITTER_WINDOW; printf \"%.2f\", (d<0.1?0.1:d)}")
        process_user "$user" "$jitter"
        (( idx++ )) || true
    done
    wait
else
    build_dockerd_pid_map
    semaphore_init 1

    process_user "$1"
    wait
fi

) 200>/root/openpanel_docker_collect_stats.lock
