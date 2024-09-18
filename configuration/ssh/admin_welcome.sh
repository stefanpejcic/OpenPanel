#!/bin/bash

##########################################################################################
# Welcome Message for OpenPanel admins                                                   #
#                                                                                        #
# This script displays a welcome message to administrators upon logging into the server. #
#                                                                                        #
# To edit and make this script executable, use:                                          #
#  nano /etc/profile.d/welcome.sh && chmod +x /etc/profile.d/welcome.sh                  #
#                                                                                        #
# Author: Stefan Pejcic (stefan@pejcic.rs)                                               #
##########################################################################################

VERSION=$(cat /usr/local/panel/version)
CONFIG_FILE_PATH='/etc/openpanel/openpanel/conf/openpanel.config'
GITHUB_CONF_REPO="https://github.com/stefanpejcic/openpanel-configuration"
GREEN='\033[0;32m'
RED='\033[0;31m'
RESET='\033[0m'

read_config() {
    local section="$1"
    if [[ -z "$section" ]]; then
        section="GENERAL"
    fi

    if [[ ! -f "$CONFIG_FILE_PATH" || ! -s "$CONFIG_FILE_PATH" ]]; then
        echo -e  "${RED}FATAL ERROR:${RESET} OpenPanel main configuration file $CONFIG_FILE_PATH is empty or missing!"
        echo -e  "OpenPanel and OpenAdmin services will not work!"    
        echo -e  ""
        echo -e  "Restore configuration from a backup or download defaults from $GITHUB_CONF_REPO"
        exit 1
    fi
    
    awk -F '=' -v section="$section" '
        BEGIN {flag=0}
        /^\[/{flag=0}
        $0 == "[" section "]" {flag=1; next}
        flag && /^[^ \t]+/ {gsub(/^[ \t]+|[ \t]+$/, "", $1); gsub(/^[ \t]+|[ \t]+$/, "", $2); print $1 "=" $2}
    ' "$CONFIG_FILE_PATH"
}

get_ssl_status() {
    config=$(read_config "DEFAULT")
    ssl_status=$(echo "$config" | grep -i 'ssl' | cut -d'=' -f2)
    [[ "$ssl_status" == "yes" ]] && echo true || echo false
}

get_license() {
    config=$(read_config "LICENSE")
    LICENSE_KEY=$(echo "$config" | grep -i 'key' | cut -d'=' -f2)
    
    if [[ $LICENSE_KEY =~ ^enterprise ]]; then
        LICENSE="${GREEN}Enterprise${RESET} edition"
    elif [ -z "$LICENSE_KEY" ]; then
        LICENSE="${RED}Community${RESET} edition"
    else 
        LICENSE="" # older versions <0.2.1
    fi
    echo "$LICENSE"
}

get_force_domain() {
    config=$(read_config "DEFAULT")
    force_domain=$(echo "$config" | grep -i 'force_domain' | cut -d'=' -f2)

    if [ -z "$force_domain" ]; then
        ip=$(get_public_ip)
        force_domain="$ip"
    fi
    echo "$force_domain"
}

get_public_ip() {
    ip=$(curl --silent --max-time 1 -4 https://ip.openpanel.co || wget --timeout=1 -qO- https://ip.openpanel.co || curl --silent --max-time 1 -4 https://ifconfig.me)
    
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

regex="^[0-9]+\.[0-9]+\.[0-9]+$"
if [[ $VERSION =~ $regex ]]; then
    VERSION="v$VERSION"
else
    VERSION=""
fi

LICENSE=$(get_license)

echo -e  "================================================================"
echo -e  ""
echo -e  "This server has installed OpenPanel ${LICENSE} ${VERSION} 🚀"
echo -e  ""
echo -e  "OPENADMIN LINK: ${GREEN}${admin_url}${RESET}"
echo -e  ""
echo -e  "Need assistance or looking to learn more? We've got you covered:"
echo -e  "        - 📚 Admin Docs: https://openpanel.com/docs/admin/intro/"
echo -e  "        - 💬 Forums: https://community.openpanel.com/"
echo -e  "        - 👉 Discord: https://discord.openpanel.com/"
echo -e  ""
echo -e  "================================================================"

timeout 1 docker cp openpanel:/usr/local/panel/version /usr/local/panel/version > /dev/null 2>&1 #1 sec max
