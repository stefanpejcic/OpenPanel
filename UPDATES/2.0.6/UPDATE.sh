#!/bin/bash

# Error: bad username; while reading /etc/cron.d/openpanel (*system*openpanel) ERROR (Syntax error, this crontab file will be ignored)
sed -i 's/^0 0 \* \* 0 \* root/0 0 * * 0 root/' /etc/cron.d/openpanel
systemctl restart cron



FILE="/etc/openpanel/openadmin/service/openadmin.service"

for line in "TimeoutStopSec=30" "KillMode=mixed"; do
    if ! grep -qxF "$line" "$FILE"; then
        sed -i "/^Restart=always$/a $line" "$FILE"
    fi
done

systemctl daemon-reload
