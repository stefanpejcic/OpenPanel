#!/bin/bash
################################################################################
# Script Name: user/email.sh
# Description: Change email for user
# Usage: opencli user-email <USERNAME> <NEW_EMAIL>
# Docs: https://docs.openpanel.com
# Author: Radovan Jecmenica
# Created: 06.12.2023
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

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Source database configuration
readonly DB_CONFIG="/usr/local/opencli/db.sh"
if [[ ! -f "$DB_CONFIG" ]]; then
    echo "Error: Database configuration file not found: $DB_CONFIG" >&2
    exit 1
fi

# shellcheck source=/usr/local/opencli/db.sh
source "$DB_CONFIG"

# Validate email format
validate_email() {
    local email="$1"
    local email_regex="^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
    
    if [[ ! "$email" =~ $email_regex ]]; then
        echo "Error: Invalid email format: $email" >&2
        return 1
    fi
}

# Check if user exists in database
user_exists() {
    local username="$1"
    local count
    
    count=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" \
        -sN -e "SELECT COUNT(*) FROM users WHERE username = '$username';")
    
    [[ "$count" -eq 1 ]]
}

# Update email in database with proper error handling
update_user_email() {
    local username="$1"
    local new_email="$2"
    
    # Check if user exists
    if ! user_exists "$username"; then
        echo "Error: User '$username' not found in database" >&2
        return 1
    fi
    
    # Validate email format
    if ! validate_email "$new_email"; then
        return 1
    fi
    
    # Check if email is already in use
    local existing_user
    existing_user=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" \
        -sN -e "SELECT username FROM users WHERE email = '$new_email' AND username != '$username';")
    
    if [[ -n "$existing_user" ]]; then
        echo "Error: Email '$new_email' is already in use by user '$existing_user'" >&2
        return 1
    fi
    
    # Perform the update
    mysql --defaults-extra-file="$config_file" -D "$mysql_database" \
        -e "UPDATE users SET email = '$new_email' WHERE username = '$username';"
}

# Display usage information
show_usage() {
    echo "Usage: opencli user-email <USERNAME> <NEW_EMAIL>"
    echo ""
    echo "Updates the email address for a specified account."
    echo ""
    echo "Arguments:"
    echo "  USERNAME   - The username of the user to update"
    echo "  NEW_EMAIL  - The new email address to assign"
    echo ""
    echo "Example:"
    echo "  opencli user-email john john.doe@newdomain.com"
}

# Main execution
main() {
    # Check argument count
    if [[ $# -ne 2 ]]; then
        show_usage
        exit 1
    fi
    
    local username="$1"
    local new_email="$2"
    
    # Validate required database variables
    if [[ -z "${config_file:-}" ]] || [[ -z "${mysql_database:-}" ]]; then
        echo "Error: Database configuration variables not properly set" >&2
        exit 1
    fi
    
    # Perform the email update
    if update_user_email "$username" "$new_email"; then
        echo "Success: Email for user '$username' updated to '$new_email'"
    else
        echo "Error: Failed to update email for user '$username'" >&2
        exit 1
    fi
}

# Execute main function with all arguments
main "$@"
