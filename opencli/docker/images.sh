#!/bin/bash
################################################################################
# Script Name: images.sh
# Description: Check and auto-update Docker images per user.
# Usage: opencli docker-images [--all|<USERNAME>] [--dry-run] [--force-update]
# Author: Stefan Pejcic
# Created: 05.05.2025
# Last Modified: 10.06.2026
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

GLOBAL_LOCK="/tmp/cup_global.lock"
DIGEST_CACHE="/tmp/cup_digest_cache.tmp"
DIGEST_CACHE_LOCK="/tmp/cup_digest_cache.lock"
MAX_PARALLEL=2
PULL_TIMEOUT=600          # seconds before a pull is killed
HEALTH_RETRY_INTERVAL=5   # seconds between health check attempts
HEALTH_MAX_WAIT=30        # seconds total to wait for healthy status
CUP_IMAGE="ghcr.io/sergi0g/cup"

DRY_RUN=false
FORCE_UPDATE=false
MANIFEST_AVAILABLE=false

usage() {
    echo "Usage: opencli docker-images [options]"
    echo ""
    echo "Options:"
    echo "  <USERNAME>       Check (and optionally update) images for a specific user."
    echo "  --all            Check (and optionally update) images for all users."
    echo "  --dry-run        Show what would be updated without making any changes."
    echo "  --force-update   Force update regardless of user auto-update setting."
    echo ""
    echo "Examples:"
    echo "  opencli docker-images stefan"
    echo "  opencli docker-images stefan --force-update"
    echo "  opencli docker-images --all"
    echo "  opencli docker-images --all --dry-run"
    echo "  opencli docker-images --all --force-update"
    exit 1
}

detect_manifest_support() {
    if docker manifest --help > /dev/null 2>&1; then
        MANIFEST_AVAILABLE=true
    else
        MANIFEST_AVAILABLE=false
    fi
}

init_log() {
    local username="$1"
    local log_dir="/home/$username/cup/logs"
    mkdir -p "$log_dir" 2>/dev/null
    LOG_FILE="$log_dir/$(date +%Y-%m-%d).log"
    RUN_START_TS=$(date +%s)

    local flags=""
    [ "$DRY_RUN"      = true ] && flags+=" [DRY-RUN]"
    [ "$FORCE_UPDATE" = true ] && flags+=" [FORCE-UPDATE]"

    {
        echo ""
        echo "════════════════════════════════════════════════════════════════"
        echo "$(date '+%Y-%m-%d %H:%M:%S') ── RUN START (pid $$)${flags}"
        echo "════════════════════════════════════════════════════════════════"
    } >> "$LOG_FILE"
}

log() {
    local level="$1"; shift
    local msg="$*"
    local ts
    ts=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$ts] [$level] $msg" | tee -a "$LOG_FILE"
}

log_info()  { log "INFO " "$@"; }
log_warn()  { log "WARN " "$@"; }
log_error() { log "ERROR" "$@"; }
log_skip()  { log "SKIP " "$@"; }
log_dry()   { log "DRY  " "$@"; }

write_summary() {
    local username="$1"
    local elapsed=$(( $(date +%s) - RUN_START_TS ))

    local update_mode="disabled (check only)"
    [ "$FORCE_UPDATE" = true ]                          && update_mode="force-update (admin override)"
    [ -f "/home/$username/cup/enabled" ]                && update_mode="enabled (user opt-in)"
    [ "$FORCE_UPDATE" = true ] && \
        [ -f "/home/$username/cup/enabled" ]            && update_mode="enabled (user opt-in + force)"

    {
        echo ""
        echo "── SUMMARY ─────────────────────────────────────────────────────"
        printf "  %-20s: %s\n" "User"              "$username"
        printf "  %-20s: %s\n" "Mode"              "$([ "$DRY_RUN" = true ] && echo "dry-run" || echo "live")"
        printf "  %-20s: %s\n" "Auto-update"       "$update_mode"
        printf "  %-20s: %s\n" "Images checked"    "${SUMMARY_CHECKED:-0}"
        printf "  %-20s: %s\n" "Updates available" "${SUMMARY_AVAILABLE:-0}"
        printf "  %-20s: %s\n" "Updates applied"   "${SUMMARY_UPDATED:-0}"
        printf "  %-20s: %s\n" "Skipped (quota)"   "${SUMMARY_QUOTA_SKIP:-0}"
        printf "  %-20s: %s\n" "Skipped (digest)"  "${SUMMARY_DIGEST_SKIP:-0}"
        printf "  %-20s: %s\n" "Failed"            "${SUMMARY_FAILED:-0}"
        printf "  %-20s: %s\n" "Errors"            "${SUMMARY_ERRORS:-0}"
        printf "  %-20s: %s\n" "Elapsed"           "${elapsed}s"
        echo "────────────────────────────────────────────────────────────────"
        echo ""
    } | tee -a "$LOG_FILE"
}

