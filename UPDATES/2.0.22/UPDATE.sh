#!/bin/bash

# env
sed -i 's/^# LITESPEED (tag is <release>-lsphpXX, the php version comes with the image)$/# LITESPEED/' /etc/openpanel/docker/compose/1.0/.env /home/*/.env 2>/dev/null

#Update account.sh
cd /usr/local/mail/openmail/openpanel && git pull


# multi-server pins the OpenAdmin cert key and shortlived IP certs get a new key every few days, so caddy keeps the key on renewal now
CADDYFILE=/etc/openpanel/caddy/Caddyfile
if [ -f "$CADDYFILE" ] && grep -q "# START HOSTNAME IP #" "$CADDYFILE" \
   && ! sed -n '/# START HOSTNAME IP #/,/# END HOSTNAME IP #/p' "$CADDYFILE" | grep -q "reuse_private_keys"; then
    cp -p "$CADDYFILE" "$CADDYFILE.bak-2.0.22"
    tmp=$(mktemp)
    sed '/# START HOSTNAME IP #/,/# END HOSTNAME IP #/ s/^\([[:space:]]*\)tls[[:space:]]*{[[:space:]]*$/&\n\1  reuse_private_keys/' "$CADDYFILE" > "$tmp"
    # cat into the file instead of mv, caddy has it bind mounted
    cat "$tmp" > "$CADDYFILE" && rm -f "$tmp"
    if podman exec caddy caddy validate --config /etc/caddy/Caddyfile </dev/null >/dev/null 2>&1; then
        podman exec caddy caddy reload --config /etc/caddy/Caddyfile </dev/null >/dev/null 2>&1
    else
        cat "$CADDYFILE.bak-2.0.22" > "$CADDYFILE"
        echo "Caddy did not accept reuse_private_keys, Caddyfile restored"
    fi
fi



# docker-proxy: add restart: unless-stopped (openpanel-configuration#9), no restarts, just the compose line
# spacing matters: the line goes in with the same 4 spaces as container_name, files with other indentation are left alone
compose_check() {
    local f="$1" dir
    dir=$(dirname "$f")
    if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
        (cd "$dir" && timeout 30 docker compose -f "$f" config -q) </dev/null >/dev/null 2>&1 || return 1
    elif command -v podman-compose >/dev/null 2>&1; then
        (cd "$dir" && timeout 30 podman-compose -f "$f" config) </dev/null >/dev/null 2>&1 || return 1
    fi
    # always also parse the yaml itself, works without any compose binary
    python3 - "$f" <<'EOF' >/dev/null 2>&1 || return 1
import sys, yaml
d = yaml.safe_load(open(sys.argv[1]))
assert isinstance(d, dict) and isinstance(d.get("services"), dict)
EOF
}

for f in /etc/openpanel/docker/compose/1.0/docker-compose.yml /home/*/docker-compose.yml; do
    [ -f "$f" ] || continue
    grep -q "^  docker-proxy:[[:space:]]*$" "$f" || continue
    sed -n '/^  docker-proxy:[[:space:]]*$/,/^  [^ ]/p' "$f" | grep -q "^    restart:" && continue
    sed -n '/^  docker-proxy:[[:space:]]*$/,/^  [^ ]/p' "$f" | grep -q "^    container_name: docker-proxy[[:space:]]*$" || continue
    if ! compose_check "$f"; then
        echo "$f fails config check before the change, skipping docker-proxy restart line"
        continue
    fi
    cp -p "$f" "$f.bak-2.0.22"
    tmp=$(mktemp)
    sed '/^  docker-proxy:[[:space:]]*$/,/^  [^ ]/ s/^\(    container_name: docker-proxy[[:space:]]*\)$/\1\n    restart: unless-stopped/' "$f" > "$tmp"
    cat "$tmp" > "$f" && rm -f "$tmp"
    if compose_check "$f" && python3 - "$f" <<'EOF' >/dev/null 2>&1
import sys, yaml
d = yaml.safe_load(open(sys.argv[1]))
assert d["services"]["docker-proxy"]["restart"] == "unless-stopped"
EOF
    then
        echo "Added restart: unless-stopped to docker-proxy in $f"
    else
        cat "$f.bak-2.0.22" > "$f"
        echo "$f failed config check after adding docker-proxy restart line, restored from $f.bak-2.0.22"
    fi
done
