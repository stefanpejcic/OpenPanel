#!/bin/bash

# docker_backup notification is new in 2.0.16, add it so it can be toggled from OpenAdmin > Settings > Notifications
NOTIFICATIONS_INI="/etc/openpanel/openadmin/config/notifications.ini"
if [ -f "$NOTIFICATIONS_INI" ] && ! grep -q "^docker_backup=" "$NOTIFICATIONS_INI"; then
    echo "Adding docker_backup to $NOTIFICATIONS_INI..."
    printf '\ndocker_backup=yes\n' >> "$NOTIFICATIONS_INI"
fi

# backups: postgres gets a dump hook instead of copying its live data dir, and db dumps get listed in a 00-meta/ folder at the start of every archive so the panel can index backups without downloading them
CONFIG_RAW="https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main"
for script in mysql/scripts/dump.sh postgres/scripts/dump.sh backups/scripts/files-list.sh; do
    tmp=$(mktemp)
    if wget --timeout=15 --tries=2 -q -O "$tmp" "$CONFIG_RAW/$script" && [ -s "$tmp" ]; then
        echo "Updating /etc/openpanel/$script..."
        mkdir -p "$(dirname "/etc/openpanel/$script")"
        # written in place, running db containers bind-mount this file and would keep the old copy after a mv
        cat "$tmp" > "/etc/openpanel/$script"
        rm -f "$tmp"
        chmod 755 "/etc/openpanel/$script"
    else
        rm -f "$tmp"
        echo "Failed to download $CONFIG_RAW/$script"
    fi
done

migrate_compose_backups() {
    local f="$1"
    [ -f "$f" ] || return 0
    grep -q "backup_meta" "$f" && return 0
    grep -q "pg_data:/backup/postgres:ro" "$f" || return 0
    [ -f /etc/openpanel/postgres/scripts/dump.sh ] || return 0
    echo "Adding postgres dumps and backup_meta volume to $f..."
    cp -a "$f" "$f.bak-2.0.16"
    sed -i \
        -e 's|^\(      - mysql_dumps:/tmp/dumps/\) *$|\1\n      - backup_meta:/tmp/backup_meta/|' \
        -e "s|^\(      - docker-volume-backup.archive-post=/bin/sh -c 'rm -rf /tmp/dumps/\*\)'$|\1 /tmp/backup_meta/mysql.json'|" \
        -e "s|^\(    container_name: postgres\)$|\1\n    labels:\n      - docker-volume-backup.archive-pre=/bin/sh -c '/dump.sh'\n      - docker-volume-backup.archive-post=/bin/sh -c 'rm -rf /tmp/dumps/* /tmp/backup_meta/postgres.json'|" \
        -e 's|^\(      - pg_data:/var/lib/postgresql/data\)$|\1\n      - pg_dumps:/tmp/dumps/\n      - backup_meta:/tmp/backup_meta/\n      - /etc/openpanel/postgres/scripts/dump.sh:/dump.sh:ro|' \
        -e 's|^      - pg_data:/backup/postgres:ro$|      - pg_dumps:/backup/postgres:ro\n      - backup_meta:/backup/00-meta:ro|' \
        -e '/^  mysql_dumps:$/i\  pg_dumps:\n    driver: local\n    labels:\n      description: "This volume holds the .sql dumps of postgres databases during backup."\n      purpose: "storage"\n  backup_meta:\n    driver: local\n    labels:\n      description: "This volume holds the list of dumped databases that goes first in every backup archive."\n      purpose: "storage"' \
        "$f"
}

# the backup container lists the website files itself, offen runs this hook in it like it does dump.sh in the db containers
migrate_compose_files_list() {
    local f="$1"
    [ -f "$f" ] || return 0
    grep -q "files-list.sh" "$f" && return 0
    grep -q "backup_meta:/backup/00-meta:ro" "$f" || return 0
    grep -q "^    image: offen/docker-volume-backup:v2$" "$f" || return 0
    [ -f /etc/openpanel/backups/scripts/files-list.sh ] || return 0
    echo "Adding the website file list hook to $f..."
    [ -f "$f.bak-2.0.16" ] || cp -a "$f" "$f.bak-2.0.16"
    sed -i \
        -e 's|^      - backup_meta:/backup/00-meta:ro$|&\n      - backup_meta:/tmp/backup_meta/\n      - /etc/openpanel/backups/scripts/files-list.sh:/files-list.sh:ro|' \
        -e "s|^    image: offen/docker-volume-backup:v2$|&\n    labels:\n      - docker-volume-backup.archive-pre=/bin/sh /files-list.sh\n      - docker-volume-backup.archive-post=/bin/sh -c 'rm -f /tmp/backup_meta/files.tsv /tmp/backup_meta/mail.tsv /tmp/backup_meta/vhosts.tsv /tmp/backup_meta/dns.tsv'|" \
        "$f"
}

