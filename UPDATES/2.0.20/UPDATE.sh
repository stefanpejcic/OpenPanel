#!/bin/bash

# opencli update runs this as wget -O - URL | bash, so a command below that reads stdin (podman) would eat the rest of the script and bash would go on mid-line; the rest is saved to a file and run from there with stdin closed
if [ -z "${BASH_SOURCE[0]:-}" ] || [ ! -f "${BASH_SOURCE[0]}" ]; then
    self=$(mktemp) && cat > "$self" && OP_UPDATE_SELF="$self" exec bash "$self" "$@" </dev/null
fi
[ -n "${OP_UPDATE_SELF:-}" ] && trap 'rm -f "$OP_UPDATE_SELF"' EXIT

# openresty.conf was mounted at /etc/openresty/nginx.conf, which openresty never reads, so edits in Server Settings did nothing and their syntax check tested the image's own nginx.conf
# it's now mounted over the file openresty reads, a running openresty is recreated with it only if the account's file passes nginx -t, otherwise that account is left as it is
# shellcheck disable=SC1091
. /usr/local/opencli/lib/podman.sh
OLD_MOUNT="./openresty.conf:/etc/openresty/nginx.conf:ro"
NEW_MOUNT="./openresty.conf:/usr/local/openresty/nginx/conf/nginx.conf:ro"
RESTY_TEST_CONF=/usr/local/openresty/nginx/conf/op-2020-test.conf

openresty_mount() {
    local f="$1" tmp
    tmp=$(mktemp)
    sed "s|^\(\s*\)- \./openresty\.conf:/etc/openresty/nginx\.conf:ro\s*$|\1- $NEW_MOUNT|" "$f" > "$tmp"
    if grep -qF -- "- $NEW_MOUNT" "$tmp" && ! grep -qF -- "- $OLD_MOUNT" "$tmp"; then
        [ -f "$f.bak-2.0.20" ] || cp -a "$f" "$f.bak-2.0.20"
        # written in place so the file keeps its owner and mode
        cat "$tmp" > "$f"
        rm -f "$tmp"
        return 0
    fi
    echo "Could not update the openresty mount in $f, left as it is."
    rm -f "$tmp"
    return 1
}

grep -qF -- "- $OLD_MOUNT" /etc/openpanel/docker/compose/1.0/docker-compose.yml 2>/dev/null \
    && openresty_mount /etc/openpanel/docker/compose/1.0/docker-compose.yml

for f in /home/*/docker-compose.yml; do
    [ -f "$f" ] || continue
    grep -qF -- "- $OLD_MOUNT" "$f" || continue
    home_dir=$(dirname "$f")
    ctx=$(basename "$home_dir")

    if ! podman_ctx_is_running "$ctx" openresty </dev/null; then
        openresty_mount "$f"
        continue
    fi

    # a copy next to the stock nginx.conf resolves the relative mime.types include the way the moved file will
    if ! out=$(podman_ctx "$ctx" exec openresty sh -c "cp /etc/openresty/nginx.conf $RESTY_TEST_CONF && nginx -t -c $RESTY_TEST_CONF; rc=\$?; rm -f $RESTY_TEST_CONF; exit \$rc" 2>&1 </dev/null); then
        echo "$ctx's openresty.conf fails nginx -t, its openresty was left as it is. Fix $home_dir/openresty.conf and run this update again:"
        echo "$out"
        continue
    fi
    openresty_mount "$f" || continue

    echo "Recreating openresty for $ctx with its openresty.conf..."
    sock=$(podman_user_socket "$ctx") || continue
    (cd "$home_dir" && CONTAINER_HOST="$sock" timeout 300 podman-compose up -d --no-deps openresty) >/dev/null 2>&1 </dev/null
    if ! podman_ctx_is_running "$ctx" openresty </dev/null; then
        echo "openresty for $ctx is not running after the change, the previous compose file is $f.bak-2.0.20"
    fi
done

# files-list.sh gives each run its own temp files, it could run in two backup containers at once and fail the backup on a clashing mv
CONFIG_RAW="https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main"
tmp=$(mktemp)
if wget --timeout=15 --tries=2 -q -O "$tmp" "$CONFIG_RAW/backups/scripts/files-list.sh" && grep -q "TMP_SUFFIX" "$tmp"; then
    echo "Updating /etc/openpanel/backups/scripts/files-list.sh..."
    # written in place, backup containers bind-mount this file
    cat "$tmp" > /etc/openpanel/backups/scripts/files-list.sh
    chmod 755 /etc/openpanel/backups/scripts/files-list.sh
else
    echo "Failed to download $CONFIG_RAW/backups/scripts/files-list.sh"
fi
rm -f "$tmp"

# logrotate failed every night (caddy-logs postrotate ran podman exec inside logrotate.service, csf's lfd file was group-writable) and syslog had no size limit, the new opencli server-logrotate fixes all three
opencli server-logrotate
