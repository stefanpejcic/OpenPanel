#!/bin/bash

echo ""
echo "📥 Updating openadmin service.."

wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py
systemctl restart admin  > /dev/null 2>&1
