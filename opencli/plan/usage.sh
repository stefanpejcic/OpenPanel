#!/bin/bash
################################################################################
# Script Name: usage.sh
# Description: Display all users that are currently using the plan.
# Usage: opencli plan-usage <PLAN_NAME> [--json]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 30.11.2023
# Last Modified: 13.07.2025
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

# --- Function to print usage instructions ---
print_usage() {
    echo "Usage: plan-usage <plan_name> [--json]"
    exit 1
}

# --- Initialize variables ---
json_output=false
plan_name=""

# --- Command-line argument processing ---
if [ "$#" -lt 1 ]; then
    print_usage
fi

plan_name=$1
shift

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

# --- Source database configuration ---
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

# --- Fetch user data based on the provided plan name ---
fetch_users_json() {
    ensure_jq_installed
    local data
    data=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "SELECT users.id, users.username, users.email, plans.name AS plan_name, users.registered_date FROM users INNER JOIN plans ON users.plan_id = plans.id WHERE plans.name = '$plan_name';" | tail -n +2)
    if [ -n "$data" ]; then
        local json
        json=$(echo "$data" | jq -R 'split("\n") | map(select(length > 0) | split("\t") | {
            id: .[0], username: .[1], email: .[2], plan_name: .[3], registered_date: .[4]
        })')
        echo "Users on plan '$plan_name':"
        echo "$json"
    else
        echo "No users on plan '$plan_name'."
    fi
}

fetch_users_table() {
    local data
    data=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" --table -e "SELECT users.id, users.username, users.email, plans.name AS plan_name, users.registered_date FROM users INNER JOIN plans ON users.plan_id = plans.id WHERE plans.name = '$plan_name';")
    if [ -n "$data" ]; then
        echo "$data"
    else
        echo "No users on plan '$plan_name'."
    fi
}

# --- Main ---
if $json_output; then
    fetch_users_json
else
    fetch_users_table
fi
