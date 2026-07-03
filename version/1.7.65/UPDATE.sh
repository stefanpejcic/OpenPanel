#!/bin/bash

wget -O /usr/local/bin/opencli https://raw.githubusercontent.com/stefanpejcic/opencli/refs/heads/main/opencli && chmod +x /usr/local/bin/opencli


CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
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
