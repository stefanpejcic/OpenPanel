#!/bin/bash
################################################################################
# Script Name: resources.sh
# Description: View services limits for user.
# Usage: opencli user-resources <CONTEXT> [--activate=<SERVICE_NAME>] [--deactivate=<SERVICE_NAME>] [--update_cpu=<FLOAT>] [--update_ram=<FLOAT>] [--service=<NAME>] [--json]
# Author: Stefan Pejcic
# Created: 26.02.2025
# Last Modified: 13.07.2025
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

# Process the service (first argument)
context=$1
env_file="/home/${context}/.env"
shift

json_output=false
FORCE_PULL=false
message=""
services_data=""
stop_service=""
new_service=""
service_to_update_cpu_ram=""
update_cpu=""
update_ram=""

usage() {
    cat <<EOF
Usage: opencli user-resources <context> [options]

Options:
  --json                         Output result in JSON format.
  --update_cpu=<value>           Update CPU allocation (global or per service).
  --update_ram=<value>           Update RAM allocation (global or per service).
  --service=<service>            Specify the service name to update.
  --activate=<service>           Start the specified service.
  --deactivate=<service>         Stop the specified service.
  --force                        Force image pull before activation.

Example:
  opencli user-resources stefan --json
  opencli user-resources stefan --service=apache --update_cpu=1.5
EOF
    exit 1
}

# --- Parse Arguments ---
for arg in "$@"; do
    case $arg in
        --json) json_output=true ;;
        --force) FORCE_PULL=true ;;
        --update_cpu=*) update_cpu="${arg#*=}" ;;
        --update_ram=*) update_ram="${arg#*=}" ;;
        --service=*) service_to_update_cpu_ram="${arg#*=}" ;;
        --activate=*) new_service="${arg#*=}" ;;
        --deactivate=*) stop_service="${arg#*=}" ;;
        *) echo "Invalid argument: $arg"; usage ;;
    esac
done

# --- Helpers ---
print_json() {
    echo "$1" | jq .
}

validate_number() {
    [[ "$1" =~ ^[0-9]*\.?[0-9]+$ ]] && awk -v n="$1" 'BEGIN {exit !(n >= 0 && n <= 512)}'
}

load_env_file() {
    [ -f "$env_file" ] && export $(grep -v '^#' "$env_file" | xargs)
}

check_context_and_env() {
    [ -z "$context" ] && echo "Missing context argument." && exit 1
    [ ! -f "$env_file" ] && echo "Missing env file: $env_file" && exit 1
}

normalize_service_name() {
    local name="$1"
    [[ "$name" == "mariadb" ]] && name="mysql"
    [[ "$name" == "phpmyadmin" ]] && name="PMA"
    [[ "$name" == "$context" ]] && name="OS"
    echo "${name//[-.]/_}" | tr '[:lower:]' '[:upper:]'
}

# --- Resource Updates ---
update_resource() {
    local type=$1 value=$2 var_name target

    [ -z "$value" ] && return

    [[ "$type" == "ram" ]] && value="${value//[gG]/}" && value="${value}g"

    validate_number "${value//g/}" || {
        echo "Invalid $type value: $value"
        exit 1
    }

    if [ -n "$service_to_update_cpu_ram" ]; then
        target=$(normalize_service_name "$service_to_update_cpu_ram")
        var_name="${target}_${type^^}"
        sed -i "s/^$var_name=\".*\"/$var_name=\"$value\"/" "$env_file"
        message+="<br>Updated $type for $service_to_update_cpu_ram to: $value"
        message+="<br>Note: disable and enable service to apply new $type limits."
    else
        var_name="TOTAL_${type^^}"
        sed -i "s/^$var_name=\".*\"/$var_name=\"$value\"/" "$env_file"
        message+="<br>Updated total $type to: $value"
    fi

}

# --- Service Start/Stop ---
check_service_exists() {
    docker --context "$context" compose -f "/home/$context/docker-compose.yml" config --services | grep -qw "$1"
}

check_service_running() {
    docker --context "$context" ps --format '{{.Names}}' | grep -qw "$1"
}

