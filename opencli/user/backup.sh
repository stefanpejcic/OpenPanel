#!/bin/bash
################################################################################
# Script Name: user/backup.sh
# Description: Creates a full account .tar.gz backup of a single OpenPanel user account.
# Usage: opencli user-backup --account <USER> [--output <DIR>] [--quiet]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 03.07.2026
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
BACKUP_FORMAT_VERSION="2"

USERNAME=""
CUSTOM_OUTPUT=""
QUIET=0

# Runtime tracking — populated during the backup, printed in summary
WARNINGS=()
BACKUP_DOMAINS=()
DOMAIN_LIST_STR="(none)"
BACKUP_FTP_COUNT=0
DISK_ESTIMATE_MB=0
DISK_FREE_MB=0
DISK_CHECK_SOURCE="unknown"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --account)  USERNAME="$2"; shift 2 ;;
        --output)   CUSTOM_OUTPUT="$2"; shift 2 ;;
        --quiet)    QUIET=1; shift ;;
        *) echo "[ERROR] Unknown option: $1"; exit 1 ;;
    esac
done

[[ -z "$USERNAME" ]] && { echo "Usage: opencli user-backup --account <USER> [--output <DIR>] [--quiet]"; exit 1; }

# DB connection  (provides $config_file and $mysql_database)
DB_CONFIG_FILE="/usr/local/opencli/db.sh"
[[ -f "$DB_CONFIG_FILE" ]] || { echo "[ERROR] DB config not found: $DB_CONFIG_FILE"; exit 1; }
# shellcheck disable=SC1090
. "$DB_CONFIG_FILE"
mysql_q() { mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -s -e "$1"; }

# Resolve account identity
USER_ID=$(mysql_q "SELECT id FROM users WHERE username = '$USERNAME';")
[[ -z "$USER_ID" ]] && { echo "[ERROR] No user found: $USERNAME"; exit 1; }

CONTEXT=$(mysql_q "SELECT server FROM users WHERE username = '$USERNAME';")
[[ -z "$CONTEXT" ]] && { echo "[ERROR] Could not resolve context (server) for '$USERNAME'"; exit 1; }

