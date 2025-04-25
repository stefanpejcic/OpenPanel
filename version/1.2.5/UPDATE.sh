#!/bin/bash

echo ""
echo "Downloading crons.ini template.."
mkdir -p /etc/openpanel/ofelia/
wget -O "/etc/openpanel/ofelia/users.ini" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ofelia/users.ini"

for dir in /home/*; do
    username=$(basename "$dir")
    CONF_PATH="$dir/crons.ini"
    USER_CONF="/etc/openpanel/$username/users.ini"

    if [ -d "$CONF_PATH" ]; then
        rm -rf "$CONF_PATH" && touch "$CONF_PATH"
    fi
done

