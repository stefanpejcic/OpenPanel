#!/bin/bash
################################################################################
# Script Name: domains/ssl.sh
# Description: Check SSL for domain, add custom certificate, view files.
# Usage: opencli domains-ssl <DOMAIN_NAME> [status|info|logs|auto|custom] [path/to/fullchain.pem path/to/key.pem]
# Author: Stefan Pejcic
# Created: 22.03.2025
# Last Modified: 20.02.2026
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

# ======================================================================
# Constants
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'


usage() {
    echo "Usage:"
    echo -e "  opencli domains-ssl <DOMAIN>                 - Display command examples for this domain."
    echo -e "  opencli domains-ssl <DOMAIN> ${GREEN}status${RESET}          - Display current status for the domain."
    echo -e "  opencli domains-ssl <DOMAIN> ${GREEN}info${RESET}            - Display certificate files."
    echo -e "  opencli domains-ssl <DOMAIN> ${GREEN}logs${RESET} [${YELLOW}1000${RESET}|${YELLOW}-f${RESET}]  - View caddy SSL-related logs for the domain."
    echo -e "  opencli domains-ssl <DOMAIN> ${GREEN}custom${RESET} ${YELLOW}<cert_path>${RESET} ${YELLOW}<key_path>${RESET} - Switch to custom SSL for the domain."
    echo -e "  opencli domains-ssl <DOMAIN> ${GREEN}auto${RESET}            - Switch back to AutoSSL for the domain."
}

# Ensure a domain name is provided
if [ -z "$1" ]; then
    echo "ERROR: Domain name is required!"
    usage
	exit 1
fi


DOMAIN="$1"
CONFIG_FILE="/etc/openpanel/caddy/domains/$DOMAIN.conf"

# Ensure the file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}Domain ${DOMAIN} does not exist.$RESET"
    exit 1
fi






hostfs_domain_tls_dir="/etc/openpanel/caddy/ssl/custom/$DOMAIN"
domain_tls_dir="/data/caddy/certificates/custom/$DOMAIN"

get_user() {
  whoowns_output=$(opencli domains-whoowns "$DOMAIN")
  user=$(echo "$whoowns_output" | awk -F "Owner of '$DOMAIN': " '{print $2}')
  if [ -n "$user" ]; then
    :
  else
      echo "Failed to determine the owner of the domain '$DOMAIN'." >&2
      exit 1
  fi
  
}

get_user_info() {
    local user="$1"
    local query="SELECT id, server FROM users WHERE username = '${user}';"
    
    # Retrieve both id and context
    user_info=$(mysql -se "$query")
    
    # Extract user_id and context from the result
    user_id=$(echo "$user_info" | awk '{print $1}')
    context=$(echo "$user_info" | awk '{print $2}')
    
    echo "$user_id,$context"
}




get_user_context() {

  result=$(get_user_info "$user")
  user_id=$(echo "$result" | cut -d',' -f1)
  context=$(echo "$result" | cut -d',' -f2)

  if [ -z "$user_id" ]; then
      echo "FATAL ERROR: user $user does not exist."
      exit 1
  fi
}


check_and_use_tls() {
	local full_cert="$1"
 	local full_key="$2"
	  
	  if [[ ! "$full_cert" =~ ^/var/www/html/ && ! "$full_key" =~ ^/var/www/html/ ]]; then
	      echo "ERROR: Paths must be inside /var/www/html/ directory."
	      exit 1
	  fi


	local cert_path="${full_cert##/var/www/html/}"
 	local key_path="${full_key##/var/www/html/}"
  
 	local real_cert_path="/home/${context}/docker-data/volumes/${context}_html_data/_data/${cert_path}"
  	local real_key_path="/home/${context}/docker-data/volumes/${context}_html_data/_data/${key_path}"
  
 	if openssl x509 -noout -checkend 0 -in "$real_cert_path" >/dev/null 2>&1; then
	    mkdir -p "$hostfs_domain_tls_dir"

	    cp "${real_cert_path}" "$hostfs_domain_tls_dir"/fullchain.pem
	    cp "${real_key_path}" "$hostfs_domain_tls_dir"/key.pem
	    
		if grep -qE "tls\s+/.*?/fullchain\.pem\s+/.*?/key\.pem" "$CONFIG_FILE"; then
		    echo "Custom SSL already configured for $DOMAIN. Updating certificate and key.."
		else
		    echo "Adding custom certificate.."
     			sed -i "/tls {/,/}/c\
tls $domain_tls_dir/fullchain.pem $domain_tls_dir/key.pem
" "$CONFIG_FILE"
		fi 
	    docker --context=default exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null
	else
	    echo "Error: $cert_path is not valid or expired!"
	    exit 1
	fi
}