read -r PLAN_ID PLAN_NAME PLAN_FEATURE_SET < <(mysql_q "
    SELECT p.id, p.name, p.feature_set
    FROM plans p JOIN users u ON p.id = u.plan_id
    WHERE u.id = $USER_ID;")

SYS_UID=$(awk -F: -v u="$CONTEXT" '$1==u{print $3}' /etc/passwd)
SYS_GID=$(awk -F: -v u="$CONTEXT" '$1==u{print $4}' /etc/passwd)

if [ -z "$SYS_UID" ] || [ -z "$SYS_GID" ]; then
    if [ -d "/home/$CONTEXT" ]; then
        SYS_UID=$(stat -c '%u' "/home/$CONTEXT")
        SYS_GID=$(stat -c '%g' "/home/$CONTEXT")
    else
        echo "$CONTEXT not found in /etc/passwd and /home/$CONTEXT does not exist" >&2
        exit 1
    fi
fi

DOMAIN_IDS=$(mysql_q "SELECT domain_id FROM domains WHERE user_id = $USER_ID;")
DOMAIN_COUNT=0
[[ -n "$DOMAIN_IDS" ]] && DOMAIN_COUNT=$(wc -l <<< "$DOMAIN_IDS")

SITE_COUNT=0
if [[ -n "$DOMAIN_IDS" ]]; then
    DOMAIN_ID_LIST=$(echo "$DOMAIN_IDS" | paste -sd "," -)
    SITE_COUNT=$(mysql_q "SELECT COUNT(*) FROM sites WHERE domain_id IN ($DOMAIN_ID_LIST);" 2>/dev/null || echo 0)
fi

# External mail fires only when admin set email_storage_location to a path outside /home.  Default for <1.7.5: mail lives in /home/CONTEXT/mail/
STORE_EMAILS_IN=$(grep -E '^email_storage_location=' /etc/openpanel/openadmin/config/admin.ini 2>/dev/null | cut -d'=' -f2- | xargs)
MAIL_EXTERNAL_PATH=""
MAIL_EXTERNAL_SIZE="n/a"
if [[ "$STORE_EMAILS_IN" == /* && -d "$STORE_EMAILS_IN" ]]; then
    MAIL_EXTERNAL_PATH="$STORE_EMAILS_IN"
    MAIL_EXTERNAL_SIZE=$(du -sh "$MAIL_EXTERNAL_PATH" 2>/dev/null | cut -f1)
fi

# Destination
DEST_DIR="${CUSTOM_OUTPUT:-/home/$CONTEXT/docker-data/volumes/${CONTEXT}_html_data/_data/_backups}"
timestamp="$(date +'%Y-%m-%d_%H-%M-%S')"
ARCHIVE_NAME="${USERNAME}_${timestamp}.tar.gz"
ARCHIVE="${DEST_DIR}/${ARCHIVE_NAME}"

# Logging
base_name="$(basename "$USERNAME")"
log_dir="/var/log/openpanel/admin/backups"
mkdir -p "$log_dir"
log_file="$log_dir/${base_name}_backup_${timestamp}.log"

log() {
    local msg="$1"; local ts; ts=$(date +'%Y-%m-%d %H:%M:%S')
    if [[ $QUIET -eq 1 ]]; then echo "[$ts] $msg" >> "$log_file"
    else echo "[$ts] $msg" | tee -a "$log_file"; fi
}
# warn() logs the message AND records it for the summary
warn() {
    local msg="$1"
    WARNINGS+=("$msg")
    log "[!] $msg"
}
die() { log "[✘] $1"; exit "${2:-1}"; }

# Summary always prints to stdout + log regardless of --quiet
slog() { echo "$1" | tee -a "$log_file"; }

# Disk space check
check_disk_space() {
    log "Checking disk space ..."
    mkdir -p "$DEST_DIR" 2>/dev/null || true

    # --- Step 1: how much disk does this account use? (= backup size estimate)
    local used_kb=0

    if command -v repquota &>/dev/null; then
        #   username  --  used_blocks  soft  hard  grace  used_inodes ...
        local rq
        rq=$(repquota -au 2>/dev/null | awk -v u="$CONTEXT" '$1==u {print $3; exit}')
        if [[ "$rq" =~ ^[0-9]+$ && "$rq" -gt 0 ]]; then
            used_kb="$rq"
            DISK_CHECK_SOURCE="repquota"
        fi
    fi

    if [[ "$used_kb" -eq 0 ]]; then
        if [[ -d "/home/$CONTEXT" ]]; then
            used_kb=$(du -sk --exclude="docker-data/volumes/${CONTEXT}_html_data/_data/_backups" "/home/$CONTEXT" 2>/dev/null | cut -f1)
            DISK_CHECK_SOURCE="du"
        fi
    fi

    if [[ ! "$used_kb" =~ ^[0-9]+$ || "$used_kb" -eq 0 ]]; then
        warn "Cannot determine disk usage for $CONTEXT — disk space check skipped."
        return
    fi
    DISK_ESTIMATE_MB=$(( used_kb / 1024 ))

    # --- Step 2: how much is free at the destination?
    local free_kb
    free_kb=$(df -k "$DEST_DIR" 2>/dev/null | awk 'NR==2 {print $4}')
    if [[ ! "$free_kb" =~ ^[0-9]+$ ]]; then
        warn "Cannot read free space at $DEST_DIR — disk space check skipped."
        return
    fi
    DISK_FREE_MB=$(( free_kb / 1024 ))

    log "  Source size  (~${DISK_ESTIMATE_MB} MB via $DISK_CHECK_SOURCE)"
    log "  Free at dest (~${DISK_FREE_MB} MB at $DEST_DIR)"

    # Abort threshold: compressed archive typically reaches 30–70 % of source.
    # If free < 1/3 of source, even a heavily compressed archive would fill the disk.
    local abort_kb=$(( used_kb / 3 ))
    if [[ "$free_kb" -lt "$abort_kb" ]]; then
        die "Not enough disk space. Source ~${DISK_ESTIMATE_MB} MB, free ~${DISK_FREE_MB} MB at $(dirname "$ARCHIVE"). Aborting to avoid a partial archive."
    fi

    # Warn threshold: free < source (uncommon but possible for binary-heavy accounts)
    if [[ "$free_kb" -lt "$used_kb" ]]; then
        warn "Low disk space: source ~${DISK_ESTIMATE_MB} MB but only ~${DISK_FREE_MB} MB free at $(dirname "$ARCHIVE"). Compressed archive is usually smaller — proceeding, but monitor progress."
    else
        log "  Disk space OK."
    fi
}


# REAL BACKUP STARTS HERE
log "Backup started  log: $log_file  (PID: $pid)"
log "Account: $USERNAME  Context: $CONTEXT  UID:${SYS_UID:-?} GID:${SYS_GID:-?}  Plan: ${PLAN_NAME:-none}"

check_disk_space

# Staging area: small metadata only; home is streamed live
STAGE=$(mktemp -d "/tmp/opbackup_${base_name}.XXXXXX")
BACKUP_OK=0
trap '[[ $BACKUP_OK -eq 0 && -f "$ARCHIVE" ]] && rm -f "$ARCHIVE"; rm -rf "$STAGE"' EXIT
mkdir -p "$STAGE"/{db,system,features,core,ftp,caddy/domains,caddy/ssl/acme,caddy/ssl/custom,caddy/domlogs,caddy/waf,caddy/stats,caddy/suspended,bind/zones,docker,emails/dkim}

# --- manifest ---
log "Writing manifest ..."
SOURCE_IP=$(curl --silent --max-time 1 -4 "https://ip.openpanel.com" 2>/dev/null || curl --silent --max-time 1 -4 "https://ifconfig.me" 2>/dev/null || ip addr | grep 'inet ' | grep global | head -n1 | awk '{print $2}' | cut -f1 -d/)
cat > "$STAGE/manifest.env" <<EOF
BACKUP_FORMAT_VERSION="$BACKUP_FORMAT_VERSION"
USERNAME="$USERNAME"
CONTEXT="$CONTEXT"
PLAN_NAME="$PLAN_NAME"
PLAN_FEATURE_SET="$PLAN_FEATURE_SET"
SOURCE_UID="$SYS_UID"
SOURCE_GID="$SYS_GID"
SOURCE_IP="$SOURCE_IP"
CREATED_AT="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
EOF

# --- panel DB ---
log "Exporting panel DB rows ..."
mysql_q "
  SELECT name, description, domains_limit, websites_limit, email_limit, ftp_limit,
         disk_limit, inodes_limit, db_limit, cpu, ram, bandwidth, feature_set,
         max_email_quota, max_hourly_email
  FROM plans WHERE id = $PLAN_ID;" > "$STAGE/db/plan.tsv"

awk 'BEGIN{FS="\t";
  print "INSERT INTO plans (name, description, domains_limit, websites_limit, email_limit, ftp_limit, disk_limit, inodes_limit, db_limit, cpu, ram, bandwidth, feature_set, max_email_quota, max_hourly_email) VALUES"}
  {printf "('\''%s'\'', '\''%s'\'', %s, %s, %s, %s, '\''%s'\'', %s, %s, '\''%s'\'', '\''%s'\'', %s, '\''%s'\'', '\''%s'\'', %s),\n",
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15}
  END{print ";"}' "$STAGE/db/plan.tsv" > "$STAGE/db/plan.sql"
rm -f "$STAGE/db/plan.tsv"

mysql_q "
  SELECT username, password, email, owner, user_domains, twofa_enabled, otp_secret,
         plan, registered_date, server, plan_id
  FROM users WHERE id = $USER_ID;" > "$STAGE/db/user.tsv"

awk 'BEGIN{FS="\t";
  print "INSERT INTO users (username, password, email, owner, user_domains, twofa_enabled, otp_secret, plan, registered_date, server, plan_id) VALUES"}
  {printf "('\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', %s, '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', %s),\n",
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11}
  END{print ";"}' "$STAGE/db/user.tsv" > "$STAGE/db/user.sql"
rm -f "$STAGE/db/user.tsv"

if [[ -n "$DOMAIN_IDS" ]]; then
    DOMAIN_ID_LIST=$(echo "$DOMAIN_IDS" | paste -sd "," -)
    # domain_url is included so restore can re-resolve the site's domain_id by name —
    # domains-add assigns a fresh auto-increment ID on restore, which rarely matches this server's.
    mysql_q "
      SELECT s.site_name, s.domain_id, s.admin_email, s.version, s.created_date, s.type, s.ports, s.path, d.domain_url
      FROM sites s JOIN domains d ON s.domain_id = d.domain_id
      WHERE s.domain_id IN ($DOMAIN_ID_LIST);" > "$STAGE/db/sites.tsv"
    if [[ -s "$STAGE/db/sites.tsv" ]]; then
        awk 'BEGIN{FS="\t";
          print "-- sites export (site_name, domain_id, admin_email, version, created_date, type, ports, path, domain_url)"}
          {printf "('\''%s'\'', %s, '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', %s, '\''%s'\'', '\''%s'\''),\n",
            $1,$2,$3,$4,$5,$6,$7,$8,$9}
          END{print ";"}' "$STAGE/db/sites.tsv" > "$STAGE/db/sites.sql"
    fi
    rm -f "$STAGE/db/sites.tsv"
fi

# domains.list: domain <tab> docroot <tab> php_version — used by restore
ALL_DOMAINS=$(opencli domains-user "$USERNAME" --docroot --php_version 2>/dev/null)
if [[ "$ALL_DOMAINS" != *"No domains found for user"* && -n "$ALL_DOMAINS" ]]; then
    echo "$ALL_DOMAINS" > "$STAGE/db/domains.list"
    while IFS=$'\t ' read -r domain _; do
        [[ -n "$domain" ]] && BACKUP_DOMAINS+=("$domain")
    done <<< "$ALL_DOMAINS"
fi

[[ ${#BACKUP_DOMAINS[@]} -gt 0 ]] && DOMAIN_LIST_STR="${BACKUP_DOMAINS[*]}"

# --- system user ---
# TODO: fails in container, do we need it for restore?
log "Capturing system user ($CONTEXT) ..."
awk -F: -v u="$CONTEXT" '$1==u{print}' /etc/passwd > "$STAGE/system/passwd.user"
awk -F: -v u="$CONTEXT" 'BEGIN{gid=""}
    $1==u{gid=$4}
    $3==gid{print}
    $1==u{print}' /etc/group | sort -u > "$STAGE/system/group.user"
grep "^${CONTEXT}:" /etc/shadow 2>/dev/null > "$STAGE/system/shadow.user" || true

# --- feature set ---
PER_USER_FEAT_FILE="/home/$CONTEXT/features.txt"
if [[ ! -f "$PER_USER_FEAT_FILE" ]]; then
    [[ -n "$PLAN_FEATURE_SET" && -f "/etc/openpanel/openpanel/features/${PLAN_FEATURE_SET}.txt" ]] && cp -a "/etc/openpanel/openpanel/features/${PLAN_FEATURE_SET}.txt" "$STAGE/features/" && log "Collected feature set for the plan"
fi

# --- per-user core config ---
CORE_DIR="/etc/openpanel/openpanel/core/users/$USERNAME"
[[ -d "$CORE_DIR" ]] && cp -a "$CORE_DIR/." "$STAGE/core/" && log "Collected system files for user"

# --- FTP ---
FTP_DIR="/etc/openpanel/ftp/users/$CONTEXT"
if [[ -d "$FTP_DIR" ]]; then
    cp -a "$FTP_DIR/." "$STAGE/ftp/" && log "Collected FTP accounts for user"
    BACKUP_FTP_COUNT=$(grep -c '.' "$FTP_DIR/users.list" 2>/dev/null || echo 0)
fi

# --- per-domain caddy + bind assets ---
if [[ -f "$STAGE/db/domains.list" ]]; then
    log "Collecting per-domain assets..."
    mapfile -t domains < "$STAGE/db/domains.list"
    for ((i=0; i<${#domains[@]}; i++)); do
        IFS=$'\t ' read -r domain docroot php_version <<< "${domains[i]}"
        [[ -z "$domain" ]] && continue

        prefix="├──"
        subprefix="│   "
        (( i == ${#domains[@]} - 1 )) && { prefix="└──"; subprefix="    "; }

        log "$prefix domain: $domain"
        [[ -f "/var/log/caddy/domlogs/$domain/access.log" ]] && cp -a "/var/log/caddy/domlogs/$domain/access.log" "$STAGE/caddy/domlogs/$domain.log" && log "${subprefix}├── Domlog"
        [[ -f "/var/log/caddy/coraza_waf/$domain.log" ]] && cp -a "/var/log/caddy/coraza_waf/$domain.log" "$STAGE/caddy/waf/" && log "${subprefix}├── WAF log"
        [[ -f "/etc/bind/zones/$domain.zone" ]] && cp -a "/etc/bind/zones/$domain.zone" "$STAGE/bind/zones/" && log "${subprefix}├── DNS zone"
        [[ -f "/etc/openpanel/caddy/suspended_domains/$domain.conf" ]] && cp -a "/etc/openpanel/caddy/suspended_domains/$domain.conf" "$STAGE/caddy/suspended/" && log "${subprefix}├── Suspended"
        [[ -d "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/$domain" ]] && cp -a "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/$domain" "$STAGE/caddy/ssl/acme/" && log "${subprefix}├── Let's Encrypt SSL"
        [[ -d "/etc/openpanel/caddy/ssl/custom/$domain" ]] && cp -a "/etc/openpanel/caddy/ssl/custom/$domain" "$STAGE/caddy/ssl/custom/" && log "${subprefix}├── Custom SSL"
        [[ -d "/usr/local/mail/openmail/docker-data/dms/config/opendkim/keys/$domain" ]] && cp -ra "/usr/local/mail/openmail/docker-data/dms/config/opendkim/keys/$domain" "$STAGE/emails/dkim/" && log "${subprefix}├── DKIM"
        [[ -f "/etc/openpanel/caddy/domains/$domain.conf" ]] && cp -a "/etc/openpanel/caddy/domains/$domain.conf" "$STAGE/caddy/domains/" && log "${subprefix}└── Caddyfile"
    done
fi

# --- Domlogs ---
[[ -d "/var/log/caddy/stats/$USERNAME" ]] && cp -a "/var/log/caddy/stats/$USERNAME" "$STAGE/caddy/stats/" && log "Collected domlogs for domains"

# --- Blocked IPs ---
[[ -f "/etc/openpanel/caddy/deny/$CONTEXT.ips" ]] && cp -a "/etc/openpanel/caddy/deny/$CONTEXT.ips" "$STAGE/caddy/blocked.ips" && log "Collected IP Blocker settings for domains"

# --- docker metadata ---
if [[ -f "/home/$CONTEXT/docker-compose.yml" ]]; then
    containers=$(docker --context="$CONTEXT" ps -a --format "{{.Names}}" 2>/dev/null | tr '\n' ' ' | sed 's/ $//')
    echo "$CONTEXT: ${containers:-no containers}" > "$STAGE/docker/containers.txt" && log "Collected list of currently active containers for user"
fi
[[ -f "/etc/apparmor.d/home.$CONTEXT.bin.rootlesskit" ]] && cp -a "/etc/apparmor.d/home.$CONTEXT.bin.rootlesskit" "$STAGE/docker/apparmor.profile" && log "Collected AppArmor profile for user"
echo "${SYS_UID}" > "$STAGE/docker/uid.txt"

# --- Emails ---
if [[ -n "$MAIL_EXTERNAL_PATH" ]]; then
    log "Archiving external mail store: $MAIL_EXTERNAL_PATH ..."
    mkdir -p "$STAGE/mail_external"
    tar -C "$(dirname "$MAIL_EXTERNAL_PATH")" --numeric-owner --acls --xattrs -cf "$STAGE/mail_external/mail.tar" "$(basename "$MAIL_EXTERNAL_PATH")" 2>>"$log_file" || warn "Failed to archive external mail store (non-fatal)."
    echo "$MAIL_EXTERNAL_PATH" > "$STAGE/mail_external/path.txt"
fi


if [[ -s "$CORE_DIR/emails.yml" ]]; then
    log "Collecting email accounts information.."

    readonly DMS_CONFIG="/usr/local/mail/openmail/docker-data/dms/config"
    readonly POSTFWD_SRC="/usr/local/mail/openmail/postfwd/postfwd.cf"

    mkdir -p "$STAGE/emails"

    DOMAIN_PATTERN=$(printf '@%s\|' $DOMAIN_LIST_STR | sed 's/\\|$//')
    REGEX_PATTERN=$(printf '/\\*@%s/|' $DOMAIN_LIST_STR | sed 's/|$//')

    : > "$STAGE/emails/postfix-accounts.cf"
    [[ -f "$DMS_CONFIG/postfix-accounts.cf" ]] && grep "$DOMAIN_PATTERN" "$DMS_CONFIG/postfix-accounts.cf" > "$STAGE/emails/postfix-accounts.cf" || true &

    : > "$STAGE/emails/postfix-regex.cf"
    [[ -f "$DMS_CONFIG/postfix-regex.cf" ]] && grep -E "$REGEX_PATTERN" "$DMS_CONFIG/postfix-regex.cf" > "$STAGE/emails/postfix-regex.cf" || true &

    : > "$STAGE/emails/dovecot-quotas.cf"
    [[ -f "$DMS_CONFIG/dovecot-quotas.cf" ]] && grep "$DOMAIN_PATTERN" "$DMS_CONFIG/dovecot-quotas.cf" > "$STAGE/emails/dovecot-quotas.cf" || true &

    : > "$STAGE/emails/postfix-receive-access.cf"
    [[ -f "$DMS_CONFIG/postfix-receive-access.cf" ]] && grep "$DOMAIN_PATTERN" "$DMS_CONFIG/postfix-receive-access.cf" > "$STAGE/emails/postfix-receive-access.cf" || true &

    : > "$STAGE/emails/postfix-send-access.cf"
    [[ -f "$DMS_CONFIG/postfix-send-access.cf" ]] && grep "$DOMAIN_PATTERN" "$DMS_CONFIG/postfix-send-access.cf" > "$STAGE/emails/postfix-send-access.cf" || true &

    : > "$STAGE/emails/postfix-virtual.cf"
    [[ -f "$DMS_CONFIG/postfix-virtual.cf" ]] && grep "$DOMAIN_PATTERN" "$DMS_CONFIG/postfix-virtual.cf" > "$STAGE/emails/postfix-virtual.cf" || true &

    : > "$STAGE/emails/postfwd.cf"
    if [[ -f "$POSTFWD_SRC" ]]; then
        while IFS= read -r line; do
            matched=0

            if [[ "$line" == id=* ]]; then
                for domain in $DOMAIN_LIST_STR; do
                    if [[ "$line" == *"@${domain}"* ]]; then
                        matched=1
                        break
                    fi
                done
            fi

            if [[ $matched -eq 1 ]]; then
                echo "$line"
                IFS= read -r next_line && echo "$next_line"
            fi
        done < "$POSTFWD_SRC" > "$STAGE/emails/postfwd.cf"
    fi

    wait  # wait for all to complete before starting compressions
    accounts_count=$(wc -l < "$STAGE/emails/postfix-accounts.cf") && log "Collected ${accounts_count} email accounts"
    alias_count=$(wc -l < "$STAGE/emails/postfix-regex.cf") && log "Collected ${alias_count} aliases"
    quotas_count=$(wc -l < "$STAGE/emails/dovecot-quotas.cf") && log "Collected dovecot quota restrictions for ${quotas_count} addresses"
    suspended_receive_count=$(wc -l < "$STAGE/emails/postfix-receive-access.cf") && log "Collected suspended incoming status for ${suspended_receive_count} addresses"
    suspended_send_count=$(wc -l < "$STAGE/emails/postfix-send-access.cf") && log "Collected suspended outgoing status for ${suspended_send_count} addresses"
    default_addresses_count=$(wc -l < "$STAGE/emails/postfix-virtual.cf") && log "Collected ${default_addresses_count} default (catch-all) addresses"
    postfwd_domain_limits_count=$(wc -l < "$STAGE/emails/postfwd.cf") && log "Collected ${postfwd_domain_limits_count} postfwd rules"
fi

STAGE_ITEMS=()
for item in manifest.env db system features core ftp caddy bind docker emails; do
    [[ -e "$STAGE/$item" ]] && STAGE_ITEMS+=("$item")
done
[[ -d "$STAGE/mail_external" ]] && STAGE_ITEMS+=("mail_external")

ESCAPED_CTX=$(printf '%s\n' "$CONTEXT" | sed 's|[|\\.*^$()+?{}]|\\&|g')

# Use pigz for multi-core compression if available, otherwise fall back to gzip
if command -v pigz &>/dev/null; then
    log "Using pigz for multi-core compression"
    TAR_ARGS=(
        -cf -
        --numeric-owner --acls --xattrs
        "--exclude=${CONTEXT}/docker-data/volumes/${CONTEXT}_html_data/_data/_backups"
        "--exclude=${CONTEXT}/sockets/*/*/*.sock"
        "--exclude=${CONTEXT}/docker-data/volumes/${CONTEXT}_mysql_data/_data/tc.log"
        "--transform=s|^${ESCAPED_CTX}/|homedir/|;s|^${ESCAPED_CTX}$|homedir|"
    )

    if [[ -d "/home/$CONTEXT" ]]; then
        log "Streaming /home/$CONTEXT  →  homedir/ ..."
        TAR_ARGS+=(-C /home "$CONTEXT")
    else
        warn "/home/$CONTEXT not found — archive will not contain home files."
    fi

    [[ ${#STAGE_ITEMS[@]} -gt 0 ]] && TAR_ARGS+=(-C "$STAGE" "${STAGE_ITEMS[@]}")

    mkdir -p "$DEST_DIR"
    tar "${TAR_ARGS[@]}" 2>>"$log_file" | pigz > "$ARCHIVE" || die "tar failed creating: $ARCHIVE"
else
    TAR_ARGS=(
        -czf "$ARCHIVE"
        --numeric-owner --acls --xattrs
        "--exclude=${CONTEXT}/docker-data/volumes/${CONTEXT}_html_data/_data/_backups"
        "--exclude=${CONTEXT}/sockets/*/*/*.sock"
        "--exclude=${CONTEXT}/docker-data/volumes/${CONTEXT}_mysql_data/_data/tc.log"
        "--transform=s|^${ESCAPED_CTX}/|homedir/|;s|^${ESCAPED_CTX}$|homedir|"
    )

    if [[ -d "/home/$CONTEXT" ]]; then
        log "Streaming /home/$CONTEXT  →  homedir/ ..."
        TAR_ARGS+=(-C /home "$CONTEXT")
    else
        warn "/home/$CONTEXT not found — archive will not contain home files."
    fi

    [[ ${#STAGE_ITEMS[@]} -gt 0 ]] && TAR_ARGS+=(-C "$STAGE" "${STAGE_ITEMS[@]}")

    mkdir -p "$DEST_DIR"
    tar "${TAR_ARGS[@]}" 2>>"$log_file" || die "tar failed creating: $ARCHIVE"
fi

# verify archive integrity ---
gzip -t "$ARCHIVE" 2>>"$log_file" || die "Archive integrity check failed: $ARCHIVE"

# Permissions + ownership
chmod 640 "$ARCHIVE"
BACKUP_UID=$(id -u "$CONTEXT" 2>/dev/null)
BACKUP_GID=$(id -g "$CONTEXT" 2>/dev/null)
if [[ -n "$BACKUP_UID" && -n "$BACKUP_GID" ]]; then
    chown "${BACKUP_UID}:${BACKUP_GID}" "$ARCHIVE" 2>/dev/null && log "Ownership: $CONTEXT (${BACKUP_UID}:${BACKUP_GID})" || warn "Could not chown archive to $CONTEXT (non-fatal)."
fi

# Summary
end_time=$(date +%s)
elapsed=$(( end_time - start_time ))
elapsed_h=$(( elapsed / 3600 ))
elapsed_m=$(( (elapsed % 3600) / 60 ))
elapsed_s=$(( elapsed % 60 ))

ARCHIVE_SIZE=$(du -h "$ARCHIVE" 2>/dev/null | cut -f1)

CONTAINERS_STR="none"
if [[ -f "$STAGE/docker/containers.txt" ]]; then
    CONTAINERS_STR=$(cut -d: -f2 "$STAGE/docker/containers.txt" 2>/dev/null | xargs)
    [[ -z "$CONTAINERS_STR" ]] && CONTAINERS_STR="none"
fi

slog ""
slog "══════════════════════════════════════════════════════════"
slog "  BACKUP COMPLETE — $ARCHIVE_NAME"
slog "══════════════════════════════════════════════════════════"
slog "$(printf "  %-24s %s\n"   "Archive:"       "$ARCHIVE")"
COMPRESS_USED="gzip"
command -v pigz &>/dev/null && COMPRESS_USED="pigz (multi-core)"
slog "$(printf "  %-24s %s\n"   "Size:"          "$ARCHIVE_SIZE")"
slog "$(printf "  %-24s %s\n"   "Compression:"   "$COMPRESS_USED")"
slog "$(printf "  %-24s %dh %dm %ds\n" "Time taken:" "$elapsed_h" "$elapsed_m" "$elapsed_s")"
slog ""
slog "  Contents:"
slog "$(printf "    %-22s %s\n" "Plan:"          "\"$PLAN_NAME\"")"
slog "$(printf "    %-22s %s\n" "Domains (${#BACKUP_DOMAINS[@]}):" "$DOMAIN_LIST_STR")"
slog "$(printf "    %-22s %s\n" "Sites:"         "$SITE_COUNT")"
slog "$(printf "    %-22s %s\n" "Feature set:"   "${PLAN_FEATURE_SET:-none}")"
slog "$(printf "    %-22s %s\n" "FTP accounts:"  "$BACKUP_FTP_COUNT")"
slog "$(printf "    %-22s %s\n" "Home dir:"      "homedir/  (source ~${DISK_ESTIMATE_MB} MB)")"
slog "$(printf "    %-22s %s\n" "Docker:"        "${CONTAINERS_STR}")"
[[ -n "$MAIL_EXTERNAL_PATH" ]] && slog "$(printf "    %-22s %s\n" "External mail:" "$MAIL_EXTERNAL_PATH  ($MAIL_EXTERNAL_SIZE)")"
slog ""
slog "  Disk check:  source ~${DISK_ESTIMATE_MB} MB (via $DISK_CHECK_SOURCE)  |  free ~${DISK_FREE_MB} MB"
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

BACKUP_OK=1
echo "$ARCHIVE"
exit 0
