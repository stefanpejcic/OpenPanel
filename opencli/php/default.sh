#!/bin/bash
################################################################################
# Script Name: php/default.sh
# Description: View or change the default PHP version used for new domains added by user.
# Usage: opencli php-default <username>
#        opencli php-default <username> --update <new_php_version>
# Author: Stefan Pejcic
# Created: 07.10.2023
# Last Modified: 05.02.2026
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
# Checks
if [ $# -lt 1 ]; then
    echo "Usage: opencli php-default <username> [--update <new_php_version>]"
    exit 1
fi

username="$1"

# ======================================================================
# Validators

validate_env_file() {
    config_file="/home/$context/.env"
    if [ ! -e "$config_file" ]; then
        echo "Configuration file for user '$username' not found."
        exit 1
    fi
}

validate_php_version() {
    local php_version="$1"
    if [[ ! "$php_version" =~ ^[0-9]\.[0-9]$ ]]; then
        echo "Invalid PHP version format. Please use the format 'number.number' (e.g., 8.1 or 5.6)."
        exit 1
    fi
}

validate_user_exists() {
    result=$(get_user_info "$username")
    user_id=$(echo "$result" | cut -d',' -f1)
    context=$(echo "$result" | cut -d',' -f2)

    if [ -z "$user_id" ]; then
        echo "FATAL ERROR: user $username does not exist."
        exit 1
    fi
}

# ======================================================================
# Helpers

get_user_info() {
    local user="$1"
    local query="SELECT id, server FROM users WHERE username = '${user}';"
    user_info=$(mysql -se "$query")
    user_id=$(echo "$user_info" | awk '{print $1}')
    context=$(echo "$user_info" | awk '{print $2}')
    echo "$user_id,$context"
}

update_php_version() {
    sed -i "s/^DEFAULT_PHP_VERSION=.*/DEFAULT_PHP_VERSION=$1/" "$config_file"
    echo "Default PHP version for user '$username' updated to: $1"
}

read_php_version() {
    php_version=$(awk -F '=' '/DEFAULT_PHP_VERSION/ {print $2}' "$config_file" | tr -d '[:space:]')
    if [[ "$php_version" == php* ]]; then # legacy for <0.3.4
        php_version="${php_version#php}"
    fi
    
    if [ -n "$php_version" ]; then
        echo "Default PHP version for user '$username' is: $php_version"
    else
        echo "Default PHP version for user: '$username' not found in the configuration file."
        exit 1
    fi
}

read_or_update_version() {
    if [ "$1" == "--update" ]; then
        if [ -z "$2" ]; then
            echo "Usage: opencli php-default <username> --update <new_php_version>"
            exit 1
        fi
        validate_php_version "$2"
        update_php_version "$2"
    else
        read_php_version
    fi
}



# ======================================================================
# Main
validate_user_exists                        # get docker context
validate_env_file                           # check if .env exists
read_or_update_version "$2" "$3"            # get/update value
