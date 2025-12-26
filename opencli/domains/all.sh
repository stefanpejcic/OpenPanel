#!/bin/bash
################################################################################
# Script Name: domains/all.sh
# Description: Lists all domain names currently hosted on the server.
# Usage: opencli domains-all [--docroot|--php_version]
# Author: Stefan Pejcic
# Created: 26.10.2023
# Last Modified: 25.12.2025
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
    local output_json=false

    # Parse optional flags
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --docroot)
                include_docroot=true
                ;;
            --php_version)
                include_php_version=true
                ;;
            --json)
                output_json=true
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
        exit 0
    fi


    if $output_json; then
        # Always select domain_url, docroot, php_version, username (owner), and server
        query_fields="d.domain_url, d.docroot, d.php_version, u.username, u.server"
        query="
            SELECT $query_fields
            FROM domains d
            LEFT JOIN users u ON d.user_id = u.id
        "
    else
        query_fields="domain_url"
        $include_docroot && query_fields+=", docroot"
        $include_php_version && query_fields+=", php_version"
        query="SELECT $query_fields FROM domains"
    fi

    domains=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query" -sN)

    if [ -z "$domains" ]; then
        echo "No domains found in the database."
        exit 0
    fi

    if ! $output_json; then
        echo "$domains"
    else
        echo -n '{"data":{'

        first=true
        while IFS=$'\t' read -r domain docroot php_version owner server; do
            dot_count=$(grep -o "\." <<<"$domain" | wc -l)
            is_main=false

            [ -z "$docroot" ] && docroot=""
            [ -z "$php_version" ] && php_version=""
            [ -z "$owner" ] && owner="null" || owner=$(printf '%s' "$owner" | sed 's/"/\\"/g')
            [ -z "$server" ] && server=""

            # Replace /var/www/html/ prefix if present
            prefix="/var/www/html/"
            if [[ "$docroot" == "$prefix"* && -n "$server" ]]; then
                replacement="/home/$server/docker-data/volumes/${server}_html_data/_data/"
                docroot="${docroot/#$prefix/$replacement}"
            fi

            $first && first=false || echo -n ','

            esc_domain=$(echo "$domain" | sed 's/"/\\"/g')
            esc_docroot=$(echo "$docroot" | sed 's/"/\\"/g')
            php_version_id="$php_version"
            php_ver_num=$(echo "$php_version" | grep -oE '[0-9]+' | tr -d '\n\r ' || echo "")
            
            if [[ "$php_ver_num" =~ ^([0-9])([0-9])$ ]]; then
                php_version_formatted="${BASH_REMATCH[1]}.${BASH_REMATCH[2]}"
            else
                php_version_formatted="$php_ver_num"
            fi
            
            # Also clean ini_path in case
            ini_path=""
            if [ -n "$server" ] && [ -n "$php_version_formatted" ]; then
                ini_path="/home/$server/php.ini/${php_version_formatted}.ini"
                ini_path=$(echo "$ini_path" | tr -d '\n\r ')
            fi
            
            echo -n "\"$esc_domain\":{"
            echo -n "\"document_root\":\"$esc_docroot\","
            echo -n "\"php\":{"
            echo -n "\"php_version_id\":\"$php_version_id\","
            echo -n "\"version\":\"$php_version_formatted\","
            echo -n "\"ini_path\":\"$ini_path\","
            echo -n "\"is_native\":true,"
            echo -n "\"handler\":\"php-fpm\""
            echo -n "},"
            echo -n "\"is_main\":$is_main,"
            if [ "$owner" = "null" ]; then
                echo -n "\"owner\":null"
            else
                echo -n "\"owner\":\"$owner\""
            fi
            echo -n "}"

        done <<< "$domains"

        echo '}, "metadata": {"result": "ok"}}'
    fi
}

get_all_domains "$@"
