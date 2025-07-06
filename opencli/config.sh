#!/bin/bash
################################################################################
# Script Name: config.sh
# Description: View / change configuration for users and set defaults for new accounts.
# Usage: opencli config get <setting_name> 
#        opencli config update <setting_name> <new_value>
# Author: Stefan Pejcic
# Created: 01.11.2023
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


config_file="/etc/openpanel/openpanel/conf/openpanel.config"


########################## UPDATE NGINX PROXY FILE FOR DOMAINS ##########################
proxy_conf_file="/etc/openpanel/nginx/vhosts/openpanel_proxy.conf"

# Function to update SSL configuration in proxy_conf_file
update_ssl_config() {
    ssl_value="$1"

    if [ "$ssl_value" = "yes" ]; then
        # Update https to http in the proxy_conf_file if it's not already present
        if grep -q 'return 301[[:space:]]\+http://' "$proxy_conf_file"; then
            sed -i 's|return 301[[:space:]]\+http:|return 301 https:|' "$proxy_conf_file"
        else
            :
            #echo "SSL is already configured as 'https' in $proxy_conf_file"
        fi
    elif [ "$ssl_value" = "no" ]; then
        # Update http to https in the proxy_conf_file if it's not already present
        if grep -q 'return 301[[:space:]]\+https://' "$proxy_conf_file"; then
            sed -i 's|return 301[[:space:]]\+https:|return 301 http:|' "$proxy_conf_file"
        else
            :
            #echo "SSL is already configured as 'http' in $proxy_conf_file"
        fi
    fi

    #echo "Updated SSL configuration in $proxy_conf_file"
}


# Function to update port configuration in proxy_conf_file
update_port_config() {
    new_port="$1"
    sed -Ei "s|(return 301 https://[^:]+:)([0-9]+;)|\1$new_port;|;s|(return 301 http://[^:]+:)([0-9]+;)|\1$new_port;|" "$proxy_conf_file"
    #echo "Updated port configuration in $proxy_conf_file to $new_port"
}

# Function to update openpanel_proxy configuration in proxy_conf_file
update_openpanel_proxy_config() {
    new_value="$1"
    # Update the value in the 2nd line after "location /$$$$ {"
    sed -i "0,/location \/openpanel/{n;s|/[^[:space:]]*|/$new_value|}" "$proxy_conf_file"
    #echo "Updated openpanel_proxy configuration in $proxy_conf_file to $new_value"
}

##############################################################################



# Function to get the current configuration value for a parameter
get_config() {
    param_name="$1"
    param_value=$(grep "^$param_name=" "$config_file" | cut -d= -f2-)
    
    if [ -n "$param_value" ]; then
        echo "$param_value"
    elif grep -q "^$param_name=" "$config_file"; then
        echo "Parameter $param_name has no value."
    else
        echo "Parameter $param_name does not exist."
    fi
}

# Function to update a configuration value
update_config() {
    param_name="$1"
    new_value="$2"

    # Check if the parameter exists in the config file
    if grep -q "^$param_name=" "$config_file"; then
        # Update the parameter with the new value
        sed -i "s/^$param_name=.*/$param_name=$new_value/" "$config_file"
        echo "Updated $param_name to $new_value"

        # Restart the panel service for all settings except autoupdate and autopatch
        if [ "$param_name" != "autoupdate" ] && [ "$param_name" != "autopatch" ]; then
            docker restart openpanel &> /dev/null &                        # run in bg, and dont show error if panel not running
            rm -rf /etc/openpanel/openpanel/core/users/*/data.json         # remove data.json files for all users
        fi
        
    else
        echo "Parameter $param_name not found in the configuration file."
    fi
}

# Main script logic
if [ "$#" -lt 2 ]; then
    echo "Usage: opencli config [get|update] <parameter_name> [new_value]"
    exit 1
fi

command="$1"
param_name="$2"

case "$command" in
    get)
        get_config "$param_name"
        ;;
    update)
        if [ "$#" -ne 3 ]; then
            echo "Usage: $0 update <parameter_name> <new_value>"
            exit 1
        fi
        new_value="$3"
        update_config "$param_name" "$new_value"
        
        case "$param_name" in
            ssl)
                update_ssl_config "$new_value"
                ;;
            port)
                update_port_config "$new_value"
                ;;
            openpanel_proxy)
                update_openpanel_proxy_config "$new_value"
                docker restart nginx &> /dev/null &
                ;;
        esac
        ;;
    *)
        echo "Invalid command. Usage: $0 [get|update] <parameter_name> [new_value]"
        exit 1
        ;;
esac
