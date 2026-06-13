#!/bin/bash
################################################################################
# Script Name: user/backup.sh
# Description: Creates a full, self-contained .tar.gz backup of a single
#              OpenPanel user account.  Home directory is streamed in a single
#              tar pass — no intermediate copy is written to disk.
#
#              Archive layout:
#                manifest.env           — identity, plan, source IP, format ver
#                db/                    — plan / user / domains / sites as SQL
#                system/                — passwd, group, shadow entries
#                features/              — plan feature-set file
#                core/                  — /etc/openpanel/.../core/users/<user>/
#                ftp/                   — FTP users.list + context config
#                caddy/                 — vhosts, ssl, domlogs, waf, stats
#                bind/zones/            — BIND DNS zone files
#                docker/                — running container list, apparmor, uid
#                mail_external/         — only present when mail lives outside
#                                         /home (admin.ini override)
#                homedir/               — /home/CONTEXT streamed directly;
#                                         always named homedir/ so the archive
#                                         is independent of the username
#
#              Default output:
#                /home/CONTEXT/docker-data/volumes/CONTEXT_html_data/_data/_backups/
#              Accessible via OpenPanel file manager / GUI.
#              _backups/ is automatically excluded from the home tree so
#              archives never nest inside each other.
#              Override the output directory with --output.
#
# Usage: opencli user-backup --account <USER> [--output <DIR>] [--dry-run] [--quiet]
# Author: Stefan Pejcic
# Based on: user/transfer.sh (openpanel.com)
################################################################################

set -o pipefail

pid=$$
start_time=$(date +%s)
BACKUP_FORMAT_VERSION="1"

USERNAME=""
CUSTOM_OUTPUT=""
QUIET=0
DRY_RUN=0

# Runtime tracking — populated during the backup, printed in summary
WARNINGS=()
BACKUP_DOMAINS=()
BACKUP_FTP_COUNT=0
DISK_ESTIMATE_MB=0
DISK_FREE_MB=0
DISK_CHECK_SOURCE="unknown"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --account)  USERNAME="$2"; shift 2 ;;
        --output)   CUSTOM_OUTPUT="$2"; shift 2 ;;
        --dry-run)  DRY_RUN=1; shift ;;
        --quiet)    QUIET=1; shift ;;
        *) echo "[ERROR] Unknown option: $1"; exit 1 ;;
    esac
done

[[ -z "$USERNAME" ]] && {
    echo "Usage: opencli user-backup --account <USER> [--output <DIR>] [--dry-run] [--quiet]"
    exit 1
}

# ---------------------------------------------------------------------------
# DB connection  (provides $config_file and $mysql_database)
# ---------------------------------------------------------------------------
DB_CONFIG_FILE="/usr/local/opencli/db.sh"
[[ -f "$DB_CONFIG_FILE" ]] || { echo "[ERROR] DB config not found: $DB_CONFIG_FILE"; exit 1; }
# shellcheck disable=SC1090
. "$DB_CONFIG_FILE"
mysql_q() { mysql --defaults-extra-file="$config_file" -D "$mysql_database" -N -s -e "$1"; }

# ---------------------------------------------------------------------------
# Resolve account identity
# ---------------------------------------------------------------------------
USER_ID=$(mysql_q "SELECT id FROM users WHERE username = '$USERNAME';")
[[ -z "$USER_ID" ]] && { echo "[ERROR] No user found: $USERNAME"; exit 1; }

CONTEXT=$(mysql_q "SELECT server FROM users WHERE username = '$USERNAME';")
[[ -z "$CONTEXT" ]] && { echo "[ERROR] Could not resolve context (server) for '$USERNAME'"; exit 1; }

PLAN_ID=$(mysql_q "SELECT plan_id FROM users WHERE id = $USER_ID;")
PLAN_NAME=$(mysql_q "SELECT name FROM plans WHERE id = $PLAN_ID;")
PLAN_FEATURE_SET=$(mysql_q "SELECT feature_set FROM plans WHERE id = $PLAN_ID;")

SYS_UID=$(awk -F: -v u="$CONTEXT" '$1==u{print $3}' /etc/passwd)
SYS_GID=$(awk -F: -v u="$CONTEXT" '$1==u{print $4}' /etc/passwd)

