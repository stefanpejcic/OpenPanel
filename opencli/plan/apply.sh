#!/bin/bash
################################################################################
# Script Name: plan/apply.sh
# Description: Change plan for a user and apply new plan limits.
# Usage: opencli plan-apply <NEW_PLAN_ID> <USERNAME> 
# Author: Petar Ćurić, Stefan Pejčić
# Created: 17.11.2023
# Last Modified: 21.04.2026
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

usage() {
    echo "Usage: opencli plan-apply <plan_id> <username1> <username2>... [--debug] [--all] [--cpu] [--ram] [--dsk] [--net] [--email]"
    exit 1
}

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
        --*)       usage; exit 1 ;;
        *)         usernames+=("$arg") ;;
    esac
done

# 1. get plan limits
source /usr/local/opencli/db.sh

IFS=$'\t' read -r cpu ram disk_limit inodes_limit max_hourly_email bandwidth < <(
    mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e "SELECT cpu, ram, disk_limit, inodes_limit, max_hourly_email, bandwidth FROM plans WHERE id = '$new_plan_id' LIMIT 1;"
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
bandwidth_text=$(limit_text "$bandwidth" " mbits bandwidth" "total")

# 2. fetch all users if --all
if $bulk; then
    mapfile -t usernames < <(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -e "SELECT username FROM users WHERE plan_id = '$new_plan_id';")
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

    # 4. get docker context and UID
    read -r current_plan_id context < <(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e "SELECT plan_id, server FROM users WHERE username = '$username'")
    user_id=$(id -u "$username")
    # user_id=$(ssh -o LogLevel=ERROR $key_flag "root@$node_ip_address" "id -u $username" 2>/dev/null)

    # 5. if cpu / ram, then create the user slice first
    user_id=$(id -u "$username")
	if ( (! $partial) || ( $docpu && $doram ) ); then
        if [ ! -f "/etc/systemd/system/user-$user_id.slice.d/override.conf" ]; then
            mkdir -p /etc/systemd/system/user-$user_id.slice.d/
            cat <<EOF > /etc/systemd/system/user-$user_id.slice.d/override.conf
[Slice]
Delegate=yes
EOF
        systemctl daemon-reload
        systemctl restart user@$user_id.service
        fi
    fi

    # 6. update limits

    # RAM
    if ! $partial || $doram; then
        sed -i "s/^TOTAL_RAM=\"[^\"]*\"/TOTAL_RAM=\"${ram}\"/" "/home/$context/.env" # legacy

        ram="${ram%G}"
        ram="${ram%g}"

        if [[ "$ram" -eq 0 ]]; then
            systemctl set-property "user-${user_id}.slice" MemoryMax=infinity
        else
            ram="${ram}G"
            systemctl set-property "user-${user_id}.slice" MemoryMax="$ram"
        fi
        echo "- Memory:     [OK]   $ram_text"
    fi
    
    # CPU
    if ! $partial || $docpu; then
        sed -i "s/^TOTAL_CPU=\"[^\"]*\"/TOTAL_CPU=\"${cpu}\"/" "/home/$context/.env" # legacy
        if [[ "$cpu" -eq 0 ]]; then
            systemctl set-property "user-${user_id}.slice" CPUQuota=
        else
            cpu_percent=$(echo "$cpu * 100" | bc)
            systemctl set-property "user-${user_id}.slice" CPUQuota="${cpu_percent}%"
        fi
        echo "- CPU:        [OK]   $cpu_text"
    fi

	# TODO: cover remote context and 
	# systemctl set-property user-1002.slice TasksMax=150 # Max processes
	# systemctl set-property user-1002.slice IOWeight=500 # I/O weight

    # Disk and Inodes
    if ! $partial || $dodsk; then
        setquota -u "$context" "$storage_in_blocks" "$storage_in_blocks" "$inodes_limit" "$inodes_limit" /
        echo "- Disk        [OK]   $disk_text"
        echo "- Inodes:     [OK]   $inodes_text"
		if (! $bulk); then
			nohup opencli docker-collect_stats "${username}" >/dev/null 2>&1 &
		    disown
		fi
    fi

    # Bandwidth (Port Speed)
    if ! $partial || $donet; then
        cd "$compose_dir" && docker --context "${username}" compose up --no-start --pull never 2>/dev/null

        USER_PID=$(pgrep -u "$username" -x dockerd | head -n 1)
        [ -z "$USER_PID" ] && { echo "- Bandwidth:[WARN]   Could not find dockerd PID for $username"; continue; }

        if [ "$bandwidth" -eq 0 ]; then
            nsenter -t "$USER_PID" -n tc qdisc del dev ifb0 root 2>/dev/null
            nsenter -t "$USER_PID" -n ip link del ifb0 2>/dev/null
            for suffix in "www" "db"; do
                full_net_name="${username}_${suffix}"
                BRIDGE_ID=$(docker --context "$username" network inspect "$full_net_name" -f '{{.Id}}' 2>/dev/null | cut -c1-12)
                if [ -n "$BRIDGE_ID" ]; then
                    IFACE="br-$BRIDGE_ID"
                    nsenter -t "$USER_PID" -n tc qdisc del dev "$IFACE" ingress 2>/dev/null
                    nsenter -t "$USER_PID" -n tc qdisc del dev "$IFACE" root 2>/dev/null
                    echo "- Bandwidth:  [OK]   $bandwidth_text ($IFACE)"
                fi
            done
        else
            nsenter -t "$USER_PID" -n ip link add ifb0 type ifb 2>/dev/null
            nsenter -t "$USER_PID" -n ip link set dev ifb0 up
            nsenter -t "$USER_PID" -n tc qdisc del dev ifb0 root 2>/dev/null
            nsenter -t "$USER_PID" -n tc qdisc add dev ifb0 root handle 1: htb default 10
            nsenter -t "$USER_PID" -n tc class add dev ifb0 parent 1: classid 1:10 htb rate "${bandwidth}mbit"
            nsenter -t "$USER_PID" -n tc qdisc add dev ifb0 parent 1:10 handle 10: sfq perturb 10
    
            IFACES=()
            for suffix in "www" "db"; do
                full_net_name="${username}_${suffix}"
                BRIDGE_ID=$(docker --context "$username" network inspect "$full_net_name" -f '{{.Id}}' 2>/dev/null | cut -c1-12)
                [ -z "$BRIDGE_ID" ] && continue
                IFACE="br-$BRIDGE_ID"
                IFACES+=("$IFACE")
    
                nsenter -t "$USER_PID" -n tc qdisc del dev "$IFACE" ingress 2>/dev/null
                nsenter -t "$USER_PID" -n tc qdisc del dev "$IFACE" root 2>/dev/null
                nsenter -t "$USER_PID" -n tc qdisc add dev "$IFACE" handle ffff: ingress
                nsenter -t "$USER_PID" -n tc filter add dev "$IFACE" parent ffff: protocol ip u32 match u32 0 0 action mirred egress redirect dev ifb0
                nsenter -t "$USER_PID" -n tc qdisc add dev "$IFACE" root handle 1: htb
                nsenter -t "$USER_PID" -n tc filter add dev "$IFACE" parent 1: protocol ip u32 match u32 0 0 action mirred egress redirect dev ifb0
            done
            if [ ${#IFACES[@]} -gt 0 ]; then
                echo "- Bandwidth:  [OK]   ${bandwidth}mbit hard cap on www and db networks (bridges: ${IFACES[*]})"
            else
                echo "- Bandwidth:[WARN]   Could not find bridge ID for any networks - does user $username have any containers running?"
            fi
        fi
    fi
done

echo "+=============================================================================+"
echo "Completed!"

# 7. refresh quotas and purge logs
if $bulk; then
    nohup opencli docker-collect_stats --all >/dev/null 2>&1 &
	disown
	find /tmp -name 'opencli_plan_apply_*' -type f -mtime +1 -exec rm {} \; >/dev/null 2>&1
fi
