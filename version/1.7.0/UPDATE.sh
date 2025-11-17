#!/bin/bash

echo 
echo "Disabling CoreRuleSet REQUEST-941-APPLICATION-ATTACK-XSS.conf"
mv /etc/openpanel/caddy/coreruleset/rules/REQUEST-941-APPLICATION-ATTACK-XSS.conf /etc/openpanel/caddy/coreruleset/rules/REQUEST-941-APPLICATION-ATTACK-XSS.conf.disabled
