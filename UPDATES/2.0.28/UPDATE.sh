#!/bin/bash

# remove the backups 2.0.22 made: docker-compose.yml files (docker-proxy restart line) and the Caddyfile (reuse_private_keys)
rm -f /etc/openpanel/docker/compose/1.0/docker-compose.yml.bak-2.0.22 /home/*/docker-compose.yml.bak-2.0.22
rm -f /etc/openpanel/caddy/Caddyfile.bak-2.0.22

# fresh themes.yml so the Saturn theme shows under Available themes, OpenAdmin falls back to this copy when github can't be reached
CONFIG_RAW="${CONFIG_RAW:-https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main}"
mkdir -p /etc/openpanel/openpanel/templates
tmp=$(mktemp)
if wget --timeout=15 --tries=2 -q -O "$tmp" "$CONFIG_RAW/openpanel/templates/themes.yml" && grep -q "^themes:" "$tmp"; then
    echo "Updating /etc/openpanel/openpanel/templates/themes.yml..."
    cat "$tmp" > /etc/openpanel/openpanel/templates/themes.yml
    chmod 644 /etc/openpanel/openpanel/templates/themes.yml
else
    echo "Failed to download $CONFIG_RAW/openpanel/templates/themes.yml"
fi
rm -f "$tmp"

# allow_user_theme=no for existing installs, users only get the theme picker once the admin turns it on
CONF=/etc/openpanel/openpanel/conf/openpanel.config
if [ -f "$CONF" ] && ! grep -q "^allow_user_theme=" "$CONF"; then
    echo "Adding allow_user_theme=no to $CONF..."
    if grep -q "^template=" "$CONF"; then
        sed -i '/^template=/a allow_user_theme=no' "$CONF"
    else
        sed -i '/^\[PANEL\]/a allow_user_theme=no' "$CONF"
    fi
fi

# blocklists sentinel checks the server IP against, only added when missing so the admin's own list stays
RBL_LIST=/etc/openpanel/openadmin/config/rbl.txt
if [ ! -f "$RBL_LIST" ]; then
    tmp=$(mktemp)
    if wget --timeout=15 --tries=2 -q -O "$tmp" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/rbl.txt" && grep -q "spamhaus" "$tmp"; then
        echo "Adding $RBL_LIST..."
        mkdir -p "$(dirname "$RBL_LIST")"
        cat "$tmp" > "$RBL_LIST"
        chmod 644 "$RBL_LIST"
    fi
    rm -f "$tmp"
fi

# the admin's own daily summary (openadmin daily-summary) is only in the cron of installs from 2.0.26 on, older servers showed "No daily summary yet"
CRON_FILE="/etc/cron.d/openpanel"
if [ -f "$CRON_FILE" ] && ! grep -q "daily-summary >>" "$CRON_FILE"; then
    echo "Adding the admin daily summary cron to $CRON_FILE..."
    echo '45 6 * * * root for b in /usr/local/admin/openadmin-amd64 /usr/local/admin/openadmin-arm64; do [ -x "$b" ] && "$b" daily-summary >> /var/log/openpanel/admin/cron.log 2>&1 && break; done' >> "$CRON_FILE"
fi

# new daily blocklist alert, on by default
NOTIF_INI=/etc/openpanel/openadmin/config/notifications.ini
if [ -f "$NOTIF_INI" ] && ! grep -q "^rbl=" "$NOTIF_INI"; then
    echo "Adding rbl=yes to $NOTIF_INI..."
    if grep -q "^mail_defer=" "$NOTIF_INI"; then
        sed -i '/^mail_defer=/a rbl=yes' "$NOTIF_INI"
    else
        sed -i '/^\[DEFAULT\]/a rbl=yes' "$NOTIF_INI"
    fi
fi

# PHP 8.6 is GA: add php-fpm-8.6 to the template and to every user that already has php-fpm-8.5, users without 8.5 are left alone.
# The 8.6 service is a copy of that file's own php-fpm-8.5 block, so it uses the same kind of image: openpanel/php-fpm:8.6
# for accounts on openpanel/php-fpm:8.5, shinsenter/php:8.6-fpm for older accounts still on shinsenter/php:8.5-fpm (whose block
# has other env and mounts). Any other 8.5 image is a custom one and is skipped. No ionCube loader for 8.6 yet, so the ionCube
# mounts are not copied, opencli update --php adds them once the loader is out.
CONFIG_RAW="${CONFIG_RAW:-https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main}"
PHP_INI_DIR=/etc/openpanel/php/ini
if [ -d "$PHP_INI_DIR" ] && [ ! -f "$PHP_INI_DIR/8.6.ini" ]; then
    tmp=$(mktemp)
    if wget --timeout=15 --tries=2 -q -O "$tmp" "$CONFIG_RAW/php/ini/8.6.ini" && grep -q "memory_limit" "$tmp"; then
        cat "$tmp" > "$PHP_INI_DIR/8.6.ini"
    elif [ -f "$PHP_INI_DIR/8.5.ini" ]; then
        cp -p "$PHP_INI_DIR/8.5.ini" "$PHP_INI_DIR/8.6.ini"
    fi
    rm -f "$tmp"
    [ -f "$PHP_INI_DIR/8.6.ini" ] && chmod 644 "$PHP_INI_DIR/8.6.ini" && echo "Added $PHP_INI_DIR/8.6.ini"
fi

FORBIDDEN=/etc/openpanel/openadmin/config/forbidden_usernames.txt
if [ -f "$FORBIDDEN" ] && ! grep -qx "php-fpm-8.6" "$FORBIDDEN"; then
    if grep -qx "php-fpm-8.5" "$FORBIDDEN"; then
        sed -i 's/^php-fpm-8\.5$/php-fpm-8.5\nphp-fpm-8.6/' "$FORBIDDEN"
    else
        echo "php-fpm-8.6" >> "$FORBIDDEN"
    fi
fi

# config -q / config validates the file with its .env, the yaml parse works without any compose binary
php86_compose_ok() {
    local f="$1" dir
    dir=$(dirname "$f")
    if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
        (cd "$dir" && timeout 30 docker compose -f "$f" config -q) </dev/null >/dev/null 2>&1 || return 1
    elif command -v podman-compose >/dev/null 2>&1; then
        (cd "$dir" && timeout 30 podman-compose -f "$f" config) </dev/null >/dev/null 2>&1 || return 1
    fi
    python3 -c 'import sys, yaml; d = yaml.safe_load(open(sys.argv[1])); assert isinstance(d.get("services"), dict)' "$f" >/dev/null 2>&1
}

# copies the php-fpm-8.5 block of $1 as php-fpm-8.6 right after it, prints the new image, exit 2 = image not ours / leftovers
php86_add_service() {
    python3 - "$1" <<'EOF'
import re, sys
path = sys.argv[1]
lines = open(path).read().split("\n")
start = next(i for i, l in enumerate(lines) if re.match(r"^  php-fpm-8\.5:\s*$", l))
end = start + 1
while end < len(lines) and not re.match(r"^(  )?[^ #\s]", lines[end]):
    end += 1
while lines[end - 1].strip() == "":
    end -= 1
block = [l for l in lines[start:end] if "ioncube" not in l]
subs = [("php-fpm-8.5", "php-fpm-8.6"), ("shinsenter/php:8.5-fpm", "shinsenter/php:8.6-fpm"),
        ("php-fpm:8.5", "php-fpm:8.6"), ("php.ini/8.5.ini", "php.ini/8.6.ini"), ("PHP_FPM_8_5_", "PHP_FPM_8_6_")]
new = []
for l in block:
    for a, b in subs:
        l = l.replace(a, b)
    new.append(l)
images = [l.split("image:", 1)[1].strip().strip("'\"") for l in new if l.startswith("    image:")]
if len(images) != 1 or not re.fullmatch(r"(docker\.io/)?(openpanel/php-fpm:8\.6|shinsenter/php:8\.6-fpm)", images[0]):
    print(images[0] if images else "no image"); sys.exit(2)
if any(re.search(r"8[._]5\b", l) for l in new):
    print("8.5 left in the block"); sys.exit(2)
lines[end:end] = [""] + new
open(path, "w").write("\n".join(lines))  # write in place, keeps owner and permissions
print(images[0])
EOF
}

# PHP_FPM_8_6_* in $1 with the values that file has for 8.5, template values when it has none
php86_add_env() {
    local env="$1" k val last
    grep -q "^PHP_FPM_8_6_" "$env" && return 0
    local block="# PHP 8.6"
    for k in RAM CPU PIDS; do
        val=$(grep -m1 "^PHP_FPM_8_5_${k}=" "$env" | cut -d= -f2-)
        if [ -z "$val" ]; then
            case "$k" in RAM) val='"1.0G"' ;; CPU) val='"1.0"' ;; PIDS) val='"1000"' ;; esac
        fi
        block="$block"$'\n'"PHP_FPM_8_6_${k}=${val}"
    done
    last=$(grep -n "^PHP_FPM_8_5_" "$env" | tail -n1 | cut -d: -f1)
    [ -n "$last" ] || last=$(grep -n "^PHP_FPM_" "$env" | tail -n1 | cut -d: -f1)
    tmp=$(mktemp)
    if [ -n "$last" ]; then
        awk -v n="$last" -v blk="$block" '{ print } NR==n { print ""; print blk }' "$env" > "$tmp"
    else
        { cat "$env"; echo; echo "$block"; } > "$tmp"
    fi
    cat "$tmp" > "$env" && rm -f "$tmp"
}

for dir in /etc/openpanel/docker/compose/1.0 /home/*; do
    f="$dir/docker-compose.yml"; env="$dir/.env"
    [ -f "$f" ] && [ -f "$env" ] || continue
    grep -q "^  php-fpm-8\.5:[[:space:]]*$" "$f" || continue
    grep -q "^  php-fpm-8\.6:[[:space:]]*$" "$f" && continue
    if ! php86_compose_ok "$f"; then
        echo "$f fails config check before the change, skipping PHP 8.6"
        continue
    fi
    cp -p "$f" "$f.bak-2.0.28"; cp -p "$env" "$env.bak-2.0.28"
    if ! image=$(php86_add_service "$f"); then
        echo "$f: php-fpm-8.5 uses a custom image or setup ($image), skipping PHP 8.6"
        cat "$f.bak-2.0.28" > "$f"; rm -f "$f.bak-2.0.28" "$env.bak-2.0.28"
        continue
    fi
    php86_add_env "$env"
    if php86_compose_ok "$f"; then
        if [ -d "$dir/php.ini" ] && [ ! -f "$dir/php.ini/8.6.ini" ] && [ -f "$PHP_INI_DIR/8.6.ini" ]; then
            cp "$PHP_INI_DIR/8.6.ini" "$dir/php.ini/8.6.ini"
            chown --reference="$dir/php.ini" "$dir/php.ini/8.6.ini"
        fi
        rm -f "$f.bak-2.0.28" "$env.bak-2.0.28"
        echo "Added php-fpm-8.6 ($image) to $f"
    else
        cat "$f.bak-2.0.28" > "$f"; cat "$env.bak-2.0.28" > "$env"
        rm -f "$f.bak-2.0.28" "$env.bak-2.0.28"
        echo "$f failed config check after adding php-fpm-8.6, restored"
    fi
done

# 8.6 is the default PHP for new accounts: only on servers still on the old default 8.5 and only once the template has php-fpm-8.6,
# an admin's own choice stays, existing accounts keep their own DEFAULT_PHP_VERSION
TPL=/etc/openpanel/docker/compose/1.0
if [ -f "$TPL/.env" ] && grep -qE "^DEFAULT_PHP_VERSION=['\"]?8\.5['\"]?[[:space:]]*$" "$TPL/.env" \
   && grep -q "^  php-fpm-8\.6:[[:space:]]*$" "$TPL/docker-compose.yml" 2>/dev/null; then
    sed -i -E "s/^DEFAULT_PHP_VERSION=(['\"]?)8\.5(['\"]?)[[:space:]]*$/DEFAULT_PHP_VERSION=\18.6\2/" "$TPL/.env"
    echo "Default PHP version for new users changed from 8.5 to 8.6 in $TPL/.env"
fi
