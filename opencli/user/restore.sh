#!/bin/bash
################################################################################
# Script Name: user/restore.sh
# Description: Restores a single OpenPanel user account from a full account .tar.gz backup.
# Usage: opencli user-restore --file <ARCHIVE> [--force] [--new-username=NAME] [--quiet] [--temp-dir=<PATH> ]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 06.07.2026
# Company: openpanel.com
# Copyright (c) openpanel.com
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
################################################################################

set -o pipefail
pid=$$
start_time=$(date +%s)
MAIL_CONTAINER="openadmin_mailserver"

ARCHIVE=""
FORCE=0
QUIET=0
NEW_USERNAME=""

WARNINGS=()
DOMAINS_ADDED=0; DOMAINS_SKIPPED=0; DOMAINS_SKIPPED_LIST=()
FTP_CREATED=0; FTP_SKIPPED=0
EMAILS_ADDED=0; EMAILS_SKIPPED=0; EMAILS_SKIPPED_LIST=()
SYSTEM_USER_STATUS="created"
PLAN_STATUS="imported"
REMAPPED_UID=""
USER_ID=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --file)             ARCHIVE="$2"; shift 2 ;;
        --force)            FORCE=1; shift ;;
        --quiet)            QUIET=1; shift ;;
        --temp-dir=*)       WORK="${1#*=}"; shift ;;
        --new-username=*)   NEW_USERNAME="${1#*=}"; shift ;;
        --new-username)     NEW_USERNAME="$2"; shift 2 ;;
        *) echo "[ERROR] Unknown option: $1"; exit 1 ;;
    esac
done

[[ -z "$ARCHIVE" ]] && { echo "Usage: opencli user-restore --file <ARCHIVE> [--force] [--new-username=NAME] [--temp-dir=/home/] [--quiet]"; exit 1; }
[[ -f "$ARCHIVE" ]] || { echo "[ERROR] Archive not found: $ARCHIVE"; exit 1; }
ARCHIVE=$(realpath "$ARCHIVE")

if [[ -z "$WORK" ]]; then
    WORK=$(mktemp -d /tmp/oprestore.XXXXXX) || { echo "[ERROR] Failed to create temporary directory"; exit 1; }
else
    mkdir -p "$WORK" || { echo "[ERROR] Cannot create WORK dir: $WORK"; exit 1; }
    [[ -n "$(ls -A "$WORK" 2>/dev/null)" ]] && { echo "[ERROR] WORK dir not empty: $WORK - remove the --temp-dir= flag to use /tmp/ instead OR create a new subdirectory (e.g. --temp-dir=/home/restore_process)"; exit 1; }
fi

# DB
DB_CONFIG_FILE="/usr/local/opencli/db.sh"
[[ -f "$DB_CONFIG_FILE" ]] || { echo "[ERROR] $DB_CONFIG_FILE not found — is OpenPanel installed?"; exit 1; }
# shellcheck disable=SC1090
. "$DB_CONFIG_FILE"
mysql_q()   { mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -s -e "$1"; }
mysql_run() { mysql --defaults-extra-file="$config_file" "$@"; }

edition_key=$(grep "^key=" /etc/openpanel/openpanel/conf/openpanel.config 2>/dev/null | cut -d'=' -f2-)

get_server_ip() {
    local ip
    ip=$(curl --silent --max-time 1 -4 "https://ip.openpanel.com" 2>/dev/null || curl --silent --max-time 1 -4 "https://ifconfig.me" 2>/dev/null)
    [[ -z "$ip" ]] && ip=$(ip addr | grep 'inet ' | grep global | head -n1 | awk '{print $2}' | cut -f1 -d/)
    echo "$ip"
}

trap 'rm -rf "$WORK"' EXIT

# ── REAL RESTORE ─────────────────────────────────────────────────────────────
timestamp="$(date +'%Y-%m-%d_%H-%M-%S')"
base_name="$(basename "$ARCHIVE" .tar.gz)"
log_dir="/var/log/openpanel/admin/restores"
mkdir -p "$log_dir"
log_file="$log_dir/${base_name}_restore_${timestamp}.log"

log() {
    local msg="$1"; local ts; ts=$(date +'%Y-%m-%d %H:%M:%S')
    if [[ $QUIET -eq 1 ]]; then echo "[$ts] $msg" >> "$log_file"
    else echo "[$ts] $msg" | tee -a "$log_file"; fi
}
warn() { WARNINGS+=("$1"); log "[!] $1"; }
die()  { log "[✘] $1"; exit "${2:-1}"; }
slog() { echo "$1" | tee -a "$log_file"; }

log "Restore started  log: $log_file  (PID: $pid)"
log "Archive: $ARCHIVE"

