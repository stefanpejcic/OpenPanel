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

[ "$(id -u)" -ne 0 ] && return

VERSION=$(opencli version)
readonly CONFIG_FILE_PATH='/etc/openpanel/openpanel/conf/openpanel.config'
readonly CADDY_FILE="/etc/openpanel/caddy/Caddyfile"
readonly CADDY_CERT_DIR="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/"
readonly GITHUB_CONF_REPO="https://github.com/stefanpejcic/openpanel-configuration"
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


get_license() {
    config=$(read_config "LICENSE")
    LICENSE_KEY=$(echo "$config" | grep -i 'key' | cut -d'=' -f2)
    
    if [[ $LICENSE_KEY =~ ^enterprise ]]; then
        LICENSE="${GREEN}Enterprise${RESET} edition"
    elif [ -z "$LICENSE_KEY" ]; then
        LICENSE="${RED}Community${RESET} edition"
    else 
        LICENSE=""
    fi
    echo "$LICENSE"
}


get_public_ip() {
    ip=$(curl --silent --max-time 1 -4 https://ip.openpanel.com || wget --timeout=1 -qO- https://ipv4.openpanel.com || curl --silent --max-time 1 -4 https://ifconfig.me)
    
    if [ -z "$ip" ] || ! [[ "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        ip=$(hostname -I | awk '{print $1}')
    fi
    
    echo "$ip"
}


get_admin_url() {
    domain_block=$(awk '/# START HOSTNAME DOMAIN #/{flag=1; next} /# END HOSTNAME DOMAIN #/{flag=0} flag {print}' "$CADDY_FILE")
    domain=$(echo "$domain_block" | sed '/^\s*$/d' | grep -v '^\s*#' | head -n1)
    domain=$(echo "$domain" | sed 's/[[:space:]]*{//' | xargs)
    domain=$(echo "$domain" | sed 's|^http[s]*://||')
        
    if [ -z "$domain" ] || [ "$domain" = "example.net" ]; then
        ip=$(get_public_ip)
        admin_url="http://${ip}:2087/"
    else
        if [ -f "${CADDY_CERT_DIR}/${domain}/${domain}.crt" ] && [ -f "${CADDY_CERT_DIR}/${domain}/${domain}.key" ]; then
            admin_url="https://${domain}:2087/"
        else
            ip=$(get_public_ip)
            admin_url="http://${ip}:2087/"
        fi
    fi
    
    echo "$admin_url"    
}


admin_url=$(get_admin_url)

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

if [[ "$LICENSE" == "Enterprise" ]]; then
    echo -e  "        - 📚 Admin Docs: https://openpanel.com/docs/admin/intro/"
    echo -e  "        - 💬 Ticketing:  https://my.openpanel.com/submitticket.php?step=2&deptid=2"
else
    echo -e  "        - 📚 Admin Docs: https://openpanel.com/docs/admin/intro/"
    echo -e  "        - 💬 Forums:     https://community.openpanel.org/"
    echo -e  "        - 👉 Discord:    https://discord.openpanel.org/"
fi

echo -e  ""
echo -e  "================================================================"
