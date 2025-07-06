#!/bin/bash
################################################################################
# Script Name: domains/ssl.sh
# Description: Check SSL for domain, add custom certificate, view files.
# Usage: opencli domains-ssl <DOMAIN_NAME> [status|info|auto|custom] [path/to/fullchain.pem path/to/key.pem]
# Author: Stefan Pejcic
# Created: 22.03.2025
# Last Modified: 04.07.2025
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




# Ensure a domain name is provided
if [ -z "$1" ]; then
    echo "Usage: opencli domains-ssl <domain> [status|info]auto|custom] [cert_path key_path]"
    exit 1
fi

DOMAIN="$1"
CONFIG_FILE="/etc/openpanel/caddy/domains/$DOMAIN.conf"

# Ensure the file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Domain does not exist."
    exit 1
fi


hostfs_domain_tls_dir="/etc/openpanel/caddy/ssl/$DOMAIN"
domain_tls_dir="/data/caddy/certificates/$DOMAIN"

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
	    mkdir -p $hostfs_domain_tls_dir

	    cp /hostfs${real_cert_path} $hostfs_domain_tls_dir/fullchain.pem
	    cp /hostfs${real_key_path} $hostfs_domain_tls_dir/key.pem
	    
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
    		cat $hostfs_domain_tls_dir/fullchain.pem
    		cat $hostfs_domain_tls_dir/key.pem
    	else
    		local cert="/hostfs/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/$DOMAIN/$DOMAIN.crt"
          	local key="/hostfs/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/$DOMAIN/$DOMAIN.key"
      		if [ -f "$cert" ]; then
			cat $cert
		fi
      		if [ -f "$key" ]; then
			cat $key
		fi

    	fi    
}


show_examples() {
	echo "Usage:"
	echo ""
	echo "Check current SSL status for domain (AutoSSL, CustomSSL or No SSL):"
	echo ""
	echo "opencli domains-ssl $DOMAIN status"
	echo ""
	echo "Display fullchain and key files for the domain:"
	echo ""
	echo "opencli domains-ssl $DOMAIN info"
	echo ""
	echo "Set free AutoSSL for the domain (default):"
	echo ""
	echo "opencli domains-ssl $DOMAIN auto"
	echo ""
	echo "Add custom certificate files for the domain:"
	echo ""
	echo "opencli domains-ssl $DOMAIN custom path/to/fullchain.pem path/to/key.pem"
	echo ""
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
    else
        echo "Invalid arguments. Usage: opencli domains-ssl <domain> [auto|custom] [cert_path key_path]"
        exit 1
    fi
else
	show_examples
	exit 0

fi


exit 0
