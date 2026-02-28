#!/bin/bash
################################################################################
# Script Name: domains/varnish.sh
# Description: Check Varnish status for domain, enable/disable Varnish caching.
# Usage: opencli domains-varnish <DOMAIN-NAME> [on|off] [--short]
# Author: Stefan Pejcic
# Created: 20.03.2025
# Last Modified: 28.02.2026
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
# ARGS
DOMAIN="$1"
ACTION="$2"
JSON_OUTPUT=false

# ======================================================================
# NO DOMAIN PROVIDED
if [[ -z "$1" ]]; then
    echo "Usage: opencli domains-varnish <domain> <on|off> <--short>"
	echo
	echo "Examples:"
	echo
	echo "opencli domains-varnish pejcic.rs               - Display Varnish Cache status for domain"
	echo "opencli domains-varnish pejcic.rs on            - Enable Varnish Cache for domain with verbose output"
	echo "opencli domains-varnish pejcic.rs               - Disable Varnish Cache for domain with verbose output"
	echo "opencli domains-varnish pejcic.rs on --short    - Enable Varnish Cache"
	echo "opencli domains-varnish pejcic.rs ff --short    - Disable Varnish Cache"
    exit 1
fi

# ======================================================================
# NO OUTPUT (used by openpanel ui which expects just on/off/error)
if [[ "$3" == "--short" ]]; then
    JSON_OUTPUT=true
fi

# ======================================================================
# CADDY FILE FOR DOMAIN
CONF_FILE="/etc/openpanel/caddy/domains/$DOMAIN.conf"

if [[ ! -f "$CONF_FILE" ]]; then
	if $JSON_OUTPUT; then
		echo "error"
	else
		echo "Error: Configuration file for $DOMAIN not found!"
	fi
    exit 1
fi


# ======================================================================
# START HELPER FUNCTIONS

get_context() {
	get_domain_owner=$(opencli domains-whoowns "$DOMAIN" --context)
	OWNER=$(echo "$get_domain_owner" | awk '{print $1}')
	context=$(echo "$get_domain_owner" | awk '{print $2}')
}

reload_caddy() {
	docker --context default restart caddy  > /dev/null	
	if ! $JSON_OUTPUT; then
		echo "Caddy container reloaded."
	fi
}

start_varnish() {
	get_context
	if docker --context $context ps -q -f name="varnish" > /dev/null; then
		if ! $JSON_OUTPUT; then
		    echo "The Varnish ontainer is running."
		fi
	else
		if ! $JSON_OUTPUT; then
	    	echo "Varnish is not running, starting.."
	    fi
	    cd /home/$context/ && docker --context $context compose up -d varnish > /dev/null
	fi
}

detect_wp_sites_and_add_mu_plugin() {

    JSON_OUTPUT=$(opencli websites-user "$OWNER" --type=wordpress --domains="$DOMAIN" --json)
    [ -z "$JSON_OUTPUT" ] && return 0    # no sites

    WP_PATHS=$(echo "$JSON_OUTPUT" | jq -r '.[].sites[]?.path // empty')
    [ -z "$WP_PATHS" ] && return 0       # no paths

 	context_uid=$(awk -F: -v user="$context" '$1 == user {print $3}' /hostfs/etc/passwd)
    [ -z "$context_uid" ] && return 0    # failed to get uid

    for WP_PATH in $WP_PATHS; do
		hostos_wp_path="/home/$context/docker-data/volumes/${context}_html_data/_data/${WP_PATH#/var/www/html/}"
        if [ -d "$WP_PATH" ]; then
            MU_DIR="${hostos_wp_path}/wp-content/mu-plugins"
            if [ ! -d "$MU_DIR" ]; then
                mkdir -p "$MU_DIR"
                chown "$context_uid:$context_uid" "$MU_DIR"
            fi

            MU_PLUGIN_FILE="$MU_DIR/opencli-helper.php"
            if [ ! -f "$MU_PLUGIN_FILE" ]; then
				cp /etc/openpanel/wordpress/mu-plugin.php "$MU_PLUGIN_FILE"
                chown "$context_uid:$context_uid" "$MU_PLUGIN_FILE"
            fi
        fi
    done
}

# END HELPER FUNCTIONS
# ======================================================================





# ======================================================================
# MAIN

case "$2" in
	on)  # ENABLE
	    sed -i '/# Handle HTTPS traffic (port 443)/,+6 s/^/#/' "$CONF_FILE"      # Comment 6 lines under "Handle HTTPS traffic (port 443)"
	    sed -i '/# Terminate TLS and pass to Varnish/,+3 s/^#//' "$CONF_FILE"    # Uncomment 3 lines under "Terminate TLS and pass to Varnish"
	    start_varnish
	    reload_caddy
		detect_wp_sites_and_add_mu_plugin                                        # https://github.com/stefanpejcic/OpenPanel/issues/853
		;;
	off) # DISABLE
	    sed -i '/# Handle HTTPS traffic (port 443)/,+6 s/^#//' "$CONF_FILE"      # Uncomment 3 lines under "Handle HTTPS traffic (port 443)"
	    sed -i '/# Terminate TLS and pass to Varnish/,+3 s/^/#/' "$CONF_FILE"    # Comment 6 lines under "Terminate TLS and pass to Varnish"
	    reload_caddy 
		;;
	*)   # DISPLAY STATUS
		if grep -q "^#.*reverse_proxy https://127.0.0.1" "$CONF_FILE"; then
			status="on"
		else
			status="off"
		fi
		
		if $JSON_OUTPUT; then
			echo "$status"
		else
			echo "Varnish is ${status^^} for domain $DOMAIN"
		fi
		;;
esac
