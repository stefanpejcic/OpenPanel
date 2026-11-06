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
