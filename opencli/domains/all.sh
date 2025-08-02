#!/bin/bash
################################################################################
# Script Name: domains/all.sh
# Description: Lists all domain names currently hosted on the server.
# Usage: opencli domains-all [--docroot|--php_version]
# Author: Stefan Pejcic
# Created: 26.10.2023
# Last Modified: 01.08.2025
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


# DB
source /usr/local/opencli/db.sh

get_all_domains() {
    local include_docroot=false
    local include_php_version=false

    # Parse optional flags
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

    # Check if the config file exists
    if [ ! -f "$config_file" ]; then
        echo "Config file $config_file not found."
        exit 1
    fi

    # Build the SELECT fields
    query_fields="domain_url"
    $include_docroot && query_fields+=", docroot"
    $include_php_version && query_fields+=", php_version"

    all_domains="SELECT $query_fields FROM domains"
    domains=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$all_domains" -sN)

    if [ -z "$domains" ]; then
        echo "No domains found in the database."
    else
        echo "$domains"
    fi
}

get_all_domains "$@"

