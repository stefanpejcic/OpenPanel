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

# notifications.log is JSON lines from 2.0.12, start empty instead of mixing in the old format
NOTIFICATIONS_LOG="/var/log/openpanel/admin/notifications.log"
if [ -f "$NOTIFICATIONS_LOG" ]; then
    echo "Clearing $NOTIFICATIONS_LOG for the new notifications format..."
    : > "$NOTIFICATIONS_LOG"
fi

# emails users that reached the hourly email limit, new in 2.0.12
if [ -f "$CRON_FILE" ] && ! grep -q "opencli email-ratelimit --notify" "$CRON_FILE"; then
    echo "Adding hourly email limit check to $CRON_FILE..."
    echo '*/30 * * * * root /usr/local/bin/opencli email-ratelimit --notify && echo "$(date) Checked for users that reached the hourly email limit" >> /var/log/openpanel/admin/cron.log' >> "$CRON_FILE"
fi

# new notification options in 2.0.12, added with their default so the Account > Notifications page and opencli agree
NEW_NOTIFICATION_KEYS=(
    notify_passkey_change=1
    notify_passkey_change_notification_disabled=1
    notify_api_token_change=1
    notify_api_token_change_notification_disabled=1
    notify_malware_found=1
    notify_email_ratelimit=1
    notify_service_failed=1
)
for prefs in /etc/openpanel/openpanel/core/users/*/notifications.yaml; do
    [ -f "$prefs" ] || continue
    # a last line without a newline would get the first new key glued onto it
    [ -n "$(tail -c1 "$prefs")" ] && echo >> "$prefs"
    for entry in "${NEW_NOTIFICATION_KEYS[@]}"; do
        grep -q "^${entry%%=*}=" "$prefs" || echo "$entry" >> "$prefs"
    done
done
