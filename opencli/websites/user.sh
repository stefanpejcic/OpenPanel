#!/bin/bash
################################################################################
# Script Name: websites/user.sh
# Description: Lists all websites and domains owned by a specific user.
# Usage: opencli websites-user <USERNAME> [--type=] [--domains=] [--json]
# Author: Stefan Pejcic
# Created: 08.07.2024
# Last Modified: 28.02.2026
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


source /usr/local/opencli/db.sh

# Display usage information with available site types (lowercased, space-separated)
usage() {
    local types
    types=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" \
            -sN -e "SELECT DISTINCT LOWER(type) FROM sites;" 2>/dev/null | xargs)

    echo "Usage: opencli websites-user <username> [--type=<type>] [--domains=<domain1,domain2,...>] [--json]"
    if [ -n "$types" ]; then
        echo "Available types: $types"
    else
        echo "Available types: (none found in database)"
    fi
    exit 1
}


# Execute MySQL query with error handling
execute_query() {
    local query="$1"
    if ! mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "$query" -sN 2>/dev/null; then
        echo "Error: Database query failed" >&2
        exit 1
    fi
}

# Get user ID from username
get_user_id() {
    local username="$1"
    local user_id
    
    user_id=$(execute_query "SELECT id FROM users WHERE username = '$username'")
    
    if [ -z "$user_id" ]; then
        echo "Error: User '$username' not found" >&2
        exit 1
    fi
    
    echo "$user_id"
}

# Get domains for user, optionally filtered by site type or domain
get_domains() {
    local user_id="$1"
    local type_filter="$2"
    local domains_filter="$3"
    local query="SELECT d.domain_id, d.domain_url, d.php_version
                 FROM domains d"

    if [ -n "$type_filter" ]; then
        type_filter="${type_filter,,}"
        query="$query
               INNER JOIN sites s ON s.domain_id = d.domain_id
               WHERE d.user_id = '$user_id' AND s.type = '$type_filter'"
    else
        query="$query WHERE d.user_id = '$user_id'"
    fi

    # Add domain filter
    if [ -n "$domains_filter" ]; then
        # Support comma-separated list
        IFS=',' read -ra DOMAIN_ARRAY <<< "$domains_filter"
        local domain_conditions=""
        for domain in "${DOMAIN_ARRAY[@]}"; do
            [ -n "$domain_conditions" ] && domain_conditions="$domain_conditions OR "
            domain_conditions="$domain_conditions d.domain_url = '$domain'"
        done
        query="$query AND ( $domain_conditions )"
    fi

    # Ensure we group by domain_id if type filtering
    if [ -n "$type_filter" ]; then
        query="$query GROUP BY d.domain_id"
    fi

    execute_query "$query"
}

format_domain_url() {
    local domain_url="$1"
    echo "$domain_url" | sed 's/\//\\\//g'
}

# Get all sites for a domain, optionally filtered by type
get_sites() {
    local domain_id="$1"
    local type_filter="$2"
    local query="SELECT site_name, type, version, ports, path, d.docroot
                 FROM sites s
                 INNER JOIN domains d ON s.domain_id = d.domain_id
                 WHERE s.domain_id = '$domain_id'"
    
    if [ -n "$type_filter" ]; then
        type_filter="${type_filter,,}"
        query="$query AND s.type = '$type_filter'"
    fi
    
    execute_query "$query"
}

# Output in table format
output_table() {
    local user_id="$1"
    local filter_type="$2"
    local domains_info sites_info
    
    domains_info=$(get_domains "$user_id" "$filter_type" "$filter_domains")
    
    if [ -z "$domains_info" ]; then
        echo "No domains found for user '$username'"
        return
    fi
    
    while read -r domain_id domain_url php_version; do
        echo "Domain: $domain_url"
        echo "PHP version: php$php_version"
        
        sites_info=$(get_sites "$domain_id" "$filter_type")
        if [ -n "$sites_info" ]; then
            echo "  Sites:"
            while read -r site_name site_type site_version ports path docroot; do
                [ -z "$path" ] && path="$docroot"
                echo "    Site Name: $site_name"
                echo "    Type: $site_type"
                echo "    Version: $site_version"
                echo "    Ports: $ports"
                echo "    Path: $path"
                echo
            done <<< "$sites_info"
        fi
        echo
    done <<< "$domains_info"
}

# Output in JSON format
output_json() {
    local user_id="$1"
    local filter_type="$2"
    local domains_info sites_info
    local json_output="" first_domain=true
    
    domains_info=$(get_domains "$user_id" "$filter_type" "$filter_domains")
    
    if [ -z "$domains_info" ]; then
        echo "{}"
        return
    fi
    
    while read -r domain_id domain_url php_version; do
        [ "$first_domain" = false ] && json_output="${json_output},"
        
        local formatted_url
        formatted_url=$(format_domain_url "$domain_url")
        json_output="${json_output}\"${formatted_url}\": {\"php\": \"php${php_version}\""
        
        sites_info=$(get_sites "$domain_id" "$filter_type")
        if [ -n "$sites_info" ]; then
            json_output="${json_output}, \"sites\": ["
            local first_site=true
            
            while read -r site_name site_type site_version ports path docroot; do
                [ -z "$path" ] && path="$docroot"
                [ "$first_site" = false ] && json_output="${json_output},"
                
                json_output="${json_output}{\"site_name\": \"$site_name\", \"type\": \"$site_type\", \"version\": \"$site_version\", \"ports\": \"$ports\", \"path\": \"$path\"}"
                first_site=false
            done <<< "$sites_info"
            
            json_output="${json_output}]"
        fi
        
        json_output="${json_output}}"
        first_domain=false
    done <<< "$domains_info"
    
    echo "{ $json_output }"
}


# Main function
main() {
    [ $# -lt 1 ] && usage

    local username="$1"
    local output_format="table"
    local filter_type=""
    local filter_domains=""

    # Parse additional flags
    shift
    while [ $# -gt 0 ]; do
        case "$1" in
            --json)
                output_format="json"
                ;;
            --type=*)
                filter_type="${1#*=}"
                ;;
            --domains=*)
                filter_domains="${1#*=}"
                ;;
            *)
                usage
                ;;
        esac
        shift
    done

    # Validate config file
    if [ ! -f "$config_file" ]; then
        echo "Error: Config file $config_file not found" >&2
        exit 1
    fi

    # Get user ID
    local user_id
    user_id=$(get_user_id "$username")

    # Output
    case "$output_format" in
        "json")
            output_json "$user_id" "$filter_type" "$filter_domains"
            ;;
        "table")
            output_table "$user_id" "$filter_type" "$filter_domains"
            ;;
    esac
}

# Run main function with all arguments
main "$@"