# Disk space check — extraction target must fit the fully uncompressed archive
check_disk_space_extract() {
    log "Checking disk space for extraction ..."

    local compressed_kb
    compressed_kb=$(( $(stat -c%s "$ARCHIVE" 2>/dev/null || echo 0) / 1024 ))

    local estimated_kb gzip_size
    gzip_size=$(gzip -l "$ARCHIVE" 2>/dev/null | awk 'NR==2{print $2}')
    if [[ "$gzip_size" =~ ^[0-9]+$ && "$gzip_size" -gt 0 ]]; then
        estimated_kb=$(( gzip_size / 1024 ))
        # gzip stores the uncompressed size mod 2^32 — if it looks smaller than
        # the compressed archive itself, the field wrapped and can't be trusted.
        [[ "$estimated_kb" -lt "$compressed_kb" ]] && estimated_kb=$(( compressed_kb * 3 ))
    else
        estimated_kb=$(( compressed_kb * 3 ))
    fi

    local free_kb
    free_kb=$(df -k "$WORK" 2>/dev/null | awk 'NR==2 {print $4}')
    if [[ ! "$free_kb" =~ ^[0-9]+$ ]]; then
        warn "Cannot read free space at $WORK — disk space check skipped."
        return
    fi

    log "  Estimated extracted size (~$(( estimated_kb / 1024 )) MB)"
    log "  Free at $WORK  (~$(( free_kb / 1024 )) MB)"

    if [[ "$free_kb" -lt "$estimated_kb" ]]; then
        die "Not enough disk space to extract archive. Estimated ~$(( estimated_kb / 1024 )) MB needed, ~$(( free_kb / 1024 )) MB free at $WORK. Aborting."
    fi
}
check_disk_space_extract

# Extract
log "Extracting archive ..."
tar -C "$WORK" --acls --xattrs -xzf "$ARCHIVE" 2>>"$log_file" || die "Failed to extract archive."

[[ -f "$WORK/manifest.env" ]] || die "manifest.env missing from archive."
# shellcheck disable=SC1090
. "$WORK/manifest.env"
[[ -n "$USERNAME" && -n "$CONTEXT" ]] || die "Invalid manifest: USERNAME/CONTEXT empty."

# --new-username override
ORIG_USERNAME="$USERNAME"
ORIG_CONTEXT="$CONTEXT"
if [[ -n "$NEW_USERNAME" ]]; then
    USERNAME="$NEW_USERNAME"
    CONTEXT="$NEW_USERNAME"
    log "Username remapped: $ORIG_USERNAME → $USERNAME  |  Context: $ORIG_CONTEXT → $CONTEXT"
fi

NEW_IP=$(get_server_ip)
log "Restoring '$USERNAME' (context: $CONTEXT)  format v${BACKUP_FORMAT_VERSION}"

# Pre-flight
username_exists=$(mysql_q "SELECT COUNT(*) FROM users WHERE username='$USERNAME';")
if [[ "${username_exists:-0}" -gt 0 ]]; then
    [[ $FORCE -eq 0 ]] && die "Username '$USERNAME' already exists. Re-run with --force to overwrite."
    warn "Username '$USERNAME' exists — overwriting because --force was given."
fi
if [[ -z "$edition_key" ]]; then
    user_count=$(mysql_q "SELECT COUNT(*) FROM users;")
    [[ "${user_count:-0}" -ge 3 && "${username_exists:-0}" -eq 0 ]] && die "Community edition limit: 3 accounts. Upgrade to Enterprise for more."
fi

