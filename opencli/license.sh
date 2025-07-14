#!/bin/bash
################################################################################
# Script Name: license.sh
# Description: Manage OpenPanel Enterprise license.
# Usage: opencli license verify 
# Author: Stefan Pejcic
# Created: 01.11.2023
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

set -euo pipefail

# Configuration
readonly CONFIG_FILE_PATH='/etc/openpanel/openpanel/conf/openpanel.config'
readonly WHMCS_URL="https://my.openpanel.com/modules/servers/licensing/verify.php"
readonly IP_SCRIPT_PATH="/usr/local/admin/core/scripts/ip_servers.sh"

# Colors
readonly GREEN='\033[0;32m'
readonly RED='\033[0;31m'
readonly RESET='\033[0m'

# Global flags
JSON="no"
NO_RESTART="no"

# Load IP servers configuration
load_ip_servers() {
    if [[ -f "$IP_SCRIPT_PATH" ]]; then
        source "$IP_SCRIPT_PATH"
    else
        IP_SERVER_1=IP_SERVER_2=IP_SERVER_3="https://ip.openpanel.com"
    fi
}

# Parse command line flags
parse_flags() {
    if [[ " $* " =~ " --json " ]]; then
        JSON="yes"
    fi
    
    if [[ " $* " =~ " --no-restart " ]]; then
        NO_RESTART="yes"
    fi
}

# Display usage information
usage() {
    cat << EOF
Usage: opencli license [options]

Commands:
  key                                           View current license key.
  enterprise-XXXXXXXXXX                         Save the license key.
  verify                                        Verify the license key.
  info                                          Display information about the license owner and expiration.
  delete                                        Delete the license key and downgrade OpenPanel to Community edition.
EOF
    exit 1
}

