#!/bin/bash

# template for n8n
wget -O /etc/openpanel/docker/compose/n8n.yml https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/compose/n8n.yml


# menu_style=classic for existing and menu_style=modern for new installs
CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
if grep -q "^menu_style" "$CONFIG_FILE"; then
    :
else
    echo "Inserting menu_style=classic into [PANEL]..."
    sed -i '/^\[PANEL\]/a menu_style=classic' "$CONFIG_FILE"
    echo "menu_style=classic added successfully."
fi
