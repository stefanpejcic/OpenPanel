#!/bin/bash
################################################################################
# Script Name: websites/user.sh
# Description: Lists all websites and domains owned by a specific user.
# Usage: opencli websites-user <USERNAME> [--json]
# Author: Stefan Pejcic
# Created: 08.07.2024
# Last Modified: 13.07.2025
# Company: openpanel.com
################################################################################

# Load database configuration
source /usr/local/opencli/db.sh

# Display usage information
usage() {
    echo "Usage: opencli websites-user <username> [--json]"
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

# Get domains for user
get_domains() {
    local user_id="$1"
    execute_query "SELECT domain_id, domain_url, php_version FROM domains WHERE user_id = '$user_id'"
}

# Get sites for domain
get_sites() {
    local domain_id="$1"
    execute_query "SELECT site_name, type, version, ports FROM sites WHERE domain_id = '$domain_id'"
}

# Format domain URL for JSON output
format_domain_url() {
    local domain_url="$1"
    echo "$domain_url" | sed 's/\//\\\//g'
}

# Output in table format
output_table() {
    local user_id="$1"
    local domains_info sites_info
    
    domains_info=$(get_domains "$user_id")
    
    if [ -z "$domains_info" ]; then
        echo "No domains found for user '$username'"
        return
    fi
    
    while read -r domain_id domain_url php_version; do
        echo "Domain: $domain_url"
        echo "PHP version: php$php_version"
        
        sites_info=$(get_sites "$domain_id")
        if [ -n "$sites_info" ]; then
            echo "  Sites:"
            while read -r site_name site_type site_version ports; do
                echo "    Site Name: $site_name"
                echo "    Type: $site_type"
                echo "    Version: $site_version"
                echo "    Ports: $ports"
                echo
            done <<< "$sites_info"
        fi
        echo
    done <<< "$domains_info"
}

# Output in JSON format
output_json() {
    local user_id="$1"
    local domains_info sites_info
    local json_output="" first_domain=true
    
    domains_info=$(get_domains "$user_id")
    
    if [ -z "$domains_info" ]; then
        echo "{}"
        return
    fi
    
    while read -r domain_id domain_url php_version; do
        if [ "$first_domain" = false ]; then
            json_output="${json_output},"
        fi
        
        local formatted_url
        formatted_url=$(format_domain_url "$domain_url")
        json_output="${json_output}\"${formatted_url}\": \"php${php_version}\""
        
        sites_info=$(get_sites "$domain_id")
        if [ -n "$sites_info" ]; then
            json_output="${json_output}, \"sites\": ["
            local first_site=true
            
            while read -r site_name site_type site_version ports; do
                if [ "$first_site" = false ]; then
                    json_output="${json_output},"
                fi
                json_output="${json_output}{\"site_name\": \"$site_name\", \"type\": \"$site_type\", \"version\": \"$site_version\", \"ports\": \"$ports\"}"
                first_site=false
            done <<< "$sites_info"
            
            json_output="${json_output}]"
        fi
        
        first_domain=false
    done <<< "$domains_info"
    
    echo "{ $json_output }"
}

# Main function
main() {
    # Validate arguments
    [ $# -lt 1 ] && usage
    
    local username="$1"
    local output_format="table"
    
    # Check for JSON flag
    [ "$2" = "--json" ] && output_format="json"
    
    # Validate config file
    if [ ! -f "$config_file" ]; then
        echo "Error: Config file $config_file not found" >&2
        exit 1
    fi
    
    # Get user ID and process output
    local user_id
    user_id=$(get_user_id "$username")
    
    case "$output_format" in
        "json")
            output_json "$user_id"
            ;;
        "table")
            output_table "$user_id"
            ;;
    esac
}

# Run main function with all arguments
main "$@"
