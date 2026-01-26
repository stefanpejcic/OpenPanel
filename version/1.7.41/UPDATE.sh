#!/bin/bash

# https://github.com/stefanpejcic/OpenPanel/issues/834
echo 
wget -O /etc/openpanel/openadmin/config/shortcuts.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/shortcuts.json

# https://community.openpanel.org/d/241-caddy-failing-after-adding-a-domain-waf-update-issue
echo
mv /etc/openpanel/caddy/coreruleset/rules/REQUEST-941-APPLICATION-ATTACK-XSS.conf.disabled /etc/openpanel/caddy/coreruleset/rules/REQUEST-941-APPLICATION-ATTACK-XSS.conf && docker restart caddy
