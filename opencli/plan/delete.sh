#!/bin/bash
################################################################################
# Script Name: plan/delete
# Description: Delete hosting plan
# Usage: opencli plan-delete <PLAN_NAME>
# Docs: https://docs.openpanel.com
# Author: Radovan Jecmenica
# Created: 01.12.2023
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

# Function to print usage instructions
print_usage() {
    script_name=$(basename "$0")
    echo "Usage: $script_name <plan_name> [--json]"
    exit 1
}

# Initialize variables
plan_name=""
output_json=0

# Command-line argument processing
if [ "$#" -lt 1 ]; then
    print_usage
fi


for arg in "$@"
do
    case $arg in
        --json)
        output_json=1
        shift # Remove --json from processing
        ;;
        *)
        if [ -z "$plan_name" ]; then
            plan_name=$arg
        else
            echo "Invalid argument: $arg"
            print_usage
        fi
        ;;
    esac
done




# Source database configuration
source /usr/local/opencli/db.sh

# Check if there are users on the plan
users_count=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "SELECT COUNT(*) FROM users INNER JOIN plans ON users.plan_id = plans.id WHERE plans.name = '$plan_name';" | tail -n +2)

if [ "$users_count" -gt 0 ]; then
    if [ "$output_json" -eq 1 ]; then
        # JSON output
        users_data=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "SELECT SUBSTRING_INDEX(username, '_', -1) as username FROM users INNER JOIN plans ON users.plan_id = plans.id WHERE plans.name = '$plan_name' AND username NOT LIKE 'SUSPENDED\_%';" -B -s)
        if [ -n "$users_data" ]; then
            echo "{\"error\": \"Cannot delete plan '$plan_name' as there are users assigned to it.\", \"users\": [$users_data]}"
        else
            echo "{\"error\": \"Cannot delete plan '$plan_name' as there are users assigned to it.\", \"users\": []}"
        fi
    else
        # Regular output
        echo "Cannot delete plan '$plan_name' as there are users assigned to it. List of users:"
        users_data=$(mysql --defaults-extra-file="$config_file" -D "$mysql_database" --table -e "SELECT SUBSTRING_INDEX(username, '_', -1) as username FROM users INNER JOIN plans ON users.plan_id = plans.id WHERE plans.name = '$plan_name' AND username NOT LIKE 'SUSPENDED\_%';")
        if [ -n "$users_data" ]; then
            echo "$users_data"
        else
            echo "No users on plan '$plan_name'."
        fi
    fi

    exit 1
else
    if [ "$output_json" -eq 1 ]; then
        # Delete the plan data
        mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "DELETE FROM plans WHERE name = '$plan_name';"
        # Delete the Docker network
        docker network rm "$plan_name" > /dev/null 2>&1
        echo "{\"message\": \"Plan '$plan_name' and Docker network '$plan_name' deleted successfully.\"}"
    else
        # Delete the plan data
        mysql --defaults-extra-file="$config_file" -D "$mysql_database" -e "DELETE FROM plans WHERE name = '$plan_name';"
        # Delete the Docker network
        docker network rm "$plan_name" > /dev/null 2>&1
        echo "Docker network '$plan_name' deleted successfully."
        echo "Plan '$plan_name' deleted successfully."
    fi
fi