# Read configuration from file
read_config() {
    awk -F '=' '
        /\[LICENSE\]/{flag=1; next} 
        /\[/{flag=0} 
        flag{
            gsub(/^[ \t]+|[ \t]+$/, "", $1); 
            gsub(/^[ \t]+|[ \t]+$/, "", $2); 
            print $1 "=" $2
        }
    ' "$CONFIG_FILE_PATH"
}

# Get license key from configuration
get_license_key() {
    local config license_key
    config=$(read_config)
    license_key=$(echo "$config" | grep -i 'key' | cut -d'=' -f2)

    if [[ -z "$license_key" ]]; then
        if [[ "$JSON" == "yes" ]]; then
            echo "No License Key"
        else
            echo -e "${RED}No License Key${RESET}"
        fi
    else
        if [[ "$JSON" == "yes" ]]; then
            echo "$license_key"
        else
            echo -e "${GREEN}$license_key${RESET}"
        fi
    fi
}

# Get public IP address
get_public_ip() {
    curl --silent --max-time 2 -4 "$IP_SERVER_1" || \
    wget --timeout=2 -qO- "$IP_SERVER_2" || \
    curl --silent --max-time 2 -4 "$IP_SERVER_3"
}

# Verify license with WHMCS
verify_license_api() {
    local license_key="$1"
    local ip_address check_token response
    
    ip_address=$(get_public_ip)
    check_token=$(openssl rand -hex 16)
    
    response=$(curl -sS -X POST -d "licensekey=$license_key&ip=$ip_address&check_token=$check_token" "$WHMCS_URL")
    echo "$response" | grep -oP '(?<=<status>).*?(?=</status>)'
}

# Extract XML field from response
extract_xml_field() {
    local response="$1"
    local field="$2"
    echo "$response" | grep -oP "(?<=<$field>).*?(?=</$field>)"
}

# Output message based on JSON flag
output_message() {
    local message="$1"
    local color="${2:-}"
    
    if [[ "$JSON" == "yes" ]]; then
        echo "$message"
    else
        echo -e "${color}${message}${RESET}"
    fi
}

# Restart services
restart_services() {
    if [[ "$NO_RESTART" == "yes" ]]; then
        echo "Please restart OpenAdmin service to enable new features."
        return 0
    fi
    
    service admin restart > /dev/null
    docker --context default ps -q -f name=openpanel | grep -q . && \
        docker --context default restart openpanel > /dev/null &
    echo "OpenPanel and OpenAdmin are restarted to apply Enterprise features."
}

# Enable emails module
enable_emails_module() {
    local enabled_modules new_modules
    enabled_modules=$(grep '^enabled_modules=' "$CONFIG_FILE_PATH" | cut -d'=' -f2)
    
    if ! echo "$enabled_modules" | grep -q 'emails'; then
        new_modules="${enabled_modules},emails"
        sed -i "s/^enabled_modules=.*/enabled_modules=${new_modules}/" "$CONFIG_FILE_PATH"
    fi
}

# Disable emails module
disable_emails_module() {
    local enabled_modules new_modules
    enabled_modules=$(grep '^enabled_modules=' "$CONFIG_FILE_PATH" | cut -d'=' -f2)
    
    if echo "$enabled_modules" | grep -q 'emails'; then
        new_modules=$(echo "$enabled_modules" | sed 's/,emails//g; s/emails,//g; s/^emails$//g')
        sed -i "s/^enabled_modules=.*/enabled_modules=${new_modules}/" "$CONFIG_FILE_PATH"
    fi
}

# Save license key to file
save_license_to_file() {
    local new_key="$1"
    
    if opencli config update key "$new_key" > /dev/null; then
        output_message "License key ${new_key} added." "$GREEN"
        enable_emails_module > /dev/null
        restart_services
    else
        output_message "License is valid, but failed to save the license key ${new_key}" "$RED"
    fi
}

# Verify and save license key
verify_and_save_license() {
    local license_key="$1"
    local license_status
    
    license_status=$(verify_license_api "$license_key")
    
    if [[ "$license_status" == "Active" ]]; then
        save_license_to_file "$license_key"
    else
        output_message "License is invalid" "$RED"
        exit 1
    fi
}

# Verify existing license
verify_existing_license() {
    local config license_key license_status
    
    config=$(read_config)
    license_key=$(echo "$config" | grep -i 'key' | cut -d'=' -f2)

    if [[ -z "$license_key" ]]; then
        output_message "No License Key. Please add the key first: opencli config update key XXXXXXXXXX" "$RED"
        exit 1
    fi

    license_status=$(verify_license_api "$license_key")
    
    if [[ "$license_status" == "Active" ]]; then
        output_message "License is valid" "$GREEN"
        restart_services
    else
        output_message "License is invalid" "$RED"
        exit 1
    fi
}

# Get and display license information
show_license_info() {
    local config license_key response license_status
    
    config=$(read_config)
    license_key=$(echo "$config" | grep -i 'key' | cut -d'=' -f2)

    if [[ -z "$license_key" ]]; then
        output_message "No License Key. Please add the key first: opencli config update key XXXXXXXXXX" "$RED"
        exit 1
    fi

    ip_address=$(get_public_ip)
    check_token=$(openssl rand -hex 16)
    
    response=$(curl -sS -X POST -d "licensekey=$license_key&ip=$ip_address&check_token=$check_token" "$WHMCS_URL")
    license_status=$(extract_xml_field "$response" "status")
    
    if [[ "$license_status" == "Active" ]]; then
        local registered_name company_name email product_name reg_date next_due_date billing_cycle valid_ip
        
        registered_name=$(extract_xml_field "$response" "registeredname")
        company_name=$(extract_xml_field "$response" "companyname")
        email=$(extract_xml_field "$response" "email")
        product_name=$(extract_xml_field "$response" "productname")
        reg_date=$(extract_xml_field "$response" "regdate")
        next_due_date=$(extract_xml_field "$response" "nextduedate")
        billing_cycle=$(extract_xml_field "$response" "billingcycle")
        valid_ip=$(extract_xml_field "$response" "validip")

        if [[ "$JSON" == "yes" ]]; then
            cat << EOF
{
    "Owner": "$registered_name",
    "Company Name": "$company_name",
    "Email": "$email",
    "License Type": "$product_name",
    "Registration Date": "$reg_date",
    "Next Due Date": "$next_due_date",
    "Billing Cycle": "$billing_cycle",
    "Valid IP": "$valid_ip"
}
EOF
        else
            cat << EOF
Owner: $registered_name
Company Name: $company_name
Email: $email
License Type: $product_name
Registration Date: $reg_date
Next Due Date: $next_due_date
Billing Cycle: $billing_cycle
Valid IP: $valid_ip
EOF
        fi
    else
        output_message "License is invalid" "$RED"
        exit 1
    fi
}

# Delete license and downgrade
delete_license() {
    opencli config update key "" > /dev/null
    disable_emails_module
    rm -rf /etc/openpanel/openpanel/core/users/*/data.json
    service admin restart
}

# Main function
main() {
    local command="${1:-}"
    
    if [[ $# -eq 0 ]]; then
        usage
    fi
    
    load_ip_servers
    parse_flags "$@"
    
    case "$command" in
        "key")
            license_key=$(get_license_key)
            echo "$license_key"
            ;;
        "info")
            show_license_info
            ;;
        "verify")
            verify_existing_license
            ;;
        "delete")
            delete_license
            ;;
        "enterprise"*)
            verify_and_save_license "$command"
            ;;
        *)
            output_message "Invalid command." "$RED"
            usage
            ;;
    esac
}

# Run main function with all arguments
main "$@"
