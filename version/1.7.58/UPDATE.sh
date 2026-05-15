#!/bin/bash

OPENMAIL_DIR="/usr/local/mail/openmail"

wget -q -O "/etc/openpanel/caddy/check.conf" https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/caddy/check.conf

# create symlinks for user files
for env_file in /home/*/.env; do
    [ -e "$env_file" ] || continue
    USERNAME=$(basename "$(dirname "$env_file")")
    HOME_DIR="/home/$USERNAME"
    TARGET_LINK="$HOME_DIR/files"
    SOURCE_DIR="$HOME_DIR/docker-data/volumes/${USERNAME}_html_data/_data/"
    if [ ! -e "$TARGET_LINK" ]; then
        ln -sfn "$SOURCE_DIR" "$TARGET_LINK"
    fi
done

if [ -d "$OPENMAIL_DIR" ]; then

    sed -i 's|$OPENMAIL_DIR/:$OPENMAIL_DIR/:ro|$OPENMAIL_DIR/:$OPENMAIL_DIR/|g' "/root/docker-compose.yml"

    if [ ! -d "$OPENMAIL_DIR/openpanel" ]; then
        mkdir -p $OPENMAIL_DIR/openpanel
        wget -O $OPENMAIL_DIR/openpanel/utils.sh https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/openpanel/utils.sh
        wget -O $OPENMAIL_DIR/openpanel/accounts.sh https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/openpanel/accounts.sh
    fi

    if [ ! -d "$OPENMAIL_DIR/roundcube-plugins" ]; then
        mkdir -p $OPENMAIL_DIR/roundcube-plugins
        wget -q -O "$OPENMAIL_DIR/roundcube-plugins/dovecot_impersonate/dovecot_impersonate.php" https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/roundcube-plugins/dovecot_impersonate/dovecot_impersonate.php
        wget -q -O "$OPENMAIL_DIR/roundcube-plugins/autologin.py" https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/roundcube-plugins/autologin.py
    fi

    if [ -f "$OPENMAIL_DIR/compose.yml" ]; then
        COMPOSE="$OPENMAIL_DIR/compose.yml"

        MISSING_ACCOUNTS=true
        MISSING_UTILS=true
        MISSING_IMPERSONATE=true
        MISSING_AUTOLOGIN=true

        # FIX: koristimo grep -F -- da spriječimo tumačenje - kao opcije
        grep -qF -- './openpanel/accounts.sh:/usr/local/bin/helpers/accounts.sh:ro' "$COMPOSE" && MISSING_ACCOUNTS=false
        grep -qF -- './openpanel/utils.sh:/usr/local/bin/helpers/utils.sh:ro' "$COMPOSE" && MISSING_UTILS=false
        grep -qF -- './roundcube-plugins/dovecot_impersonate:/var/www/html/plugins/dovecot_impersonate:ro' "$COMPOSE" && MISSING_IMPERSONATE=false
        grep -qF -- './roundcube-plugins/autologin.php:/var/www/html/public_html/autologin.php:ro' "$COMPOSE" && MISSING_AUTOLOGIN=false

        # --- mailserver volumes ---
        if $MISSING_ACCOUNTS || $MISSING_UTILS; then
            MAILSERVER_LINE=$(grep -n '^\s\{2\}mailserver:' "$COMPOSE" | cut -d: -f1)
            NEXT_SERVICE_LINE=$(awk "NR > $MAILSERVER_LINE && /^  [a-z]/ { print NR; exit }" "$COMPOSE")

            VOLUMES_LINE=$(awk "NR > $MAILSERVER_LINE && NR < ${NEXT_SERVICE_LINE:-9999} && /^\s*volumes:/ { print NR; exit }" "$COMPOSE")

            if [ -z "$VOLUMES_LINE" ]; then
                echo "ERROR: Could not find volumes: section under mailserver service" >&2
                exit 1
            fi

            VOLUMES_END_LINE=$(awk "NR > $VOLUMES_LINE && /^\s{4}[a-z_]/ { print NR; exit }" "$COMPOSE")
            if [ -z "$VOLUMES_END_LINE" ]; then
                VOLUMES_END_LINE="${NEXT_SERVICE_LINE:-9999}"
            fi

            LAST_VOLUME_LINE=$(awk "NR > $VOLUMES_LINE && NR < $VOLUMES_END_LINE && /^\s*- / { last=NR } END { print last+0 }" "$COMPOSE")

            if [ "$LAST_VOLUME_LINE" -eq 0 ]; then
                echo "ERROR: Could not find any volume entries under mailserver volumes:" >&2
                exit 1
            fi

            # FIX: koristimo privremeni fajl umjesto sed inline append da izbjegnemo problem sa - u stringu
            INSERT=""
            $MISSING_ACCOUNTS && INSERT="${INSERT}      - ./openpanel/accounts.sh:/usr/local/bin/helpers/accounts.sh:ro\n"
            $MISSING_UTILS    && INSERT="${INSERT}      - ./openpanel/utils.sh:/usr/local/bin/helpers/utils.sh:ro\n"
            INSERT="${INSERT%\\n}"

            python3 -c "
import sys
lines = open('$COMPOSE').readlines()
insert = '''$(printf '%b' "$INSERT")'''
insert_lines = [l + '\n' for l in insert.split('\n') if l]
lines = lines[:$LAST_VOLUME_LINE] + insert_lines + lines[$LAST_VOLUME_LINE:]
open('$COMPOSE', 'w').writelines(lines)
"
            echo "Added volume mounts to mailserver in $COMPOSE"
            echo "Restarting mailserver to use the overwrites for storage.."
            cd $OPENMAIL_DIR && docker --context=default compose down mailserver && docker --context=default compose up -d mailserver
            echo "done"
        fi

        # --- roundcube volumes ---
        if $MISSING_IMPERSONATE || $MISSING_AUTOLOGIN; then
            RC_LINE=$(grep -n '^\s\{2\}roundcube:' "$COMPOSE" | cut -d: -f1)
            RC_NEXT_SERVICE_LINE=$(awk "NR > $RC_LINE && /^  [a-z]/ { print NR; exit }" "$COMPOSE")

            RC_VOLUMES_LINE=$(awk "NR > $RC_LINE && NR < ${RC_NEXT_SERVICE_LINE:-9999} && /^\s*volumes:/ { print NR; exit }" "$COMPOSE")

            if [ -z "$RC_VOLUMES_LINE" ]; then
                echo "ERROR: Could not find volumes: section under roundcube service" >&2
                exit 1
            fi

            RC_VOLUMES_END_LINE=$(awk "NR > $RC_VOLUMES_LINE && /^\s{4}[a-z_]/ { print NR; exit }" "$COMPOSE")
            if [ -z "$RC_VOLUMES_END_LINE" ]; then
                RC_VOLUMES_END_LINE="${RC_NEXT_SERVICE_LINE:-9999}"
            fi

            RC_LAST_VOLUME_LINE=$(awk "NR > $RC_VOLUMES_LINE && NR < $RC_VOLUMES_END_LINE && /^\s*- / { last=NR } END { print last+0 }" "$COMPOSE")

            if [ "$RC_LAST_VOLUME_LINE" -eq 0 ]; then
                echo "ERROR: Could not find any volume entries under roundcube volumes:" >&2
                exit 1
            fi

            RC_INSERT=""
            $MISSING_IMPERSONATE && RC_INSERT="${RC_INSERT}      - ./roundcube-plugins/dovecot_impersonate:/var/www/html/plugins/dovecot_impersonate:ro\n"
            $MISSING_AUTOLOGIN   && RC_INSERT="${RC_INSERT}      - ./roundcube-plugins/autologin.php:/var/www/html/public_html/autologin.php:ro\n"
            RC_INSERT="${RC_INSERT%\\n}"

            python3 -c "
import sys
lines = open('$COMPOSE').readlines()
insert = '''$(printf '%b' "$RC_INSERT")'''
insert_lines = [l + '\n' for l in insert.split('\n') if l]
lines = lines[:$RC_LAST_VOLUME_LINE] + insert_lines + lines[$RC_LAST_VOLUME_LINE:]
open('$COMPOSE', 'w').writelines(lines)
"
            echo "Added volume mounts to roundcube in $COMPOSE"
        fi

        # --- roundcube ROUNDCUBEMAIL_PLUGINS env var ---
        if $MISSING_IMPERSONATE; then
            PLUGINS_LINE=$(grep -n 'ROUNDCUBEMAIL_PLUGINS=' "$COMPOSE" | cut -d: -f1)
            if [ -n "$PLUGINS_LINE" ]; then
                sed -i "${PLUGINS_LINE}s/ROUNDCUBEMAIL_PLUGINS=\(.*\)/ROUNDCUBEMAIL_PLUGINS=\1,dovecot_impersonate/" "$COMPOSE"
                echo "Added dovecot_impersonate to ROUNDCUBEMAIL_PLUGINS in $COMPOSE"
            else
                echo "WARNING: ROUNDCUBEMAIL_PLUGINS= line not found in $COMPOSE" >&2
            fi
        fi

        if $MISSING_IMPERSONATE || $MISSING_AUTOLOGIN; then
            echo "Restarting roundcube to apply changes.."
            cd $OPENMAIL_DIR && docker --context=default compose down roundcube && docker --context=default compose up -d roundcube
            echo "done"
        fi
    fi
fi


# ADD WWW NETWORK FOR BACKUP SERVICE
COMPOSE="/etc/openpanel/docker/compose/1.0/docker-compose.yml"

if [ -f "$COMPOSE" ]; then
    BACKUP_LINE=$(grep -n '^\s\{2\}backup:' "$COMPOSE" | cut -d: -f1)
    NEXT_SERVICE_LINE=$(awk "NR > $BACKUP_LINE && /^  [a-z]/ { print NR; exit }" "$COMPOSE")

    HAS_WWW=$(awk "NR > $BACKUP_LINE && NR < ${NEXT_SERVICE_LINE:-9999} && /^\s*- www$/ { found=1 } END { print found+0 }" "$COMPOSE")

    if [ "$HAS_WWW" -eq 0 ]; then
        BACKUP_NETWORKS_LINE=$(awk "NR > $BACKUP_LINE && NR < ${NEXT_SERVICE_LINE:-9999} && /^\s*networks:/ { print NR; exit }" "$COMPOSE")

        if [ -z "$BACKUP_NETWORKS_LINE" ]; then
            echo "ERROR: Could not find networks: section under backup service" >&2
            exit 1
        fi

        LAST_NETWORK_LINE=$(awk "NR > $BACKUP_NETWORKS_LINE && NR < ${NEXT_SERVICE_LINE:-9999} && /^\s*- / { last=NR } END { print last+0 }" "$COMPOSE")

        if [ "$LAST_NETWORK_LINE" -eq 0 ]; then
            echo "ERROR: Could not find any network entries under backup" >&2
            exit 1
        fi

        sed -i "${LAST_NETWORK_LINE}a\\      - www" "$COMPOSE"
        echo "Added 'www' network to backup service in $COMPOSE"
    else
        echo "backup service already has 'www' network, no changes needed."
    fi
fi
