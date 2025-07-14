#!/bin/bash
################################################################################
# Script Name: php/default.sh
# Description: View or change the default PHP version used for new domains added by user.
# Usage: opencli php-default <username>
#        opencli php-default <username> --update <new_php_version>
# Author: Stefan Pejcic
# Created: 07.10.2023
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

# Check if username argument is provided
if [ $# -lt 1 ]; then
    echo "Usage: opencli php-default <username> [--update <new_php_version>]"
    exit 1
fi

username="$1"
config_file="/home/$username/.env"


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

# Function to update PHP version in the configuration file
update_php_version() {
    local new_php_version="$1"
    local config_file="$2"


    result=$(get_user_info "$username")
    user_id=$(echo "$result" | cut -d',' -f1)
    context=$(echo "$result" | cut -d',' -f2)

    if [ -z "$user_id" ]; then
        echo "FATAL ERROR: user $username does not exist."
        exit 1
    fi

    sed -i "s/^DEFAULT_PHP_VERSION=.*/DEFAULT_PHP_VERSION=$new_php_version/" "$config_file"
    echo "Default PHP version for user '$username' updated to: $new_php_version"

}

# Function to validate the PHP version format
validate_php_version() {
    local php_version="$1"
    if [[ ! "$php_version" =~ ^[0-9]\.[0-9]$ ]]; then
        echo "Invalid PHP version format. Please use the format 'number.number' (e.g., 8.1 or 5.6)."
        exit 1
    fi
}



# Check if the configuration file exists
if [ ! -e "$config_file" ]; then
    echo "Configuration file for user '$username' not found."
    exit 1
fi

if [ "$2" == "--update" ]; then
    # Check if a new PHP version is provided
    if [ -z "$3" ]; then
        echo "Usage: $0 <username> --update <new_php_version>"
        exit 1
    fi

    new_php_version="$3"
    validate_php_version "$new_php_version"
    update_php_version "$new_php_version" "$config_file"
else
    # Use awk to extract the PHP version from the YAML file
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
fi
