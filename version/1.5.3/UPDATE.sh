#!/bin/bash

sed -i 's|/usr/local/admin/service/notifications.sh|/usr/local/opencli sentinel|g' /etc/cron.d/openpanel

wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py
