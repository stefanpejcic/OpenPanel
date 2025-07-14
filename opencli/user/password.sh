#!/bin/bash
################################################################################
# Script Name: user/password.sh
# Description: Reset password for a user.
# Usage: opencli user-password <USERNAME> <NEW_PASSWORD | RANDOM> [--ssh]
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

# Function to generate a random password
generate_random_password() {
    tr -dc 'a-zA-Z0-9' < /dev/urandom | head -c 12
}

# Function to print usage
print_usage() {
    echo "Usage: $0 <username> <new_password | random> [--ssh]"
    exit 1
}

# Check if username and new password are provided as arguments
if [ $# -lt 2 ]; then
    print_usage
fi

# Parse command line options
username=$1
new_password=$2
ssh_flag=false
random_flag=false  # Flag to check if the new password is initially set as "random"
DEBUG=false  # Default value for DEBUG

# Parse optional flags to enable debug mode when needed!
for arg in "$@"; do
    case $arg in
        --debug)
            DEBUG=true
            ;;
        *)
            ;;
    esac
done

for arg in "$@"; do
    case $arg in
        --ssh)
            ssh_flag=true
            ;;
    esac
done

# DB
source /usr/local/opencli/db.sh

# Check if new password should be randomly generated
if [ "$new_password" == "random" ]; then
    new_password=$(generate_random_password)
    random_flag=true
fi



#Insert data into the database

# Hash password
hashed_password=$(/usr/local/admin/venv/bin/python3 -c "from werkzeug.security import generate_password_hash; print(generate_password_hash('$new_password'))")

# Insert hashed password into MySQL database
mysql_query="UPDATE users SET password='$hashed_password' WHERE username='$username';"

mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$mysql_query"

if [ $? -eq 0 ]; then
    delete_sessions_query="DELETE FROM active_sessions WHERE user_id=(SELECT id FROM users WHERE username='$username');"
    mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$delete_sessions_query"

    if [ $? -eq 0 ]; then
        :
    else
        echo "WARNING: Failed to terminate existing sessions for the user."
    fi
    
    if [ "$random_flag" = true ]; then
        echo "Successfully changed password for user $username, new generated password is: $new_password"
    else
        echo "Successfully changed password for user $username."
    fi
    
else
    echo "Error: Data insertion failed."
    exit 1
fi


# Check if --ssh flag is provided
if [ "$ssh_flag" = true ]; then

    
# get user ID from the database
get_user_info() {
    local user="$1"
    local query="SELECT id, server FROM users WHERE username = '${user}';"
    
    # Retrieve both id and context
    user_info=$(mysql -se "$query")
    
    # Extract user_id and context from the result
    user_id=$(echo "$user_info" | awk '{print $1}')
    context=$(echo "$user_info" | awk '{print $2}')
    
    echo "$user_id,$context"
}


result=$(get_user_info "$username")
user_id=$(echo "$result" | cut -d',' -f1)
context=$(echo "$result" | cut -d',' -f2)

#echo "User ID: $user_id"
#echo "Context: $context"

if [ -z "$context" ]; then
    echo "FATAL ERROR: Docker context for user $username does not exist or mysql is not running."
    exit 1
fi
    
    : '
    if [ "$DEBUG" = true ]; then
        # Change the user password in the Docker container
        echo "root:$new_password" | docker --context $context exec $context -i root chpasswd
        if [ "$random_flag" = true ]; then
            echo "SSH root user in container now also have password: $new_password"
        else
            echo "SSH user root password changed."
        fi
    else
        # Change the user password in the Docker container
        echo "root:$new_password" | docker --context $context exec $context -i root chpasswd
    fi
    '
    
fi
