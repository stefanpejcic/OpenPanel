#!/bin/bash

CRON_FILE="/etc/cron.d/openpanel"
if [ -f "$CRON_FILE" ]; then
    # opencli waf-update doesn't exist, so the monthly OWASP CRS update never ran
    if grep -q "opencli waf-update" "$CRON_FILE"; then
        echo "Fixing OWASP CRS update cron in $CRON_FILE..."
        sed -i 's#opencli waf-update#opencli waf update#g' "$CRON_FILE"
    fi

    # saving OpenAdmin > Scheduled Actions rewrote sentinel to a removed script, which stopped sentinel from running
    if grep -q "/bin/bash /usr/local/admin/service/notifications.sh" "$CRON_FILE"; then
        echo "Fixing sentinel crons in $CRON_FILE..."
        sed -i 's#/bin/bash /usr/local/admin/service/notifications.sh#/usr/local/bin/opencli sentinel#g' "$CRON_FILE"
    fi
fi
