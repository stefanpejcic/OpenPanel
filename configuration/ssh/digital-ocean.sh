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

VERSION=$(cat /usr/local/panel/version)
CONFIG_FILE_PATH='/etc/openpanel/openpanel/conf/openpanel.config'
GREEN='\033[0;32m'
RED='\033[0;31m'
RESET='\033[0m'

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
    ip=$(curl -s https://ip.openpanel.co)
    
    # If curl fails, try wget
    if [ -z "$ip" ]; then
        ip=$(wget -qO- https://ip.openpanel.co)
    fi
    
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
echo -e  "        - 💬 Forums: https://community.openpanel.com/"
echo -e  "        - 👉 Discord: https://discord.openpanel.com/"
echo -e  ""
echo -e  "================================================================"
