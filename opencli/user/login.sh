#!/bin/bash
################################################################################
# Script Name: user/login.sh
# Description: Generate an auto-login link for OpenPanel user.
# Usage: opencli user-login <username> [--open|--delete]
# Author: Stefan Pejcic
# Created: 27.01.2026
# Last Modified: 29.01.2026
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

NOW_FLAG=false
DELETE_FLAG=false
USERNAME=$1

if [ -z "$USERNAME" ]; then
    echo "Usage: opencli user-login <username>"
    exit 1
fi

for arg in "$@"; do
    if [ "$arg" == "--open" ]; then
        NOW_FLAG=true
    fi
    if [ "$arg" == "--delete" ]; then
        DELETE_FLAG=true
    fi
done


# ======================================================================
# Helpers

urlencode() {
    local LANG=C
    local length="${#1}"
    for (( i = 0; i < length; i++ )); do
        local c="${1:i:1}"
        case $c in
            [a-zA-Z0-9.~_-]) printf "$c" ;;
            *) printf '%%%02X' "'$c"
        esac
    done
}

get_public_ip() {
    ip=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 --tries=1 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3)
    if [ -z "$ip" ] || ! [[ "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        ip=$(hostname -I | awk '{print $1}')
    fi
    echo "$ip"
}

get_openpanel_url() {

  PORT=$(opencli port)
  PORT=${PORT:-2083}

	SCRIPT_PATH="/usr/local/opencli/ip_servers.sh"
	if [ -f "$SCRIPT_PATH" ]; then
	    source "$SCRIPT_PATH"
	else
	    IP_SERVER_1=IP_SERVER_2=IP_SERVER_3="https://ip.openpanel.com"
	fi

    caddyfile="/etc/openpanel/caddy/Caddyfile"
    domain_block=$(awk '/# START HOSTNAME DOMAIN #/{flag=1; next} /# END HOSTNAME DOMAIN #/{flag=0} flag {print}' "$caddyfile")
    domain=$(echo "$domain_block" | sed '/^\s*$/d' | grep -v '^\s*#' | head -n1)
    domain=$(echo "$domain" | sed 's/[[:space:]]*{//' | xargs)
    domain=$(echo "$domain" | sed 's|^http[s]*://||')
        
    if [ -z "$domain" ] || [ "$domain" = "example.net" ]; then
        ip=$(get_public_ip)
        openpanel_url="http://${ip}:$PORT/"
    else
		# ---------------------- letsencrypt
		local cert_path_on_hosts="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/${domain}/${domain}.crt"
		local key_path_on_hosts="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/${domain}/${domain}.key"
	
		# ---------------------- custom ssl
		local fallback_cert_path="/etc/openpanel/caddy/ssl/custom/${domain}/${domain}.crt"
		local fallback_key_path="/etc/openpanel/caddy/ssl/custom/${domain}/${domain}.key"
	 
		if { [ -f "$cert_path_on_hosts" ] && [ -f "$key_path_on_hosts" ]; } || \
		   { [ -f "$fallback_cert_path" ] && [ -f "$fallback_key_path" ]; }; then
		    openpanel_url="https://${domain}:$PORT/"
        else
            ip=$(get_public_ip)
            openpanel_url="http://${ip}:$PORT/"
        fi
    fi

    echo "$openpanel_url"
}




# ======================================================================
# Main

# 1. get username
USER_DIR="/etc/openpanel/openpanel/core/users/${USERNAME}"
if [ ! -d "$USER_DIR" ]; then
    # TODO: handle case when suer is suspended
    echo "[✘] Error: Username '$USERNAME' does not exist or was not properly created (missing files)."
    exit 1
fi


# 2. Read existing or generate a new token
TOKEN_FILE="$USER_DIR/logintoken.txt"

if [[ -f "$TOKEN_FILE" ]]; then
  # read existing token
  ADMIN_TOKEN=$(<"$TOKEN_FILE")
  # delete token if '--delete' flag
  if [ "$DELETE_FLAG" = true ]; then
    rm -rf $TOKEN_FILE
    echo "Auto-login token '$ADMIN_TOKEN' for user ${USERNAME} is now invalidated."
    exit 0
  fi
else
  # show msg that there is no token if '--delete' flag
  if [ "$DELETE_FLAG" = true ]; then
    echo "No auto-login token exists for user ${USERNAME}."
    exit 0
  fi
  # generate new token
  mkdir -p "$(dirname "$TOKEN_FILE")"
  ADMIN_TOKEN=$(tr -dc 'A-Za-z0-9' </dev/urandom | head -c 30)
  echo "$ADMIN_TOKEN" > "$TOKEN_FILE"
fi


# 3. Format login link with the token
openpanel_url=$(get_openpanel_url)
BASE_URL="${openpanel_url}login_autologin"
QUERY="admin_token=$(urlencode "$ADMIN_TOKEN")&username=$(urlencode "$USERNAME")"
LOGIN_URL="${BASE_URL}?${QUERY}"


# 4. Display link
echo "$LOGIN_URL"


# 5. if '--open' flag, then try to open the link
if [ "$NOW_FLAG" = true ]; then
    if command -v xdg-open >/dev/null 2>&1; then
        xdg-open "$LOGIN_URL"
    else
        echo "xdg-open not found, cannot open URL automatically."
    fi
fi