# ── 1) System user ───────────────────────────────────────────────────────────
restore_system_user() {
    local PW="$WORK/system/passwd.user"
    local GR="$WORK/system/group.user"
    local SH="$WORK/system/shadow.user"
    if [[ ! -s "$PW" ]]; then
        warn "No system user data in backup — creating '$CONTEXT' fresh (no password hash available, account will be locked)."
        if id "$CONTEXT" &>/dev/null; then
            REMAPPED_UID=$(id -u "$CONTEXT")
            SYSTEM_USER_STATUS="already existed"
        else
            useradd -m -d "/home/$CONTEXT" -s /bin/bash "$CONTEXT"
            usermod -L "$CONTEXT"
            REMAPPED_UID=$(id -u "$CONTEXT")
            SYSTEM_USER_STATUS="created fresh (UID $REMAPPED_UID, locked — no password restored)"
        fi
        return
    fi

    log "Restoring system user ($CONTEXT) ..."

    [[ -f "$GR" ]] && while IFS=: read -r group _ gid _; do
        [[ -z "$group" ]] && continue
        # When username changed, rename the group entry to the new context name
        local target_group="$group"
        [[ "$group" == "$ORIG_CONTEXT" ]] && target_group="$CONTEXT"
        getent group "$target_group" >/dev/null 2>&1 || groupadd -g "$gid" "$target_group" 2>/dev/null || true
    done < <(cut -d: -f1,2,3 "$GR")

    used_uids=$(getent passwd | cut -d: -f3)
    find_free_uid() { local u=$1; while echo "$used_uids" | grep -qw "$u"; do u=$((u+1)); done; echo "$u"; }

    while IFS=: read -r user uid gid comment home shell; do
        [[ "$user" != "$ORIG_CONTEXT" ]] && continue
        # Map to new context if --new-username was given
        local target_user="$CONTEXT"
        local target_home="/home/$CONTEXT"
        if id "$target_user" &>/dev/null; then
            log "System user $target_user already exists — skipping creation."
            REMAPPED_UID=$(id -u "$target_user")
            SYSTEM_USER_STATUS="already existed"
        else
            local free_uid; free_uid=$(find_free_uid "$uid")
            useradd -u "$free_uid" -g "$gid" -c "$comment" -d "$target_home" -s "$shell" "$target_user" 2>/dev/null || useradd -u "$free_uid" -c "$comment" -d "$target_home" -s "$shell" "$target_user"
            REMAPPED_UID="$free_uid"
            [[ "$free_uid" != "$uid" ]] && warn "UID remapped for $target_user: $uid → $free_uid"
            SYSTEM_USER_STATUS="created (UID $REMAPPED_UID)"
        fi
    done < <(cut -d: -f1,3,4,5,6,7 "$PW")

    [[ -f "$SH" ]] && while IFS=: read -r user hash _; do
        [[ "$user" == "$ORIG_CONTEXT" && -n "$hash" ]] && usermod -p "$hash" "$CONTEXT" 2>/dev/null || true
    done < "$SH"
}
restore_system_user

# ── 2) Home directory ────────────────────────────────────────────────────────
restore_home() {
    [[ -d "$WORK/homedir" ]] || { warn "No homedir/ in archive — skipping file restore."; return; }

    log "Checking disk space for home directory restore ..."
    local home_used_kb
    home_used_kb=$(du -sk "$WORK/homedir" 2>/dev/null | cut -f1)
    mkdir -p /home
    if [[ "$home_used_kb" =~ ^[0-9]+$ && "$home_used_kb" -gt 0 ]]; then
        local home_free_kb
        home_free_kb=$(df -k /home 2>/dev/null | awk 'NR==2 {print $4}')
        if [[ "$home_free_kb" =~ ^[0-9]+$ ]]; then
            log "  Home directory size (~$(( home_used_kb / 1024 )) MB)"
            log "  Free at /home  (~$(( home_free_kb / 1024 )) MB)"
            if [[ "$home_free_kb" -lt "$home_used_kb" ]]; then
                die "Not enough disk space at /home. Needed ~$(( home_used_kb / 1024 )) MB, free ~$(( home_free_kb / 1024 )) MB. Aborting to avoid a partial restore."
            fi
        else
            warn "Cannot read free space at /home — disk space check skipped."
        fi
    else
        warn "Cannot determine home directory size — disk space check skipped."
    fi

    log "Restoring /home/$CONTEXT ..."
    # Pipe through rename transform: homedir → CONTEXT
    tar -C "$WORK" --numeric-owner --acls --xattrs --transform "s,^homedir,${CONTEXT}," -cf - homedir | tar -C /home --numeric-owner --acls --xattrs -xf - 2>>"$log_file" || die "Failed to restore home directory."

    # so containers can later start without permission issues
    rm -f /home/"$CONTEXT"/sockets/*/*.sock /home/"$CONTEXT"/sockets/*/*.pid
    
    # MARIADB ERROR: Bad magic header in tc log
    rm -f /home/"$CONTEXT"/volumes/"${CONTEXT}_mysql_data"/_data/tc.log


    local gid; gid=$(id -g "$CONTEXT" 2>/dev/null)
    if [[ -n "$REMAPPED_UID" && -n "$SOURCE_UID" && "$REMAPPED_UID" != "$SOURCE_UID" ]]; then
        log "UID changed ($SOURCE_UID → $REMAPPED_UID); chowning /home/$CONTEXT ..."
        chown -R "${REMAPPED_UID}:${gid:-$REMAPPED_UID}" "/home/$CONTEXT"
    else
        log "UID not changed ($gid); chowning /home/$CONTEXT/docker-data/volumes/ ..."
        chown -R "${gid}:${gid}" "/home/$CONTEXT/docker-data/volumes/" &
        chown -R "${gid}:${gid}" "/home/$CONTEXT/sockets" &
    fi

    if [[ -f "$WORK/mail_external/path.txt" && -f "$WORK/mail_external/mail.tar" ]]; then
        local mp; mp=$(cat "$WORK/mail_external/path.txt")
        log "Restoring external mail → $mp ..."
        mkdir -p "$(dirname "$mp")"
        tar -C "$(dirname "$mp")" --numeric-owner --acls --xattrs -xf "$WORK/mail_external/mail.tar" 2>>"$log_file" || warn "External mail restore failed."
    fi
}
restore_home

