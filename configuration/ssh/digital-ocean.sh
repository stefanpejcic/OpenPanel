#!/bin/bash

##########################################################################################
# Welcome Message for OpenPanel admins                                                   #
#                                                                                        #
# This script displays a welcome message to administrators upon logging into the server. #
#                                                                                        #
# For DigitalOcean droplets it also creates admin user                                   #
#                                                                                        #
# To edit and make this script executable, use:                                          #
#  nano /etc/profile.d/welcome.sh && chmod +x /etc/profile.d/welcome.sh                  #
#                                                                                        #
# Author: Stefan Pejcic (stefan@pejcic.rs)                                               #
##########################################################################################

docker cp openpanel:/usr/local/panel/version /usr/local/panel/version > /dev/null 2>&1

VERSION=$(opencli version)
CONFIG_FILE_PATH='/etc/openpanel/openpanel/conf/openpanel.config'
CADDY_FILE="/etc/openpanel/caddy/Caddyfile"
CADDY_CERT_DIR="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/"
GREEN='\033[0;32m'
RED='\033[0;31m'
RESET='\033[0m'

read_config() {
    config=$(awk -F '=' '/\[DEFAULT\]/{flag=1; next} /\[/{flag=0} flag{gsub(/^[ \t]+|[ \t]+$/, "", $1); gsub(/^[ \t]+|[ \t]+$/, "", $2); print $1 "=" $2}' $CONFIG_FILE_PATH)
    echo "$config"
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

# Path to the SQLite database
DB_PATH="/etc/openpanel/openadmin/users.db"

echo -e  "================================================================"
echo -e  ""
echo -e  "This server has installed OpenPanel v${VERSION} 🚀"
echo -e  ""
echo -e  "OPENADMIN LINK: ${GREEN}${admin_url}${RESET}"
echo -e  ""

user_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM user;" 2>/dev/null)

# If the user table doesn't exist or no users are found, create a new user
if [[ $? -ne 0 || $user_count -eq 0 ]]; then
    # Download the shell script to generate a random username
    wget -O /tmp/generate.sh https://gist.githubusercontent.com/stefanpejcic/905b7880d342438e9a2d2ffed799c8c6/raw/a1cdd0d2f7b28f4e9c3198e14539c4ebb9249910/random_username_generator_docker.sh > /dev/null 2>&1

    # Source the shell script to use its functions
    source /tmp/generate.sh

    # Generate a random username
    new_username="$random_name"

    # Generate a random password of 16 characters
    new_password=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 16)

    # Create the user table if it doesn't exist
    sqlite3 "$DB_PATH" "CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', is_active BOOLEAN DEFAULT 1 NOT NULL);"  > /dev/null 2>&1 && 

    # Add the new user with the generated username and password
    opencli admin new "$new_username" "$new_password" > /dev/null 2>&1

    # Output the new user's credentials
    echo -e ""
    echo -e "Username: $new_username"
    echo -e "Password: $new_password"
    echo -e ""
fi
echo -e  "Need assistance or looking to learn more? We've got you covered:"
echo -e  "        - 📚 Admin Docs: https://openpanel.com/docs/admin/intro/"
echo -e  "        - 💬 Forums: https://community.openpanel.org/"
echo -e  "        - 👉 Discord: https://discord.openpanel.org/"
echo -e  ""
echo -e  "================================================================"
