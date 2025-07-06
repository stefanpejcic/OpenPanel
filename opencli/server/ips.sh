#!/bin/bash
################################################################################
# Script Name: server/ips.sh
# Description: Generates a file that contians a list of users with dedicated IPs
# Usage: opencli server-ips
#        opencli server-ips <USERNAME>
# Author: Stefan Pejcic
# Created: 16.01.2024
# Last Modified: 04.07.2025
# Company: openpanel.co
# Copyright (c) openpanel.co
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

# IP SERVERS
SCRIPT_PATH="/usr/local/admin/core/scripts/ip_servers.sh"
if [ -f "$SCRIPT_PATH" ]; then
    source "$SCRIPT_PATH"
else
    IP_SERVER_1=IP_SERVER_2=IP_SERVER_3="https://ip.openpanel.com"
fi


if [ -z "$1" ]; then
    # If no username provided, get all active users
    usernames=$(opencli user-list --json | grep -v 'SUSPENDED' | awk -F'"' '/username/ {print $4}')
else
    # If username provided, process only for that user!
    usernames=("$1")
fi

current_server_main_ip=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3)

get_context_for_user() {
     USERNAME="$1"
     source /usr/local/opencli/db.sh
        username_query="SELECT server FROM users WHERE username = '$USERNAME'"
        context=$(mysql -D "$mysql_database" -e "$username_query" -sN)
        if [ -z "$context" ]; then
            context=$USERNAME
        fi
}

# Create or overwrite the JSON file for user
create_ip_file() {
    USERNAME=$1
    IP=$2
    JSON_FILE="/etc/openpanel/openpanel/core/users/$USERNAME/ip.json"
    echo "{ \"ip\": \"$IP\" }" > "$JSON_FILE"
}

for username in $usernames; do
    get_context_for_user $username
    user_ip=$(docker --context $context exec "$username" bash -c "curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3")
    echo $username - $user_ip
    if [[ "$user_ip" != "$current_server_main_ip" ]]; then
        create_ip_file "$username" "$user_ip"
    fi
done