# ── 3) Panel DB: patch SQL if new username, then import ──────────────────────
restore_database() {
    log "Restoring panel DB rows ..."

    # Patch user.sql for --new-username: replace old username + old context values
    if [[ -n "$NEW_USERNAME" && -f "$WORK/db/user.sql" ]]; then
        log "Patching user.sql: $ORIG_USERNAME → $USERNAME"
        sed -i "s|'${ORIG_USERNAME}'|'${USERNAME}'|g" "$WORK/db/user.sql"
        [[ "$ORIG_CONTEXT" != "$ORIG_USERNAME" ]] && sed -i "s|'${ORIG_CONTEXT}'|'${CONTEXT}'|g" "$WORK/db/user.sql"
    fi

    # Plan
    local plan_id
    local plan_name; plan_name=$(awk -F"'" '/INSERT INTO plans/{getline; print $2; exit}' "$WORK/db/plan.sql" 2>/dev/null)
    plan_id=$(mysql_q "SELECT id FROM plans WHERE name='$plan_name' LIMIT 1;")
    if [[ -n "$plan_id" ]]; then
        log "Plan '$plan_name' exists (ID: $plan_id) — reusing."
        PLAN_STATUS="reused (ID $plan_id)"
    elif [[ -f "$WORK/db/plan.sql" ]]; then
        sed -i -E ':a;N;$!ba;s/,\s*;\s*/;/g' "$WORK/db/plan.sql"
        ( echo "USE \`$mysql_database\`;"; cat "$WORK/db/plan.sql" ) | mysql_run
        plan_id=$(mysql_q "SELECT id FROM plans WHERE name='$plan_name' LIMIT 1;")
        log "Plan '$plan_name' imported (ID: $plan_id)."
        PLAN_STATUS="imported (ID $plan_id)"
    fi

    # User row
    if [[ -f "$WORK/db/user.sql" ]]; then
        sed -i -E ':a;N;$!ba;s/,\s*;\s*/;/g' "$WORK/db/user.sql"
        local tmp="$WORK/db/_user_local.sql"
        sed -E "s/,[[:space:]]*[0-9]+\);$/,${plan_id:-NULL});/" "$WORK/db/user.sql" | sed "s/'NULL'/NULL/g" > "$tmp"

        if [[ "${username_exists:-0}" -gt 0 && $FORCE -eq 1 ]]; then
            local old_id; old_id=$(mysql_q "SELECT id FROM users WHERE username='$USERNAME';")
            if [[ -n "$old_id" ]]; then
                mysql_q "DELETE FROM sites WHERE domain_id IN (SELECT domain_id FROM domains WHERE user_id=$old_id);"
                mysql_q "DELETE FROM domains WHERE user_id=$old_id;"
                mysql_q "DELETE FROM users WHERE id=$old_id;"
            fi
        fi
        ( echo "USE \`$mysql_database\`;"; cat "$tmp" ) | mysql_run
    fi

    USER_ID=$(mysql_q "SELECT id FROM users WHERE username='$USERNAME';")
    [[ -z "$USER_ID" ]] && die "Failed to import user row for '$USERNAME'."
    log "User row ready (ID: $USER_ID)."
}
restore_database

# ── 4) Feature set + core config ─────────────────────────────────────────────
restore_static_config() {
    local FDIR="/etc/openpanel/openpanel/features"
    if compgen -G "$WORK/features/*.txt" >/dev/null 2>&1; then
        mkdir -p "$FDIR"
        for f in "$WORK/features/"*.txt; do
            local bn; bn=$(basename "$f")
            [[ -f "$FDIR/$bn" ]] && log "Feature set $bn present — keeping." || { cp -a "$f" "$FDIR/"; log "Feature set $bn installed."; }
        done
    fi
    if [[ -d "$WORK/core" && -n "$(ls -A "$WORK/core" 2>/dev/null)" ]]; then
        mkdir -p "/etc/openpanel/openpanel/core/users/$USERNAME"
        cp -a "$WORK/core/." "/etc/openpanel/openpanel/core/users/$USERNAME/"
        log "Core config restored."
    fi
}
restore_static_config

