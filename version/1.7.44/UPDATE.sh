#!/bin/bash

opencli config update dev_mode off

# services
curl -L -o /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py
curl -L -o /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py
curl -L -o /etc/systemd/system/admin.service https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/openadmin.service

systemctl daemon-reload

# caddy
sed -i '/interval 1h/d' /etc/openpanel/caddy/Caddyfile && docker --copntext=default restart caddy


