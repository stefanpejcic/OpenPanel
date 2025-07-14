#!/bin/bash
################################################################################
# Script Name: plans.sh
# Description: Display all plans: id, name, description, limits..
# Usage: opencli plan-list [--json]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 30.11.2023
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

# --- Usage function ---
print_usage() {
    echo "Usage: opencli plan-list [--json]"
    exit 1
}

# --- Command line argument handling ---
json_output=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --json)
            json_output=true
            shift
            ;;
        *)
            print_usage
            ;;
    esac
done

# --- Source DB config ---
source /usr/local/opencli/db.sh

# --- Ensure jq is installed ---
ensure_jq_installed() {
    if ! command -v jq &> /dev/null; then
        if command -v apt-get &> /dev/null; then
            sudo apt-get update -qq > /dev/null
            sudo apt-get install -y -qq jq > /dev/null
        elif command -v yum &> /dev/null; then
            sudo yum install -y -q jq > /dev/null
        elif command -v dnf &> /dev/null; then
            sudo dnf install -y -q jq > /dev/null
        else
            echo "Error: No compatible package manager found. Please install jq manually and try again."
            exit 1
        fi
        if ! command -v jq &> /dev/null; then
            echo "Error: jq installation failed. Please install jq manually and try again."
            exit 1
        fi
    fi
}

# --- Fetch and output plans ---
fetch_plans_json() {
    ensure_jq_installed
    local data
    data=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "SELECT * FROM plans;" | tail -n +2)
    if [ -z "$data" ]; then
        echo "No plans."
        return
    fi
    local json
    json=$(echo "$data" | jq -R '
        split("\n") | map(select(length > 0) | split("\t") | {
            id: .[0], name: .[1], description: .[2], email_limit: .[3], ftp_limit: .[4],
            domains_limit: .[5], websites_limit: .[6], disk_limit: .[7],
            inodes_limit: .[8], db_limit: .[9], cpu: .[10], ram: .[11],
            bandwidth: .[12], feature_set: .[13]
        })'
    )
    echo "Plans:"
    echo "$json"
}

fetch_plans_table() {
    local data
    data=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" --table -e "SELECT * FROM plans;")
    if [ -n "$data" ]; then
        echo "$data"
    else
        echo "No plans."
    fi
}

# --- Main ---
if $json_output; then
    fetch_plans_json
else
    fetch_plans_table
fi
