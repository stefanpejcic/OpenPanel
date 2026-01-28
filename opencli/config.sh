#!/bin/bash
################################################################################
# Script Name: config.sh
# Description: View / change configuration for users and set defaults for new accounts.
# Usage: opencli config get <setting_name> 
#        opencli config update <setting_name> <new_value>
# Author: Stefan Pejcic
# Created: 01.11.2023
# Last Modified: 27.01.2026
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

config_file="/etc/openpanel/openpanel/conf/openpanel.config"

if [ "$#" -lt 2 ]; then
    echo "Usage: opencli config [get|update] <parameter_name> [new_value]"
    exit 1
fi

command="$1"
param_name="$2"


# ======================================================================
# Helper functions

read_config() {
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

update_config() {
    param_name="$1"
    new_value="$2"

    if grep -q "^$param_name=" "$config_file"; then

        # update value, respect quotes
        sed -i "s|^\($param_name=\)\".*\"|\1\"$new_value\"|" "$config_file"
        sed -i "s|^\($param_name=\)[^\"].*|\1$new_value|" "$config_file"
        echo "Updated $param_name to $new_value"

        # restart openpanel container
        if [[ ! "$param_name" =~ ^(email|autoupdate|autopatch|key|default_php_version|max_cpu|max_ram)$ ]]; then
            docker --context=default restart openpanel &> /dev/null &
        fi
    else
        echo "ERROR: Parameter '$param_name' was not found in the configuration file."
        echo
        echo "Usage: opencli config update <parameter_name> <new_value>"
        echo "See the list of available parameters here: https://dev.openpanel.com/cli/config.html#Available-Options"
    fi
}





# ======================================================================
# Main
case "$command" in
    get)
        read_config "$param_name"
        ;;
    update)
        if [ "$#" -ne 3 ]; then
            echo "ERROR: New value is not provided for the parameter '$param_name'."
            echo
            echo "Usage: opencli config update <parameter_name> <new_value>"
            echo "See the list of available parameters with examples here: https://dev.openpanel.com/cli/config.html#Update"
            exit 1            
        fi
        new_value="$3"
        update_config "$param_name" "$new_value"
        
        case "$param_name" in
            email)
                # special case for 'email' - need to update CSF conf also 
                if [[ "$new_value" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
                    sed -i "s/LF_ALERT_TO = \"\"/LF_ALERT_TO = \"$new_value\"/" /etc/csf/csf.conf
                fi
                ;;
        esac
        ;;
    *)
        echo "ERROR: Invalid command, only 'get' and 'update' are allowed."
        echo
        echo "Usage: opencli config [get|update] <parameter_name> [new_value]"
        echo "See the list of available commands here: https://dev.openpanel.com/cli/config.html"
        exit 1
        ;;
esac
