#!/bin/bash
# Description:
# This patch fixes the 500 error on user pages when Services module is enabled.
# It is safe to run multiple times.


wget -O /etc/openpanel/ftp/start_vsftpd.sh https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ftp/start_vsftpd.sh

chmod +x /etc/openpanel/ftp/start_vsftpd.sh

opencli update --cli

FILE="/root/docker-compose.yml"
SERVICE="openadmin_ftp"
VOLUME_LINE="      - /etc/openpanel/ftp/users/:/etc/openpanel/ftp/users/"

if grep -q "/etc/openpanel/ftp/users/" "$FILE"; then
    echo "Volume already exists. Skipping modification."
else
    echo "Adding volume to $SERVICE..."

    awk -v service="$SERVICE" -v volume="$VOLUME_LINE" '
    $0 ~ service ":" { in_service=1 }
    in_service && /volumes:/ { print; print volume; in_service=0; next }
    { print }
    ' "$FILE" > "${FILE}.tmp" && mv "${FILE}.tmp" "$FILE"

    echo "Volume added."
fi


docker compose down "$SERVICE"
docker compose up -d "$SERVICE"

