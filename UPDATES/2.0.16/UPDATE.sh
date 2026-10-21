#!/bin/bash

# docker_backup notification is new in 2.0.16, add it so it can be toggled from OpenAdmin > Settings > Notifications
NOTIFICATIONS_INI="/etc/openpanel/openadmin/config/notifications.ini"
if [ -f "$NOTIFICATIONS_INI" ] && ! grep -q "^docker_backup=" "$NOTIFICATIONS_INI"; then
    echo "Adding docker_backup to $NOTIFICATIONS_INI..."
    printf '\ndocker_backup=yes\n' >> "$NOTIFICATIONS_INI"
fi

# backups: postgres gets a dump hook instead of copying its live data dir, and db dumps get listed in a 00-meta/ folder at the start of every archive so the panel can index backups without downloading them
CONFIG_RAW="https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main"
for script in mysql/scripts/dump.sh postgres/scripts/dump.sh backups/scripts/files-list.sh php/fpm-tuning.sh systemd/iocost.sh; do
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

# ENABLE_TUNING_FPM sized php-fpm from host cores (18 cores -> 90 children in a 1G container), so one busy site thrashed and took the server down; fpm-tuning.sh sizes it from the container's own memory limit instead
migrate_compose_fpm_tuning() {
    local f="$1"
    [ -f "$f" ] || return 0
    grep -q "^      - ENABLE_TUNING_FPM=1$" "$f" || return 0
    [ -x /etc/openpanel/php/fpm-tuning.sh ] || return 0
    echo "Replacing ENABLE_TUNING_FPM with container-sized php-fpm tuning in $f..."
    [ -f "$f.bak-2.0.16" ] || cp -a "$f" "$f.bak-2.0.16"
    sed -i \
        -e '/^      - ENABLE_TUNING_FPM=1$/d' \
        -e 's|^\(      - /etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp:ro\) *$|\1\n      - /etc/openpanel/php/fpm-tuning.sh:/etc/hooks/bootstrap/05-openpanel-fpm-tuning:ro|' \
        "$f"
}

migrate_compose_backups /etc/openpanel/docker/compose/1.0/docker-compose.yml
migrate_compose_files_list /etc/openpanel/docker/compose/1.0/docker-compose.yml
migrate_backup_env /etc/openpanel/backups/backup.env
migrate_compose_fpm_tuning /etc/openpanel/docker/compose/1.0/docker-compose.yml

source /usr/local/opencli/db.sh
contexts=$(mariadb --defaults-extra-file="$config_file" -D "$mysql_database" -N -e "SELECT DISTINCT server FROM users;" 2>/dev/null)
for ctx in $contexts; do
    case "$ctx" in ""|default|root) continue ;; esac
    migrate_compose_backups "/home/$ctx/docker-compose.yml"
    migrate_compose_files_list "/home/$ctx/docker-compose.yml"
    migrate_backup_env "/home/$ctx/backup.env"
    migrate_compose_fpm_tuning "/home/$ctx/docker-compose.yml"
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
# everything but the website files goes in front of html/ in the archive as backup/01-<name>, so restoring a database, vhosts or mail stops reading early
# runs after the mail mounts above, opencli older than 2.0.16 only finds the plain mail_data line
migrate_compose_early_dirs() {
    local f="$1"
    [ -f "$f" ] || return 1
    grep -qE "^      - (mysql_dumps|pg_dumps|mail_data|webserver_data):/backup/(mysql|postgres|mail|vhosts):ro$" "$f" || return 1
    echo "Moving db dumps, mail and vhosts in front of html/ in the backup archive in $f..."
    [ -f "$f.bak-2.0.16" ] || cp -a "$f" "$f.bak-2.0.16"
    sed -i -E 's#^      - (mysql_dumps|pg_dumps|mail_data|webserver_data):/backup/(mysql|postgres|mail|vhosts):ro$#      - \1:/backup/01-\2:ro#' "$f"
}

# BACKUP_SOURCES pointing at one of the moved folders has to follow it
migrate_backup_env_early_dirs() {
    local f="$1"
    [ -f "$f" ] || return 0
    grep -qE "^BACKUP_SOURCES=/backup/(mysql|postgres|mail|vhosts)/?$" "$f" || return 0
    echo "Pointing BACKUP_SOURCES at the moved folder in $f..."
    sed -i -E 's#^BACKUP_SOURCES=/backup/(mysql|postgres|mail|vhosts)/?$#BACKUP_SOURCES=/backup/01-\1#' "$f"
}

migrate_compose_early_dirs /etc/openpanel/docker/compose/1.0/docker-compose.yml
migrate_backup_env_early_dirs /etc/openpanel/backups/backup.env
mariadb --defaults-extra-file="$config_file" -D "$mysql_database" -N -e "SELECT username, server FROM users;" 2>/dev/null \
    | while IFS=$'\t' read -r username ctx; do
        case "$ctx" in ""|default|root) continue ;; esac
        migrate_backup_env_early_dirs "/home/$ctx/backup.env"
        # a user-scheduled backup container keeps running with the old mounts until it is recreated
        if migrate_compose_early_dirs "/home/$ctx/docker-compose.yml" && declare -F backup_service_reload >/dev/null; then
            backup_service_reload "$username" "$ctx"
        fi
    done

