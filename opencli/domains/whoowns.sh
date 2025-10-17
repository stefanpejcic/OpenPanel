#!/bin/bash
################################################################################
# Script Name: domains/whoowns.sh
# Description: Check which username owns a certain domain name.
# Usage: opencli domains-whoowns <DOMAIN-NAME> [--context] [--docroot]
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 16.10.2025
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

# DB
source /usr/local/opencli/db.sh

# Function to fetch the owner username of a domain
get_domain_owner() {
  
    if [ ! -f "$config_file" ]; then
        echo "Config file $config_file not found."
        exit 1
    fi
    
    user_id_query="SELECT user_id, docroot FROM domains WHERE domain_url = '$domain'"
    read user_id docroot <<< $(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$user_id_query" -sN)

    if [ -z "$user_id" ]; then
        echo "Domain '$domain' not found in the database."
    else
        if $context_flag; then
            query="SELECT username, server FROM users WHERE id = '$user_id'"
            user_info=$(mysql -se "$query")
            # Extract user_id and context from the result
            username=$(echo "$user_info" | awk '{print $1}')
            context=$(echo "$user_info" | awk '{print $2}')
            
            if [ -z "$context" ]; then
                echo "Error, no context '$user_id'."
            else
                if $docroot_flag; then
                    echo "$username $context $docroot"
                else
                    echo "$username $context"
                fi
            fi
        else
            query="SELECT username FROM users WHERE id = '$user_id'"
            username=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query" -sN)
            
            if [ -z "$username" ]; then
                echo "User does not exist with that ID '$user_id'."
            else
                if $docroot_flag; then
                    echo "Owner of '$domain': $username | docroot: $docroot"
                else
                    echo "Owner of '$domain': $username"
                fi
            fi
           
        fi
    fi
}


if [ $# -lt 1 ] || [ $# -gt 3 ]; then
    echo "Usage: opencli domains-whoowns <domain_name> [--docroot] [--context]"
    exit 1
fi

docroot_flag=false
context_flag=false
domain="$1"

shift
while [ $# -gt 0 ]; do
    case "$1" in
        --docroot)
            docroot_flag=true
            ;;
        --context)
            context_flag=true
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: opencli domains-whoowns <domain_name> [--docroot] [--context]"
            exit 1
            ;;
    esac
    shift
done

get_domain_owner
