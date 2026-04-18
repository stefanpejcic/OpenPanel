#!/bin/bash


wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py
wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py



echo "Adding patch for shortlived Let's Encrypt SSL certificates for IP address.."
cd /root && docker compose down caddy && docker image rm openpanel/caddy-coraza && docker compose up -d caddy
