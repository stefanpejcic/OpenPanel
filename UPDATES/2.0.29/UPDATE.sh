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
