#!/bin/bash

# opencli update runs this as wget -O - URL | bash, so a command below that reads stdin (apt-get, podman) would eat the rest of the script and bash would go on mid-line; the rest is saved to a file and run from there with stdin closed
if [ -z "${BASH_SOURCE[0]:-}" ] || [ ! -f "${BASH_SOURCE[0]}" ]; then
    self=$(mktemp) && cat > "$self" && OP_UPDATE_SELF="$self" exec bash "$self" "$@" </dev/null
fi
[ -n "${OP_UPDATE_SELF:-}" ] && trap 'rm -f "$OP_UPDATE_SELF"' EXIT

# PHP2Shell (CVE-2026-109618, CVE-2026-109619): https://github.com/stefanpejcic/OpenPanel/issues/1203
# 2.0.16 pulled shinsenter/php before upstream rebuilt 8.3 and 8.2 (and on servers that updated early, 8.5 and 8.4 too), and didn't cover phpMyAdmin and Roundcube
# docker-update now checks the PHP version inside each image and skips the recreate, reporting a failure, while an image still has the old PHP
DOCKER_UPDATE=/usr/local/opencli/docker/update.sh
if [ ! -x "$DOCKER_UPDATE" ] || ! grep -q "^readonly PHP_FIXED=" "$DOCKER_UPDATE" || ! grep -q "^pull_image()" "$DOCKER_UPDATE"; then
    echo "opencli docker-update has no PHP version check, PHP images not updated - run opencli update first, then run this update again."
    exit 0
fi

# 2.0.16's shared store pulls used root's runroot, which unmounts every running root container's layer: they keep running, but their next restart or recreate fails ("executable file '/init' not found", "removing mount point ... directory not empty")
# podman mount puts the layer back for a running container without stopping it, containers whose layer is still mounted are left alone
podman ps --format '{{.Names}}' 2>/dev/null | while read -r name; do
    merged=$(podman inspect -f '{{.GraphDriver.Data.MergedDir}}' "$name" 2>/dev/null </dev/null)
    [ -n "$merged" ] || continue
    awk -v m="$merged" '$2 == m { found = 1 } END { exit !found }' /proc/mounts && continue
    echo "Remounting $name, its layer was unmounted by an earlier shared store pull..."
    podman mount "$name" >/dev/null 2>&1 </dev/null || echo "Failed to remount $name, restart it with: podman stop $name && podman umount --force $name && podman start $name"
done

