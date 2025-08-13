#!/bin/bash
################################################################################
# Script Name: email/webmail.sh
# Description: Display webmail domain or choose webmail software
# Usage: opencli email-webmail [roundcube|snappymail|sogo] [--debug]
# Docs: https://docs.openpanel.com
# Author: 27.08.2024
# Created: 18.08.2024
# Last Modified: 12.08.2025
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


usage() {
    echo "Usage: opencli email-webmail {roundcube|snappymail|sogo}"
    echo
    echo "Examples:"
    echo "  opencli email-webmail                            # Display webmail domain."
    echo "  opencli email-webmail domain webmail.example.com # Set webmail.example.com as domain."
    echo "  opencli email-webmail roundcube                  # Set RoundCube as webmail."
    echo "  opencli email-webmail snappymail                 # Set SnappyMail as webmail [DEPRECATED]."
    echo "  opencli email-webmail sogo                       # Set SoGo as webmail [DEPRECATED]."
    echo ""
    exit 1
}



if [ "$#" -gt 2 ]; then
    usage
fi

DEBUG=false  # Default value for DEBUG
SNAPPYMAIL=false
ROUNDCUBE=false
SOGO=false
WEBMAIL_PORT="8080" # TODO: 8080 should be disabled and instead allow domain proxy only!



# ENTERPRISE
ENTERPRISE="/usr/local/opencli/enterprise.sh"
PANEL_CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
PROXY_FILE="/etc/openpanel/caddy/redirects.conf"
key_value=$(grep "^key=" $PANEL_CONFIG_FILE | cut -d'=' -f2-)

# Check if 'enterprise edition'
if [ -n "$key_value" ]; then
    :
else
    echo "Error: OpenPanel Community edition does not support emails. Please consider purchasing the Enterprise version that allows unlimited number of email addresses."
    source $ENTERPRISE
    echo "$ENTERPRISE_LINK"
    exit 1
fi


is_valid_domain() {
    local domain="$1"
    [[ "$domain" =~ ^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$ ]]
}

update_webmail_domain() {
    # basic
    if ! is_valid_domain "$new_domain"; then
        log_error "Invalid domain format. Please provide a valid domain."
        usage
        exit 1
    fi
    
    if [[ ! -f "$PROXY_FILE" ]]; then
        echo "Error: Caddy file not found at $PROXY_FILE"
        exit 1
    fi

    if grep -q "^redir @webmail" "$PROXY_FILE"; then
        if [[ "$DEBUG" = true ]]; then
            echo "Updating webmail redirect domain to: $new_domain"
        fi
        # Determine protocol
        if [[ "$new_domain" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            proto="http"
        else
            proto="https"
        fi

        sed -i -E "s|^redir @webmail https?://[^ ]+|redir @webmail ${proto}://${new_domain}|" "$PROXY_FILE"
   
        readonly DOMAINS_DIR="/etc/openpanel/caddy/domains/"
        readonly CADDYFILE="/etc/openpanel/caddy/Caddyfile"

        mkdir -p "$DOMAINS_DIR"
        touch "${DOMAINS_DIR}${new_domain}.conf" # so caddy can generate le

        if [[ -f "$CADDYFILE" ]]; then
            if grep -q "# START WEBMAIL DOMAIN #" "$CADDYFILE"; then
                if [[ "$DEBUG" = true ]]; then
                    echo "Updating Caddyfile webmail block to: ${new_domain}"
                fi
        
                # Escape dots for sed
                escaped_domain=$(echo "$new_domain" | sed 's/\./\\./g')
        
                # Replace the domain inside the block
sed -i "/# START WEBMAIL DOMAIN #/,/# END WEBMAIL DOMAIN #/c\\
# START WEBMAIL DOMAIN #\\
${new_domain} {\\
    reverse_proxy localhost:8080\\
}\\
# END WEBMAIL DOMAIN #" "$CADDYFILE"


            else
                echo "Warning: Webmail domain block not found in $CADDYFILE"
            fi
        else
            echo "Warning: $CADDYFILE does not exist"
        fi

        echo "Webmail domain updated to ${proto}://${new_domain}"

        # Restart Caddy
        cd /root && docker --context default compose down caddy >/dev/null 2>&1
        cd /root && docker --context default compose up -d caddy >/dev/null 2>&1
        # TODO: RELOAD!
        exit 0
    else
        echo "Error: No redir @webmail line found in $PROXY_FILE"
        exit 1
    fi
}


while [[ "$#" -gt 0 ]]; do
    case $1 in
        --debug)
            echo ""
            echo "----------------- DISPLAY DEBUG INFORMATION ------------------"
            echo ""
            DEBUG=true
            ;;
        domain)
            shift
            if [[ -z "$1" ]]; then
                echo "Error: No domain specified after 'domain'."
                echo "Usage: opencli email-webmail domain mail.example.com"
                exit 1
            fi
            new_domain="$1"
            update_webmail_domain "$new_domain"
            ;;
        roundcube)
            echo "Setting RoundCube as Webmail software:"
            SNAPPYMAIL=false
            ROUNDCUBE=true
            SOGO=false
            ;;
        snappymail)
            echo "Setting SnappyMail as Webmail software:"
            SNAPPYMAIL=true
            ROUNDCUBE=false
            SOGO=false
            ;;
        sogo)
            echo "Setting SoGo as Webmail software"
            SNAPPYMAIL=false
            ROUNDCUBE=false
            SOGO=true
            ;;
        *)
            echo "Invalid option: $1"
            usage
            exit 1
            ;;
    esac
    shift
