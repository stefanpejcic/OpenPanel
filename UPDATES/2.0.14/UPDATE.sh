#!/bin/bash

# fresh installs on 2.0.12 and 2.0.13 still got the cron with opencli waf-update, so the monthly OWASP CRS update never ran
CRON_FILE="/etc/cron.d/openpanel"
if [ -f "$CRON_FILE" ] && grep -q "opencli waf-update" "$CRON_FILE"; then
    echo "Fixing OWASP CRS update cron in $CRON_FILE..."
    sed -i 's#opencli waf-update#opencli waf update#g' "$CRON_FILE"
fi

# new wp manager rules disable_wp_admin and mitigate_spam_logins, plus fixes for env_files and author_scan, https://github.com/stefanpejcic/OpenPanel/issues/1167
WP_RULES="/etc/openpanel/caddy/templates/wp.rules"
echo "Replacing WP Manager security rules in $WP_RULES..."
mkdir -p "$(dirname "$WP_RULES")"
if wget --timeout=15 --tries=3 -q -O "$WP_RULES.new" https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/main/caddy/templates/wp.rules && grep -q "^(wp_manager_" "$WP_RULES.new"; then
    [ -f "$WP_RULES" ] && cp "$WP_RULES" "$WP_RULES.bak-2.0.14"
    cat "$WP_RULES.new" > "$WP_RULES"
    if podman ps --format '{{.Names}}' | grep -qx caddy; then
        # put the old file back if caddy doesn't accept the new one
        if podman exec caddy caddy validate --config /etc/caddy/Caddyfile >/dev/null 2>&1; then
            podman exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1
        elif [ -f "$WP_RULES.bak-2.0.14" ]; then
            echo "Caddy config is not valid with the new rules, restoring the old $WP_RULES"
            cat "$WP_RULES.bak-2.0.14" > "$WP_RULES"
        fi
    fi
else
    echo "Failed to download wp.rules, keeping the existing file"
fi
rm -f "$WP_RULES.new"

# docker-mailserver v16 switched to Dovecot 2.4, which ignores the old quota_rule our accounts.sh wrote, so mailbox quotas were never enforced
ACCOUNTS_SH="/usr/local/mail/openmail/openpanel/accounts.sh"
if [ -f "$ACCOUNTS_SH" ]; then
    echo "Updating email quota handling in $ACCOUNTS_SH..."
    if wget --timeout=15 --tries=3 -q -O "$ACCOUNTS_SH.new" https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/openpanel/accounts.sh && grep -q "userdb_quota_storage_size" "$ACCOUNTS_SH.new" && bash -n "$ACCOUNTS_SH.new"; then
        cp "$ACCOUNTS_SH" "$ACCOUNTS_SH.bak-2.0.14"
        # cat keeps the inode, the file is bind-mounted into the mailserver container
        cat "$ACCOUNTS_SH.new" > "$ACCOUNTS_SH"
        # a container restart skips docker-mailserver's account setup, so reload the change detector with the new file and make it think dovecot-quotas.cf changed, it then rebuilds every account's dovecot entry and existing quotas apply
        if podman ps --format '{{.Names}}' | grep -qx openadmin_mailserver; then
            podman exec openadmin_mailserver supervisorctl restart changedetector >/dev/null 2>&1
            podman exec openadmin_mailserver sed -i 's#^[0-9a-f]*  /tmp/docker-mailserver/dovecot-quotas.cf$#0  /tmp/docker-mailserver/dovecot-quotas.cf#' /tmp/docker-mailserver-config-chksum
        fi
    else
        echo "Failed to download a fixed accounts.sh, keeping the existing file"
    fi
    rm -f "$ACCOUNTS_SH.new"
fi

# enforce <context>_ prefix on new mysql dbs/users, new in 2.0.14 and off by default
OPENPANEL_CONFIG="/etc/openpanel/openpanel/conf/openpanel.config"
if [ -f "$OPENPANEL_CONFIG" ] && ! grep -q "^mysql_enforce_username_prefix=" "$OPENPANEL_CONFIG"; then
    echo "Adding mysql_enforce_username_prefix to $OPENPANEL_CONFIG..."
    if grep -q "^\[PANEL\]" "$OPENPANEL_CONFIG"; then
        sed -i '/^\[PANEL\]/a mysql_enforce_username_prefix=no' "$OPENPANEL_CONFIG"
    else
        printf '\n[PANEL]\nmysql_enforce_username_prefix=no\n' >> "$OPENPANEL_CONFIG"
    fi
fi
