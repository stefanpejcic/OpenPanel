#!/bin/bash

# docker_backup notification is new in 2.0.16, add it so it can be toggled from OpenAdmin > Settings > Notifications
NOTIFICATIONS_INI="/etc/openpanel/openadmin/config/notifications.ini"
if [ -f "$NOTIFICATIONS_INI" ] && ! grep -q "^docker_backup=" "$NOTIFICATIONS_INI"; then
    echo "Adding docker_backup to $NOTIFICATIONS_INI..."
    printf '\ndocker_backup=yes\n' >> "$NOTIFICATIONS_INI"
fi
