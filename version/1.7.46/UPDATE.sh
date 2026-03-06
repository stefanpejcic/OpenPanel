#!/bin/bash

echo "Applying patch for WAF error starting Caddy.."
sed -i '/^SecRequestBodyJsonDepthLimit/ s/^/#/' /etc/openpanel/caddy/coraza_rules.conf
docker --context=default restart caddy
