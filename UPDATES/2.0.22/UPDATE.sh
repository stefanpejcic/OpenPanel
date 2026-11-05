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
