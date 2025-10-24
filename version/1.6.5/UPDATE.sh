#!/bin/bash

#wget -O /etc/openpanel/docker/compose/nodejs.yml https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/compose/nodejs.yml
#wget -O /etc/openpanel/docker/compose/python.yml https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/compose/python.yml


CONFIG_FILE="/etc/openpanel/openadmin/config/notifications.ini"
LINE="dns=yes"
grep -qxF "$LINE" "$CONFIG_FILE" || echo "$LINE" >> "$CONFIG_FILE"
