#!/bin/bash
################################################################################
# Script Name: webserver/get_webserver_for_user.sh
# Description: View cached or check the installed webserver inside user container.
# Usage: opencli webserver-get_webserver_for_user <USERNAME>
#        opencli webserver-get_webserver_for_user <USERNAME> --update
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 23.04.2026
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

determine_web_server() {
    container_name=$(docker --context "$context" ps --filter "status=running" --format "{{.Names}}")
    echo "$container_name" | grep -Eo 'nginx|apache|openresty|openlitespeed|litespeed' || echo "unknown"
}


get_user_info() {
    local query="SELECT id, server FROM users WHERE username = '${username}';"   
    user_info=$(mysql -se "$query")
    
    user_id=$(echo "$user_info" | awk '{print $1}')
    context=$(echo "$user_info" | awk '{print $2}')
    
    if [ -z "$user_id" ]; then
        echo "FATAL ERROR: user ${username##*_} does not exist."
        exit 1
    fi
}


if [ $# -lt 1 ]; then
    echo "Usage: webserver-get_webserver_for_user <username> [--update]"
    exit 1
fi

username="$1"

get_user_info

USERNAME="${USERNAME##*_}"
config_file="/home/$context/.env"

if [ ! -f "$config_file" ]; then
    echo "Configuration file not found for user $USERNAME"
    exit 1
fi

if [ "$2" == "--update" ]; then
    current_web_server=$(determine_web_server)
    if [ "$current_web_server" == "unknown" ]; then
        echo "Unable to determine the running web server for user $USERNAME."
        exit 1
    fi
    sed -i "s/^WEB_SERVER=.*/WEB_SERVER=\"$current_web_server\"/" "$config_file"        
    echo "Web Server for user $USERNAME updated to: $current_web_server"
else
    web_server=$(grep "^WEB_SERVER=" "$config_file" | awk -F '=' '{print $2}' | tr -d '[:space:]' | sed 's/^"\(.*\)"$/\1/')
    if [ -n "$web_server" ]; then
        echo "Web Server for user $USERNAME: $web_server"
    else
        echo "Web Server not found for user $USERNAME"
    fi
fi
