#!/bin/bash

echo "Applying patch for WAF error starting Caddy.."
sed -i '/^SecRequestBodyJsonDepthLimit/ s/^/#/' /etc/openpanel/caddy/coraza_rules.conf
docker --context=default restart caddy


echo "Updating MOTD for admins to support custom OpenAdmin port.."
wget -O /etc/openpanel/ssh/admin_welcome.sh https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ssh/admin_welcome.sh


echo "Adding support for Webhooks on OpenAdmin Notifications.."
file="/etc/openpanel/openadmin/config/notifications.ini"

if ! grep -q "webhook_url" "$file"; then
    echo "webhook_url=" >> "$file"
    echo "webhook_url option appended to $file."
else
    echo "webhook_url option already exists in $file."
fi