DOMAIN_IDS=$(mysql_q "SELECT domain_id FROM domains WHERE user_id = $USER_ID;")
DOMAIN_COUNT=0
[[ -n "$DOMAIN_IDS" ]] && DOMAIN_COUNT=$(echo "$DOMAIN_IDS" | wc -l | tr -d ' ')

SITE_COUNT=0
if [[ -n "$DOMAIN_IDS" ]]; then
    DOMAIN_ID_LIST=$(echo "$DOMAIN_IDS" | paste -sd "," -)
    SITE_COUNT=$(mysql_q "SELECT COUNT(*) FROM sites WHERE domain_id IN ($DOMAIN_ID_LIST);" 2>/dev/null || echo 0)
fi

# External mail — fires only when admin set email_storage_location to a path
# outside /home.  Default: mail lives in /home/CONTEXT/mail/ and is captured
# automatically with the home directory.
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

# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------
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

# ---------------------------------------------------------------------------
# Disk space check
# Uses repquota to read how much disk the account actually occupies; that
# figure is our estimate for how much space the backup will need.
# Falls back to du if repquota is unavailable or returns nothing.
# Checks free space at the destination filesystem — not at /home — because
# they may be on different partitions.
# ---------------------------------------------------------------------------
check_disk_space() {
    log "Checking disk space ..."
    mkdir -p "$DEST_DIR" 2>/dev/null || true

    # --- Step 1: how much disk does this account use? (= backup size estimate)
    local used_kb=0

    if command -v repquota &>/dev/null; then
        # repquota -au  reports all quota-enabled filesystems, 1 KB blocks
        # Output line format:
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

# ---------------------------------------------------------------------------
# DRY RUN — analyse and report; exit without writing anything
# ---------------------------------------------------------------------------
if [[ $DRY_RUN -eq 1 ]]; then
    mkdir -p "$DEST_DIR" 2>/dev/null || true

    # Disk estimate for dry-run (same logic, non-fatal)
    DRY_USED_KB=0; DRY_USED_SOURCE="unknown"
    if command -v repquota &>/dev/null; then
        rq=$(repquota -au 2>/dev/null | awk -v u="$CONTEXT" '$1==u {print $3; exit}')
        if [[ "$rq" =~ ^[0-9]+$ && "$rq" -gt 0 ]]; then
            DRY_USED_KB="$rq"; DRY_USED_SOURCE="repquota"
        fi
    fi
    if [[ "$DRY_USED_KB" -eq 0 && -d "/home/$CONTEXT" ]]; then
        DRY_USED_KB=$(du -sk --exclude="docker-data/volumes/${CONTEXT}_html_data/_data/_backups" "/home/$CONTEXT" 2>/dev/null | cut -f1)
        DRY_USED_SOURCE="du"
    fi
    DRY_USED_MB=$(( ${DRY_USED_KB:-0} / 1024 ))
    DRY_FREE_KB=$(df -k "$DEST_DIR" 2>/dev/null | awk 'NR==2 {print $4}')
    DRY_FREE_MB=$(( ${DRY_FREE_KB:-0} / 1024 ))

    echo ""
    echo "╔══════════════════════════════════════════════════════════╗"
    echo "║  DRY RUN — Backup of '$USERNAME'"
    echo "║  No files will be written."
    echo "╚══════════════════════════════════════════════════════════╝"
    echo ""
    printf "  %-22s %s\n" "Username:"   "$USERNAME"
    printf "  %-22s %s\n" "Context:"    "$CONTEXT"
    printf "  %-22s %s\n" "Plan:"       "${PLAN_NAME:-none}"
    printf "  %-22s %s\n" "UID / GID:"  "${SYS_UID:-?} / ${SYS_GID:-?}"
    echo ""

    echo "  Database:"
    printf "    %-20s %s\n" "users row:"  "1"
    printf "    %-20s %s\n" "plans row:"  "1  (\"${PLAN_NAME}\")"
    printf "    %-20s %s\n" "domains:"    "$DOMAIN_COUNT"
    printf "    %-20s %s\n" "sites:"      "$SITE_COUNT"
    echo ""

    echo "  Domains:"
    ALL_DOMAINS_DRY=$(opencli domains-user "$USERNAME" --docroot --php_version 2>/dev/null)
    if [[ "$ALL_DOMAINS_DRY" == *"No domains found"* || -z "$ALL_DOMAINS_DRY" ]]; then
        echo "    (none)"
    else
        while IFS=$'\t ' read -r domain docroot php_version; do
            [[ -z "$domain" ]] && continue
            cm="-"; zm="-"; sm="-"
            [[ -f "/etc/openpanel/caddy/domains/$domain.conf" ]] && cm="✓"
            [[ -f "/etc/bind/zones/$domain.zone" ]]              && zm="✓"
            { [[ -d "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/$domain" ]] || [[ -d "/etc/openpanel/caddy/ssl/custom/$domain" ]]; } && sm="✓"
            printf "    %-42s [caddy %s] [zone %s] [ssl %s]\n" "$domain" "$cm" "$zm" "$sm"
        done <<< "$ALL_DOMAINS_DRY"
    fi
    echo ""

    echo "  FTP accounts:"
    FTP_LIST_DRY="/etc/openpanel/ftp/users/$CONTEXT/users.list"
    FTP_COUNT_DRY=0
    [[ -f "$FTP_LIST_DRY" ]] && FTP_COUNT_DRY=$(grep -c '.' "$FTP_LIST_DRY" 2>/dev/null || echo 0)
    if [[ $FTP_COUNT_DRY -eq 0 ]]; then
        echo "    (none)"
    else
        while IFS='|' read -r ftpuser _; do
            [[ -n "$ftpuser" ]] && echo "    - $ftpuser"
        done < "$FTP_LIST_DRY"
    fi
    echo ""

    echo "  Feature set:"
    if [[ -n "$PLAN_FEATURE_SET" ]]; then
        FEAT_FILE="/etc/openpanel/openpanel/features/${PLAN_FEATURE_SET}.txt"
        FEAT_S="[not found — will be skipped]"
        [[ -f "$FEAT_FILE" ]] && FEAT_S="[found ✓]"
        printf "    %-30s %s\n" "$PLAN_FEATURE_SET" "$FEAT_S"
    else
        echo "    (none)"
    fi
    echo ""

    echo "  Home directory:"
    if [[ -d "/home/$CONTEXT" ]]; then
        printf "    %-22s %s\n" "Path:"       "/home/$CONTEXT/"
        printf "    %-22s %s MB  (via %s)\n"  "Estimated size:" "$DRY_USED_MB" "$DRY_USED_SOURCE"
        printf "    %-22s %s\n" "Excluding:"  "_data/_backups/ (previous backups)"
    else
        echo "    [!] /home/$CONTEXT not found — home files would NOT be archived"
    fi
    echo ""

    [[ -n "$MAIL_EXTERNAL_PATH" ]] && {
        echo "  External mail store:"
        printf "    %-22s %s  (%s)\n" "Path:" "$MAIL_EXTERNAL_PATH" "$MAIL_EXTERNAL_SIZE"
        echo "    (captured as mail_external/ nested tar)"
        echo ""
    }

    echo "  Disk space check:"
    printf "    %-22s ~%s MB  (from %s)\n"  "Estimated need:"  "$DRY_USED_MB"  "$DRY_USED_SOURCE"
    printf "    %-22s ~%s MB  at %s\n"       "Free at dest:"    "$DRY_FREE_MB"  "$DEST_DIR"
    if [[ "${DRY_FREE_KB:-0}" -lt $(( ${DRY_USED_KB:-1} / 3 )) ]]; then
        echo "    [✘] WOULD ABORT: insufficient space (free < 1/3 of estimated source)."
    elif [[ "${DRY_FREE_KB:-0}" -lt "${DRY_USED_KB:-0}" ]]; then
        echo "    [!] Low — compressed archive is usually smaller, but space is tight."
    else
        echo "    [✓] OK"
    fi
    echo ""

    echo "  Archive destination: $ARCHIVE"
    echo ""
    echo "  No files have been written."
    echo "  Re-run without --dry-run to create the backup."
    echo ""
    exit 0
fi

# ---------------------------------------------------------------------------
# REAL BACKUP
# ---------------------------------------------------------------------------
log "Backup started  log: $log_file  (PID: $pid)"
log "Account: $USERNAME  Context: $CONTEXT  UID:${SYS_UID:-?} GID:${SYS_GID:-?}  Plan: ${PLAN_NAME:-none}"

check_disk_space

# ---------------------------------------------------------------------------
# Staging area — small metadata only; home is streamed live
# ---------------------------------------------------------------------------
STAGE=$(mktemp -d "/tmp/opbackup_${base_name}.XXXXXX")
trap 'rm -rf "$STAGE"' EXIT
mkdir -p "$STAGE"/{db,system,features,core,ftp,\
caddy/domains,caddy/ssl/acme,caddy/ssl/custom,caddy/domlogs,caddy/waf,caddy/stats,\
bind/zones,docker}

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
    mysql_q "
      SELECT site_name, domain_id, admin_email, version, created_date, type, ports, path
      FROM sites WHERE domain_id IN ($DOMAIN_ID_LIST);" > "$STAGE/db/sites.tsv"
    if [[ -s "$STAGE/db/sites.tsv" ]]; then
        awk 'BEGIN{FS="\t";
          print "INSERT INTO sites (site_name, domain_id, admin_email, version, created_date, type, ports, path) VALUES"}
          {printf "('\''%s'\'', %s, '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', %s, '\''%s'\''),\n",
            $1,$2,$3,$4,$5,$6,$7,$8}
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

