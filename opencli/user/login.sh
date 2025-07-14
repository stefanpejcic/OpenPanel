#!/bin/bash
################################################################################
# Script Name: user/login.sh
# Description: Login as a user container.
# Usage: opencli user-login <USERNAME>
# Author: Stefan Pejcic
# Created: 21.10.2023
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

install_fzf() {
    if ! command -v fzf &> /dev/null; then
        echo "Attempting to install fzf..."
        apt install -y fzf > /dev/null 2>&1 || dnf install -y fzf
        if ! command -v fzf &> /dev/null; then
            echo "Failed to install fzf. Please install it manually."
            exit 1
        fi
    fi   
}


get_all_users(){
    users=$(mysql -Bse "SELECT username FROM users")
    if [ -z "$users" ]; then
      echo "No users found in the database."
      exit 1
    fi
}


if [ $# -gt 0 ]; then
    selected_user="$1"
    if [ -z "$selected_user" ]; then
        echo "Invalid user."
        exit 1
    fi
else
    install_fzf
    get_all_users
    selected_user=$(echo "$users" | fzf --prompt="Select a user: ")
    if [[ -z "$selected_user" || ! "$users" =~ (^|[[:space:]])"$selected_user"($|[[:space:]]) ]]; then
        echo "Invalid selection or no user selected."
        exit 1
    fi
fi





# get user ID from the database
get_user_info() {
    local user="$1"
    local query="SELECT id, server FROM users WHERE username = '${user}';"
    
    # Retrieve both id and context
    user_info=$(mysql -se "$query")
    
    # Extract user_id and context from the result
    user_id=$(echo "$user_info" | awk '{print $1}')
    context=$(echo "$user_info" | awk '{print $2}')
    
    echo "$user_id,$context"
}

result=$(get_user_info "$selected_user")
user_id=$(echo "$result" | cut -d',' -f1)
context=$(echo "$result" | cut -d',' -f2)



if [ -z "$user_id" ]; then
    echo "FATAL ERROR: user $selected_user does not exist."
    exit 1
else
    if id "$selected_user" &>/dev/null; then     
       USER_UID=$(id -u)
        exec docker run --rm -it \
            --cpus="0.1" \
            --memory="100m" \
            --pids-limit="10" \
            --security-opt no-new-privileges \
            -v /hostfs/run/user/$USER_UID/docker.sock:/var/run/docker.sock \
            openpanel/lazydocker
    else
        echo "Neither container nor the user $selected_user exist on the server."
        exit 1
    fi
fi