cat_certificate_files() {
    	if grep -q "fullchain.pem" "$CONFIG_FILE" && grep -q "key.pem" "$CONFIG_FILE"; then
	 		# custom ssl
    		cat "$hostfs_domain_tls_dir"/fullchain.pem
    		cat "$hostfs_domain_tls_dir"/key.pem
    	else
	 		# letsencrypt
    		local cert="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/$DOMAIN/$DOMAIN.crt"
          	local key="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/$DOMAIN/$DOMAIN.key"
			[ -f "$cert" ] && cat "$cert"
			[ -f "$key" ] && cat "$key"
    	fi    
}


show_ssl_logs() {
    local lines=1000
    local follow=0

    # $3 is a number or '-f'
    if [[ "$3" =~ ^[0-9]+$ ]]; then
        lines="$3"
    elif [[ "$3" == "-f" ]]; then
        follow=1
    fi

    echo "Showing SSL-related log lines for $DOMAIN"	
    echo "-------------------------------------------------------"
	# docker --context=default logs --tail 1000 caddy 2>&1  | grep "$DOMAIN" | grep -Ei 'tls|acme|certificate|renew|obtain|challenge'

    local log_path
    log_path=$(docker inspect --format='{{.LogPath}}' caddy 2>/dev/null)

    if [ -z "$log_path" ] || [ ! -f "$log_path" ]; then
        echo "ERROR: Could not determine Caddy log file path."
        exit 1
    fi

    if [ "$follow" -eq 1 ]; then
        tail -n "$lines" -f "$log_path"
    else
        tail -n "$lines" "$log_path"
    fi | jq -r '.log' | jq -R -c --arg domain "$DOMAIN" '
        fromjson? |
        select(
            .
            and
            (
                .request.host? == $domain or
                .request.tls.server_name? == $domain or
                .server_name? == $domain or
                (.identifiers?[]? == $domain)
            )
            and
            (
                (.logger? | test("tls|acme")) or
                (.msg? | test("certificate|renew|obtain|challenge"; "i"))
            )
        )
    '
}





show_examples() {
	echo -e "Usage examples for domain ${YELLOW}$DOMAIN${RESET}:"
	echo ""
	echo "- Check current SSL status for domain (AutoSSL, CustomSSL or No SSL):"
	echo -e "  opencli domains-ssl ${YELLOW}$DOMAIN${RESET} ${GREEN}status${RESET}"
	echo "- Display fullchain and key files for the domain:"
	echo -e "  opencli domains-ssl ${YELLOW}$DOMAIN${RESET} ${GREEN}info${RESET}"
	echo "- Set AutoSSL for the domain (default):"
	echo -e "  opencli domains-ssl ${YELLOW}$DOMAIN${RESET} ${GREEN}auto${RESET}"
	echo "- Add custom certificate for the domain:"
	echo -e "  opencli domains-ssl ${YELLOW}$DOMAIN${RESET} ${GREEN}custom${RESET} ${RED}/var/www/html/fullchain.pem /var/www/html/key.pem${RESET}"
	echo "- View SSL-related lines for the domain from Caddy logs:"
	echo -e "  opencli domains-ssl ${YELLOW}$DOMAIN${RESET} ${GREEN}logs${RESET}"
}


check_custom_ssl_or_auto() {   
	if grep -q "fullchain.pem" "$CONFIG_FILE"; then
	    echo "Custom SSL"
	elif grep -q "on_demand" "$CONFIG_FILE"; then
	    echo "AutoSSL"
	else
	    echo "Unknown"
	fi
}


if [ -n "$2" ]; then

  get_user
  get_user_context

    if [ "$2" == "info" ]; then
	cat_certificate_files
	exit 0
    elif [ "$2" == "status" ]; then
    	check_custom_ssl_or_auto
    	exit 0
    elif [ "$2" == "auto" ]; then
        sed -i -E "s|tls\s+/.*?/fullchain\.pem\s+/.*?/key\.pem|  tls {\n    on_demand\n  }|g" "$CONFIG_FILE"
		docker --context=default exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null
        echo "Updated $DOMAIN to use AutoSSL."
        exit 0
    elif [ "$2" == "custom" ] && [ -n "$3" ] && [ -n "$4" ]; then        
        check_and_use_tls "$3" "$4"
        echo "Updated $DOMAIN to use custom SSL."
        exit 0
	elif [ "$2" == "logs" ]; then
	    show_ssl_logs "$@"
	    exit 0
    else
	    echo "ERROR: Invalid arguments provided for domain!"
	    usage	
        exit 1
    fi
else
	show_examples
	exit 0
fi
