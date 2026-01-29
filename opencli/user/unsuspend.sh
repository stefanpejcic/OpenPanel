#!/bin/bash
################################################################################
# Script Name: user/unsuspend.sh
# Description: Unsuspend user: start all containers and unsuspend domains
# Usage: opencli user-unsuspend <USERNAME>
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 28.01.2026
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

# ======================================================================
# Constants
SUSPENDED_DIR="/etc/openpanel/caddy/suspended_domains/"
CADDY_VHOST_DIR="/etc/openpanel/caddy/domains"



# ======================================================================
# Variables
DEBUG=false
USERNAME="$1"



# ======================================================================
# Validations
if [[ "$#" -lt 1 || "$#" -gt 2 ]]; then
    echo "Usage: opencli user-unsuspend <username> [--debug]"
    exit 1
fi

for arg in "$@"; do
    [[ "$arg" == "--debug" ]] && DEBUG=true
done


source "/usr/local/opencli/db.sh"


# ======================================================================
# Functions
get_docker_context() {
    local query="SELECT server FROM users WHERE username LIKE 'SUSPENDED\_%${USERNAME}';"
    local server_name
    server_name=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query" -N)
    CONTEXT_FLAG="--context $server_name"
}

reload_caddy_if_valid() {
    if docker --context default exec caddy caddy validate --config /etc/caddy/Caddyfile > /dev/null 2>&1; then
        docker --context default exec caddy caddy reload --config /etc/caddy/Caddyfile > /dev/null 2>&1
        return 0
    fi
    return 1
}

unsuspend_user_domains() {

    # 1. get list of all user domains
    local domain_list
    domain_list=$(opencli domains-user "$USERNAME")

    # 2. move all vhosts
    for domain_name in $domain_list; do
        # from /etc/openpanel/caddy/suspended_domains/ to /etc/openpanel/caddy/domains/
        [[ -f "${SUSPENDED_DIR}${domain_name}.conf" ]] && cp "${SUSPENDED_DIR}${domain_name}.conf" "${CADDY_VHOST_DIR}/${domain_name}.conf"
    done

    # 3. validate caddy
    reload_caddy_if_valid
}

start_user_containers() {
    local cores
    jobs=$(( $(nproc) * 2 ))
    $DEBUG && echo "Starting containers for user: $USERNAME"

    docker $CONTEXT_FLAG ps -a --format "{{.Names}}" |
        xargs -r -n1 -P "$jobs" bash -c '
            '"$DEBUG"' && echo "- Starting container: $0"
            docker '"$CONTEXT_FLAG"' start "$0" > /dev/null 2>&1
        '
}

rename_user_in_db() {
    local query="UPDATE users SET username='${USERNAME}' WHERE username LIKE 'SUSPENDED\\_%_${USERNAME}';"

    if mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query"; then
        echo "User '$USERNAME' unsuspended successfully."
    else
        echo "ERROR: Failed to unsuspend user '$USERNAME'."
        exit 1
    fi
}



# ======================================================================
# Main
get_docker_context
unsuspend_user_domains
start_user_containers
rename_user_in_db
