#!/bin/bash

# rate limits: IPs in csf.allow and csf.ignore skip them now (#1366), the snippets get the exemption and the list is synced every 5 minutes
CRON=/etc/cron.d/openpanel
SYNC='*/5 * * * * root /usr/local/bin/opencli domains-ratelimit trusted sync --quiet'
if [ -f "$CRON" ] && ! grep -q "domains-ratelimit trusted sync" "$CRON"; then
    echo "Adding the rate limit trusted IPs sync to $CRON..."
    printf '%s\n' "$SYNC" >> "$CRON"
fi
# rewrites ratelimit.conf with the exemption only where rate limiting is set up, caddy validates it first and the old files go back if it says no
if [ -x /usr/local/bin/opencli ]; then
    /usr/local/bin/opencli domains-ratelimit trusted sync || echo "Failed to apply the rate limit trusted IPs, run: opencli domains-ratelimit trusted sync"
fi