done



get_domain_for_webmail() {
    # Check if the file exists
    if [[ ! -f "$PROXY_FILE" ]]; then
        echo "Configuration file not found at $PROXY_FILE"
        exit 1
    fi
    
    # Use grep and awk to extract the domain from the /webmail block
    domain=$(grep "^redir @webmail" "$PROXY_FILE" | awk '{print $3}' | sed -e 's|https://||' -e 's|http://||' -e 's|:.*||')
    
    if [[ -n "$domain" ]]; then
        echo "$domain"
    else
        echo "No domain configured for /webmail"
        exit 0
    fi
}




cd /usr/local/mail/openmail

if [ "$SNAPPYMAIL" = true ]; then
  if [ "$DEBUG" = true ]; then
      echo ""
      echo "----------------- STOPPING EXISTING WEBMAIL SOFTWARE ------------------"
      echo ""
      echo "Stopping RoundCube:"
    docker --context default compose rm -s -v roundcube
      echo "Stopping SoGO:"
    docker --context default compose rm -s -v sogo
      echo ""
      echo "----------------- STARTING SNAPPYMAIL ------------------"
      echo ""
    docker --context default compose up -d snappymail
  else
    docker --context default compose rm -s -v roundcube >/dev/null 2>&1
    docker --context default compose rm -s -v sogo >/dev/null 2>&1
    docker --context default compose up -d snappymail >/dev/null 2>&1
  fi
elif [ "$ROUNDCUBE" = true ]; then
    docker --context default compose rm -s -v snappymail >/dev/null 2>&1
    docker --context default compose rm -s -v sogo >/dev/null 2>&1
    docker --context default compose up -d roundcube
elif [ "$SOGO" = true ]; then
    docker --context default compose rm -s -v roundcube >/dev/null 2>&1
    docker --context default compose rm -s -v snappymail >/dev/null 2>&1
    docker --context default compose up -d sogo
else
    get_domain_for_webmail    # display domain only
fi







function open_port_csf() {
    local port=$1
    local csf_conf="/etc/csf/csf.conf"
    
    # Check if port is already open
    port_opened=$(grep "TCP_IN = .*${port}" "$csf_conf")
    if [ -z "$port_opened" ]; then
        # Open port
      if [ "$DEBUG" = true ]; then
          echo ""
          echo "Opening port on ConfigServer Firewall"
          echo ""
          sed -i "s/TCP_IN = \"\(.*\)\"/TCP_IN = \"\1,${port}\"/" "$csf_conf"
          echo ""
      else
          sed -i "s/TCP_IN = \"\(.*\)\"/TCP_IN = \"\1,${port}\"/" "$csf_conf" >/dev/null 2>&1
      fi
      ports_opened=1
    else
      if [ "$DEBUG" = true ]; then
          echo "Port ${port} is already open in CSF."
      else
          echo "Port ${port} is already open in CSF." >/dev/null 2>&1
      fi
    fi
}


if [ "$DEBUG" = true ]; then
    echo ""
    echo "----------------- OPENING PORT 8080 ON FIREWALL ------------------"
fi
# CSF
if command -v csf >/dev/null 2>&1; then
    open_port_csf $WEBMAIL_PORT    
else
      if [ "$DEBUG" = true ]; then
          echo "Warning: CSF is not installed. In order for Webmail to work, make sure port 8080 is opened on external firewall.."
      else
          :
      fi
fi




