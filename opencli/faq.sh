#!/bin/bash
################################################################################
# Script Name: faq.sh
# Description: Display answers to most frequently asked questions.
# Usage: opencli faq
# Author: Stefan Pejcic
# Created: 20.05.2024
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

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
PURPLE='\033[0;35m'
NC='\033[0m' #reset
CONFIG_FILE_PATH='/etc/openpanel/openpanel/conf/openpanel.config'
service_name="admin"

# IP SERVERS
SCRIPT_PATH="/usr/local/admin/core/scripts/ip_servers.sh"
if [ -f "$SCRIPT_PATH" ]; then
    source "$SCRIPT_PATH"
else
    IP_SERVER_1=IP_SERVER_2=IP_SERVER_3="https://ip.openpanel.com"
fi


read_config() {
    config=$(awk -F '=' '/\[DEFAULT\]/{flag=1; next} /\[/{flag=0} flag{gsub(/^[ \t]+|[ \t]+$/, "", $1); gsub(/^[ \t]+|[ \t]+$/, "", $2); print $1 "=" $2}' $CONFIG_FILE_PATH)
    echo "$config"
}

get_ssl_status() {
    config=$(read_config)
    ssl_status=$(echo "$config" | grep -i 'ssl' | cut -d'=' -f2)
    [[ "$ssl_status" == "yes" ]] && echo true || echo false
}

get_force_domain() {
    config=$(read_config)
    force_domain=$(echo "$config" | grep -i 'force_domain' | cut -d'=' -f2)

    if [ -z "$force_domain" ]; then
        ip=$(get_public_ip)
        force_domain="$ip"
    fi
    echo "$force_domain"
}

get_public_ip() {
    ip=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3)
    
    # Check if IP is empty or not a valid IPv4
    if [ -z "$ip" ] || ! [[ "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        ip=$(hostname -I | awk '{print $1}')
    fi
    echo "$ip"
}

if [ "$(get_ssl_status)" == true ]; then
        hostname=$(get_force_domain)
        admin_url="https://${hostname}:2087/"
else
        ip=$(get_public_ip)
        admin_url="http://${ip}:2087/"
fi

echo -e "
Frequently Asked Questions

${PURPLE}1.${NC} What is the login link for admin panel?

LINK: ${GREEN}${admin_url}${NC}
${BLUE}------------------------------------------------------------${NC}
${PURPLE}2.${NC} How to restart OpenAdmin or OpenPanel services?

- OpenPanel: ${RED}docker restart openpanel${NC}
- OpenAdmin: ${RED}service admin restart${NC}
${BLUE}------------------------------------------------------------${NC}
${PURPLE}3.${NC} How to reset admin password?

execute command ${GREEN}opencli admin password USERNAME NEW_PASSWORD${NC}
${BLUE}------------------------------------------------------------${NC}
${PURPLE}4.${NC} How to create new admin account ?

execute command ${GREEN}opencli admin new USERNAME PASSWORD${NC}
${BLUE}------------------------------------------------------------${NC}
${PURPLE}5.${NC} How to list admin accounts ?

execute command ${GREEN}opencli admin list${NC}
${BLUE}------------------------------------------------------------${NC}
${PURPLE}6.${NC} How to check OpenPanel version ?

execute command ${GREEN}opencli --version${NC}
${BLUE}------------------------------------------------------------${NC}
${PURPLE}7.${NC} How to update OpenPanel ?

execute command ${GREEN}opencli update --force${NC}
${BLUE}------------------------------------------------------------${NC}
${PURPLE}8.${NC} How to disable automatic updates?

execute command ${GREEN}opencli config update autoupdate off${NC}
${BLUE}------------------------------------------------------------${NC}
${PURPLE}9.${NC} Where are the logs?

- User panel errors: ${GREEN}/var/log/openpanel/user/error.log${NC}
- User panel access log: ${GREEN}/var/log/openpanel/user/access.log${NC}
- Admin panel errors: ${GREEN}/var/log/openpanel/admin/error.log${NC}
- Admin panel access log: ${GREEN}/var/log/openpanel/admin/access.log${NC}
- Admin panel API access: ${GREEN}/var/log/openpanel/admin/api.log${NC}
- Admin panel logins: ${GREEN}/var/log/openpanel/admin/api.log${NC}
- Admin panel alerts: ${GREEN}/var/log/openpanel/admin/notifications.log${NC}
- Admin panel crons: ${GREEN}/var/log/openpanel/admin/cron.log${NC}

${BLUE}------------------------------------------------------------${NC}
"
