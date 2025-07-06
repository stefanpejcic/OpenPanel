#!/bin/bash
################################################################################
# Script Name: websites/user.sh
# Description: Lists all websites and domains owned by a specific user.
# Usage: opencli websites-user <USERNAME>
# Author: Stefan Pejcic
# Created: 08.07.2024
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

# Function to fetch sites information for a given username
get_sites_info_for_user() {
    local username="$1"
    local output_format="$2"  # Capture the output format option
    local domains_output=""   # Initialize the output variable

    # Check if the config file exists
    if [ ! -f "$config_file" ]; then
        echo "Config file $config_file not found."
        exit 1
    fi
    
    # Query to fetch the user_id for the specified username
    username_query="SELECT id FROM users WHERE username = '$username'"
    
    # Execute the query and fetch the user_id
    user_id=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$username_query" -sN)
    if [ -z "$user_id" ]; then
        echo "User '$username' not found in the database."
    else
        # Query to fetch domains with domain_id, domain_url, php_version, and "php" type for domains without sites
        domains_info_query="SELECT d.domain_id, d.domain_url, d.php_version, 'php' as type FROM domains d WHERE d.user_id = '$user_id'"
        domains_info=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$domains_info_query" -s)
        
        if [ -z "$domains_info" ]; then
            echo "No domains found for user '$username'."
        else
            # Loop through each domain_info
            while read -r domain_id domain_url php_version type; do
                if [ "$output_format" = "json" ]; then
                    if [ -n "$domains_output" ]; then
                        domains_output="${domains_output},"
                    fi
                    
                    # Format domain_url if it contains '/'
                    formatted_domain_url=$(echo "$domain_url" | sed 's/\//\\\//g')
                    
                    # Append to domains_output
                    domains_output="${domains_output}\"${formatted_domain_url}\": \"${type}${php_version}\""
                    
                    # Query to fetch sites (websites) for the current domain_id
                    sites_query="SELECT site_name, type, version, ports FROM sites WHERE domain_id = '$domain_id'"
                    sites_info=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$sites_query" -s)

                    if [ -n "$sites_info" ]; then
                        # Append sites information
                        domains_output="${domains_output}, \"sites\": ["
                        first_site=true
                        while read -r site_name site_type site_version ports; do
                            if [ "$first_site" = false ]; then
                                domains_output="${domains_output},"
                            fi
                            domains_output="${domains_output}{\"site_name\": \"$site_name\", \"type\": \"$site_type\", \"version\": \"$site_version\", \"ports\": \"$ports\"}"
                            first_site=false
                        done <<< "$sites_info"
                        domains_output="${domains_output}]"
                    fi
                else
                    # Output in table format by default
                    echo "Domain: $domain_url"
                    echo "PHP version: ${type}${php_version}"
                    
                    # Fetch and output sites information in table format
                    sites_query="SELECT site_name, type, version, ports FROM sites WHERE domain_id = '$domain_id'"
                    sites_info=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$sites_query" -s)

                    if [ -n "$sites_info" ]; then
                        echo "  Sites:"
                        while read -r site_name site_type site_version ports; do
                            echo "    Site Name: $site_name"
                            echo "    Type: $site_type"
                            echo "    Version: $site_version"
                            echo "    Ports: $ports"
                            echo ""
                        done <<< "$sites_info"
                    fi
                    
                    echo ""  # Add a blank line between domains
                fi
            done <<< "$domains_info"
            
            if [ "$output_format" = "json" ]; then
                domains_output="{ $domains_output }"
                echo "$domains_output"
            fi
        fi
    fi
}

# Check for the username argument
if [ $# -lt 1 ]; then
    echo "Usage: $0 <username> [--json]"
    exit 1
fi

username="$1"
output_format="table"

# Check if --json flag is provided
if [ "$2" = "--json" ]; then
    output_format="json"
fi

get_sites_info_for_user "$username" "$output_format"
