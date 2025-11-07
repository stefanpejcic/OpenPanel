#!/bin/bash

echo
echo "📥 Updating features to include Mautic (BETA)"
wget -O /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json
