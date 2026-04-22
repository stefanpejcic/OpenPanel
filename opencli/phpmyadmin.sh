#!/bin/bash
################################################################################
# Script Name: phpmyadmin.sh
# Description: View and set domain/ip for phpmyadmin access.
# Usage: opencli phpmyadmin [set <domain_name> | ip]
# Author: Stefan Pejcic
# Created: 30.03.2026
# Last Modified: 21.04.2026
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

readonly CADDYFILE="/etc/openpanel/caddy/Caddyfile"
START_MARKER="# START PHPMYADMIN DOMAIN #"
END_MARKER="# END PHPMYADMIN DOMAIN #"

is_valid_domain() {
    local domain="$1"
    [[ "$domain" =~ ^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$ ]]
}

get_domain() {
    grep -A1 "START PHPMYADMIN DOMAIN" "$CADDYFILE" | tail -n1 | awk '{print $1}'
}

remove_block() {
    if grep -q "$START_MARKER" "$CADDYFILE" && grep -q "$END_MARKER" "$CADDYFILE"; then
        sed -i "/$START_MARKER/,/$END_MARKER/d" "$CADDYFILE"
    fi
}

set_domain() {
    NEW_DOMAIN="$1"

    if [ -z "$NEW_DOMAIN" ]; then
        echo "Usage: opencli phpmyadmin set <domain>"
        exit 1
    fi

    if [ "$NEW_DOMAIN" = "ip" ]; then
        remove_block
        update_env_for_existing_users "disable"
        update_env_for_new_users "disable"
        echo "Done"
        return
    fi

    if ! is_valid_domain "$NEW_DOMAIN"; then
        echo "ERROR: Invalid domain format. Please provide a valid domain."
        exit 1
    fi


  if opencli domains-whoowns "$NEW_DOMAIN" | grep -q "not found in the database."; then
     if [ -f "$CADDYFILE" ] || [ -f "$CONFIG_FILE" ]; then
         if grep -q -E "^\s*$NEW_DOMAIN\s*\{" "$CADDYFILE" 2>/dev/null; then
             echo "ERROR: $NEW_DOMAIN is already configured for another system service and can not be set for phpMyAdmin access."
             exit 1
         fi
     fi
  else
      echo "ERROR: Domain $NEW_DOMAIN is already in used by an OpenPanel account and can not be set for phpMyAdmin access."
      exit 1
  fi


    NEW_BLOCK=$(cat <<EOF
$START_MARKER
${NEW_DOMAIN}: {
    @dynamicPort path_regexp port ^/(\\d+)(/.*)?\$
    handle @dynamicPort {
        uri strip_prefix /{http.regexp.port.1}
        reverse_proxy localhost:{http.regexp.port.1}
    }

    handle {
        respond "Please use your unique phpmyadmin path shown on OpenPanel > MySQL > phpMyAdmin page."
    }
}
$END_MARKER
EOF
)

    echo "Creating proxy for the $NEW_DOMAIN domain.."
    if grep -q "$START_MARKER" "$CADDYFILE" && grep -q "$END_MARKER" "$CADDYFILE"; then
        # edit
        sed -i "/$START_MARKER/,/$END_MARKER/c\\
$NEW_BLOCK
" "$CADDYFILE"
        echo "Updated existing PHPMYADMIN domain block."
    else
        # insert
        echo -e "\n$NEW_BLOCK" >> "$CADDYFILE"
        echo "Added new PHPMYADMIN domain block."
    fi
    update_env_for_existing_users "$NEW_DOMAIN"
    update_env_for_new_users "$NEW_DOMAIN"
    echo "Done"
}


update_env_for_existing_users() {
  DOMAIN="$1"

  echo "Updating PMA_ABSOLUTE_URI for all existing users..."
  for file in /home/*/.env; do
    context=$(basename "$(dirname "$file")")
    

    if [ "$DOMAIN" = "disable" ]; then
      sed -i -E 's|^PMA_ABSOLUTE_URI=.*|PMA_ABSOLUTE_URI=\"\"|' "$file"
    else
      port=$(grep -E '^PMA_PORT=' "$file" | sed -E 's/.*"([0-9]+):.*/\1/')
      if [ -n "$port" ]; then
        sed -i -E "s|^PMA_ABSOLUTE_URI=.*|PMA_ABSOLUTE_URI=\"https://$DOMAIN/${port}/\"|" "$file"
      fi
    fi

    if docker --context "$context" compose ps --services 2>/dev/null | grep "^phpmyadmin$"; then
      echo "Restarting running phpmyadmin service for docker context: $context"
      docker --context "$context" compose down phpmyadmin &>/dev/null
      docker --context "$context" compose up -d phpmyadmin &>/dev/null
    fi  
  done
}

update_env_for_new_users() {
    echo "Updating PMA_ABSOLUTE_URI in template for new users.."
    local file="/etc/openpanel/docker/compose/1.0/.env"
    if [ "$DOMAIN" = "disable" ]; then
        sed -i -E 's|^PMA_ABSOLUTE_URI=.*|PMA_ABSOLUTE_URI=\"\"|' "$file"
    else
      sed -i -E "s|^PMA_ABSOLUTE_URI=.*|PMA_ABSOLUTE_URI=\"https://$DOMAIN/{PMA_PORT}/\"|" "$file"
    fi


    touch "/etc/openpanel/caddy/domains/$DOMAIN.conf"

    if docker --context default compose ps -q caddy >/dev/null 2>&1; then
        nohup docker --context default restart caddy >/dev/null 2>&1 &
        disown
    else
        nohup bash -c "cd /root && docker --context default compose up -d caddy" >/dev/null 2>&1 &
        disown
    fi
}

if [ "$#" -eq 0 ]; then
    get_domain
elif [ "$1" = "set" ]; then
    set_domain "$2"
else
    echo "Usage:"
    echo "  opencli phpmyadmin               # Show current phpMyAdmin link"
    echo "  opencli phpmyadmin set <domain>  # Set https://DOMAIN/PORT/ for phpMyAdmin access."
    echo "  opencli phpmyadmin set ip        # Set http://IP:PORT       for phpMyAdmin access."
fi