# --- system user ---
log "Capturing system user ($CONTEXT) ..."
awk -F: -v u="$CONTEXT" '$1==u{print}' /etc/passwd > "$STAGE/system/passwd.user"
awk -F: -v u="$CONTEXT" 'BEGIN{gid=""}
    $1==u{gid=$4}
    $3==gid{print}
    $1==u{print}' /etc/group | sort -u > "$STAGE/system/group.user"
grep "^${CONTEXT}:" /etc/shadow 2>/dev/null > "$STAGE/system/shadow.user" || true

# --- feature set ---
FEATURES_DIR="/etc/openpanel/openpanel/features"
[[ -n "$PLAN_FEATURE_SET" && -f "$FEATURES_DIR/${PLAN_FEATURE_SET}.txt" ]] && cp -a "$FEATURES_DIR/${PLAN_FEATURE_SET}.txt" "$STAGE/features/"

# --- per-user core config ---
CORE_DIR="/etc/openpanel/openpanel/core/users/$USERNAME"
[[ -d "$CORE_DIR" ]] && cp -a "$CORE_DIR/." "$STAGE/core/"

# --- FTP ---
FTP_DIR="/etc/openpanel/ftp/users/$CONTEXT"
if [[ -d "$FTP_DIR" ]]; then
    cp -a "$FTP_DIR/." "$STAGE/ftp/"
    BACKUP_FTP_COUNT=$(grep -c '.' "$FTP_DIR/users.list" 2>/dev/null || echo 0)
