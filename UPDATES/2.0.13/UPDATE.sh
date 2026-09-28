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
