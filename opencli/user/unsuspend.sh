#!/bin/bash
################################################################################
# Script Name: user/unsuspend.sh
# Description: Unsuspend user: start all containers and unsuspend domains
# Usage: opencli user-unsuspend <USERNAME>
# Author: Stefan Pejcic
# Created: 01.10.2023
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

# Constants
SUSPENDED_DIR="/etc/openpanel/caddy/suspended_domains/"
CADDY_VHOST_DIR="/etc/openpanel/caddy/domains"
CONFIG_FILE="/usr/local/opencli/db.sh"

# Globals
DEBUG=false
USERNAME="$1"

# Usage check
if [[ "$#" -lt 1 || "$#" -gt 2 ]]; then
    echo "Usage: opencli user-unsuspend <username> [--debug]"
    exit 1
fi

# Parse optional flags
for arg in "$@"; do
    [[ "$arg" == "--debug" ]] && DEBUG=true
done

# Load database configuration
source "$CONFIG_FILE"

# Retrieve Docker context for suspended user
get_docker_context() {
    local query="SELECT server FROM users WHERE username LIKE 'SUSPENDED\_%${USERNAME}';"
    local server_name
    server_name=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query" -N)
    CONTEXT_FLAG="--context $server_name"
}

# Reload Caddy config if valid
reload_caddy_if_valid() {
    if docker --context default exec caddy caddy validate --config /etc/caddy/Caddyfile > /dev/null 2>&1; then
        docker --context default exec caddy caddy reload --config /etc/caddy/Caddyfile > /dev/null 2>&1
        return 0
    fi
    return 1
}

# Ensure Caddy is running
ensure_caddy_running() {
    if ! docker ps -q -f name=caddy > /dev/null; then
        cd /root && docker --context default compose up -d caddy > /dev/null 2>&1
    fi
}

# Restore original domain configs from suspended backups
unsuspend_user_domains() {
    local user_id domain_name domain_vhost

    user_id=$(mysql "$mysql_database" -e "SELECT id FROM users WHERE username LIKE 'SUSPENDED\_%${USERNAME}';" -N)
    if [[ -z "$user_id" ]]; then
        echo "ERROR: User '$USERNAME' not found in the database"
        exit 1
    fi

    local domain_list
    domain_list=$(mysql -D "$mysql_database" -e "SELECT domain_url FROM domains WHERE user_id='$user_id';" -N)

    for domain_name in $domain_list; do
        domain_vhost="${CADDY_VHOST_DIR}/${domain_name}.conf"
        suspended_conf="${SUSPENDED_DIR}${domain_name}.conf"

        if [[ -f "$suspended_conf" ]]; then
            cp "$suspended_conf" "$domain_vhost"
        fi
    done

    ensure_caddy_running
    reload_caddy_if_valid || {
        for domain_name in $domain_list; do
            mv "${CADDY_VHOST_DIR}/${domain_name}.conf" "${SUSPENDED_DIR}${domain_name}.conf" > /dev/null 2>&1
        done
        reload_caddy_if_valid
    }
}

# Start Docker containers for the user
start_user_containers() {
    [ "$DEBUG" = true ] && echo "Starting containers for user: $USERNAME"

    docker $CONTEXT_FLAG ps -a --format "{{.Names}}" | while read -r container; do
        [ "$DEBUG" = true ] && echo "- Starting container: $container"
        docker $CONTEXT_FLAG start "$container" > /dev/null 2>&1
    done
}

# Rename user from suspended to active
rename_user_in_db() {
    local query="UPDATE users SET username='${USERNAME}' WHERE username LIKE 'SUSPENDED\\_%_${USERNAME}';"

    if mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query"; then
        echo "User '$USERNAME' unsuspended successfully."
    else
        echo "ERROR: Failed to unsuspend user '$USERNAME'."
        exit 1
    fi
}

# === MAIN EXECUTION FLOW ===

get_docker_context
unsuspend_user_domains
start_user_containers
rename_user_in_db