fi

# --- per-domain caddy + bind assets ---
if [[ -f "$STAGE/db/domains.list" ]]; then
    log "Collecting per-domain caddy + bind assets ..."
    while IFS=$'\t ' read -r domain docroot php_version; do
        [[ -z "$domain" ]] && continue
        [[ -f "/etc/openpanel/caddy/domains/$domain.conf" ]] && cp -a "/etc/openpanel/caddy/domains/$domain.conf" "$STAGE/caddy/domains/"
        [[ -f "/var/log/caddy/domlogs/$domain" ]] && cp -a "/var/log/caddy/domlogs/$domain" "$STAGE/caddy/domlogs/"
        [[ -f "/var/log/caddy/coraza_waf/$domain.log" ]] && cp -a "/var/log/caddy/coraza_waf/$domain.log" "$STAGE/caddy/waf/"
        [[ -f "/etc/bind/zones/$domain.zone" ]] && cp -a "/etc/bind/zones/$domain.zone" "$STAGE/bind/zones/"
        [[ -d "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/$domain" ]] && cp -a "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/$domain" "$STAGE/caddy/ssl/acme/"
        [[ -d "/etc/openpanel/caddy/ssl/custom/$domain" ]] && cp -a "/etc/openpanel/caddy/ssl/custom/$domain" "$STAGE/caddy/ssl/custom/"
    done < "$STAGE/db/domains.list"
