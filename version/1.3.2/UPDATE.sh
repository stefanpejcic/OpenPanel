#!/bin/bash


mkdir -p /etc/openpanel/openpanel/features

opencli config update session_duration 100

echo ""
echo "📥 Downloading dfault feature set for new users.."
wget -O /etc/openpanel/openpanel/features/default.txt https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/features/default.txt

echo ""
echo "Setting default feature set for all existing users.."


for dir in /home/*; do
  if [ -d "$dir" ]; then
    if [ -f "$dir/.env" ]; then
      ln -sf /etc/openpanel/openpanel/features/default.txt "$dir/features.txt"
      echo "- Default feature set applied for user $dir"
    fi
  fi
done