# ── 5) Domains — skip if already owned ───────────────────────────────────────
restore_domains() {
    [[ -f "$WORK/db/domains.list" ]] || { log "No domains in archive."; return; }
    log "Restoring domains ..."
    mkdir -p /etc/openpanel/caddy/domains/ /var/log/caddy/domlogs/ /var/log/caddy/coraza_waf/ /etc/bind/zones/ /etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/ /etc/openpanel/caddy/ssl/custom/ /var/log/caddy/stats/

    if [[ -f "$WORK/caddy/blocked.ips" ]]; then
        mkdir -p /etc/openpanel/caddy/deny/
        cp -a "$WORK/caddy/blocked.ips" "/etc/openpanel/caddy/deny/${CONTEXT}.ips"
        log "Blocked IPs restored."
    fi

    while IFS=$'\t ' read -r domain docroot php_version; do
        [[ -z "$domain" ]] && continue
        local owner
        owner=$(opencli domains-whoowns "$domain" 2>/dev/null | awk -F "Owner of '$domain': " '{print $2}')
        if [[ -n "$owner" ]]; then
            warn "Domain '$domain' exists (owner: $owner) — skipped."
            DOMAINS_SKIPPED=$((DOMAINS_SKIPPED+1))
            DOMAINS_SKIPPED_LIST+=("$domain")
            continue
        fi

        if ! opencli domains-add "$domain" "$USERNAME" --docroot "$docroot" --php_version "$php_version" --skip_caddy --skip_vhost --skip_containers --skip_dns --skip-sentinel >>"$log_file" 2>&1; then
            warn "opencli domains-add failed for '$domain' — skipping assets."
            DOMAINS_SKIPPED=$((DOMAINS_SKIPPED+1))
            continue
        fi
        DOMAINS_ADDED=$((DOMAINS_ADDED+1))

        [[ -f "$WORK/caddy/domains/$domain.conf" ]] && cp -a "$WORK/caddy/domains/$domain.conf" /etc/openpanel/caddy/domains/
        [[ -f "$WORK/caddy/domlogs/$domain" ]]      && cp -a "$WORK/caddy/domlogs/$domain" /var/log/caddy/domlogs/
        [[ -f "$WORK/caddy/waf/$domain.log" ]]      && cp -a "$WORK/caddy/waf/$domain.log" /var/log/caddy/coraza_waf/
        [[ -d "$WORK/caddy/ssl/acme/$domain" ]]     && cp -a "$WORK/caddy/ssl/acme/$domain" /etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/
        [[ -d "$WORK/caddy/ssl/custom/$domain" ]]   && cp -a "$WORK/caddy/ssl/custom/$domain" /etc/openpanel/caddy/ssl/custom/
        [[ -f "$WORK/caddy/suspended/$domain.conf" ]] && { mkdir -p /etc/openpanel/caddy/suspended_domains/; cp -a "$WORK/caddy/suspended/$domain.conf" /etc/openpanel/caddy/suspended_domains/; }

        if [[ -f "$WORK/bind/zones/$domain.zone" ]]; then
            cp -a "$WORK/bind/zones/$domain.zone" /etc/bind/zones/
            [[ -n "$SOURCE_IP" && -n "$NEW_IP" && "$SOURCE_IP" != "$NEW_IP" ]] && sed -i "s/$SOURCE_IP/$NEW_IP/g" "/etc/bind/zones/$domain.zone"
            grep -q "\"$domain\"" /etc/bind/named.conf.local 2>/dev/null || echo "zone \"$domain\" IN { type master; file \"/etc/bind/zones/$domain.zone\"; };" >> /etc/bind/named.conf.local
        fi
        log "Domain restored: $domain"
    done < "$WORK/db/domains.list"

    docker --context default exec openpanel_dns rndc reconfig >/dev/null 2>&1 || true
    (cd /root && docker --context default compose up -d bind9 >/dev/null 2>&1) || true

    # Sites table
    if [[ -f "$WORK/db/sites.sql" ]]; then
        tail -n +2 "$WORK/db/sites.sql" | sed 's/),/)\n/g' | while read -r line; do
            local clean; clean=$(echo "$line" | sed "s/[()']//g" | sed 's/,$//')
            [[ -z "$clean" || "$clean" == ";" ]] && continue
            local sname; sname=$(echo "$clean" | cut -d',' -f1 | xargs)
            local email; email=$(echo "$clean" | cut -d',' -f3 | xargs)
            local ver;   ver=$(echo "$clean"   | cut -d',' -f4 | xargs)
            local typ;   typ=$(echo "$clean"   | cut -d',' -f6 | xargs)
            local ports; ports=$(echo "$clean" | cut -d',' -f7 | xargs)
            local path;  path=$(echo "$clean"  | cut -d',' -f8 | xargs)
            local domain_url; domain_url=$(echo "$clean" | cut -d',' -f9 | xargs)

            # domains-add assigns a fresh domain_id on restore, so resolve the current
            # one by name rather than trusting the source server's ID.
            local did; did=$(mysql_q "SELECT domain_id FROM domains WHERE domain_url='$domain_url' AND user_id=$USER_ID;" 2>/dev/null)
            if [[ -z "$did" ]]; then
                warn "Site '$sname' — could not resolve destination domain_id, skipped."
                continue
            fi
            mysql_q "INSERT INTO sites (site_name,domain_id,admin_email,version,type,ports,path) VALUES ('$sname',$did,'$email','$ver','$typ',$ports,'$path');" || true
        done
    fi
}

