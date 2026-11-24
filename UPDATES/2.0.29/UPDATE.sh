#!/bin/bash

# limited MCP tokens (no deletes, only some features) need mode and allowed_features on mcp_tokens, new installs get them from plans.sql
# the panel also adds them on first use, but only when its db user may ALTER, so add them here and old tokens keep working as before
DB_CNF=/etc/my.cnf
DB_NAME=panel
if [ -f "$DB_CNF" ] && command -v mariadb >/dev/null 2>&1; then
    has_table=$(mariadb --defaults-extra-file="$DB_CNF" -D "$DB_NAME" -N -B -e "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'mcp_tokens';" 2>/dev/null)
    if [ "$has_table" = "1" ]; then
        # information_schema instead of ADD COLUMN IF NOT EXISTS, that one is MariaDB only
        for col in "mode:VARCHAR(10) NULL DEFAULT NULL" "allowed_features:TEXT NULL DEFAULT NULL"; do
            name="${col%%:*}"
            def="${col#*:}"
            has_col=$(mariadb --defaults-extra-file="$DB_CNF" -D "$DB_NAME" -N -B -e "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'mcp_tokens' AND COLUMN_NAME = '$name';" 2>/dev/null)
            if [ "$has_col" = "0" ]; then
                echo "Adding $name to the mcp_tokens table..."
                mariadb --defaults-extra-file="$DB_CNF" -D "$DB_NAME" -e "ALTER TABLE mcp_tokens ADD COLUMN $name $def;" \
                    || echo "Failed to add $name to mcp_tokens, limited MCP tokens won't be available until it exists"
            fi
        done
    else
        echo "No mcp_tokens table, skipping the MCP token columns - run the 2.0.15 update first."
    fi
else
    echo "$DB_CNF or the mariadb client is missing, could not add the MCP token columns."
fi