# podman --env-file keeps quotes in values, so offen got 'true' for BACKUP_FILENAME_EXPAND and every backup failed
# full account backups from opencli user-backup (.tar.gz in html_data/_backups) shouldn't go into every volume backup
migrate_backup_env() {
    local f="$1"
    [ -f "$f" ] || return 0
    if grep -qE "^(# )?[A-Z0-9_]+=(['\"]).*['\"]$" "$f"; then
        echo "Removing quotes from values in $f..."
        sed -i -E "s/^((# )?[A-Z0-9_]+=)(['\"])(.*)\3$/\1\4/" "$f"
    fi
    if grep -qxF 'BACKUP_EXCLUDE_REGEXP=\.log$' "$f"; then
        echo "Excluding .tar.gz files in $f..."
        sed -i 's#^BACKUP_EXCLUDE_REGEXP=.*#BACKUP_EXCLUDE_REGEXP=\\.(log|tar\\.gz)$#' "$f"
    fi
}

migrate_compose_backups /etc/openpanel/docker/compose/1.0/docker-compose.yml
migrate_compose_files_list /etc/openpanel/docker/compose/1.0/docker-compose.yml
migrate_backup_env /etc/openpanel/backups/backup.env

source /usr/local/opencli/db.sh
contexts=$(mariadb --defaults-extra-file="$config_file" -D "$mysql_database" -N -e "SELECT DISTINCT server FROM users;" 2>/dev/null)
for ctx in $contexts; do
    case "$ctx" in ""|default|root) continue ;; esac
    migrate_compose_backups "/home/$ctx/docker-compose.yml"
    migrate_compose_files_list "/home/$ctx/docker-compose.yml"
    migrate_backup_env "/home/$ctx/backup.env"
done

# mail_data only has symlinks to the real mail dirs, which the backup container stored as links without any mail, so each domain's mail dir gets mounted at /backup/emails/<domain>, and its dns zone at /backup/dns/<domain>.zone
if [ -f /usr/local/opencli/lib/backups.sh ]; then
    . /usr/local/opencli/lib/backups.sh
    mariadb --defaults-extra-file="$config_file" -D "$mysql_database" -N -e "SELECT u.username, u.server, d.domain_url FROM domains d JOIN users u ON u.id = d.user_id;" 2>/dev/null \
        | while IFS=$'\t' read -r username ctx domain; do
            case "$ctx" in ""|default|root) continue ;; esac
            dir=$(backup_mail_dir "$ctx" "$domain")
            if [ -d "$dir" ] && ! grep -qF ":/backup/emails/$domain:ro" "/home/$ctx/docker-compose.yml" 2>/dev/null; then
                echo "Adding mail backup mount for $domain to /home/$ctx/docker-compose.yml..."
                backup_mount_add "$ctx" "$dir" "/backup/emails/$domain" && echo "$username $ctx" >> /tmp/op_2016_mail_reload
            fi
            zone="/etc/bind/zones/$domain.zone"
            if [ -f "$zone" ] && ! grep -qF ":/backup/dns/$domain.zone:ro" "/home/$ctx/docker-compose.yml" 2>/dev/null; then
                echo "Adding dns zone backup mount for $domain to /home/$ctx/docker-compose.yml..."
                backup_mount_add "$ctx" "$zone" "/backup/dns/$domain.zone" && echo "$username $ctx" >> /tmp/op_2016_mail_reload
            fi
        done
    # one reload per account, not per domain
    sort -u /tmp/op_2016_mail_reload 2>/dev/null | while read -r username ctx; do backup_service_reload "$username" "$ctx"; done
    rm -f /tmp/op_2016_mail_reload
else
    echo "opencli lib/backups.sh missing, mail backup mounts not added - run opencli update first."
fi
# running db containers pick up the new volume on their next recreate, opencli docker-backup does that before it backs the user up
