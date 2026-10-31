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
CONFIG_RAW="${CONFIG_RAW:-https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main}"
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

# caddy logged every coraza "key not found in collection" warning to syslog, gigabytes a day: default logger at ERROR and coraza debug log at errors only
CADDYFILE=/etc/openpanel/caddy/Caddyfile
CORAZA_RULES=/etc/openpanel/caddy/coraza_rules.conf
caddy_restart=no
if [ -f "$CADDYFILE" ] && ! awk '/^{$/{g=1} g && /^}$/{exit} g && /^    log \{/{f=1} END{exit !f}' "$CADDYFILE"; then
    cp -a "$CADDYFILE" "$CADDYFILE.bak-2.0.20"
    tmp=$(mktemp)
    awk '!done && /^{$/ {print; print "    # default logger only, site access logs have their own; the WAF warns \"key not found in collection\" on most requests and filled syslog"; print "    log {"; print "        level ERROR"; print "    }"; print ""; done=1; next} {print}' "$CADDYFILE" > "$tmp"
    # written in place, the caddy container bind-mounts this file
    cat "$tmp" > "$CADDYFILE"
    rm -f "$tmp"
    caddy_restart=yes
fi
if [ -f "$CORAZA_RULES" ] && ! grep -q "^SecDebugLogLevel" "$CORAZA_RULES"; then
    sed -i 's|^# LOGS$|# LOGS\n# errors only, at the default level coraza warns "key not found in collection" on most requests and that filled syslog\nSecDebugLogLevel 1|' "$CORAZA_RULES"
    caddy_restart=yes
fi
if [ "$caddy_restart" = yes ]; then
    # a reload keeps the waf logging at the old level, only a restart picks it up
    if podman exec caddy caddy validate --config /etc/caddy/Caddyfile >/dev/null 2>&1 </dev/null; then
        echo "Restarting caddy with ERROR log level..."
        podman restart caddy >/dev/null 2>&1 </dev/null
    else
        echo "Caddyfile fails caddy validate, restoring $CADDYFILE.bak-2.0.20"
        [ -f "$CADDYFILE.bak-2.0.20" ] && cat "$CADDYFILE.bak-2.0.20" > "$CADDYFILE"
    fi
fi

# LiteSpeed Enterprise is a new webserver choice: it reads Apache style vhosts, runs php in its bundled lsphp and starts on a trial license
CONFIG_RAW="${CONFIG_RAW:-https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main}"
TEMPLATE=/etc/openpanel/docker/compose/1.0/docker-compose.yml
TEMPLATE_ENV=/etc/openpanel/docker/compose/1.0/.env

fetch() {
    local src="$1" dst="$2" mode="$3" tmp
    tmp=$(mktemp)
    if wget --timeout=15 --tries=2 -q -O "$tmp" "$CONFIG_RAW/$src" && [ -s "$tmp" ]; then
        mkdir -p "$(dirname "$dst")"
        cat "$tmp" > "$dst"
        chmod "$mode" "$dst"
        rm -f "$tmp"
        return 0
    fi
    echo "Failed to download $CONFIG_RAW/$src"
    rm -f "$tmp"
    return 1
}

echo "Adding LiteSpeed Enterprise files..."
fetch litespeed/httpd.conf /etc/openpanel/litespeed/httpd.conf 644 || exit 0
fetch litespeed/start.sh /etc/openpanel/litespeed/start.sh 755 || exit 0
fetch nginx/vhosts/1.1/docker_litespeed_domain.conf /etc/openpanel/nginx/vhosts/1.1/docker_litespeed_domain.conf 644 || exit 0
NEW_TEMPLATE=$(mktemp)
trap 'rm -f "$OP_UPDATE_SELF" "$NEW_TEMPLATE"' EXIT
fetch docker/compose/1.0/docker-compose.yml "$NEW_TEMPLATE" 644 || exit 0

# adds the litespeed service and its volume to a compose file, the ports lines are taken from the file's own openlitespeed block so a varnish setup stays as it is
add_litespeed_service() {
    local f="$1" tmp
    grep -q "^  litespeed:" "$f" && return 0
    tmp=$(mktemp)
    if ! python3 - "$NEW_TEMPLATE" "$f" "$tmp" <<'EOF'
import re, sys
new, cur, out = open(sys.argv[1]).read(), open(sys.argv[2]).read(), sys.argv[3]

def block(text, name):
    m = re.search(r'(?ms)^  %s:\n.*?(?=^  \S|^\S|\Z)' % re.escape(name), text)
    return m.group(0) if m else None

ls = block(new, 'litespeed')
if not ls:
    sys.exit(1)
ols = block(cur, 'openlitespeed')
if ols:
    ports = re.search(r'(?ms)^    ports:\n(?:      - .*\n)+', ols)
    if ports:
        ls = re.sub(r'(?ms)^    ports:\n(?:      - .*\n)+', lambda _: ports.group(0), ls, count=1)
    cur = cur.replace(ols, ols + ls, 1)
else:
    cur = re.sub(r'(?m)^services:\n', lambda m: m.group(0) + '\n' + ls, cur, count=1)

vol = block(new.split('\nvolumes:\n', 1)[1] if '\nvolumes:\n' in new else '', 'litespeed_conf')
if vol and not re.search(r'(?m)^  litespeed_conf:', cur.split('\nvolumes:\n', 1)[-1]):
    if '\nvolumes:\n' in cur:
        cur = cur.replace('\nvolumes:\n', '\nvolumes:\n' + vol, 1)
    else:
        cur = cur.rstrip('\n') + '\n\nvolumes:\n' + vol
open(out, 'w').write(cur)
EOF
    then
        echo "Could not add the litespeed service to $f, left as it is."
        rm -f "$tmp"
        return 1
    fi
    if grep -q "^  litespeed:" "$tmp" && grep -q "^  litespeed_conf:" "$tmp"; then
        [ -f "$f.bak-2.0.20" ] || cp -a "$f" "$f.bak-2.0.20"
        # written in place so the file keeps its owner and mode
        cat "$tmp" > "$f"
    else
        echo "Could not add the litespeed service to $f, left as it is."
    fi
    rm -f "$tmp"
}

add_litespeed_env() {
    local f="$1"
    [ -f "$f" ] || return 0
    grep -q "^LITESPEED_VERSION=" "$f" && return 0
    printf '\n# LITESPEED (tag is <release>-lsphpXX, the php version comes with the image)\nLITESPEED_VERSION="latest"\nLITESPEED_CPU="1.0"\nLITESPEED_RAM="1.0G"\nLITESPEED_PIDS="2500"\n' >> "$f"
}

add_litespeed_service "$TEMPLATE"
add_litespeed_env "$TEMPLATE_ENV"

for f in /home/*/docker-compose.yml; do
    [ -f "$f" ] || continue
    home_dir=$(dirname "$f")
    ctx=$(basename "$home_dir")
    [ -f "$home_dir/.env" ] || continue
    grep -q "^  litespeed:" "$f" && [ -f "$home_dir/litespeed.conf" ] && continue
    add_litespeed_service "$f"
    add_litespeed_env "$home_dir/.env"
    if [ ! -f "$home_dir/litespeed.conf" ]; then
        cp /etc/openpanel/litespeed/httpd.conf "$home_dir/litespeed.conf"
        chown --reference="$home_dir/.env" "$home_dir/litespeed.conf"
    fi
    echo "LiteSpeed is now available to $ctx."
done
