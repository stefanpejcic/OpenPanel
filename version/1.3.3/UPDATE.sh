#!/bin/bash


echo ""
echo "📥 Editing SSH welcome message for administrators.."
wget -O /etc/openpanel/ssh/admin_welcome.sh https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ssh/admin_welcome.sh
chmod +x /etc/openpanel/ssh/admin_welcome.sh



mkdir -p /etc/openpanel/openpanel/features

echo ""
echo "📥 Downloading default feature set for new users.."
wget -O /etc/openpanel/openpanel/features/default.txt https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/features/default.txt

echo ""
echo "Setting default feature set for all existing users.."

for dir in /home/*; do
  if [ -d "$dir" ]; then
    if [ -f "$dir/.env" ]; then
      rm -rf "$dir/features.txt"
      echo "- Default feature set applied for user $dir"
    fi
  fi
done
