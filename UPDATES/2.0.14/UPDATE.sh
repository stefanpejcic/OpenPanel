#!/bin/bash

# fresh installs on 2.0.12 and 2.0.13 still got the cron with opencli waf-update, so the monthly OWASP CRS update never ran
CRON_FILE="/etc/cron.d/openpanel"
if [ -f "$CRON_FILE" ] && grep -q "opencli waf-update" "$CRON_FILE"; then
    echo "Fixing OWASP CRS update cron in $CRON_FILE..."
    sed -i 's#opencli waf-update#opencli waf update#g' "$CRON_FILE"
fi