# OpenLiteSpeed runs the lsphp of its own image, the 2.0.27 PHP update only covered php-fpm: https://github.com/stefanpejcic/OpenPanel/issues/1368
# upstream has the PHP fixes of 19 November 2026 only in 1.9.4-lsphpXX and latest-lsphpXX, older X.Y.Z-lsphpXX tags never get rebuilt, so
# a .env pinned to an older release moves to 1.9.4 with the same lsphp (lsphp82-85, the others have no fixed build), then docker-update
# pulls the tag into the shared store and recreates the running openlitespeed containers on it
OLS_FIXED="1.9.4"
OLS_REPO="litespeedtech/openlitespeed"
ols_images=""
ols_users=""
for env in /etc/openpanel/docker/compose/1.0/.env /home/*/.env; do
    [ -f "$env" ] || continue
    val=$(grep -m1 "^OPENLITESPEED_VERSION=" "$env" | cut -d= -f2- | tr -d "\"' \r")
    if [[ "$val" =~ ^([0-9]+\.[0-9]+\.[0-9]+)-(lsphp[0-9]+)$ ]] && [ "${BASH_REMATCH[1]}" != "$OLS_FIXED" ] \
       && [ "$(printf '%s\n%s\n' "${BASH_REMATCH[1]}" "$OLS_FIXED" | sort -V | head -n1)" = "${BASH_REMATCH[1]}" ]; then
        case "${BASH_REMATCH[2]}" in
            lsphp8[2-5])
                new="$OLS_FIXED-${BASH_REMATCH[2]}"
                tmp=$(mktemp)
                # written in place, keeps owner and permissions
                sed -E "s/^OPENLITESPEED_VERSION=.*/OPENLITESPEED_VERSION=\"$new\"/" "$env" > "$tmp" && cat "$tmp" > "$env"
                rm -f "$tmp"
                echo "OPENLITESPEED_VERSION $val -> $new in $env"
                val="$new"
                ;;
            *) echo "$env: OPENLITESPEED_VERSION=$val, upstream has no ${BASH_REMATCH[2]} build with the PHP fixes, left alone" ;;
        esac
    fi
    case "$env" in /home/*) ;; *) continue ;; esac
    grep -qE "^WEB_SERVER=[\"']?openlitespeed" "$env" || continue
    user=$(basename "$(dirname "$env")")
    image="$OLS_REPO:${val:-latest}"
    ols_users="$ols_users $user"
    case " $ols_images " in *" $image "*) ;; *) ols_images="$ols_images $image" ;; esac
    # a copy of the tag in the user's own store wins over the shared one when the container is recreated, so docker-update would leave it on the old image
    sock="/run/user/$(id -u "$user" 2>/dev/null)/podman/podman.sock"
    if [ -S "$sock" ] && CONTAINER_HOST="unix://$sock" timeout 30 podman --remote images --format '{{.Repository}}:{{.Tag}} {{.ReadOnly}}' 2>/dev/null \
            | grep -qxE "(docker\.io/)?$OLS_REPO:${val:-latest} false"; then
        CONTAINER_HOST="unix://$sock" timeout 30 podman --remote untag "docker.io/$OLS_REPO:${val:-latest}" >/dev/null 2>&1 \
            && echo "Removed the $OLS_REPO:${val:-latest} tag from the own image store of $user, the shared store has the current build"
    fi
done
if [ -n "$ols_images" ]; then
    if command -v opencli >/dev/null 2>&1 && opencli docker-update --help 2>&1 | grep -q -- "--no-php-move"; then
        OLS_LOG=/var/log/openpanel/updates/2.0.29-openlitespeed.log
        mkdir -p "$(dirname "$OLS_LOG")"
        OLS_RUN='for image in'"$ols_images"'; do echo "Updating $image..."; opencli docker-update "$image" -y --no-php-move || echo "Failed to update $image, run: opencli docker-update $image -y"; done
for user in'"$ols_users"'; do v=$(CONTAINER_HOST="unix:///run/user/$(id -u "$user")/podman/podman.sock" timeout 30 podman --remote exec openlitespeed /usr/local/lsws/fcgi-bin/lsphp -v 2>/dev/null | head -n1); echo "$user: ${v:-openlitespeed is not running}"; done
echo "OpenLiteSpeed update finished."'
        echo "Updating OpenLiteSpeed images in the background ($ols_images ), follow it with: tail -f $OLS_LOG"
        if ! systemd-run --unit="openpanel-ols-update-$(date +%s)" --collect --quiet --setenv=HOME=/root \
                -p StandardOutput="append:$OLS_LOG" -p StandardError="append:$OLS_LOG" \
                /bin/bash -c "$OLS_RUN" 2>/dev/null; then
            setsid nohup /bin/bash -c "$OLS_RUN" >> "$OLS_LOG" 2>&1 </dev/null &
        fi
    else
        echo "opencli docker-update is missing or too old, OpenLiteSpeed images not updated - run opencli update --cli first, then: for i in$ols_images; do opencli docker-update \$i -y; done"
    fi
fi

# PHP 8.1 and older get no security fixes, session upload progress is one of the paths of CVE-2026-109619 (and CVE-2026-115560) and
# almost nothing uses it, so it's turned off in their php.ini: the template in /etc/openpanel/php/ini and every account's php.ini/<version>.ini
# (mounted into php-fpm-<version> and into openlitespeed's lsphp). A file that already sets it, On or Off, is the admin's or user's choice and is left alone.
# Running php-fpm containers of those versions are reloaded (USR2 to the php-fpm master, no downtime) so the change takes effect
UP_VERSIONS="5.6 7.0 7.1 7.2 7.3 7.4 8.0 8.1"
UP_LINE="session.upload_progress.enabled = Off"
UP_NOTE="; Off on PHP 8.1 and older: no security fixes for them, and upload progress is a path of CVE-2026-109619"
# turns upload progress off in $1, exit 1 when there's nothing to do or the file has no place for it
upload_progress_off() {
    local f="$1" tmp
    [ -f "$f" ] || return 1
    grep -qE "^[[:space:]]*session\.upload_progress\.enabled[[:space:]]*=" "$f" && return 1
    tmp=$(mktemp)
    if grep -qE "^[[:space:]]*;[[:space:]]*session\.upload_progress\.enabled[[:space:]]*=" "$f"; then
        awk -v note="$UP_NOTE" -v line="$UP_LINE" '!done && /^[[:space:]]*;[[:space:]]*session\.upload_progress\.enabled[[:space:]]*=/ { print note; print line; done=1; next } { print }' "$f" > "$tmp"
    elif grep -qE "^[[:space:]]*\[Session\][[:space:]]*$" "$f"; then
        awk -v note="$UP_NOTE" -v line="$UP_LINE" '{ print } !done && /^[[:space:]]*\[Session\][[:space:]]*$/ { print note; print line; done=1 }' "$f" > "$tmp"
    else
        rm -f "$tmp"
        echo "$f has no [Session] section, session.upload_progress.enabled not added"
        return 1
    fi
    # in place: the file is a single-file bind mount into the running containers
    cat "$tmp" > "$f"
    rm -f "$tmp"
    echo "session.upload_progress.enabled = Off in $f"
}
for v in $UP_VERSIONS; do
    upload_progress_off "/etc/openpanel/php/ini/$v.ini"
done
for dir in /home/*; do
    [ -d "$dir/php.ini" ] || continue
    user=$(basename "$dir")
    changed=""
    for v in $UP_VERSIONS; do
        upload_progress_off "$dir/php.ini/$v.ini" && changed="$changed $v"
    done
    [ -n "$changed" ] || continue
    sock="/run/user/$(id -u "$user" 2>/dev/null)/podman/podman.sock"
    [ -S "$sock" ] || continue
    for v in $changed; do
        c="php-fpm-$v"
        [ "$(CONTAINER_HOST="unix://$sock" timeout 30 podman --remote inspect -f '{{.State.Status}}' "$c" 2>/dev/null)" = "running" ] || continue
        # only when pid 1 is the php-fpm master (openpanel/php-fpm), other images run a supervisor there
        if [ "$(CONTAINER_HOST="unix://$sock" timeout 30 podman --remote exec "$c" cat /proc/1/comm 2>/dev/null)" = "php-fpm" ] \
           && CONTAINER_HOST="unix://$sock" timeout 30 podman --remote kill --signal USR2 "$c" >/dev/null 2>&1; then
            echo "Reloaded $c for $user"
        else
            echo "Could not reload $c for $user, restart it to apply: opencli podman $user restart $c"
        fi
    done
done

# rate limits: IPs in csf.allow and csf.ignore skip them now (#1366), the snippets get the exemption and the list is synced every 5 minutes
CRON=/etc/cron.d/openpanel
SYNC='*/5 * * * * root /usr/local/bin/opencli domains-ratelimit trusted sync --quiet'
if [ -f "$CRON" ] && ! grep -q "domains-ratelimit trusted sync" "$CRON"; then
    echo "Adding the rate limit trusted IPs sync to $CRON..."
    printf '%s\n' "$SYNC" >> "$CRON"
fi
# rewrites ratelimit.conf with the exemption only where rate limiting is set up, caddy validates it first and the old files go back if it says no
if [ -x /usr/local/bin/opencli ]; then
    /usr/local/bin/opencli domains-ratelimit trusted sync || echo "Failed to apply the rate limit trusted IPs, run: opencli domains-ratelimit trusted sync"
fi

# dashboard Websites apps: the 2.0.13 default left out git, so the Git tile was gone from the dashboard, add it back where the default was never changed
OPENPANEL_CONFIG="/etc/openpanel/openpanel/conf/openpanel.config"
if [ -f "$OPENPANEL_CONFIG" ] && grep -qx "applications_dashboard_items=websites wordpress autoinstaller" "$OPENPANEL_CONFIG"; then
    echo "Adding git to applications_dashboard_items in $OPENPANEL_CONFIG..."
    sed -i 's/^applications_dashboard_items=websites wordpress autoinstaller$/applications_dashboard_items=websites wordpress autoinstaller git/' "$OPENPANEL_CONFIG"
fi
