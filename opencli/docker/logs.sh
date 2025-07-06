#!/bin/bash
################################################################################
# Script Name: logs.sh
# Description: Display log sizes for user and ssytem containers
# Usage: opencli docker-logs [--all|system|<USERNAME>]
# Author: Stefan Pejcic
# Created: 28.05.2025
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

print_logs() {
    local context=$1
    local log_dir

    if [ -z "$context" ] || [ "$context" == "default" ]; then
        log_dir="/var/lib/docker/containers"
        echo "System Containers"
    else
        log_dir="/home/$context/docker-data/containers"
        echo "Context: ${context}"
    fi

    if [ ! -d "$log_dir" ]; then
        echo "- Log directory not found for context '$context' ($log_dir)"
        return
    fi
    echo "Container Name | Log Size"
    echo "----------------------------------------------------"

    tmpfile=$(mktemp)

    for log_file in "$log_dir"/*/*-json.log; do
        [ -e "$log_file" ] || continue

        container_id=$(basename "$(dirname "$log_file")")

        container_name=$(docker --context=${context:-default} inspect --format='{{.Name}}' "$container_id" 2>/dev/null | sed 's/^\/\(.*\)/\1/')

        if [ -z "$container_name" ]; then
            continue
        fi

        log_size_bytes=$(stat -c%s "$log_file")
        log_size_human=$(du -h "$log_file" | cut -f1)

        echo "$log_size_bytes $container_name | $log_size_human" >> "$tmpfile"
    done

    sort -nr "$tmpfile" | cut -d' ' -f2-

    rm "$tmpfile"

    echo ""
}





# Display usage information
usage() {
    echo "Usage: opencli docker-logs [options]"
    echo ""
    echo "Options:"
    echo "  <USERNAME>                                    Display log sizes for specified user."
    echo "  --system                                      Display log sizes just for system containers."
    echo "  --users                                       Display log sizes just for user containers."
    echo "  --all                                         Display log sizes for all user and system containers."
    echo ""
    echo "Examples:"
    echo "  opencli docker-logs stefan"
    echo "  opencli docker-logs --users"
    echo "  opencli docker-logs --system"
    echo "  opencli docker-logs --all"
    exit 1
}



# Main logic
if [ "$1" == "--all" ]; then
    # including system
    contexts=$(docker context ls --format '{{.Name}}')
    for ctx in $contexts; do
        print_logs "$ctx"
    done
elif [ "$1" == "--users" ]; then 
    # exclude system
    contexts=$(docker context ls --format '{{.Name}}' | grep -v '^default$')
    for ctx in $contexts; do
        print_logs "$ctx"
    done
elif [ "$1" == "--system" ]; then 
    # just system
    print_logs "default"
elif [ -n "$1" ]; then
    # just one user
    print_logs "$1"
else
    usage
fi
