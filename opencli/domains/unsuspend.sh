#!/bin/bash
################################################################################
# Script Name: domains/unsuspend.sh
# Description: Unsuspend a domain name
# Usage: opencli domains-unsuspend <DOMAIN-NAME>
# Author: Stefan Pejcic
# Created: 04.11.2024
# Last Modified: 04.04.2026
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
    echo "Usage: opencli domains-unsuspend <domain_name>"
    exit 1
fi

domain_name="$1"
domain_vhost="/etc/openpanel/caddy/domains/$domain_name.conf"
suspended_vhost="/etc/openpanel/caddy/suspended_domains/$domain_name.conf"

[ -f "$suspended_vhost" ] || { echo "Domain $domain_name is not suspended."; exit 0; }

# 1. restore vhost
cp "$suspended_vhost" "$domain_vhost"  > /dev/null 2>&1

# 2. reload caddy
nohup docker --context=default exec caddy sh -c "caddy validate && caddy reload" > /dev/null 2>&1 &
disown

# 3. notify
nohup opencli sentinel --action=domains_status --title="Domain name $domain_name unsuspended" --message="Domain name $domain_name has been unsuspended." >/dev/null 2>&1 &
disown

echo "Domain unsuspended successfully."
