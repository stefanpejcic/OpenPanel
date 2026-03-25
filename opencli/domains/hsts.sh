#!/bin/bash
################################################################################
# Script Name: hsts.sh
# Description: Manage HSTS for a domain
# Usage: opencli hsts <domain> [on|off] 
# Author: Stefan Pejcic
# Created: 22.05.2025
# Last Modified: 24.03.2026
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
# Helpers

usage() {
    echo "Usage: opencli domains-hsts <domain> [options]"
    echo ""
    echo "Examples:"
    echo "  opencli domains-hsts pejcic.rs          - Check HSTS status for domain pejcic.rs"
    echo "  opencli domains-hsts pejcic.rs enable   - Enable HSTS for domain pejcic.rs"
    echo "  opencli domains-hsts pejcic.rs disable  - Disable HSTS for domain pejcic.rs"
    exit 1
}


check_domain() {
    local domain="$1"

    [[ -f "$file" ]] || { echo "Domain not found!"; exit 1; }

    if grep -iq 'Strict-Transport-Security' "$file"; then
        echo "Strict-Transport-Security header is set for domain $domain"
    else
        echo "Strict-Transport-Security header is NOT set for domain $domain"
    fi
    exit 0
}

set_hsts_for_domain() {
    local domain="$1"
    local action="$2"   # "enable" or "disable"

    [[ -f "$file" ]] || { echo "Domain not found!"; exit 1; }

    if [[ "$action" == "enable" ]]; then

        if grep -iq 'Strict-Transport-Security' "$file"; then
            echo "HSTS already enabled for $domain"
            return
        fi

        awk '
        {
            print
        
            # detect tls block start
            if ($0 ~ /^[[:space:]]*tls[[:space:]]*\{/) {
                in_tls_block=1
            }
        
            # AutoSSL format: tls { }
            else if (in_tls_block && $0 ~ /^[[:space:]]*\}/) {
                print ""
                print "  # HSTS"
                print "  header {"
                print "    Strict-Transport-Security \"max-age=2592000; preload\""
                print "  }"
                in_tls_block=0
            }
        
            # Custom SSL format: tls path/to/cert /path/to/key
            else if ($0 ~ /^[[:space:]]*tls[[:space:]]+[^ {]/) {
                print ""
                print "  # HSTS"
                print "  header {"
                print "    Strict-Transport-Security \"max-age=2592000; preload\""
                print "  }"
            }
        }
        ' "$file" > "${file}.HSTS.tmp" && mv "${file}.HSTS.tmp" "$file"

        echo "HSTS enabled for $domain"
    elif [[ "$action" == "disable" ]]; then
        sed -i '/# HSTS/,+3d' "$file"
        echo "HSTS disabled for $domain"
    else
        echo "Invalid action. Use 'enable' or 'disable'."
        exit 1
    fi

    nohup docker --context=default exec caddy sh -c "caddy validate && caddy reload" > /dev/null 2>&1 &
    disown
}



# ======================================================================
# Main
domain="$1"
action="$2"

[ -n "$domain" ] || { usage; exit 1; }

file="/etc/openpanel/caddy/domains/${domain}.conf"

[ -z "$action" ] && check_domain "$domain"

case "$action" in
    enable|disable)
        set_hsts_for_domain "$domain" "$action"
        nohup opencli sentinel --action=domains_hsts --title="HSTS ${action}d for domain $domain" --message="HSTS status has been changed to:'${action}d' for domain: '$domain'." >/dev/null 2>&1 &
        disown
        ;;
    *)
        echo "Invalid action: $action"
        usage
        exit 1
        ;;
esac