# running php-fpm containers keep their old pool size until they are recreated, a plain restart keeps the old mounts
# running db containers pick up the new volume on their next recreate, opencli docker-backup does that before it backs the user up

# domains-unsuspend used to copy the vhost back and leave it in suspended_domains, which made the next domains-suspend think the domain was already suspended
for copy in /etc/openpanel/caddy/suspended_domains/*.conf; do
    [ -f "$copy" ] || continue
    live="/etc/openpanel/caddy/domains/$(basename "$copy")"
    if [ -f "$live" ] && ! grep -qE '^# suspended_by:|suspended_domain\.html' "$live"; then
        echo "Removing stale suspended copy $copy..."
        rm -f "$copy"
    fi
done

# suspension reason pages are new in 2.0.16, only fetch missing ones so admin edits are kept
mkdir -p /etc/openpanel/caddy/templates/suspended
for reason in dmca billing phishing; do
    dest="/etc/openpanel/caddy/templates/suspended/$reason.html"
    [ -f "$dest" ] && continue
    tmp=$(mktemp)
    if wget --timeout=15 --tries=2 -q -O "$tmp" "$CONFIG_RAW/caddy/templates/suspended/$reason.html" && [ -s "$tmp" ]; then
        echo "Adding suspension reason page $dest..."
        mv "$tmp" "$dest"
        chmod 644 "$dest"
    else
        rm -f "$tmp"
        echo "Failed to download $CONFIG_RAW/caddy/templates/suspended/$reason.html"
    fi
done

# tenants run rootless in user.slice, sshd and OpenAdmin in system.slice, caddy/openpanel/mysql as rootful containers in machine.slice
# weights only matter under contention, so one overloaded account can't starve the host services; MemoryLow keeps their memory from being reclaimed for it
# set-property applies live and persists in /etc/systemd/system.control, no service is restarted
ram_mb=$(awk '/^MemTotal:/{print int($2/1024)}' /proc/meminfo)
low_mb=$(( ram_mb / 8 ))
(( low_mb > 2048 )) && low_mb=2048
(( low_mb < 256 )) && low_mb=256
echo "Prioritising system.slice and machine.slice over user.slice (MemoryLow=${low_mb}M)..."
systemctl set-property system.slice CPUWeight=1000 IOWeight=1000 MemoryLow=${low_mb}M
systemctl set-property machine.slice CPUWeight=1000 IOWeight=1000 MemoryLow=${low_mb}M
systemctl set-property user.slice CPUWeight=100 IOWeight=100

# systemd-oomd kills the thrashing cgroup under user.slice once its memory pressure stays above the limit, instead of the whole server freezing
if [ ! -r /proc/pressure/memory ]; then
    echo "Kernel PSI is disabled (add psi=1 to the kernel cmdline), skipping systemd-oomd."
else
    if ! systemctl cat systemd-oomd.service >/dev/null 2>&1; then
        echo "Installing systemd-oomd..."
        if command -v apt-get >/dev/null 2>&1; then
            # --no-upgrade so it never pulls a newer systemd, NEEDRESTART_SUSPEND so needrestart doesn't restart running services
            DEBIAN_FRONTEND=noninteractive NEEDRESTART_SUSPEND=1 apt-get install -y --no-upgrade systemd-oomd >/dev/null 2>&1
        elif command -v dnf >/dev/null 2>&1; then
            dnf install -y systemd-oomd >/dev/null 2>&1
        fi
    fi
    if systemctl cat systemd-oomd.service >/dev/null 2>&1; then
        echo "Enabling systemd-oomd for user.slice..."
        systemctl set-property user.slice ManagedOOMMemoryPressure=kill ManagedOOMMemoryPressureLimit=60%
        systemctl enable --now systemd-oomd >/dev/null 2>&1
    else
        echo "systemd-oomd is not available on this OS, a thrashing account will not be killed automatically."
    fi
fi

# no swap for accounts, a container at its memory limit gets oom-killed and restarted instead of swapping and saturating the disk for every account
# set-property applies live to running containers, nothing is restarted; new accounts and plan changes get it from opencli plan-apply / user-change_plan
for ctx in $contexts; do
    case "$ctx" in ""|default|root) continue ;; esac
    [ -d "/home/$ctx" ] || continue
    uid=$(stat -c '%u' "/home/$ctx")
    echo "Disabling swap for user-${uid}.slice ($ctx)..."
    systemctl set-property "user-${uid}.slice" MemorySwapMax=0
done

# accounts on plans with unlimited cpu/ram got no slice limit at all, so one busy site could take the whole server; give them the same safety ceiling opencli plan-apply now sets
if [ -f /usr/local/opencli/lib/podman.sh ] && grep -q "^unlimited_slice_ceiling()" /usr/local/opencli/lib/podman.sh && grep -q "^derive_io_weight()" /usr/local/opencli/lib/podman.sh; then
    . /usr/local/opencli/lib/podman.sh
    read -r cpu_ceiling mem_high_mb mem_max_mb < <(unlimited_slice_ceiling)
    mariadb --defaults-extra-file="$config_file" -D "$mysql_database" -N -B -e "SELECT u.server, p.cpu, p.ram FROM users u JOIN plans p ON p.id = u.plan_id;" 2>/dev/null \
        | while IFS=$'\t' read -r ctx cpu ram; do
            case "$ctx" in ""|default|root) continue ;; esac
            [ -d "/home/$ctx" ] || continue
            uid=$(stat -c '%u' "/home/$ctx")
            if awk -v v="${cpu%[Gg]}" 'BEGIN{exit !(v+0 == 0)}'; then
                echo "Setting CPU safety ceiling ${cpu_ceiling}% for user-${uid}.slice ($ctx)..."
                systemctl set-property "user-${uid}.slice" CPUQuota="${cpu_ceiling}%"
            fi
            systemctl set-property "user-${uid}.slice" IOWeight="$(derive_io_weight "$cpu")"
            if awk -v v="${ram%[Gg]}" 'BEGIN{exit !(v+0 == 0)}'; then
                echo "Setting RAM safety ceiling ${mem_high_mb}M/${mem_max_mb}M for user-${uid}.slice ($ctx)..."
                systemctl set-property "user-${uid}.slice" MemoryHigh="${mem_high_mb}M" MemoryMax="${mem_max_mb}M" TasksMax="$(derive_tasks_max $(( mem_max_mb / 1024 )) "$ctx")"
            fi
        done
else
    echo "opencli lib/podman.sh has no unlimited_slice_ceiling, safety ceilings and io weights not set - run opencli update first."
fi

# io.weight is ignored by the none/mq-deadline schedulers NVMe and VPS disks use, io.cost enforces the slice and per-plan IOWeight on any scheduler
# enabling it is a live write to /sys/fs/cgroup/io.cost.qos, nothing is restarted; the unit reapplies it at boot
unit_tmp=$(mktemp)
if wget --timeout=15 --tries=2 -q -O "$unit_tmp" "$CONFIG_RAW/systemd/openpanel-iocost.service" && [ -s "$unit_tmp" ] && [ -f /etc/openpanel/systemd/iocost.sh ]; then
    echo "Enabling io.cost so per-account IOWeight is enforced..."
    install -m 644 "$unit_tmp" /etc/systemd/system/openpanel-iocost.service
    systemctl daemon-reload
    systemctl enable --now openpanel-iocost >/dev/null 2>&1
    systemctl status openpanel-iocost --no-pager -n 5 2>/dev/null | grep -E "io.cost|Could not" || true
else
    echo "Failed to download $CONFIG_RAW/systemd/openpanel-iocost.service, io.cost not enabled."
fi
rm -f "$unit_tmp"

# rate limiting: openpanel/caddy-coraza now has the caddy-ratelimit module, new installs ship 300/min per ip and 6000/min per domain
# existing servers stay off until an admin runs opencli domains-ratelimit: the default snippet is empty, so the domain imports added below are no-ops,
# they are valid on the old caddy image too, and caddy is neither reloaded nor restarted here; the new image is only pulled for the next recreate
RL_CONF=/etc/openpanel/caddy/ratelimit.conf
if [ -f /etc/openpanel/caddy/Caddyfile ]; then
    if [ ! -f "$RL_CONF" ]; then
        echo "Adding $RL_CONF with rate limiting off..."
        printf '# opencli domains-ratelimit\n# default: 0 0\n(ratelimit) {\n}\n' > "$RL_CONF"
    fi
    if ! grep -q "import $RL_CONF" /etc/openpanel/caddy/Caddyfile; then
        echo "Adding the rate limit snippets import to /etc/openpanel/caddy/Caddyfile..."
        # written in place, caddy bind-mounts this single file and would keep seeing the old inode after sed -i, then fail its next reload on the domain imports
        sed "s|^import /etc/openpanel/caddy/domains/\*|# import rate limit snippets (opencli domains-ratelimit)\nimport $RL_CONF\n\n&|" /etc/openpanel/caddy/Caddyfile > /tmp/op_2016_caddyfile \
            && cat /tmp/op_2016_caddyfile > /etc/openpanel/caddy/Caddyfile
        rm -f /tmp/op_2016_caddyfile
    fi
    for conf in /etc/openpanel/caddy/domains/*.conf; do
        [ -f "$conf" ] || continue
        grep -qE '^[[:space:]]*(import[[:space:]]+ratelimit|# ratelimit off)' "$conf" && continue
        grep -q 'import /etc/openpanel/caddy/redirects.conf' "$conf" || continue
        domain=$(basename "$conf" .conf)
        sed -i -E "s|^([[:space:]]*)import /etc/openpanel/caddy/redirects.conf[[:space:]]*\$|&\n\n\1# rate limit (opencli domains-ratelimit)\n\1import ratelimit $domain|" "$conf"
    done
    podman pull docker.io/openpanel/caddy-coraza:latest >/dev/null 2>&1 || echo "Failed to pull the new caddy image, it will be pulled on the next caddy recreate."
fi