restore_domains


# ── 6) FTP ───────────────────────────────────────────────────────────────────
restore_ftp() {
    local LDIR="/etc/openpanel/ftp/users/$CONTEXT"
    [[ -d "$WORK/ftp" && -n "$(ls -A "$WORK/ftp" 2>/dev/null)" ]] || { log "No FTP in archive."; return; }
    log "Restoring FTP accounts ..."
    mkdir -p "$LDIR"; cp -a "$WORK/ftp/." "$LDIR/"
    [[ -f "$LDIR/users.list" ]] || { log "No users.list — skipping container step."; return; }

    [[ -z "$(docker ps -q -f name=openadmin_ftp 2>/dev/null)" ]] && { (cd /root && docker --context default compose up -d openadmin_ftp >/dev/null 2>&1) || true; sleep 2; }

    local GID; GID=$(stat -c '%u' "/home/$CONTEXT" 2>/dev/null)
    [[ "$GID" =~ ^[0-9]+$ ]] || { warn "Cannot get GID for $CONTEXT — FTP container step skipped."; return; }
    docker exec openadmin_ftp sh -c "getent group '$GID' >/dev/null 2>&1" || docker exec openadmin_ftp addgroup -g "$GID" "$CONTEXT" 2>/dev/null || true

    # TODO: check if UID already taken and not our username, in which case assign new UID and update in $LDIR/users.list 
    while IFS='|' read -r fu hp dir uid gid; do
        [[ -z "$fu" ]] && continue
        if docker exec openadmin_ftp id "$fu" >/dev/null 2>&1; then
            log "[skip] FTP user '$fu' exists."; FTP_SKIPPED=$((FTP_SKIPPED+1)); continue
        fi
        local rp="/home/${CONTEXT}/docker-data/volumes/${CONTEXT}_html_data/_data/"
        local nd="${rp}${dir##/var/www/html/}"
        mkdir -p "$nd"; chown -R "$GID:$GID" "$nd"; chmod -R 2775 "$nd"
        docker exec openadmin_ftp useradd -d "$nd" -s /sbin/nologin -g "$CONTEXT" -M "$fu" --badname 2>/dev/null || true
        docker exec openadmin_ftp sh -c "usermod -p '$hp' '$fu'"
        log "FTP user restored: $fu"; FTP_CREATED=$((FTP_CREATED+1))
    done < "$LDIR/users.list"
}
restore_ftp

