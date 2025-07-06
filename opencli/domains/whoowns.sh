#!/bin/bash
################################################################################
# Script Name: domains/whoowns.sh
# Description: Check which username owns a certain domain name.
# Usage: opencli domains-whoowns <DOMAIN-NAME> [--context]
# Author: Stefan Pejcic
# Created: 01.10.2023
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

# DB
source /usr/local/opencli/db.sh

# Function to fetch the owner username of a domain
get_domain_owner() {
    local domain="$1"
    
    # Check if the config file exists
    if [ ! -f "$config_file" ]; then
        echo "Config file $config_file not found."
        exit 1
    fi
    
    # Query to fetch the user_id for the specified domain
    user_id_query="SELECT user_id FROM domains WHERE domain_url = '$domain'"
    
    # Execute the query and fetch the user_id
    user_id=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$user_id_query" -sN)

    if [ -z "$user_id" ]; then
        echo "Domain '$domain' not found in the database."
    else
        if [ "$context" = "--context" ]; then
            query="SELECT username, server FROM users WHERE id = '$user_id'"
            user_info=$(mysql -se "$query")
            # Extract user_id and context from the result
            username=$(echo "$user_info" | awk '{print $1}')
            context=$(echo "$user_info" | awk '{print $2}')
            
            if [ -z "$context" ]; then
                echo "Error, no context '$user_id'."
            else
                echo "$username $context"
            fi
                    
        else
            query="SELECT username FROM users WHERE id = '$user_id'"
            username=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query" -sN)
            
            if [ -z "$username" ]; then
                echo "User does not exist with that ID '$user_id'."
            else
                echo "Owner of '$domain': $username"
            fi
           
        fi
    fi
}

# Check for the domain argument
if [ $# -lt 1 ] || [ $# -gt 2 ]; then
    echo "Usage: opencli domains-whoowns <domain_name> [--context]"
    exit 1
fi

# Get the domain name from the command line argument
domain_name="$1"
context="$2"

# Call the function to fetch the owner of the domain
get_domain_owner "$domain_name" "$context"
