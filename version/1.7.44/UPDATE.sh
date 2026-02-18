#!/bin/bash

echo
wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py
wget -O /etc/systemd/system/admin.service https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/openadmin.service

systemctl daemon-reload

