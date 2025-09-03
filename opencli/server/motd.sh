#!/bin/bash

##########################################################################################
# Welcome Message for OpenPanel users                                                    #
#                                                                                        #
# This script displays a welcome message to users upon logging into the server.          #
#                                                                                        #
# To edit and make this script executable, use:                                          #
# nano /etc/openpanel/skeleton/welcome.sh && chmod +x /etc/openpanel/skeleton/welcome.sh #
#                                                                                        #
# Author: Stefan Pejcic (stefan@pejcic.rs)                                               #
##########################################################################################

VERSION=$(opencli version)
CONFIG_FILE_PATH='/etc/openpanel/openpanel/conf/openpanel.config'
CADDY_FILE="/etc/openpanel/caddy/Caddyfile"
CADDY_CERT_DIR=(
    "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory"
    "/etc/openpanel/caddy/ssl/custom"
)
OUTPUT_FILE='/etc/openpanel/skeleton/motd'
GREEN='\033[0;32m'
RED='\033[0;31m'
RESET='\033[0m'

DOCS_LINK="https://openpanel.com/docs/user/intro/"
FORUM_LINK="https://community.openpanel.org/"
DISCORD_LINK="https://discord.openpanel.org/"


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
        LICENSE="" # older versions <0.2.1
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


get_user_url() {
    domain_block=$(awk '/# START HOSTNAME DOMAIN #/{flag=1; next} /# END HOSTNAME DOMAIN #/{flag=0} flag {print}' "$CADDY_FILE")
    domain=$(echo "$domain_block" | sed '/^\s*$/d' | grep -v '^\s*#' | head -n1)
    domain=$(echo "$domain" | sed 's/[[:space:]]*{//' | xargs)
    domain=$(echo "$domain" | sed 's|^http[s]*://||')
    port="2083" # TODO!
    if [ -z "$domain" ] || [ "$domain" = "example.net" ]; then
        ip=$(get_public_ip)
        user_url="http://${ip}:${port}/"
    else
        if [ -f "${CADDY_CERT_DIR}/${domain}/${domain}.crt" ] && [ -f "${CADDY_CERT_DIR}/${domain}/${domain}.key" ]; then
            user_url="https://${domain}:${port}/"
        else
            ip=$(get_public_ip)
            user_url="http://${ip}:${port}/"
        fi
    fi
    
    echo "$user_url"    
}

user_url=$(get_user_url)

regex="^[0-9]+\.[0-9]+\.[0-9]+$"
if [[ $VERSION =~ $regex ]]; then
    VERSION="v$VERSION"
else
    VERSION=""
fi

LICENSE=$(get_license)



mkdir -p /etc/openpanel/skeleton/
touch $OUTPUT_FILE



{
    echo -e  "================================================================"
    echo -e  ""
    echo -e  "This server has installed OpenPanel ${LICENSE} ${VERSION} 🚀"
    echo -e  ""
    echo -e  "OPENPANEL LINK: ${GREEN}${user_url}${RESET}"
    echo -e  ""
    echo -e  "Need assistance or looking to learn more? We've got you covered:"
    echo -e  "        - 📚 User Docs: $DOCS_LINK"
    echo -e  "        - 💬 Forums:    $FORUM_LINK"
    echo -e  "        - 👉 Discord:   $DISCORD_LINK"
    echo -e  ""
    echo -e  "================================================================"
} > $OUTPUT_FILE

chmod +x /etc/openpanel/ssh/lazydocker.sh > /dev/null 2>&1
