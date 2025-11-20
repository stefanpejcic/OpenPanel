#!/bin/bash


echo 
echo "Updating service file to show timestamp in docker log for openpanel service.."
wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py
docker restart openpanel

echo 
echo "Disabling CoreRuleSet REQUEST-941-APPLICATION-ATTACK-XSS.conf"
mv /etc/openpanel/caddy/coreruleset/rules/REQUEST-941-APPLICATION-ATTACK-XSS.conf /etc/openpanel/caddy/coreruleset/rules/REQUEST-941-APPLICATION-ATTACK-XSS.conf.disabled


echo 
echo "Adding 'ssl' module for Community edition.."
wget -O /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json

echo 
echo "Increasing activity_items_per_page value from 25 to 100"
opencli config update activity_items_per_page 100
