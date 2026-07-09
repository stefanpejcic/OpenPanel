#!/bin/bash

wget -O /usr/local/bin/opencli https://raw.githubusercontent.com/stefanpejcic/opencli/refs/heads/main/opencli && chmod +x /usr/local/bin/opencli
wget -O  /etc/openpanel/openadmin/config/features.json https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json

CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
cp "$CONFIG_FILE" /etc/openpanel/openpanel/conf/openpanel.config_1765BACKUP
if grep -q '^\[USERS\]' "$CONFIG_FILE" && ! grep -q '^password_strength=' "$CONFIG_FILE"; then
  awk '
  BEGIN {done=0}
  /^\[USERS\]/ {
    print
    print "password_strength=50"
    done=1
    next
  }
  {print}
  ' "$CONFIG_FILE" > "${CONFIG_FILE}.tmp" && mv "${CONFIG_FILE}.tmp" "$CONFIG_FILE"
fi
