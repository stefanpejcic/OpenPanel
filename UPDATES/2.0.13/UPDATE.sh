#!/bin/bash

# apps shown in the Websites section on the OpenPanel dashboard, new in 2.0.13
OPENPANEL_CONFIG="/etc/openpanel/openpanel/conf/openpanel.config"
if [ -f "$OPENPANEL_CONFIG" ] && ! grep -q "^applications_dashboard_items=" "$OPENPANEL_CONFIG"; then
    echo "Adding applications_dashboard_items to $OPENPANEL_CONFIG..."
    if grep -q "^\[PANEL\]" "$OPENPANEL_CONFIG"; then
        sed -i '/^\[PANEL\]/a applications_dashboard_items=websites wordpress autoinstaller' "$OPENPANEL_CONFIG"
    else
        printf '\n[PANEL]\napplications_dashboard_items=websites wordpress autoinstaller\n' >> "$OPENPANEL_CONFIG"
    fi
fi

# weakpass check uses the Weakpass top 1M list from 2.0.13, drop the old cached English dictionary
rm -f /tmp/weakpass_dictionary.txt

# waf buffered every html response for the leak checks, adding ~230ms per page, off from 2.0.13
CORAZA_RULES="/etc/openpanel/caddy/coraza_rules.conf"
if [ -f "$CORAZA_RULES" ] && grep -q "^SecResponseBodyAccess On" "$CORAZA_RULES"; then
    echo "Disabling SecResponseBodyAccess in $CORAZA_RULES..."
    sed -i 's/^SecResponseBodyAccess On/SecResponseBodyAccess Off/' "$CORAZA_RULES"
    if podman ps --format '{{.Names}}' | grep -qx caddy; then
        echo "Reloading Caddy..."
        podman exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1
    fi
fi
