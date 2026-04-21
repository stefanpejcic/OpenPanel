#!/bin/bash
################################################################################
# Script Name: domains/suspend.sh
# Description: Suspend a domain name
# Usage: opencli domains-suspend <DOMAIN-NAME>
# Author: Stefan Pejcic
# Created: 04.11.2024
# Last Modified: 20.04.2026
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

if [ $# -ne 1 ]; then
    echo "Usage: opencli domains-suspend <domain_name>"
    exit 1
fi

domain_name="$1"
domain_vhost="/etc/openpanel/caddy/domains/$domain_name.conf"
suspended_dir="/etc/openpanel/caddy/suspended_domains/"
conf_template="/etc/openpanel/caddy/templates/suspended.conf"

[ -f "${suspended_dir}${domain_name}.conf" ] && { echo "Domain is already suspended."; exit 0; }
[ -f "$domain_vhost" ] || { echo "ERROR: vhost file for domain $domain_name does not exist"; exit 1; }

# 1. cp vhost to suspended-dir
mkdir -p $suspended_dir
cp $domain_vhost $suspended_dir  > /dev/null 2>&1

# 2. create file base don suspended domain template
domain_conf=$(cat "$conf_template" | sed -e "s|<DOMAIN_NAME>|$domain_name|g")
echo "$domain_conf" > "$domain_vhost"

# 3. reload caddy
nohup docker --context=default exec caddy sh -c "caddy validate && caddy reload" > /dev/null 2>&1 &
disown

# 4. notify
nohup opencli sentinel --action=domains_status --title="Domain name $domain_name suspended" --message="Domain name $domain_name has been suspended." >/dev/null 2>&1 &
disown

echo "Domain suspended successfully."
