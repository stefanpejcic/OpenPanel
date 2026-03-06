#!/bin/bash

echo "Applying patch for WAF error starting Caddy.."
sed -i '/^SecRequestBodyJsonDepthLimit/ s/^/#/' /etc/openpanel/caddy/coraza_rules.conf
docker --context=default restart caddy


echo "Updating MOTD for admins to support custom OpenAdmin port.."
wget -O /etc/openpanel/ssh/admin_welcome.sh https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ssh/admin_welcome.sh
