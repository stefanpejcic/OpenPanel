#!/bin/bash

# cloudflare-only snippet now also allows the server's own ips, rebuild it so wp-cron and loopback requests stop getting 403
CF_SNIPPET="/etc/openpanel/caddy/templates/cloudflare.only"
if [ -f "$CF_SNIPPET" ] && ! grep -q "server ip" "$CF_SNIPPET"; then
    echo "Regenerating $CF_SNIPPET..."
    opencli domains-cloudflare
    if podman ps --format '{{.Names}}' | grep -qx caddy; then
        echo "Reloading Caddy..."
        podman exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1
    fi
fi
