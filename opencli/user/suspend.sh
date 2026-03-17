#!/bin/bash
################################################################################
# Script Name: user/suspend.sh
# Description: Suspend user: stop all containers and suspend domains.
# Usage: opencli user-suspend <USERNAME>
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 16.03.2026
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
readonly SUSPENDED_DIR="/etc/openpanel/caddy/suspended_domains/"
readonly TEMPLATE_CONF="/etc/openpanel/caddy/templates/suspended_user.conf"
readonly CADDY_VHOST_DIR="/etc/openpanel/caddy/domains"



# ======================================================================
# Variables
DEBUG=false
USERNAME="$1"



# ======================================================================
# Helpers
confirm() {
    read -p "Are you sure you want to suspend OpenPanel user '$USERNAME'? (y/N) " answer
    case "$answer" in
        [yY]|[yY][eE][sS]) return 0 ;;
        *) echo "Operation cancelled."; exit 1 ;;
    esac
}



# ======================================================================
# Validations
if [[ "$#" -lt 1 || "$#" -gt 3 ]]; then
    echo "Usage: opencli user-suspend <username> [-y] [--debug]"
    exit 1
fi

for arg in "$@"; do
    [[ "$arg" == "--debug" ]] && DEBUG=true
done

if [ "$2" != "-y" ]; then
    confirm
fi


source "/usr/local/opencli/db.sh"


# ======================================================================
# Functions

get_docker_context() {
    local query="SELECT id, server FROM users WHERE username = '$USERNAME';"
    user_info=$(mysql -se "$query")
    
    user_id=$(echo "$user_info" | awk '{print $1}')
    context=$(echo "$user_info" | awk '{print $2}')
}

suspend_user_domains() {
    mkdir -p "$SUSPENDED_DIR"

    # 1. get list of all user domains
    local domain_list
    domain_list=$(opencli domains-user "$USERNAME" 2>/dev/null)

    if [[ -z "$domain_list" || "$domain_list" =~ ^(User|No\ domains\ found) ]]; then
        echo "No domains found for user '$USERNAME'."
        return 0
    fi

    domain_regex='\.'

    # 2. move all vhosts
    for domain_name in $domain_list; do
        if [[ ! $domain_name =~ $domain_regex ]]; then
            echo "Skipping invalid domain name: $domain_name"
            continue
        fi

        # from /etc/openpanel/caddy/domains/ to /etc/openpanel/caddy/suspended_domains/
        src="${CADDY_VHOST_DIR}/${domain_name}.conf"
        dest="${SUSPENDED_DIR}/${domain_name}.conf"
    
        if [[ -f "$src" ]]; then
            if [[ ! -f "$dest" ]]; then
                cp "$src" "$dest"
            fi
        fi

        # /etc/openpanel/caddy/templates/suspended_user.html
        sed "s|<DOMAIN_NAME>|$domain_name|g" "$TEMPLATE_CONF" > "${CADDY_VHOST_DIR}/${domain_name}.conf"
    done

    # 3. reload caddy
    nohup docker --context=default exec caddy sh -c "caddy validate && caddy reload" >/dev/null 2>&1 &
    disown   
}

stop_user_containers() {
    local cores
    jobs=$(( $(nproc) * 2 ))
    $DEBUG && echo "Stopping containers for user: $USERNAME"

    docker --context $context ps --format "{{.Names}}" |
        xargs -r -n1 -P "$jobs" -I{} docker --context $context stop {} > /dev/null 2>&1       
}

rename_user_in_db() {
    local new_username="SUSPENDED_$(date +'%Y%m%d%H%M%S')_${USERNAME}"
    local query="UPDATE users SET username='${new_username}' WHERE username='${USERNAME}';"

    [ -n "$user_id" ] && query+="DELETE FROM active_sessions WHERE user_id='$user_id'; "

    if mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query"; then
        echo "User '$USERNAME' suspended successfully."
    else
        echo "ERROR: Failed to suspend user '$USERNAME'."
        exit 1
    fi
    nohup opencli sentinel --action=user_status --title="User account suspended" --message="User account '$USERNAME' has been suspended." >/dev/null 2>&1 &
    disown
}


# ======================================================================
# Main
get_docker_context
suspend_user_domains
stop_user_containers
rename_user_in_db