start_service() {
    $FORCE_PULL && docker --context "$context" compose -f "/home/$context/docker-compose.yml" pull "$1" >/dev/null 2>&1
    docker --context "$context" compose -f "/home/$context/docker-compose.yml" up -d "$1" >/dev/null 2>&1
}

stop_service() {
    docker --context "$context" compose -f "/home/$context/docker-compose.yml" down "$1" >/dev/null 2>&1
}

# --- Main Logic ---
check_context_and_env
update_resource "cpu" "$update_cpu"
update_resource "ram" "$update_ram"
load_env_file

TOTAL_CPU=$(echo "$TOTAL_CPU" | awk '{print int($1)}')
TOTAL_RAM=$(echo "$TOTAL_RAM" | sed 's/[gG]//' | awk '{print int($1)}')
TOTAL_USED_CPU=0
TOTAL_USED_RAM=0

RUNNING_SERVICES=$(docker --context "$context" ps --format '{{.Names}}')
for service in $RUNNING_SERVICES; do
    norm_name=$(normalize_service_name "$service")
    cpu_var="${norm_name}_CPU"
    ram_var="${norm_name}_RAM"
    
    cpu="${!cpu_var}"
    ram="${!ram_var}"
    ram=${ram//[gG]/}

    TOTAL_USED_CPU=$(awk "BEGIN {print $TOTAL_USED_CPU + $cpu}")
    TOTAL_USED_RAM=$(awk "BEGIN {print $TOTAL_USED_RAM + $ram}")

    [ "$json_output" = true ] && services_data+="{\"name\": \"$service\", \"cpu\": \"$cpu\", \"ram\": \"$ram\"}," || echo "- $service: CPU $cpu, RAM ${ram}G"
done

# Remove trailing comma
services_data=${services_data%,}

# STOP
if [ -n "$stop_service" ]; then
    check_service_exists "$stop_service" || {
        echo "Service $stop_service not found in compose file."
        exit 1
    }
    stop_service "$stop_service"
    check_service_running "$stop_service" && message+="<br>Failed to stop $stop_service." || message+="<br>Stopped $stop_service."
fi

# START
if [ -n "$new_service" ]; then
    check_service_exists "$new_service" || {
        echo "Service $new_service not found."
        exit 1
    }

    norm_name=$(normalize_service_name "$new_service")
    cpu_var="${norm_name}_CPU"
    ram_var="${norm_name}_RAM"
    cpu="${!cpu_var}"
    ram="${!ram_var}"

    ram=${ram//[gG]/}

    projected_cpu=$(awk "BEGIN {print $TOTAL_USED_CPU + $cpu}")
    projected_ram=$(awk "BEGIN {print $TOTAL_USED_RAM + $ram}")

    if [ "$TOTAL_CPU" -ne 0 ] && awk -v a="$projected_cpu" -v b="$TOTAL_CPU" 'BEGIN {exit !(a > b)}'; then
        message+="<br>CPU limit exceeded by starting $new_service."
    elif [ "$TOTAL_RAM" -ne 0 ] && awk -v a="$projected_ram" -v b="$TOTAL_RAM" 'BEGIN {exit !(a > b)}'; then
        message+="<br>RAM limit exceeded by starting $new_service."
    else
        start_service "$new_service"
        check_service_running "$new_service" && message+="<br>Started $new_service." || message+="<br>Failed to start $new_service."
    fi
fi

# --- Final Output ---
json_data=$(cat <<EOF
{
  "context": "$context",
  "services": [${services_data}],
  "limits": {
    "cpu": { "used": $TOTAL_USED_CPU, "total": $TOTAL_CPU },
    "ram": { "used": $TOTAL_USED_RAM, "total": $TOTAL_RAM }
  },
  "message": "$message"
}
EOF
)

if [ "$json_output" = true ]; then
    print_json "$json_data"
else
    echo "$message"
    echo ""
    echo "Total CPU: $TOTAL_USED_CPU / $TOTAL_CPU"
    echo "Total RAM: $TOTAL_USED_RAM / $TOTAL_RAM"
fi

opencli docker-collect_stats "$context" >/dev/null 2>&1 &
disown

exit 0

