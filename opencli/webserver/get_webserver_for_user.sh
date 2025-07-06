#!/bin/bash
################################################################################
# Script Name: webserver/get_webserver_for_user.sh
# Description: View cached or check the installed webserver inside user container.
# Usage: opencli webserver-get_webserver_for_user <USERNAME>
#        opencli webserver-get_webserver_for_user <USERNAME> --update
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 04.07.2025
# Company: openpanel.com
# Copyright (c) openpanel.com
# 
# THE SOFTWARE.
################################################################################

# Function to determine the current web server inside the user's container
determine_web_server() {

    result=$(get_user_info "$username")
    user_id=$(echo "$result" | cut -d',' -f1)
    context=$(echo "$result" | cut -d',' -f2)

    if [ -z "$user_id" ]; then
        echo "FATAL ERROR: user $username does not exist."
        exit 1
    fi


    
    # Check which container is running nginx, apache, or litespeed
    container_name=$(docker --context "$context" ps --filter "status=running" --format "{{.Names}}")
    if [[ "$container_name" == *"nginx"* ]]; then
        echo "nginx"
    elif [[ "$container_name" == *"apache"* ]]; then
        echo "apache"
    elif [[ "$container_name" == *"openresty"* ]]; then
        echo "openresty"
    elif [[ "$container_name" == *"litespeed"* ]]; then
        echo "litespeed"
    else
        echo "unknown"
    fi


}


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


# Check if the username is provided as a command-line argument
if [ $# -lt 1 ]; then
    echo "Usage: $0 <username> [--update]"
    exit 1
fi

# Get the username from the command-line argument
username="$1"

# Construct the path to the configuration file
config_file="/home/$username/.env"

# Check if the --update flag is provided
if [ "$2" == "--update" ]; then

    # Determine the current web server
    current_web_server=$(determine_web_server)

    if [ "$current_web_server" == "unknown" ]; then
        echo "Unable to determine the web server in the container named $username."
        exit 1
    fi

    # Check if the file exists
    if [ -f "$config_file" ]; then
        # Update the web_server value in the configuration file
        sed -i "s/^WEB_SERVER=.*/WEB_SERVER=\"$current_web_server\"/" "$config_file"        
        echo "Web Server for user $username updated to: $current_web_server"
    else
        echo "Configuration file not found for user $username"
    fi
else
    # Check if the file exists
    if [ -f "$config_file" ]; then
        # Use grep and awk to extract the value of web_server
        web_server=$(grep "^WEB_SERVER=" "$config_file" | awk -F '=' '{print $2}' | tr -d '[:space:]' | sed 's/^"\(.*\)"$/\1/')

        # Check if web_server is not empty
        if [ -n "$web_server" ]; then
            echo "Web Server for user $username: $web_server"
        else
            echo "Web Server not found for user $username"
        fi
    else
        echo "Configuration file not found for user $username"
    fi
fi
