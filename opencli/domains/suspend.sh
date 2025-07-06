#!/bin/bash
################################################################################
# Script Name: domains/suspend.sh
# Description: Suspend a domain name
# Usage: opencli domains-suspend <DOMAIN-NAME>
# Author: Stefan Pejcic
# Created: 04.11.2024
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


check_and_add_to_enabled() {
    if docker exec caddy caddy validate --config /etc/caddy/Caddyfile 2>&1 | grep -q "Valid configuration"; then
        docker exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1
        return 0
    else
        return 1
    fi
}

validate_conf() {

  	if [ $(docker ps -q -f name=caddy) ]; then
        :
	else
        cd /root && docker compose up -d caddy  >/dev/null 2>&1
     fi
     
	check_and_add_to_enabled
 
	if [ $? -eq 0 ]; then
		echo "Domain suspended successfully."
	else
        mv ${suspended_dir}${domain_name}.conf  $domain_vhost > /dev/null 2>&1
        check_and_add_to_enabled
		echo "ERROR: Failed to validate conf after suspend, changes reverted."
	fi
}

edit_caddy_vhosts() {
       domain_vhost="/etc/openpanel/caddy/domains/$domain_name.conf"
       suspended_dir="/etc/openpanel/caddy/suspended_domains/"
       conf_template="/etc/openpanel/caddy/templates/suspended.conf"
       mkdir -p $suspended_dir
       if [ -f "${suspended_dir}${domain_name}.conf" ]; then
		echo "Domain is already suspended."
	else
	       if [ -f "$domain_vhost" ]; then      
	            cp $domain_vhost $suspended_dir  > /dev/null 2>&1
	            domain_conf=$(cat "$conf_template" | sed -e "s|<DOMAIN_NAME>|$domain_name|g")
	            echo "$domain_conf" > "$domain_vhost"
	            validate_conf
	         
	        else
	            echo "ERROR: vhost file for domain $domain_name does not exist"
	        fi
	fi
}



# Check for the domain argument
if [ $# -ne 1 ]; then
    echo "Usage: $0 <domain_name>"
    exit 1
fi

# Get the domain name from the command line argument
domain_name="$1"
edit_caddy_vhosts
