#!/bin/bash

# FTP passive mode: the start script looked the public IP up once with a 3s timeout and fell back to the
# container's 172.20.0.6 when that failed (e.g. network not up yet after a reboot), so logins worked but transfers didn't.
# Now it retries for ~60s and takes PASV_ADDRESS from FTP_PASV_ADDRESS in /root/.env when set.
# Files only, the FTP container is NOT restarted or recreated here: the new start script is used on its next start,
# the env var on its next recreate (sentinel recreates it on boot when the IP changed).

# 1. start script
FTP_START="/etc/openpanel/ftp/start_vsftpd.sh"
if [ -f "$FTP_START" ] && ! grep -q "valid_ipv4" "$FTP_START"; then
    echo "Updating $FTP_START..."
    tmp=$(mktemp)
    if wget -q -O "$tmp" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ftp/start_vsftpd.sh" && grep -q "valid_ipv4" "$tmp" && sh -n "$tmp"; then
        cp -p "$FTP_START" "${FTP_START}.bak-2.0.22"
        # write in place, the file is bind-mounted into the container
        cat "$tmp" > "$FTP_START"
        chmod +x "$FTP_START"
    else
        echo "Failed to download the new $FTP_START, update it by hand from the openpanel-configuration repo"
    fi
    rm -f "$tmp"
fi

# 2. FTP_PASV_ADDRESS in /root/.env, only a public IPv4, a private one (NAT) is left for the container to detect
ENV_FILE="/root/.env"
if [ -f "$ENV_FILE" ] && ! grep -q "^FTP_PASV_ADDRESS=" "$ENV_FILE"; then
    ip=$(curl --silent --max-time 3 -4 "https://ip.openpanel.com" 2>/dev/null || curl --silent --max-time 3 -4 "https://ifconfig.me/ip" 2>/dev/null)
    if ! [[ "$ip" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]] \
       || [[ "$ip" =~ ^(10\.|127\.|192\.168\.|172\.(1[6-9]|2[0-9]|3[01])\.|100\.(6[4-9]|[7-9][0-9]|1[01][0-9]|12[0-7])\.) ]]; then
        ip=""
    fi
    printf '\n# public IPv4 for FTP passive mode, leave empty to detect it on start\nFTP_PASV_ADDRESS="%s"\n' "$ip" >> "$ENV_FILE"
    echo "Added FTP_PASV_ADDRESS=\"$ip\" to $ENV_FILE"
fi

# 3. pass it to the container in /root/docker-compose.yml, no restart, just the compose lines
# spacing matters: the lines go in with the same 4 spaces as container_name, files with other indentation are left alone
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

COMPOSE="/root/docker-compose.yml"
if [ -f "$COMPOSE" ] && grep -q "^  openadmin_ftp:[[:space:]]*$" "$COMPOSE" && ! grep -q 'PASV_ADDRESS=\${FTP_PASV_ADDRESS' "$COMPOSE"; then
    block=$(sed -n '/^  openadmin_ftp:[[:space:]]*$/,/^  [^ ]/p' "$COMPOSE")
    if grep -q "^    environment:" <<< "$block"; then
        echo "openadmin_ftp in $COMPOSE already has an environment section, add - PASV_ADDRESS=\${FTP_PASV_ADDRESS:-} to it by hand"
    elif ! grep -q "^    container_name: openadmin_ftp[[:space:]]*$" <<< "$block"; then
        echo "openadmin_ftp in $COMPOSE has unexpected indentation, skipping PASV_ADDRESS"
    elif ! compose_check "$COMPOSE"; then
        echo "$COMPOSE fails config check before the change, skipping PASV_ADDRESS"
    else
        cp -p "$COMPOSE" "$COMPOSE.bak-2.0.23"
        tmp=$(mktemp)
        sed '/^  openadmin_ftp:[[:space:]]*$/,/^  [^ ]/ s/^\(    container_name: openadmin_ftp[[:space:]]*\)$/\1\n    environment:\n      - PASV_ADDRESS=${FTP_PASV_ADDRESS:-}/' "$COMPOSE" > "$tmp"
        cat "$tmp" > "$COMPOSE" && rm -f "$tmp"
        if compose_check "$COMPOSE" && python3 - "$COMPOSE" <<'EOF' >/dev/null 2>&1
import sys, yaml
d = yaml.safe_load(open(sys.argv[1]))
assert "PASV_ADDRESS=${FTP_PASV_ADDRESS:-}" in d["services"]["openadmin_ftp"]["environment"]
EOF
        then
            echo "Added PASV_ADDRESS to openadmin_ftp in $COMPOSE, it applies the next time the FTP container is recreated"
        else
            cat "$COMPOSE.bak-2.0.23" > "$COMPOSE"
            echo "$COMPOSE failed config check after adding PASV_ADDRESS, restored from $COMPOSE.bak-2.0.23"
        fi
    fi
fi
