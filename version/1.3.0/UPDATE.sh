#!/bin/bash

echo ""
echo "📥 Updating openadmin service.."
wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py
systemctl restart admin  > /dev/null 2>&1


echo ""
echo "📥 Updating rootless.json.."
wget -O /etc/openpanel/docker/daemon/rootless.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/daemon/rootless.json
