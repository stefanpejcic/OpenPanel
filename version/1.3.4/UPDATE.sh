#!/bin/bash

wget -O /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json

wget -O /etc/openpanel/openpanel/features/default.txt https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/features/default.txt

CADDYFILE="/etc/openpanel/caddy/Caddyfile"

if grep -q "^ *admin off" "$CADDYFILE"; then
    sed -i 's/^ *admin off/#admin off/' "$CADDYFILE"
    echo "'admin off' has been commented out."
else
    echo "No 'admin off' line found or it's already commented."
fi
