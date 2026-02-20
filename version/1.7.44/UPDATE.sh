#!/bin/bash

echo
wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py

wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py


wget -O /etc/systemd/system/admin.service https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/openadmin.service

sed -i '/interval 1h/d' /etc/openpanel/caddy/Caddyfile && docker --copntext=default restart caddy

systemctl daemon-reload