acquire_lock() {
    local lockfile="$1"
    local label="$2"
    if ! mkdir "$lockfile" 2>/dev/null; then
        local owner
        owner=$(cat "$lockfile/pid" 2>/dev/null || echo "unknown")
        echo "ERROR: Could not acquire $label lock (held by pid $owner). Exiting."
        exit 1
    fi
    echo $$ > "$lockfile/pid"
}

release_lock() {
    local lockfile="$1"
    rm -rf "$lockfile" 2>/dev/null
}

user_lock_path() {
    echo "/tmp/cup_user_${1}.lock"
}

cache_get() {
    local image="$1"
    flock -s "$DIGEST_CACHE_LOCK" grep -m1 "^${image}|" "$DIGEST_CACHE" 2>/dev/null | cut -d'|' -f2 || true
}

cache_set() {
    local image="$1"
    local digest="$2"
    (
        flock -x 200
        local tmp
        tmp=$(grep -v "^${image}|" "$DIGEST_CACHE" 2>/dev/null || true)
        echo "$tmp" > "$DIGEST_CACHE"
        echo "${image}|${digest}" >> "$DIGEST_CACHE"
    ) 200>"$DIGEST_CACHE_LOCK"
}

get_remote_digest() {
    local image="$1"

    # Cache hit?
    local cached
    cached=$(cache_get "$image")
    if [ -n "$cached" ]; then
        echo "$cached"
        return 0
    fi

    # Fetch from registry
    local digest
    digest=$(docker manifest inspect --verbose "$image" 2>/dev/null | python3 -c "
import json, sys
data = json.load(sys.stdin)
# manifest inspect --verbose returns either a single object or a list (multi-arch)
if isinstance(data, list):
    data = data[0]
d = (data.get('Descriptor') or data.get('SchemaV2Manifest') or {})
print(d.get('digest', ''))
" 2>/dev/null || true)

    if [ -n "$digest" ]; then
        cache_set "$image" "$digest"
        echo "$digest"
    fi
}

get_local_digest() {
    local ctx="$1"
    local image="$2"
    docker --context="$ctx" image inspect "$image" --format='{{index .RepoDigests 0}}' 2>/dev/null | awk -F'@' '{print $2}' || true
}

image_has_remote_update() {
    local ctx="$1"
    local image="$2"

    if [ "$MANIFEST_AVAILABLE" = false ]; then
        log_warn "docker manifest not available — cannot verify digest for $image. Skipping."
        return 1
    fi

    local remote_digest local_digest
    remote_digest=$(get_remote_digest "$image")

    if [ -z "$remote_digest" ]; then
        log_warn "Could not fetch remote digest for $image — skipping."
        return 1
    fi

    local_digest=$(get_local_digest "$ctx" "$image")

    if [ "$remote_digest" = "$local_digest" ]; then
        return 1    # No update
    fi

    return 0        # Update available
}

check_registry_reachable() {
    if ! timeout 10 bash -c 'echo > /dev/tcp/registry-1.docker.io/443' 2>/dev/null; then
        log_error "Docker Hub registry is unreachable. Aborting."
        return 1
    fi
    return 0
}

validate_compose_file() {
    local ctx="$1"
    local username="$2"
    local compose_file="/home/$username/docker-compose.yml"
    if [ ! -f "$compose_file" ]; then
        log_error "Compose file not found: $compose_file"
        return 1
    fi
    if ! docker --context="$ctx" compose -f "$compose_file" config --quiet 2>/dev/null; then
        log_error "Compose file is invalid: $compose_file"
        return 1
    fi
    return 0
}

check_user_quota() {
    local username="$1"

    local quota_line
    quota_line=$(repquota -a 2>/dev/null | awk -v u="$username" '$1 == u {print}' | head -1)

    if [ -z "$quota_line" ]; then
        log_warn "Could not read repquota for $username — skipping quota enforcement."
        QUOTA_BLOCK_AVAIL=999999999
        QUOTA_INODE_AVAIL=999999999
        return 0
    fi

    # repquota columns: 1=user 2=flags 3=blk_used 4=blk_soft 5=blk_hard 6=blk_grace 7=ino_used 8=ino_soft 9=ino_hard
    local block_used block_soft block_hard inode_used inode_soft inode_hard
    block_used=$(echo "$quota_line" | awk '{print $3}')
    block_soft=$(echo "$quota_line" | awk '{print $4}')
    block_hard=$(echo "$quota_line" | awk '{print $5}')
    inode_used=$(echo "$quota_line" | awk '{print $7}')
    inode_soft=$(echo "$quota_line" | awk '{print $8}')
    inode_hard=$(echo "$quota_line" | awk '{print $9}')

    local block_limit inode_limit
    if [ "${block_hard:-0}" -gt 0 ] 2>/dev/null; then
        block_limit=$block_hard
    elif [ "${block_soft:-0}" -gt 0 ] 2>/dev/null; then
        block_limit=$block_soft
    else
        block_limit=0
    fi

    if [ "${inode_hard:-0}" -gt 0 ] 2>/dev/null; then
        inode_limit=$inode_hard
    elif [ "${inode_soft:-0}" -gt 0 ] 2>/dev/null; then
        inode_limit=$inode_soft
    else
        inode_limit=0
    fi

    QUOTA_BLOCK_AVAIL=$([ "$block_limit" -gt 0 ] && echo $(( block_limit - block_used )) || echo 999999999)
    QUOTA_INODE_AVAIL=$([ "$inode_limit" -gt 0 ] && echo $(( inode_limit - inode_used )) || echo 999999999)

    if [ "$QUOTA_BLOCK_AVAIL" -le 0 ] || [ "$QUOTA_INODE_AVAIL" -le 0 ]; then
        log_error "User $username is already at or over quota (blocks avail: ${QUOTA_BLOCK_AVAIL}KB, inodes avail: ${QUOTA_INODE_AVAIL})."
        return 1
    fi

    log_info "Quota — blocks available: ${QUOTA_BLOCK_AVAIL}KB, inodes available: ${QUOTA_INODE_AVAIL}"
    return 0
}

get_image_size_kb() {
    local ctx="$1"
    local image="$2"
    docker --context="$ctx" image inspect "$image" \
        --format='{{.Size}}' 2>/dev/null | awk '{printf "%d", $1/1024}'
}

wait_for_healthy() {
    local ctx="$1"
    local container_id="$2"
    local waited=0

    while [ "$waited" -lt "$HEALTH_MAX_WAIT" ]; do
        local health status
        health=$(docker --context="$ctx" inspect \
            --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' \
            "$container_id" 2>/dev/null)
        status=$(docker --context="$ctx" inspect \
            --format='{{.State.Status}}' \
            "$container_id" 2>/dev/null)

        if   [ "$health" = "healthy" ]; then
            return 0
        elif [ "$health" = "none" ] && [ "$status" = "running" ]; then
            return 0
        elif [ "$health" = "unhealthy" ]; then
            return 1
        fi

        sleep "$HEALTH_RETRY_INTERVAL"
        waited=$(( waited + HEALTH_RETRY_INTERVAL ))
    done

    return 1
}

update_image() {
    local ctx="$1"
    local username="$2"
    local image="$3"
    local service="$4"
    local compose_file="/home/$username/docker-compose.yml"

    log_info "[$service] Checking $image ..."

    if ! image_has_remote_update "$ctx" "$image"; then
        log_info "[$service] Digest unchanged on registry — no update needed."
        SUMMARY_DIGEST_SKIP=$(( ${SUMMARY_DIGEST_SKIP:-0} + 1 ))
        return
    fi

    log_info "[$service] Remote digest differs — update available."
    SUMMARY_AVAILABLE=$(( ${SUMMARY_AVAILABLE:-0} + 1 ))

    local image_size_kb
    image_size_kb=$(get_image_size_kb "$ctx" "$image")
    log_info "[$service] Current image size: ${image_size_kb}KB — quota available: ${QUOTA_BLOCK_AVAIL}KB"

    if [ "${image_size_kb:-0}" -gt "${QUOTA_BLOCK_AVAIL:-0}" ]; then
        log_skip "[$service] Not enough quota to pull $image (need ${image_size_kb}KB, have ${QUOTA_BLOCK_AVAIL}KB)."
        SUMMARY_QUOTA_SKIP=$(( ${SUMMARY_QUOTA_SKIP:-0} + 1 ))
        return
    fi

    if [ "$DRY_RUN" = true ]; then
        log_dry "[$service] Would pull $image → compose down/up '$service' → health check."
        SUMMARY_UPDATED=$(( ${SUMMARY_UPDATED:-0} + 1 ))
        return
    fi

    local old_digest
    old_digest=$(get_local_digest "$ctx" "$image")

    log_info "[$service] Pulling $image ..."
    if ! timeout "$PULL_TIMEOUT" docker --context="$ctx" pull "$image" >> "$LOG_FILE" 2>&1; then
        log_error "[$service] Pull failed or timed out for $image."
        SUMMARY_FAILED=$(( ${SUMMARY_FAILED:-0} + 1 ))
        SUMMARY_ERRORS=$(( ${SUMMARY_ERRORS:-0} + 1 ))
        return
    fi

    log_info "[$service] Stopping service ..."
    if ! docker --context="$ctx" compose -f "$compose_file" stop "$service" >> "$LOG_FILE" 2>&1; then
        log_error "[$service] Failed to stop service."
        SUMMARY_FAILED=$(( ${SUMMARY_FAILED:-0} + 1 ))
        SUMMARY_ERRORS=$(( ${SUMMARY_ERRORS:-0} + 1 ))
        return
    fi
    docker --context="$ctx" compose -f "$compose_file" rm -f "$service" >> "$LOG_FILE" 2>&1 || true

    log_info "[$service] Starting service with new image ..."
    if ! docker --context="$ctx" compose -f "$compose_file" up -d "$service" >> "$LOG_FILE" 2>&1; then
        log_error "[$service] Failed to start service after update."
        SUMMARY_FAILED=$(( ${SUMMARY_FAILED:-0} + 1 ))
        SUMMARY_ERRORS=$(( ${SUMMARY_ERRORS:-0} + 1 ))
        return
    fi

    local container_id
    container_id=$(docker --context="$ctx" compose -f "$compose_file" ps -q "$service" 2>/dev/null | head -1)

    if wait_for_healthy "$ctx" "$container_id"; then
        log_info "[$service] Healthy after update. ✓"
        SUMMARY_UPDATED=$(( ${SUMMARY_UPDATED:-0} + 1 ))

        QUOTA_BLOCK_AVAIL=$(( QUOTA_BLOCK_AVAIL - image_size_kb ))

        if [ -n "$old_digest" ]; then
            log_info "[$service] Removing old digest: $old_digest"
            docker --context="$ctx" rmi "$old_digest" >> "$LOG_FILE" 2>&1 \
                || log_warn "[$service] Could not remove $old_digest (may be shared)."
        fi
    else
        log_error "[$service] Failed health check after update — leaving new image in place."
        SUMMARY_FAILED=$(( ${SUMMARY_FAILED:-0} + 1 ))
    fi
}

run_for_user() {
    local username="$1"

    # 1. lock first
    local user_lock
    user_lock=$(user_lock_path "$username")
    if ! mkdir "$user_lock" 2>/dev/null; then
        local owner
        owner=$(cat "$user_lock/pid" 2>/dev/null || echo "unknown")
        echo "SKIP: $username already being processed (pid $owner)."
        return
    fi
    echo $$ > "$user_lock/pid"
    # shellcheck disable=SC2064
    trap "release_lock '$user_lock'" RETURN INT TERM

    # 2. check if skip_cup.flag
    if [ -f "/home/$username/skip_cup.flag" ]; then
        echo "SKIP: $username has skip_cup.flag — skipping entirely."
        return
    fi

    init_log "$username"

    SUMMARY_CHECKED=0
    SUMMARY_AVAILABLE=0
    SUMMARY_UPDATED=0
    SUMMARY_QUOTA_SKIP=0
    SUMMARY_DIGEST_SKIP=0
    SUMMARY_FAILED=0
    SUMMARY_ERRORS=0

    local flags=""
    [ "$DRY_RUN"      = true ] && flags+=" [DRY-RUN]"
    [ "$FORCE_UPDATE" = true ] && flags+=" [FORCE-UPDATE]"
    log_info "Starting for user: $username$flags"

    # 3. check if context exists
    if ! docker context inspect "$username" > /dev/null 2>&1; then
        log_error "Docker context '$username' does not exist."
        SUMMARY_ERRORS=$(( SUMMARY_ERRORS + 1 ))
        write_summary "$username"
        return 1
    fi

    # 4. check if all services are stopped = suspended user
    local running_containers
    running_containers=$(docker --context="$username" ps -q 2>/dev/null)
    if [ -z "$running_containers" ]; then
        log_skip "No running containers — skipping."
        write_summary "$username"
        return
    fi

    # 5. validate compose file so we dont get errors on compose up later
    if ! validate_compose_file "$username" "$username"; then
        SUMMARY_ERRORS=$(( SUMMARY_ERRORS + 1 ))
        write_summary "$username"
        return 1
    fi

    local compose_file="/home/$username/docker-compose.yml"

    # 6. list active services for user
    local active_services
    active_services=$(docker --context="$username" compose -f "$compose_file" ps \
        --filter "status=running" --format "{{.Service}}" 2>/dev/null)

    if [ -z "$active_services" ]; then
        log_skip "No active compose services — skipping."
        write_summary "$username"
        return
    fi

    log_info "Active services: $(echo "$active_services" | tr '\n' ' ')"

    # 7. check if docker hub is reachable
    if ! check_registry_reachable; then
        SUMMARY_ERRORS=$(( SUMMARY_ERRORS + 1 ))
        write_summary "$username"
        return 1
    fi

    # 8. get socket for step 10
    local user_id mount_flag
    user_id=$(stat -c %u "/home/$username" 2>/dev/null)
    if [ -n "$user_id" ] && [ -S "/hostfs/run/user/$user_id/docker.sock" ]; then
        mount_flag="-v /hostfs/run/user/$user_id/docker.sock:/var/run/docker.sock:ro"
    elif [ -S "/var/run/docker.sock" ]; then
        mount_flag="-v /var/run/docker.sock:/var/run/docker.sock:ro"
    else
        log_error "No Docker socket available."
        SUMMARY_ERRORS=$(( SUMMARY_ERRORS + 1 ))
        write_summary "$username"
        return 1
    fi

    # 9. build image→service map for active services
    local compose_json
    compose_json=$(docker --context="$username" compose -f "$compose_file" \
        config --format json 2>/dev/null)

    declare -A image_to_service
    for svc in $active_services; do
        local svc_image
        svc_image=$(echo "$compose_json" | python3 -c "
import json, sys
cfg = json.load(sys.stdin)
svc = cfg.get('services', {}).get('$svc', {})
print(svc.get('image', ''))
" 2>/dev/null)
        if [ -n "$svc_image" ]; then
            image_to_service["$svc_image"]="$svc"
        fi
    done

    SUMMARY_CHECKED=${#image_to_service[@]}
    log_info "Active images to check: $SUMMARY_CHECKED"

    # 10. run cup
    local cup_output_dir="/home/$username/docker-data/cup"
    mkdir -p "$cup_output_dir"
    local cup_json="$cup_output_dir/cup.json"

    log_info "Running cup check ..."
    if ! docker --context="$username" run --rm $mount_flag "$CUP_IMAGE" check -r > "$cup_json" 2>>"$LOG_FILE"; then
        log_error "cup check failed."
        SUMMARY_ERRORS=$(( SUMMARY_ERRORS + 1 ))
        write_summary "$username"
        return 1
    fi
    docker --context="$username" rmi -f "$CUP_IMAGE" >> "$LOG_FILE" 2>&1 || true
    log_info "cup output saved to $cup_json"

    # 11. extract images cup flagged as having updates
    local cup_flagged
    cup_flagged=$(python3 -c "
import json, sys
data = json.load(open('$cup_json'))
for img, info in data.get('images', {}).items():
    if info.get('has_update', False):
        print(img)
" 2>/dev/null)

    # 12. autoupdate?
    local do_update=false
    if [ "$FORCE_UPDATE" = true ]; then
        do_update=true
        log_info "Force-update active — updates will proceed regardless of user setting."
    elif [ -f "/home/$username/cup/enabled" ]; then
        do_update=true
        log_info "Auto-update enabled for $username."
    else
        log_info "Auto-update not enabled — check only (pass --force-update to override)."
    fi

    # 13. if no autoupdate, just log cup and exit
    if [ "$do_update" = false ]; then
        local count=0
        for img in $cup_flagged; do
            [ -n "${image_to_service[$img]+_}" ] && count=$(( count + 1 ))
        done
        SUMMARY_AVAILABLE=$count
        log_info "Updates available for active services: $count (not applying — check only)."
        write_summary "$username"
        return
    fi

    # 14. check quota for user to avoid reaching limits on pull
    if ! check_user_quota "$username"; then
        SUMMARY_ERRORS=$(( SUMMARY_ERRORS + 1 ))
        write_summary "$username"
        return 1
    fi

    # 15. process each image cup flagged, limited to active services
    for img in $cup_flagged; do
        local svc="${image_to_service[$img]:-}"
        if [ -z "$svc" ]; then
            log_info "Image $img flagged by cup but not in active services — skipping."
            continue
        fi
        update_image "$username" "$username" "$img" "$svc"
    done

    # 16. prune remaining dangling images
    if [ "$DRY_RUN" = false ]; then
        log_info "Pruning dangling images ..."
        docker --context="$username" image prune -f >> "$LOG_FILE" 2>&1 || true
    fi

    # 17. create summary to be later used on notifications
    write_summary "$username"
}

# ALL USERS
run_for_all_users() {
    local contexts
    contexts=$(docker context ls --format '{{.Name}}' | grep -v '^default$')

    local pids=()

    for ctx in $contexts; do
        while [ "${#pids[@]}" -ge "$MAX_PARALLEL" ]; do
            local new_pids=()
            for pid in "${pids[@]}"; do
                kill -0 "$pid" 2>/dev/null && new_pids+=("$pid")
            done
            pids=("${new_pids[@]}")
            [ "${#pids[@]}" -ge "$MAX_PARALLEL" ] && sleep 1
        done

        run_for_user "$ctx" &
        pids+=($!)
    done

    for pid in "${pids[@]}"; do
        wait "$pid"
    done

    echo "Done processing all users."
}

TARGET=""
for arg in "$@"; do
    case "$arg" in
        --dry-run)      DRY_RUN=true ;;
        --force-update) FORCE_UPDATE=true ;;
        --all)          TARGET="--all" ;;
        --help|-h)      usage ;;
        -*)             echo "Unknown option: $arg"; usage ;;
        *)              TARGET="$arg" ;;
    esac
done

[ -z "$TARGET" ] && usage

detect_manifest_support
log_info_global() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO ] $*"; }

if [ "$MANIFEST_AVAILABLE" = true ]; then
    log_info_global "docker manifest is available — digest pre-check enabled (no-pull comparison)."
else
    log_info_global "docker manifest not available — images without digest verification will be skipped."
fi

: > "$DIGEST_CACHE"
: > "$DIGEST_CACHE_LOCK"

if [ "$TARGET" = "--all" ]; then
    acquire_lock "$GLOBAL_LOCK" "global"
    trap "release_lock '$GLOBAL_LOCK'; rm -f '$DIGEST_CACHE' '$DIGEST_CACHE_LOCK'" EXIT
    run_for_all_users
else
    trap "rm -f '$DIGEST_CACHE' '$DIGEST_CACHE_LOCK'" EXIT
    run_for_user "$TARGET"
fi
