#!/bin/bash
################################################################################
# Script Name: firewall/reset.sh
# Description: Deletes all docker related ports from CSF and opens exposed ports.
#              Use: opencli firewall-reset
# Author: Stefan Pejcic
# Created: 01.11.2023
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





# Check for CSF
if command -v csf >/dev/null 2>&1; then
    echo "Checking ConfigServer Firewall configuration.."
    echo ""
    FIREWALL="CSF"
else
    echo "Error: CSF is not installed, all user ports will be exposed to the internet, without any protection."
    exit 1
fi




function open_port_csf() {
    local port=$1
    local csf_conf="/etc/csf/csf.conf"
    
    # Check if port is already open
    port_opened=$(grep "TCP_IN = .*${port}" "$csf_conf")
    if [ -z "$port_opened" ]; then
        # Open port
        sed -i "s/TCP_IN = \"\(.*\)\"/TCP_IN = \"\1,${port}\"/" "$csf_conf"
        echo "Port ${port} opened in CSF."
        ports_opened=1
    else
        echo "Port ${port} is already open in CSF."
    fi
}



# Function to extract port number from a file
function extract_port_from_file() {
    local file_path=$1
    local pattern=$2
    local port=$(grep -Po "(?<=${pattern}[ =])\d+" "$file_path")
    echo "$port"
}


function open_out_port_csf() {
        port="3306"
        local csf_conf="/etc/csf/csf.conf"
        # Check if port is already open
        port_opened=$(grep "TCP_OUT = .*${port}" "$csf_conf")
        if [ -z "$port_opened" ]; then
            # Open port
            sed -i "s/TCP_OUT = \"\(.*\)\"/TCP_OUT = \"\1,${port}\"/" "$csf_conf"
            ports_opened=1
        fi
}



















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



# Variable to track whether any ports were opened
ports_opened=0
echo "Opening ports:"
echo ""

# Check and open ports
if [ "$FIREWALL" = "CSF" ]; then

    # TODO: reuse install_csf() from https://github.com/stefanpejcic/OpenPanel/edit/main/version/0.2.3/INSTALL.sh
    
    open_port_csf 53 #dns
    open_port_csf 80 #http
    open_port_csf 443 #https
    
    ######for emails we wil add:
    # open_port_csf 25
    # open_port_csf 587
    # open_port_csf 465
    # open_port_csf 993

    open_port_csf 21 #ftp
    open_port_csf 21000-21010 #passive ftp

    
    open_port_csf $(extract_port_from_file "/etc/openpanel/openpanel/conf/openpanel.config" "port") #openpanel
    open_port_csf 2087
    open_port_csf $(extract_port_from_file "/etc/ssh/sshd_config" "Port") #ssh
    open_port_csf 32768:60999 #docker
    
    open_out_port_csf #mysql out
        
fi

if [ $ports_opened -eq 1 ]; then
    echo "Restarting $FIREWALL"
    if [ "$FIREWALL" = "CSF" ]; then
        csf -r
    fi
fi

echo ""
