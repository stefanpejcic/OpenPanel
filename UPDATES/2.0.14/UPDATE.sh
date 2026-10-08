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

# new MySQL settings with usage based suggestions on the MySQL Configuration page, new in 2.0.14
MYSQL_KEYS="/etc/openpanel/mysql/keys.txt"
if [ -f "$MYSQL_KEYS" ]; then
    for key in table_open_cache table_definition_cache innodb_flush_log_at_trx_commit; do
        if ! grep -qx "$key" "$MYSQL_KEYS"; then
            echo "Adding $key to $MYSQL_KEYS..."
            [ -n "$(tail -c1 "$MYSQL_KEYS")" ] && echo >> "$MYSQL_KEYS"
            echo "$key" >> "$MYSQL_KEYS"
        fi
    done
fi

# nginx and openresty saw every visitor as the container's own 10.x address, rootless podman forwards the published port from inside the container network and only 172.x was trusted for X-Forwarded-For
realip_main_conf() {
    local f="$1"
    [ -f "$f" ] || return 1
    if grep -qE "set_real_ip_from[[:space:]]+10\.0\.0\.0/8" "$f"; then
        return 1
    elif grep -qE "^[[:space:]]*set_real_ip_from[[:space:]]+172\.16\.0\.0/12;" "$f"; then
        sed -i -E 's#^([[:space:]]*)set_real_ip_from([[:space:]]+)172\.16\.0\.0/12;#\1set_real_ip_from\210.0.0.0/8;\n&#' "$f"
    elif ! grep -q "set_real_ip_from" "$f" && grep -qE "^[[:space:]]*http[[:space:]]*\{" "$f"; then
        sed -i -E '0,/^[[:space:]]*http[[:space:]]*\{/s##&\n    set_real_ip_from  10.0.0.0/8;\n    set_real_ip_from  172.16.0.0/12;\n    set_real_ip_from  127.0.0.1;\n    real_ip_header    X-Forwarded-For;#' "$f"
    else
        return 1
    fi
}
realip_vhost() {
    local f="$1"
    grep -qE "set_real_ip_from[[:space:]]+172\.17\.0\.0/16;" "$f" && ! grep -qE "set_real_ip_from[[:space:]]+10\.0\.0\.0/8" "$f" || return 1
    sed -i -E 's#^([[:space:]]*)set_real_ip_from([[:space:]]+)172\.17\.0\.0/16;#\1set_real_ip_from\210.0.0.0/8;\n\1set_real_ip_from\2172.16.0.0/12;\n\1set_real_ip_from\2127.0.0.1;#' "$f"
}

echo "Trusting the container network for the visitor IP in nginx and openresty templates..."
for f in /etc/openpanel/nginx/user-nginx.conf /etc/openpanel/nginx/nginx.conf /etc/openpanel/openresty/nginx.conf; do
    realip_main_conf "$f" && echo "  updated $f"
done
for f in /etc/openpanel/nginx/vhosts/1.1/docker_nginx_domain.conf /etc/openpanel/nginx/vhosts/1.1/docker_openresty_domain.conf; do
    [ -f "$f" ] && realip_vhost "$f" && echo "  updated $f"
done

for home in /home/*/; do
    ctx=$(basename "$home")
    [ -f "$home.env" ] || continue
    changed=()
    for f in "${home}nginx.conf" "${home}openresty.conf"; do
        [ -f "$f" ] || continue
        # backup only kept when the file changes, so a rerun doesn't drop the first run's one
        cp -p "$f" "$f.realip-tmp"
        if realip_main_conf "$f"; then mv -f "$f.realip-tmp" "$f.bak-2.0.14"; changed+=("$f"); else rm -f "$f.realip-tmp"; fi
    done
    for f in "${home}docker-data/volumes/${ctx}_webserver_data/_data/"*.conf; do
        [ -f "$f" ] || continue
        # backup only kept when the file changes, so a rerun doesn't drop the first run's one
        cp -p "$f" "$f.realip-tmp"
        if realip_vhost "$f"; then mv -f "$f.realip-tmp" "$f.bak-2.0.14"; changed+=("$f"); else rm -f "$f.realip-tmp"; fi
    done
    [ ${#changed[@]} -eq 0 ] && continue
    echo "Fixed visitor IP for $ctx in ${#changed[@]} file(s)"

    ws=$(grep -E '^WEB_SERVER=' "${home}.env" | cut -d= -f2 | tr -d '"')
    [ "$ws" = nginx ] || [ "$ws" = openresty ] || continue
    uid=$(stat -c '%u' "$home")
    sock="unix:///run/user/${uid}/podman/podman.sock"
    [ "$(CONTAINER_HOST=$sock timeout 10 podman --remote inspect "$ws" --format '{{.State.Status}}' 2>/dev/null)" = running ] || continue
    if CONTAINER_HOST=$sock timeout 20 podman --remote exec "$ws" nginx -t >/dev/null 2>&1; then
        CONTAINER_HOST=$sock timeout 20 podman --remote exec "$ws" nginx -s reload >/dev/null 2>&1
    else
        # leave the user exactly as before if the new config doesn't pass
        echo "  $ws config test failed for $ctx, restoring the previous files"
        for f in "${changed[@]}"; do mv -f "$f.bak-2.0.14" "$f"; done
    fi
done
