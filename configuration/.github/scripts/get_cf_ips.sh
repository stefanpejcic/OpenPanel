#!/bin/bash

cf_ips="$(curl -fsLm5 --retry 3 https://api.cloudflare.com/client/v4/ips)"

if [ -n "$cf_ips" ] && [ "$(echo "$cf_ips" | jq -r '.success')" = "true" ]; then
    cf_inc="nginx/cloudflare.inc"

    echo "[ * ] Updating Cloudflare IP Ranges for Nginx..."
    echo "# Cloudflare IP Ranges" > $cf_inc
    echo "" >> $cf_inc
    echo "# IPv4" >> $cf_inc
    for ipv4 in $(echo "$cf_ips" | jq -r '.result.ipv4_cidrs[]' | sort); do
        echo "set_real_ip_from $ipv4;" >> $cf_inc
    done
    echo "" >> $cf_inc
    echo "# IPv6" >> $cf_inc
    for ipv6 in $(echo "$cf_ips" | jq -r '.result.ipv6_cidrs[]' | sort); do
        echo "set_real_ip_from $ipv6;" >> $cf_inc
    done
    echo "" >> $cf_inc
    echo "real_ip_header CF-Connecting-IP;" >> $cf_inc
fi