# every php version in use, versions whose image isn't rebuilt yet are reported and left as they are
php_images=$(grep -hoE "^    image:[[:space:]]*['\"]?(docker\.io/)?shinsenter/php:[0-9.]+-fpm" /etc/openpanel/docker/compose/1.0/docker-compose.yml /home/*/docker-compose.yml 2>/dev/null \
    | sed -E "s/^    image:[[:space:]]*['\"]?(docker\.io\/)?//" | sort -uV)
for image in $php_images; do
    echo "Updating $image..."
    opencli docker-update "$image" -y || echo "Not all containers on $image were updated, run again later: opencli docker-update $image"
done

# phpMyAdmin is in root's stack, docker-update recreates it if it's running
if podman container exists phpmyadmin; then
    pma_image=$(podman inspect -f '{{.ImageName}}' phpmyadmin 2>/dev/null)
    pma_image="${pma_image:-phpmyadmin:latest}"
    echo "Updating $pma_image..."
    opencli docker-update "$pma_image" -y || echo "phpMyAdmin was not updated, run again later: opencli docker-update $pma_image"
fi

# Roundcube runs from the mail stack, which docker-update doesn't manage; its sqlite db (contacts, identities, settings) is in the container layer, so it's moved to the host first and kept from now on
MAIL_DIR=/usr/local/mail/openmail
MAIL_COMPOSE="$MAIL_DIR/compose.yml"
RC_CONTAINER=openadmin_roundcube
RC_DB_DIR="$MAIL_DIR/docker-data/roundcube/db"
RC_PLUGIN_LINE="      - ./roundcube-plugins/autologin.php:/var/www/html/public_html/autologin.php:ro"
update_roundcube() {
    podman container exists "$RC_CONTAINER" || return 0
    [ -f "$MAIL_COMPOSE" ] || return 0
    if [ "$(podman inspect -f '{{.State.Status}}' "$RC_CONTAINER" 2>/dev/null)" != "running" ]; then
        echo "Roundcube is not running, left alone - it gets updated by the next run of this update after it's started."
        return 0
    fi
    command -v skopeo >/dev/null 2>&1 || { echo "skopeo is missing, Roundcube not updated - run opencli docker-update once to install it, then this update again."; return 0; }
    local image current_id new_id version fixed
    image=$(podman inspect -f '{{.ImageName}}' "$RC_CONTAINER" 2>/dev/null)
    image="${image:-docker.io/roundcube/roundcubemail:latest}"

    # the registry config has the exact PHP_VERSION for roundcube's php base image, nothing is pulled or recreated until upstream has the fix
    . <(sed -n '/^readonly PHP_FIXED=/p;/^php_outdated() {/,/^}/p' "$DOCKER_UPDATE")
    version=$(skopeo inspect --config "docker://$image" 2>/dev/null | jq -r '.config.Env[]? | select(startswith("PHP_VERSION=")) | sub("PHP_VERSION="; "")')
    fixed=$(php_outdated "$version")
    if [ -z "$version" ]; then
        echo "Could not read the PHP version of $image from the registry, Roundcube not updated - run this update again later."
        return 0
    elif [ -n "$fixed" ]; then
        echo "$image still has PHP $version (security fixes are in $fixed), Roundcube not updated - run this update again once it's rebuilt."
        return 0
    fi

    current_id=$(podman inspect -f '{{.Image}}' "$RC_CONTAINER" 2>/dev/null)
    echo "Pulling $image (PHP $version)..."
    if ! podman pull -q "$image" >/dev/null 2>&1; then
        echo "Failed to pull $image, Roundcube not updated."
        return 0
    fi
    new_id=$(podman image inspect -f '{{.Id}}' "$image" 2>/dev/null)

    if ! grep -qF ":/var/roundcube/db" "$MAIL_COMPOSE"; then
        grep -qxF "$RC_PLUGIN_LINE" "$MAIL_COMPOSE" || { echo "Roundcube volumes in $MAIL_COMPOSE are not the default ones, not recreating it - its sqlite db would be lost."; return 0; }
        echo "Moving the Roundcube database to $RC_DB_DIR..."
        mkdir -p "$RC_DB_DIR"
        if ! podman cp "$RC_CONTAINER:/var/roundcube/db/." "$RC_DB_DIR/"; then
            echo "Failed to copy the Roundcube database out of the container, Roundcube not updated."
            return 0
        fi
        # www-data in the roundcube image
        chown -R 33:33 "$RC_DB_DIR"
        cp -a "$RC_DB_DIR" "$RC_DB_DIR.bak-2.0.17"
        cp -a "$MAIL_COMPOSE" "$MAIL_COMPOSE.bak-2.0.17"
        sed -i "s|^$(printf '%s' "$RC_PLUGIN_LINE" | sed 's/[.[\*^$|]/\\&/g')\$|&\n      - ./docker-data/roundcube/db:/var/roundcube/db|" "$MAIL_COMPOSE"
        grep -qF "./docker-data/roundcube/db:/var/roundcube/db" "$MAIL_COMPOSE" || { echo "Failed to add the Roundcube database volume to $MAIL_COMPOSE, Roundcube not updated."; return 0; }
    elif [ "$current_id" = "$new_id" ]; then
        echo "Roundcube already runs $image (PHP $version)."
        return 0
    fi

    # podman-compose 1.0.6 recreates every service of a project whose compose file changed and ignores --no-deps, so the mail stack is taken down and started as a whole; mail is unavailable for about a minute
    echo "Restarting the mail server with the new Roundcube..."
    (cd "$MAIL_DIR" && timeout 120 podman-compose down) >/dev/null 2>&1
    (cd "$MAIL_DIR" && timeout 300 podman-compose up -d) >/dev/null 2>&1
    local name
    for name in openadmin_mailserver "$RC_CONTAINER"; do
        if [ "$(podman inspect -f '{{.State.Status}}' "$name" 2>/dev/null)" != "running" ]; then
            echo "$name is not running after the restart, start the mail server with: opencli email-server restart"
            return 0
        fi
    done
    echo "Roundcube now runs PHP $(podman exec "$RC_CONTAINER" php -r 'echo PHP_VERSION;' 2>/dev/null)."
}
update_roundcube
