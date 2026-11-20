#!/bin/bash

# remove the backups 2.0.22 made: docker-compose.yml files (docker-proxy restart line) and the Caddyfile (reuse_private_keys)
rm -f /etc/openpanel/docker/compose/1.0/docker-compose.yml.bak-2.0.22 /home/*/docker-compose.yml.bak-2.0.22
rm -f /etc/openpanel/caddy/Caddyfile.bak-2.0.22

# fresh themes.yml so the Saturn theme shows under Available themes, OpenAdmin falls back to this copy when github can't be reached
CONFIG_RAW="${CONFIG_RAW:-https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main}"
mkdir -p /etc/openpanel/openpanel/templates
tmp=$(mktemp)
if wget --timeout=15 --tries=2 -q -O "$tmp" "$CONFIG_RAW/openpanel/templates/themes.yml" && grep -q "^themes:" "$tmp"; then
    echo "Updating /etc/openpanel/openpanel/templates/themes.yml..."
    cat "$tmp" > /etc/openpanel/openpanel/templates/themes.yml
    chmod 644 /etc/openpanel/openpanel/templates/themes.yml
else
    echo "Failed to download $CONFIG_RAW/openpanel/templates/themes.yml"
fi
rm -f "$tmp"

# allow_user_theme=no for existing installs, users only get the theme picker once the admin turns it on
CONF=/etc/openpanel/openpanel/conf/openpanel.config
if [ -f "$CONF" ] && ! grep -q "^allow_user_theme=" "$CONF"; then
    echo "Adding allow_user_theme=no to $CONF..."
    if grep -q "^template=" "$CONF"; then
        sed -i '/^template=/a allow_user_theme=no' "$CONF"
    else
        sed -i '/^\[PANEL\]/a allow_user_theme=no' "$CONF"
    fi
fi

# blocklists sentinel checks the server IP against, only added when missing so the admin's own list stays
RBL_LIST=/etc/openpanel/openadmin/config/rbl.txt
if [ ! -f "$RBL_LIST" ]; then
    tmp=$(mktemp)
    if wget --timeout=15 --tries=2 -q -O "$tmp" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/rbl.txt" && grep -q "spamhaus" "$tmp"; then
        echo "Adding $RBL_LIST..."
        mkdir -p "$(dirname "$RBL_LIST")"
        cat "$tmp" > "$RBL_LIST"
        chmod 644 "$RBL_LIST"
    fi
    rm -f "$tmp"
fi

# the admin's own daily summary (openadmin daily-summary) is only in the cron of installs from 2.0.26 on, older servers showed "No daily summary yet"
CRON_FILE="/etc/cron.d/openpanel"
if [ -f "$CRON_FILE" ] && ! grep -q "daily-summary >>" "$CRON_FILE"; then
    echo "Adding the admin daily summary cron to $CRON_FILE..."
    echo '45 6 * * * root for b in /usr/local/admin/openadmin-amd64 /usr/local/admin/openadmin-arm64; do [ -x "$b" ] && "$b" daily-summary >> /var/log/openpanel/admin/cron.log 2>&1 && break; done' >> "$CRON_FILE"
fi

# new daily blocklist alert, on by default
NOTIF_INI=/etc/openpanel/openadmin/config/notifications.ini
if [ -f "$NOTIF_INI" ] && ! grep -q "^rbl=" "$NOTIF_INI"; then
    echo "Adding rbl=yes to $NOTIF_INI..."
    if grep -q "^mail_defer=" "$NOTIF_INI"; then
        sed -i '/^mail_defer=/a rbl=yes' "$NOTIF_INI"
    else
        sed -i '/^\[DEFAULT\]/a rbl=yes' "$NOTIF_INI"
    fi
fi
