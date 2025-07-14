#!/bin/bash
################################################################################
# Script Name: images.sh
# Description: Check images for updates.
# Usage: opencli docker-images [--all|<USERNAME>]
# Author: Stefan Pejcic
# Created: 05.05.2025
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

# Display usage information
usage() {
    echo "Usage: opencli docker-images [options]"
    echo ""
    echo "Options:"
    echo "  <USERNAME>                                    Check image updates for specified user."
    echo "  --all                                         Check image updates for all users."
    echo ""
    echo "Examples:"
    echo "  opencli docker-images stefan"
    echo "  opencli docker-images --all"
    exit 1
}

# Core logic for running Cup for a given context/user
run_for_context() {
    local ctx="$1"
    local mode_desc="$2"   # "user" or "context" for messaging

    # Check if the skip_cup.flag file exists for this context/user
    local skip_flag="/home/$ctx/skip_cup.flag"
    if [ -f "$skip_flag" ]; then
        echo "Skipping $mode_desc $ctx because $skip_flag exists."
        return
    fi

    # Check if the context exists
    if ! docker context inspect "$ctx" > /dev/null 2>&1; then
        echo "Error: Docker context '$ctx' does not exist."
        return 1
    fi

    # Check if there are any running containers in the context
    local running_containers
    running_containers=$(docker --context="$ctx" ps -q)
    if [ -z "$running_containers" ]; then
        echo "No containers running in context $ctx."
        return
    fi

    echo "Running Cup in context: $ctx"

    local user_id mount_flag
    user_id=$(id -u "$ctx" 2>/dev/null)

    if [ -n "$user_id" ]; then
        if [ -S "/hostfs/run/user/$user_id/docker.sock" ]; then
            mount_flag="-v /hostfs/run/user/$user_id/docker.sock:/var/run/docker.sock:ro"
        else
            mount_flag=""
        fi
    else
        mount_flag="-v /var/run/docker.sock:/var/run/docker.sock:ro"
    fi

    if [ -z "$mount_flag" ] && [ "$ctx" == "default" ]; then
        echo "Skipping context $ctx because mount_flag is empty."
        return
    fi

    # Define the output directory
    local output_dir="/hostfs/home/$ctx/docker-data/cup"
    # Check if the directory is writable
    if ! mkdir -p "$output_dir" || ! [ -w "$output_dir" ]; then
        echo "Error: Cannot write to output directory $output_dir."
        return 1
    fi

    local output_file="$output_dir/cup.json"

    # Run the Docker container
    if docker --context=$ctx run --rm $mount_flag ghcr.io/sergi0g/cup check -r > "$output_file"; then
        echo "Output for $mode_desc $ctx saved to $output_file"
        echo ""
        cat $output_file
        echo  ""
    else
        echo "Error running Cup for $mode_desc $ctx"
        return 1
    fi

    # Remove the Docker image after the run
    if ! docker --context=$ctx rmi -f ghcr.io/sergi0g/cup; then
        echo "Error removing Docker image for $mode_desc $ctx"
        return 1
    fi

    echo "Done processing $mode_desc $ctx."
}

run_for_all_users() {
    local contexts
    contexts=$(docker context ls --format '{{.Name}}' | grep -v '^default$')

    for ctx in $contexts; do
        run_for_context "$ctx" "context"
    done

    echo "Done processing all contexts."
}

run_for_single_user() {
    local user="$1"
    run_for_context "$user" "user"
}

# Main logic
if [ "$1" == "--all" ]; then
    run_for_all_users
elif [ -n "$1" ]; then
    run_for_single_user "$1"
else
    usage
fi