# ── 7) Email — skip existing mailboxes ───────────────────────────────────────
restore_email() {
    local EMAILS_DIR="$WORK/emails"
    local DMS_CONFIG="/usr/local/mail/openmail/docker-data/dms/config"
    local POSTFWD_DST="/usr/local/mail/openmail/postfwd/postfwd.cf"

    # DKIM keys
    if [[ -d "$EMAILS_DIR/dkim" && -n "$(ls -A "$EMAILS_DIR/dkim" 2>/dev/null)" ]]; then
        local DKIM_DIR="/usr/local/mail/openmail/docker-data/dms/config/opendkim/keys"
        mkdir -p "$DKIM_DIR"
        cp -a "$EMAILS_DIR/dkim/." "$DKIM_DIR/"
        log "DKIM keys restored."
    fi

    # postfix-accounts.cf — read from staged emails/ dir (not homedir)
    local asrc="$EMAILS_DIR/postfix-accounts.cf"
    [[ -s "$asrc" ]] || { log "No postfix-accounts.cf in archive — skipping email merge."; return; }

    [[ -z "$(docker ps -q -f name=${MAIL_CONTAINER} 2>/dev/null)" ]] && {
        [[ -d /usr/local/mail/openmail ]] && (cd /usr/local/mail/openmail && docker --context default compose up -d mailserver roundcube >/dev/null 2>&1) || true
        sleep 2
    }

    local MNT
    MNT=$(docker inspect "$MAIL_CONTAINER" --format '{{ range .Mounts }}{{ if eq .Destination "/tmp/docker-mailserver" }}{{ .Source }}{{ end }}{{ end }}' 2>/dev/null)

    # Helper: merge a simple line-based config file, skipping lines whose key (up to first |/space) already exists
    merge_cf() {
        local src="$1" dst="$2" label="$3"
        [[ -s "$src" ]] || return
        [[ -z "$dst" ]] && { warn "Cannot resolve path for $label — skipped."; return; }
        touch "$dst" 2>/dev/null || true
        local added=0 skipped=0
        while IFS= read -r line; do
            [[ -z "$line" ]] && continue
            local key="${line%%|*}"
            if grep -qF "${key}|" "$dst" 2>/dev/null; then
                skipped=$((skipped+1))
            else
                echo "$line" >> "$dst"
                added=$((added+1))
            fi
        done < "$src"
        log "$label: $added added, $skipped skipped."
    }

    local cfg_dst="${MNT:-$DMS_CONFIG}"
    local adst="$cfg_dst/postfix-accounts.cf"

    log "Merging email accounts ..."
    while IFS='|' read -r addr hash; do
        [[ -z "$addr" ]] && continue
        if grep -q "^${addr}|" "$adst" 2>/dev/null; then
            log "[skip] Mailbox '$addr' exists."
            EMAILS_SKIPPED=$((EMAILS_SKIPPED+1)); EMAILS_SKIPPED_LIST+=("$addr")
            continue
        fi
        echo "${addr}|${hash}" >> "$adst"
        EMAILS_ADDED=$((EMAILS_ADDED+1)); log "Mailbox restored: $addr"
    done < "$asrc"

    # Merge the remaining DMS config files
    merge_cf "$EMAILS_DIR/postfix-regex.cf"          "$cfg_dst/postfix-regex.cf"          "aliases (postfix-regex.cf)"
    merge_cf "$EMAILS_DIR/dovecot-quotas.cf"         "$cfg_dst/dovecot-quotas.cf"         "quotas (dovecot-quotas.cf)"
    merge_cf "$EMAILS_DIR/postfix-receive-access.cf" "$cfg_dst/postfix-receive-access.cf" "receive-access (postfix-receive-access.cf)"
    merge_cf "$EMAILS_DIR/postfix-send-access.cf"    "$cfg_dst/postfix-send-access.cf"    "send-access (postfix-send-access.cf)"
    merge_cf "$EMAILS_DIR/postfix-virtual.cf"        "$cfg_dst/postfix-virtual.cf"        "catch-all (postfix-virtual.cf)"

    # postfwd.cf — append non-duplicate rule blocks (two-line id=…/action=… pairs)
    if [[ -s "$EMAILS_DIR/postfwd.cf" && -f "$POSTFWD_DST" ]]; then
        local pf_added=0
        while IFS= read -r id_line; do
            IFS= read -r action_line
            [[ -z "$id_line" ]] && continue
            if ! grep -qF "$id_line" "$POSTFWD_DST" 2>/dev/null; then
                printf '%s\n%s\n' "$id_line" "$action_line" >> "$POSTFWD_DST"
                pf_added=$((pf_added+1))
            fi
        done < "$EMAILS_DIR/postfwd.cf"
        log "postfwd.cf: $pf_added rule(s) added."
    fi

    [[ $EMAILS_ADDED -gt 0 ]] && { docker exec "$MAIL_CONTAINER" postfix reload >/dev/null 2>&1 || docker restart "$MAIL_CONTAINER" >/dev/null 2>&1 || true; }
}
restore_email

# ── 8) Docker ────────────────────────────────────────────────────────────────
restore_docker() {
    [[ -f "$WORK/docker/apparmor.profile" ]] && {
        cp -a "$WORK/docker/apparmor.profile" "/etc/apparmor.d/home.$CONTEXT.bin.rootlesskit"
        systemctl restart apparmor.service >/dev/null 2>&1 || true
    }
    if [[ -d "/home/$CONTEXT/.docker" ]]; then
        local uid_now; uid_now=$(id -u "$CONTEXT" 2>/dev/null)
        if [[ -n "$uid_now" ]]; then
            docker context create "$CONTEXT" --docker "host=unix:///hostfs/run/user/${uid_now}/docker.sock" --description "$CONTEXT" >/dev/null 2>&1 || true
            loginctl enable-linger "$CONTEXT" >/dev/null 2>&1 || true
            machinectl shell "${CONTEXT}@" /bin/bash -c 'systemctl --user daemon-reload' >/dev/null 2>&1 || true
            machinectl shell "${CONTEXT}@" /bin/bash -c 'systemctl --user --quiet restart docker' >/dev/null 2>&1 || true
        fi
    fi
    if [[ -f "$WORK/docker/containers.txt" && -f "/home/$CONTEXT/docker-compose.yml" ]]; then
        local ctx containers
        IFS=':' read -r ctx containers < "$WORK/docker/containers.txt"
        ctx=$(echo "$ctx" | xargs); containers=$(echo "$containers" | xargs)
        if [[ -n "$ctx" && "$containers" != "no containers" && -n "$containers" ]]; then
            log "Starting containers for $ctx ..."
            docker --context="$ctx" compose -f "/home/$ctx/docker-compose.yml" down >/dev/null 2>&1 || true
            docker --context="$ctx" compose -f "/home/$ctx/docker-compose.yml" up -d $containers >/dev/null 2>&1 || true
        fi
    fi
}
restore_docker

