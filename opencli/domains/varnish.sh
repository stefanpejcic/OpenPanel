#!/bin/bash
################################################################################
# Script Name: domains/varnish.sh
# Description: Check Varnish status for domain, enable/disable Varnish caching.
# Usage: opencli domains-varnish <DOMAIN-NAME> [on|off] [--short]
# Author: Stefan Pejcic
# Created: 20.03.2025
# Last Modified: 13.07.2025
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

# Ensure at least the domain name is provided
if [[ -z "$1" ]]; then
    echo "Usage: $0 <domain> [on|off] [--short]"
    exit 1
fi


# Check for the --short flag
JSON_OUTPUT=false
if [[ "$2" == "--short" || "$3" == "--short" ]]; then
    JSON_OUTPUT=true
fi


DOMAIN="$1"
CONF_FILE="/etc/openpanel/caddy/domains/$DOMAIN.conf"

# Check if the configuration file exists
if [[ ! -f "$CONF_FILE" ]]; then
        if $JSON_OUTPUT; then
		echo "error"
        else
    		echo "Error: Configuration file for $DOMAIN not found!"
    	fi
    exit 1
fi



if grep -q "^#.*reverse_proxy https://127.0.0.1" "$CONF_FILE"; then    
        if $JSON_OUTPUT; then
            echo "on"
        else
            echo "Varnish is ON for domain $DOMAIN"
        fi       
else
        if $JSON_OUTPUT; then
            echo "off"
        else
            echo "Varnish is OFF for domain $DOMAIN"
        fi    
fi




ACTION="$2"

get_context() {
	get_domain_owner=$(opencli domains-whoowns "$DOMAIN" --context)
	context=$(echo "$get_domain_owner" | awk '{print $2}')
}


reload_caddy() {
		# Reload Caddy container
		docker --context default restart caddy  > /dev/null


	    if $JSON_OUTPUT; then
	    	:
	    else
		echo "Caddy container reloaded."
	    fi
}

start_varnish() {
	
	get_context
	
	if docker --context $context ps -q -f name="varnish" > /dev/null; then
	    if $JSON_OUTPUT; then
	    	:
	    else
	    	echo "The Varnish ontainer is running."
	    fi
	else
	    if $JSON_OUTPUT; then
	    	:
	    else	
	    	echo "Varnish is not running, starting.."
	    fi
	    cd /home/$context/ && docker --context $context compose up -d varnish > /dev/null
	fi
}


if [[ "$ACTION" == "on" ]]; then
    # Comment the lines under "Handle HTTPS traffic (port 443) with on_demand SSL"
    sed -i '/# Handle HTTPS traffic (port 443) with on_demand SSL/,+6 s/^/#/' "$CONF_FILE"

    # Uncomment the lines under "Terminate TLS and pass to Varnish"
    sed -i '/# Terminate TLS and pass to Varnish/,+3 s/^#//' "$CONF_FILE"
    
    start_varnish
    
    reload_caddy

elif [[ "$ACTION" == "off" ]]; then
    # Uncomment the lines under "Handle HTTPS traffic (port 443) with on_demand SSL"
    sed -i '/# Handle HTTPS traffic (port 443) with on_demand SSL/,+6 s/^#//' "$CONF_FILE"

    # Comment the lines under "Terminate TLS and pass to Varnish"
    sed -i '/# Terminate TLS and pass to Varnish/,+3 s/^/#/' "$CONF_FILE"
    
    reload_caddy
fi



