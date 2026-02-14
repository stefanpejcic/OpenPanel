#!/bin/bash
################################################################################
# Script Name: faq.sh
# Description: Display answers to most frequently asked questions.
# Usage: opencli faq
# Author: Stefan Pejcic
# Created: 20.05.2024
# Last Modified: 13.02.2026
# Company: openpanel.comm
# Copyright (c) openpanel.comm
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
BLUE='\033[0;34m'
RED='\033[0;31m'
PURPLE='\033[0;35m'
NC='\033[0m' #reset
readonly CONFIG_FILE_PATH='/etc/openpanel/openpanel/conf/openpanel.config'
readonly service_name="admin"

# ======================================================================
# Helper functions
SCRIPT_PATH="/usr/local/opencli/ip_servers.sh"
if [ -f "$SCRIPT_PATH" ]; then
    source "$SCRIPT_PATH"
else
    IP_SERVER_1=IP_SERVER_2=IP_SERVER_3="https://ip.openpanel.com"
fi


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
    ip=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 --tries=1 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3)
    if [ -z "$ip" ] || ! [[ "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        ip=$(hostname -I | awk '{print $1}')
    fi
    echo "$ip"
}

get_openpanel_openadmin_links() {
    readonly caddyfile="/etc/openpanel/caddy/Caddyfile"

    local domain
    domain="$(get_force_domain 2>/dev/null)"

    domain="$(echo "$domain" \
        | sed -e 's|^https\?://||' -e 's/[[:space:]]*{//' \
        | xargs)"
    domain="${domain%%:*}"

    if [[ "$domain" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        domain=""
    fi

    if [ -z "$domain" ] || [ "$domain" = "example.net" ]; then
        local domain_block
        domain_block="$(awk '
            /# START HOSTNAME DOMAIN #/ {flag=1; next}
            /# END HOSTNAME DOMAIN #/   {flag=0}
            flag {print}
        ' "$caddyfile" 2>/dev/null)"

        domain="$(echo "$domain_block" \
            | sed '/^\s*$/d' \
            | grep -v '^\s*#' \
            | head -n1 \
            | sed -e 's/[[:space:]]*{//' -e 's|^https\?://||' \
            | xargs)"
        domain="${domain%%:*}"
        domain="${domain%%,*}"
    fi

    if [ -z "$domain" ] || [ "$domain" = "example.net" ]; then
        domain="$(awk '
            function clean(h) {
                gsub(/^https?:\/\//,"",h)
                gsub(/\{.*/,"",h)
                gsub(/,.*/,"",h)
                gsub(/:[0-9]+$/,"",h)
                gsub(/^[ \t]+|[ \t]+$/,"",h)
                return h
            }
            # capture likely site label lines that end with "{"
            /^[^#].*\{[ \t]*$/ && $0 !~ /^[ \t]*\{/ {
                site=$0
                site=clean(site)
                next
            }
            # if this site block proxies to 2087, we want that hostname
            /reverse_proxy[ \t]+.*:2087/ {
                if (site != "" && site != "example.net") {
                    print site
                    exit
                }
            }
        ' "$caddyfile" 2>/dev/null)"
    fi

    local port
    port="$(opencli port 2>/dev/null)"

    if [ -z "$domain" ] || [ "$domain" = "example.net" ]; then
        local ip
        ip="$(get_public_ip)"
        user_url="http://${ip}:${port}/"
        admin_url="http://${ip}:2087/"
        return 0
    fi

    local cert_path_on_hosts="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/${domain}/${domain}.crt"
    local key_path_on_hosts="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/${domain}/${domain}.key"
    local fallback_cert_path="/etc/openpanel/caddy/ssl/custom/${domain}/${domain}.crt"
    local fallback_key_path="/etc/openpanel/caddy/ssl/custom/${domain}/${domain}.key"

    local has_cert=false
	if { [ -f "$cert_path_on_hosts" ] && [ -f "$key_path_on_hosts" ]; } || \
	   { [ -f "$fallback_cert_path" ] && [ -f "$fallback_key_path" ]; }; then
        has_cert=true
    fi

    if [ "$has_cert" = true ]; then
        user_url="https://${domain}:${port}/"
        admin_url="https://${domain}:2087/"
    else
        user_url="http://${domain}:${port}/"
        admin_url="http://${domain}:2087/"
    fi
}

get_openpanel_openadmin_links


# ======================================================================
# Q
questions=(
    "What is the login link for admin panel?"
    "What is the login link for user panel?"
    "How to restart OpenAdmin or OpenPanel services?"
    "How to reset admin password?"
    "How to create new admin account?"
    "How to list admin accounts?"
    "How to check OpenPanel version?"
    "How to update OpenPanel?"
    "How to disable automatic updates?"
    "Where are the logs?"
    "How to enable detailed logs?"
)

# ======================================================================
# A
answers=(
    "LINK: ${GREEN}${admin_url}${NC}"
    "LINK: ${GREEN}${user_url}${NC}"
    "- OpenPanel: ${RED}docker restart openpanel${NC}\n- OpenAdmin: ${RED}service admin restart${NC}"
    "execute command ${GREEN}opencli admin password USERNAME NEW_PASSWORD${NC}"
    "execute command ${GREEN}opencli admin new USERNAME PASSWORD${NC}"
    "execute command ${GREEN}opencli admin list${NC}"
    "execute command ${GREEN}opencli --version${NC}"
    "execute command ${GREEN}opencli update --force${NC}"
    "execute command ${GREEN}opencli config update autoupdate off${NC}"
    "- User panel errors:      ${GREEN}/var/log/openpanel/user/error.log${NC}\n- User panel access log:  ${GREEN}/var/log/openpanel/user/access.log${NC}\n- Admin panel errors:     ${GREEN}/var/log/openpanel/admin/error.log${NC}\n- Admin panel access log: ${GREEN}/var/log/openpanel/admin/access.log${NC}\n- Admin panel API access: ${GREEN}/var/log/openpanel/admin/api.log${NC}\n- Admin panel logins:     ${GREEN}/var/log/openpanel/admin/login.log${NC}\n- Admin panel alerts:     ${GREEN}/var/log/openpanel/admin/notifications.log${NC}\n- Admin panel crons:      ${GREEN}/var/log/openpanel/admin/cron.log${NC}"
    "- ${GREEN}opencli config update dev_mode yes${NC}"
)



# ======================================================================
# Output
echo -e "\nFrequently Asked Questions\n"

for i in "${!questions[@]}"; do
    qnum=$((i + 1))
    echo -e "${PURPLE}${qnum}.${NC} ${questions[$i]}\n"
    echo -e "${answers[$i]}"
    echo -e "${BLUE}------------------------------------------------------------${NC}\n"
done
