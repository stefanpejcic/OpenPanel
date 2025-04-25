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

echo ""
echo "Adding demo_mode option for both OpenAdmin and OpenPanel.."
FILE="/etc/openpanel/openpanel/conf/openpanel.config"

# Check if dev_mode exists under the [PANEL] section
if ! grep -q "dev_mode" "$FILE"; then
    if grep -q "\[PANEL\]" "$FILE"; then
        sed -i '/\[PANEL\]/a dev_mode=off' "$FILE"
        echo "dev_mode=off added under [PANEL]."
    else
        echo "error: PANEL section doesn't exist!"
    fi
else
    echo "dev_mode already exists."
fi
