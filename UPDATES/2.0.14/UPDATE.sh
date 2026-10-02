#!/bin/bash

# fresh installs on 2.0.12 and 2.0.13 still got the cron with opencli waf-update, so the monthly OWASP CRS update never ran
CRON_FILE="/etc/cron.d/openpanel"
if [ -f "$CRON_FILE" ] && grep -q "opencli waf-update" "$CRON_FILE"; then
    echo "Fixing OWASP CRS update cron in $CRON_FILE..."
    sed -i 's#opencli waf-update#opencli waf update#g' "$CRON_FILE"
fi

# new wp manager rules disable_wp_admin and mitigate_spam_logins, plus fixes for env_files and author_scan, https://github.com/stefanpejcic/OpenPanel/issues/1167
WP_RULES="/etc/openpanel/caddy/templates/wp.rules"
echo "Replacing WP Manager security rules in $WP_RULES..."
mkdir -p "$(dirname "$WP_RULES")"
if wget --timeout=15 --tries=3 -q -O "$WP_RULES.new" https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/main/caddy/templates/wp.rules && grep -q "^(wp_manager_" "$WP_RULES.new"; then
    [ -f "$WP_RULES" ] && cp "$WP_RULES" "$WP_RULES.bak-2.0.14"
    cat "$WP_RULES.new" > "$WP_RULES"
    if podman ps --format '{{.Names}}' | grep -qx caddy; then
        # put the old file back if caddy doesn't accept the new one
        if podman exec caddy caddy validate --config /etc/caddy/Caddyfile >/dev/null 2>&1; then
            podman exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1
        elif [ -f "$WP_RULES.bak-2.0.14" ]; then
            echo "Caddy config is not valid with the new rules, restoring the old $WP_RULES"
            cat "$WP_RULES.bak-2.0.14" > "$WP_RULES"
        fi
    fi
else
    echo "Failed to download wp.rules, keeping the existing file"
fi
rm -f "$WP_RULES.new"