# ── 9) Reload services + quotas ──────────────────────────────────────────────
[[ -d "$WORK/caddy/stats/$ORIG_USERNAME" ]] && { mkdir -p /var/log/caddy/stats/; cp -a "$WORK/caddy/stats/$ORIG_USERNAME" /var/log/caddy/stats/; }

log "Reloading services ..."
(cd /root && docker compose up -d openpanel bind9 caddy >/dev/null 2>&1; systemctl restart admin >/dev/null 2>&1) || true
docker --context default exec caddy caddy reload >/dev/null 2>&1 || true
systemctl daemon-reload >/dev/null 2>&1 || true
log "Recalculating quotas ..."
opencli user-quota --update "$USERNAME" >/dev/null 2>&1 || true

# ── Summary ──────────────────────────────────────────────────────────────────
end_time=$(date +%s)
elapsed=$(( end_time - start_time ))
elapsed_h=$(( elapsed / 3600 ))
elapsed_m=$(( (elapsed % 3600) / 60 ))
elapsed_s=$(( elapsed % 60 ))

HOME_SIZE=$(du -sh "/home/$CONTEXT" 2>/dev/null | cut -f1)

CONTAINERS_STARTED="none"
if [[ -f "$WORK/docker/containers.txt" ]]; then
    CONTAINERS_STARTED=$(cut -d: -f2 "$WORK/docker/containers.txt" 2>/dev/null | xargs)
    [[ -z "$CONTAINERS_STARTED" ]] && CONTAINERS_STARTED="none"
fi

slog ""
slog "══════════════════════════════════════════════════════════"
slog "  RESTORE COMPLETE — $USERNAME"
slog "══════════════════════════════════════════════════════════"
slog "$(printf "  %-24s %s\n"       "Archive:"     "$(basename "$ARCHIVE")")"
slog "$(printf "  %-24s %dh %dm %ds\n" "Time taken:" "$elapsed_h" "$elapsed_m" "$elapsed_s")"
if [[ -n "$NEW_USERNAME" ]]; then
    slog "$(printf "  %-24s %s  →  %s\n" "Username:" "$ORIG_USERNAME" "$USERNAME")"
    slog "$(printf "  %-24s %s  →  %s\n" "Context:"  "$ORIG_CONTEXT"  "$CONTEXT")"
fi
slog ""
slog "  Restored:"
slog "$(printf "    %-22s %s\n" "System user:"     "$CONTEXT  [$SYSTEM_USER_STATUS]")"
slog "$(printf "    %-22s %s\n" "Home directory:"  "/home/$CONTEXT/  ($HOME_SIZE)")"
slog "$(printf "    %-22s %s\n" "Plan:"            "\"$PLAN_NAME\"  [$PLAN_STATUS]")"
slog "$(printf "    %-22s %s\n" "DB user ID:"      "$USER_ID")"
slog "$(printf "    %-22s %s added, %s skipped\n"  "Domains:"       "$DOMAINS_ADDED"  "$DOMAINS_SKIPPED")"
slog "$(printf "    %-22s %s created, %s skipped\n" "FTP accounts:" "$FTP_CREATED"    "$FTP_SKIPPED")"
slog "$(printf "    %-22s %s added, %s skipped\n"  "Email accounts:" "$EMAILS_ADDED"  "$EMAILS_SKIPPED")"
slog "$(printf "    %-22s %s\n" "Containers:"      "$CONTAINERS_STARTED")"
slog ""

warn_count=${#WARNINGS[@]}
slog "  Warnings: $warn_count"
if [[ $warn_count -gt 0 ]]; then
    for w in "${WARNINGS[@]}"; do slog "    [!] $w"; done
else
    slog "    none"
fi
slog "══════════════════════════════════════════════════════════"
slog ""
exit 0
