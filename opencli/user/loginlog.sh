#!/bin/bash
################################################################################
# Script Name: user/loginlog.sh
# Description: View user's .lastlogin file with last 20 successfull logins.
# Usage: opencli user-loginlog <USERNAME> [--table|--text|--json]
# Author: Stefan Pejcic
# Created: 16.11.2023
# Last Modified: 12.02.2026
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

# ======================================================================
# Helpers

print_usage() {
    echo "Usage: opencli user-loginlog <username> [--json | --table]"
    exit 1
}

ensure_jq_installed() {
    if ! command -v jq &>/dev/null; then
        echo "Error: jq is required for --json output. Please install it."
        exit 1
    fi
}


# ======================================================================
# Flags
json_output=false
table_output=true
text_output=false


# ======================================================================
# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        --json)
            table_output=false
            json_output=true
            ;;
        --table)
            table_output=true
            ;;
        --text)
            table_output=false
            text_output=true
            ;;
        -*)
            print_usage
            ;;
        *)
            username="$1"
            ;;
    esac
    shift
done


# ======================================================================
# Validations

[ -z "$username" ] && print_usage

login_log_file="/etc/openpanel/openpanel/core/users/$username/.lastlogin"

if [ ! -f "$login_log_file" ]; then
    echo "Login log file not found for user: $username"
    exit 1
fi


# ======================================================================
# Main
if [ "$table_output" = true ]; then    # TABLE (default since 1.7.41)
    {
        echo -e "IP\tCountry\tTime"
        awk '
        {
            split($0, a, " - ");
            split(a[1], ip, ": ");
            split(a[2], country, ": ");
            split(a[3], time, ": ");
            printf "%s\t%s\t%s\n", ip[2], country[2], time[2]
        }' "$login_log_file"
    } | column -t -s $'\t'    
elif [ "$json_output" = true ]; then   # JSON
    ensure_jq_installed
    json_data=$(awk 'BEGIN {print "["} {if(NR>1)print ","; split($0, arr, " - "); split(arr[1], ipArr, ": "); split(arr[2], countryArr, ": "); split(arr[3], timeArr, ": "); print "{\n\t\"ip\": \""ipArr[2]"\",\n\t\"country\": \""countryArr[2]"\",\n\t\"time\": \""timeArr[2] "\"\n}"} END {print "\n]"}' "$login_log_file")
    echo -e "$json_data" | jq .
elif [ "$text_output" = true ]; then   # TEXT
    cat "$login_log_file"
    echo
fi
