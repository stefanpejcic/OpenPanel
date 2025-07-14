#!/bin/bash
################################################################################
# Script Name: user/suspend.sh
# Description: Suspend user: stop all containers and suspend domains.
# Usage: opencli user-suspend <USERNAME>
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
TEMPLATE_CONF="/etc/openpanel/caddy/templates/suspended_user.conf"
CADDY_VHOST_DIR="/etc/openpanel/caddy/domains"
CONFIG_FILE="/usr/local/opencli/db.sh"

# Globals
DEBUG=false
USERNAME="$1"

# Usage check
if [[ "$#" -lt 1 || "$#" -gt 2 ]]; then
    echo "Usage: opencli user-suspend <username> [--debug]"
    exit 1
fi

# Parse optional flags
for arg in "$@"; do
    [[ "$arg" == "--debug" ]] && DEBUG=true
done

# Load database configuration
source "$CONFIG_FILE"

# Retrieve Docker context for user
get_docker_context() {
    local query="SELECT server FROM users WHERE username='$USERNAME';"
    local server_name
    server_name=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query" -N)
    CONTEXT_FLAG="--context $server_name"
}

# Validate Caddy configuration and reload if valid
reload_caddy_if_valid() {
    if docker exec caddy caddy validate --config /etc/caddy/Caddyfile 2>&1 | grep -q "Valid configuration"; then
        docker --context default exec caddy caddy reload --config /etc/caddy/Caddyfile > /dev/null 2>&1
        return 0
    fi
    return 1
}

# Ensure Caddy container is running
ensure_caddy_running() {
    if ! docker --context default ps -q -f name=caddy > /dev/null; then
        cd /root && docker --context default compose up -d caddy > /dev/null 2>&1
    fi
}

# Replace user domain configs with suspended template
suspend_user_domains() {
    local user_id domain_name domain_vhost suspended_conf

    user_id=$(mysql "$mysql_database" -e "SELECT id FROM users WHERE username='$USERNAME';" -N)
    if [[ -z "$user_id" ]]; then
        echo "ERROR: User '$USERNAME' not found in the database"
        exit 1
    fi

    mkdir -p "$SUSPENDED_DIR"
    local domain_list
    domain_list=$(mysql -D "$mysql_database" -e "SELECT domain_url FROM domains WHERE user_id='$user_id';" -N)

    for domain_name in $domain_list; do
        domain_vhost="${CADDY_VHOST_DIR}/${domain_name}.conf"
        suspended_conf="${SUSPENDED_DIR}${domain_name}.conf"

        [[ ! -f "$suspended_conf" ]] && cp "$domain_vhost" "$suspended_conf"

        sed "s|<DOMAIN_NAME>|$domain_name|g" "$TEMPLATE_CONF" > "$domain_vhost"
    done

    ensure_caddy_running
    reload_caddy_if_valid || {
        for domain_name in $domain_list; do
            mv "${SUSPENDED_DIR}${domain_name}.conf" "${CADDY_VHOST_DIR}/${domain_name}.conf" > /dev/null 2>&1
        done
        reload_caddy_if_valid
    }
}

# Stop all Docker containers related to the user
stop_user_containers() {
    $DEBUG && echo "Stopping containers for user: $USERNAME"
    docker $CONTEXT_FLAG ps --format "{{.Names}}" | while read -r container; do
        $DEBUG && echo "- Stopping container: $container"
        docker $CONTEXT_FLAG stop "$container" > /dev/null 2>&1
    done
}

# Rename user in the database
rename_user_in_db() {
    local new_username="SUSPENDED_$(date +'%Y%m%d%H%M%S')_${USERNAME}"
    local query="UPDATE users SET username='${new_username}' WHERE username='${USERNAME}';"

    if mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query"; then
        echo "User '$USERNAME' suspended successfully."
    else
        echo "ERROR: Failed to suspend user '$USERNAME'."
        exit 1
    fi
}

# === MAIN EXECUTION FLOW ===

get_docker_context
suspend_user_domains
stop_user_containers
rename_user_in_db
