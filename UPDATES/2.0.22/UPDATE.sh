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

# the FTP container's start script chowned account folders as root through symlinks
FTP_START="/etc/openpanel/ftp/start_vsftpd.sh"
if [ -f "$FTP_START" ] && ! grep -q "prepare_ftp_folder" "$FTP_START"; then
    echo "Updating $FTP_START..."
    tmp=$(mktemp)
    if wget -q -O "$tmp" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ftp/start_vsftpd.sh" && grep -q "prepare_ftp_folder" "$tmp" && sh -n "$tmp"; then
        cp -p "$FTP_START" "${FTP_START}.bak-2.0.21"
        # write in place, the file is bind-mounted into the container
        cat "$tmp" > "$FTP_START"
        chmod +x "$FTP_START"

        if podman ps --format '{{.Names}}' | grep -qx openadmin_ftp; then
            echo "Restarting the FTP service..."
            source /usr/local/opencli/lib/podman.sh
            podman restart -t 10 openadmin_ftp >/dev/null 2>&1
            # a restart can leave the container stuck in "stopping", recreate it then
            podman_ensure_running openadmin_ftp /root openadmin_ftp 60 || echo "FTP service did not come back up, start it with: cd /root && podman-compose up -d openadmin_ftp"
        fi
    else
        echo "Failed to download the new $FTP_START, update it by hand from the openpanel-configuration repo"
    fi
    rm -f "$tmp"
fi

# DNS record types users can add and edit, new in 2.0.22
OPENPANEL_CONFIG="/etc/openpanel/openpanel/conf/openpanel.config"
if [ -f "$OPENPANEL_CONFIG" ] && ! grep -q "^dns_record_types=" "$OPENPANEL_CONFIG"; then
    echo "Adding dns_record_types to $OPENPANEL_CONFIG..."
    if grep -q "^\[USERS\]" "$OPENPANEL_CONFIG"; then
        sed -i '/^\[USERS\]/a dns_record_types=A AAAA CAA CNAME MX SRV TLSA TXT' "$OPENPANEL_CONFIG"
    else
        printf '\n[USERS]\ndns_record_types=A AAAA CAA CNAME MX SRV TLSA TXT\n' >> "$OPENPANEL_CONFIG"
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