fi
[[ -d "/var/log/caddy/stats/$USERNAME" ]] && cp -a "/var/log/caddy/stats/$USERNAME" "$STAGE/caddy/stats/"

# --- docker metadata ---
if [[ -f "/home/$CONTEXT/docker-compose.yml" ]]; then
    containers=$(docker --context="$CONTEXT" ps -a --format "{{.Names}}" 2>/dev/null | tr '\n' ' ' | sed 's/ $//')
    echo "$CONTEXT: ${containers:-no containers}" > "$STAGE/docker/containers.txt"
fi
[[ -f "/etc/apparmor.d/home.$CONTEXT.bin.rootlesskit" ]] && cp -a "/etc/apparmor.d/home.$CONTEXT.bin.rootlesskit" "$STAGE/docker/apparmor.profile"
echo "${SYS_UID}" > "$STAGE/docker/uid.txt"

# --- external mail store ---
if [[ -n "$MAIL_EXTERNAL_PATH" ]]; then
    log "Archiving external mail store: $MAIL_EXTERNAL_PATH ..."
    mkdir -p "$STAGE/mail_external"
    tar -C "$(dirname "$MAIL_EXTERNAL_PATH")" --numeric-owner --acls --xattrs -cf "$STAGE/mail_external/mail.tar" "$(basename "$MAIL_EXTERNAL_PATH")" 2>>"$log_file" || warn "Failed to archive external mail store (non-fatal)."
    echo "$MAIL_EXTERNAL_PATH" > "$STAGE/mail_external/path.txt"
fi

# ---------------------------------------------------------------------------
# Single streaming tar pass
#   -C /home CONTEXT          streams home directly
#   --transform               renames CONTEXT/ → homedir/
#   --exclude                 blocks _backups/ from archiving itself
#   -C STAGE items...         appends all staged metadata
# ---------------------------------------------------------------------------
STAGE_ITEMS=()
for item in manifest.env db system features core ftp caddy bind docker; do
    [[ -e "$STAGE/$item" ]] && STAGE_ITEMS+=("$item")
done
[[ -d "$STAGE/mail_external" ]] && STAGE_ITEMS+=("mail_external")

ESCAPED_CTX=$(printf '%s\n' "$CONTEXT" | sed 's|[|\\.*^$()+?{}]|\\&|g')

TAR_ARGS=(
    -czf "$ARCHIVE"
    --numeric-owner --acls --xattrs
    "--exclude=${CONTEXT}/docker-data/volumes/${CONTEXT}_html_data/_data/_backups"
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

# ---------------------------------------------------------------------------
# Permissions + ownership
# ---------------------------------------------------------------------------
chmod 640 "$ARCHIVE"
BACKUP_UID=$(id -u "$CONTEXT" 2>/dev/null)
BACKUP_GID=$(id -g "$CONTEXT" 2>/dev/null)
if [[ -n "$BACKUP_UID" && -n "$BACKUP_GID" ]]; then
    chown "${BACKUP_UID}:${BACKUP_GID}" "$ARCHIVE" 2>/dev/null && log "Ownership: $CONTEXT (${BACKUP_UID}:${BACKUP_GID})" || warn "Could not chown archive to $CONTEXT (non-fatal)."
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
end_time=$(date +%s)
elapsed=$(( end_time - start_time ))
elapsed_h=$(( elapsed / 3600 ))
elapsed_m=$(( (elapsed % 3600) / 60 ))
elapsed_s=$(( elapsed % 60 ))

ARCHIVE_SIZE=$(du -h "$ARCHIVE" 2>/dev/null | cut -f1)

DOMAIN_LIST_STR="(none)"
[[ ${#BACKUP_DOMAINS[@]} -gt 0 ]] && DOMAIN_LIST_STR="${BACKUP_DOMAINS[*]}"

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
slog "$(printf "  %-24s %s\n"   "Size:"          "$ARCHIVE_SIZE")"
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

echo "$ARCHIVE"
exit 0
