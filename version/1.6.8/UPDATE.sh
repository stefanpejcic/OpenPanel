#!/bin/bash

echo
echo "📥 Updating features to include Mautic (BETA)"
wget -O /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json

echo
echo "Updating OpenPanel configuraiton to show errors regardless of dev_mode setting"
wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py

set -euo pipefail

echo
echo "Adding configuration option for SSH logins on 'OpenAdmin > Notifications'."

FILE="/etc/openpanel/openadmin/config/notifications.ini"

if grep -xq 'ssh=yes' "$FILE"; then
  echo "ssh=yes already present. No changes needed."
  exit 0
fi

if grep -q '^\[DEFAULT\]' "$FILE"; then
  awk 'BEGIN{added=0}
       /^\[DEFAULT\]/ && added==0 { print; print "ssh=yes"; added=1; next }
       { print }
       END { if(added==0) print "ssh=yes" }' "$FILE" > "${FILE}.tmp" && mv "${FILE}.tmp" "$FILE"
  echo "ssh=yes added immediately after [DEFAULT]."
else
  echo "" >> "$FILE"
  echo "ssh=yes" >> "$FILE"
  echo "DEFAULT section not found — ssh=yes appended at end of file."
fi

exit 0
