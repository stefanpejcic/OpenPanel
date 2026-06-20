#!/bin/bash
################################################################################
# Script Name: autostart.sh
# Description: Set services to auto-start for user on acocunt creation.
# Usage: opencli docker-autostart
# Author: Stefan Pejcic
# Created: 14.05.2026
# Last Modified: 19.06.2026
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

readonly COMPOSE_FILE="/etc/openpanel/docker/compose/1.0/docker-compose.yml"
AUTOSTART_FILE="/etc/openpanel/docker/compose/1.0/autostart.services"

usage() {
    echo "Usage: opencli docker-autostart [--set <service1,service2,...> | --add <service> | --remove <service> | --clear]"
    echo ""
    echo "  (no args)              Show current autostart services"
    echo "  --set svc1,svc2        Replace autostart list with given services"
    echo "  --add svc              Add a service to autostart"
    echo "  --remove svc           Remove a service from autostart"
    echo "  --clear                Clear all autostart services"
    exit 1
}

get_available() {
    awk '/^services:/{f=1; next} /^[a-zA-Z]/{f=0} f && /^  [a-zA-Z]/{print $1}' "$COMPOSE_FILE" 2>/dev/null | tr -d ':'
}

get_current() {
    [[ -f "$AUTOSTART_FILE" ]] && grep -v '^\s*#' "$AUTOSTART_FILE" | grep -v '^\s*$' | sort
}

save_list() {
    # $1: newline-separated list!
    local dir
    dir=$(dirname "$AUTOSTART_FILE")
    mkdir -p "$dir"
    echo "$1" | grep -v '^\s*$' | sort -u > "$AUTOSTART_FILE"
}

validate_service() {
    local svc="$1"
    local available
    available=$(get_available)
    if ! echo "$available" | grep -qx "$svc"; then
        echo "Error: '$svc' is not a known service." >&2
        echo "Available: $(echo "$available" | tr '\n' ' ')" >&2
        exit 1
    fi
}

show_status() {
    local available current

    [[ -f "$COMPOSE_FILE" ]] || { echo "Compose file not found: $COMPOSE_FILE" >&2; exit 1; }

    available=$(get_available)
    current=$(get_current)

    echo "Available services:"
    while IFS= read -r svc; do
        if echo "$current" | grep -qx "$svc"; then
            echo "  [x] $svc"
        else
            echo "  [ ] $svc"
        fi
    done <<< "$available"

    echo ""
    if [[ -z "$current" ]]; then
        echo "Autostart: (none)"
    else
        echo "Autostart: $(echo "$current" | tr '\n' ' ')"
    fi
}

# No args = show status
if [[ $# -eq 0 ]]; then
    show_status
    exit 0
fi

case "$1" in
    --set)
        [[ -z "$2" ]] && usage
        IFS=',' read -ra services <<< "$2"
        for svc in "${services[@]}"; do
            svc="${svc// /}"
            validate_service "$svc"
        done
        new_list=$(printf '%s\n' "${services[@]}" | tr -d ' ')
        save_list "$new_list"
        echo "Autostart set to: $(echo "$new_list" | tr '\n' ' ')"
        ;;

    --add)
        [[ -z "$2" ]] && usage
        svc="${2// /}"
        validate_service "$svc"
        current=$(get_current)
        if echo "$current" | grep -qx "$svc"; then
            echo "'$svc' is already in autostart."
        else
            save_list "$(printf '%s\n%s' "$current" "$svc")"
            echo "Added '$svc' to autostart."
        fi
        ;;

    --remove)
        [[ -z "$2" ]] && usage
        svc="${2// /}"
        current=$(get_current)
        new_list=$(echo "$current" | grep -vx "$svc")
        save_list "$new_list"
        echo "Removed '$svc' from autostart."
        ;;

    --clear)
        save_list ""
        echo "Autostart list cleared."
        ;;

    *)
        usage
        ;;
esac
