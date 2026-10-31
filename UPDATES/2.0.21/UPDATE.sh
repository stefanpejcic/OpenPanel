#!/bin/bash

# opencli update runs this as wget -O - URL | bash, so a command below that reads stdin (podman) would eat the rest of the script and bash would go on mid-line; the rest is saved to a file and run from there with stdin closed
if [ -z "${BASH_SOURCE[0]:-}" ] || [ ! -f "${BASH_SOURCE[0]}" ]; then
    self=$(mktemp) && cat > "$self" && OP_UPDATE_SELF="$self" exec bash "$self" "$@" </dev/null
fi
[ -n "${OP_UPDATE_SELF:-}" ] && trap 'rm -f "$OP_UPDATE_SELF"' EXIT

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
        [ -f "$f.bak-2.0.21" ] || cp -a "$f" "$f.bak-2.0.21"
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
    add_litespeed_service "$f"
    add_litespeed_env "$home_dir/.env"
    if [ ! -f "$home_dir/litespeed.conf" ]; then
        cp /etc/openpanel/litespeed/httpd.conf "$home_dir/litespeed.conf"
        chown --reference="$home_dir/.env" "$home_dir/litespeed.conf"
    fi
    echo "LiteSpeed is now available to $ctx."
done
