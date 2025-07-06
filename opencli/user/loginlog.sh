#!/bin/bash
################################################################################
# Script Name: user/loginlog.sh
# Description: View users .loginlog that shows last 20 successfull logins.
# Usage: opencli user-loginlog <USERNAME> [--json]
# Author: Stefan Pejcic
# Created: 16.11.2023
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

# Function to print usage
print_usage() {
    echo "Usage: $0 <username> [--json]"
    exit 1
}

# Check if username is provided
if [ -z "$1" ]; then
    print_usage
fi

# Parse command-line options
json_output=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --json)
            json_output=true
            shift
            ;;
        *)
            username=$1
            shift
            ;;
    esac
done



ensure_jq_installed() {
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        # Detect the package manager and install jq
        if command -v apt-get &> /dev/null; then
            sudo apt-get update > /dev/null 2>&1
            sudo apt-get install -y -qq jq > /dev/null 2>&1
        elif command -v yum &> /dev/null; then
            sudo yum install -y -q jq > /dev/null 2>&1
        elif command -v dnf &> /dev/null; then
            sudo dnf install -y -q jq > /dev/null 2>&1
        else
            echo "Error: No compatible package manager found. Please install jq manually and try again."
            exit 1
        fi

        # Check if installation was successful
        if ! command -v jq &> /dev/null; then
            echo "Error: jq installation failed. Please install jq manually and try again."
            exit 1
        fi
    fi
}


# Check if the login log file exists for the user
login_log_file="/etc/openpanel/openpanel/core/users/$username/.lastlogin"
if [ ! -f "$login_log_file" ]; then
    echo "Login log file not found for user: $username"
    exit 1
fi

# Display the content of the login log file
if [ "$json_output" = true ]; then
    ensure_jq_installed
    # Format data as JSON
    json_data=$(awk 'BEGIN {print "["} {if(NR>1)print ","; split($0, arr, " - "); split(arr[1], ipArr, ": "); split(arr[2], countryArr, ": "); split(arr[3], timeArr, ": "); print "{\n\t\"ip\": \""ipArr[2]"\",\n\t\"country\": \""countryArr[2]"\",\n\t\"time\": \""timeArr[2] "\"\n}"} END {print "\n]"}' "$login_log_file")

    echo -e "$json_data" | jq .
else
    # Display data as is
    cat "$login_log_file"
    echo ""
fi
