#!/bin/bash

echo
echo "📥 Updating features to include Mautic (BETA)"
wget -O /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json


set -euo pipefail

FILE="/etc/openpanel/openadmin/config/notifications.ini"
BACKUP="${FILE}.$(date +%Y%m%d%H%M%S).bak"

if [ ! -f "$FILE" ]; then
  echo "Error: file does not exist: $FILE" >&2
  exit 2
fi

cp -p "$FILE" "$BACKUP"

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
