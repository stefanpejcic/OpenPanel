#!/bin/bash

echo ""
echo "Downloading crons.ini template.."
wget -q -O "/etc/openpanel/ofelia/users.ini" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ofelia/users.ini"

for dir in /home/*; do
    username=$(basename "$dir")
    CONF_PATH="$dir/crons.ini"
    USER_CONF="/etc/openpanel/$username/users.ini"

    if [ -d "$CONF_PATH" ]; then
        echo "Found directory: $CONF_PATH - removing and creating a file."
        rm -rf "$CONF_PATH" && touch "$CONF_PATH"
    fi
done

