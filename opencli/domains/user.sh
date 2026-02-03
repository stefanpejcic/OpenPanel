#!/bin/bash
################################################################################
# Script Name: domains/user.sh
# Description: Lists all domain names currently owned by a specific user.
# Usage: opencli domains-user <USERNAME> [--docroot|--php_version]
# Author: Stefan Pejcic
# Created: 26.10.2023
# Last Modified: 02.02.2026
# Company: openpanel.comm
# Copyright (c) openpanel.comm
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
# MYSQL
source /usr/local/opencli/db.sh

get_domains() {
    local username="$1"
    shift
    local include_docroot=false
    local include_php_version=false

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --docroot)
                include_docroot=true
                ;;
            --php_version)
                include_php_version=true
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
        shift
    done

    if [ ! -f "$config_file" ]; then
        echo "Config file $config_file not found."
        exit 1
    fi

    username_query="SELECT id FROM users WHERE username = '$username'"
    user_id=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$username_query" -sN)

    if [ -z "$user_id" ]; then
        echo "User '$username' not found in the database."
    else
        query_fields="domain_url"
        $include_docroot && query_fields+=", docroot"
        $include_php_version && query_fields+=", php_version"

        domains_query="SELECT $query_fields FROM domains WHERE user_id = '$user_id'"
        domains=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$domains_query" -sN)

        if [ -z "$domains" ]; then
            echo "No domains found for user '$username'."
        else
            echo "$domains"
        fi
    fi
}

# ======================================================================
# USAGE
if [ $# -lt 1 ]; then
    echo "Usage: $0 <username> [--docroot] [--php_version]"
    exit 1
fi

# ======================================================================
# MAIN
get_domains "$@"
