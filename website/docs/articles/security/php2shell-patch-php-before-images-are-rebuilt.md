---
sidebar_label: "PHP2Shell: patch PHP before the images are rebuilt"
description: "How to patch PHP2Shell (CVE-2026-109618, CVE-2026-109619) on an OpenPanel server before phpMyAdmin, Roundcube and shinsenter/php ship fixed images: rebuild PHP inside each image from php.net source."
keywords: [PHP2Shell, CVE-2026-109618, CVE-2026-109619, PHP, PHP-FPM, phpMyAdmin, Roundcube, shinsenter, OpenPanel, security]
---

# PHP2Shell (CVE-2026-109618, CVE-2026-109619): patch PHP before the images are rebuilt

PHP 8.5.12, 8.4.27, 8.3.36 and 8.2.35 fix PHP2Shell, two unauthenticated remote code execution bugs in PHP itself ([blog post](/blog/php-security-update-openpanel)). On OpenPanel, PHP serves requests in three places, and all three come from images OpenPanel does not build:

| Where | Image | Runs PHP as |
|---|---|---|
| Every user's websites | `shinsenter/php:<version>-fpm` | PHP-FPM |
| phpMyAdmin (port 2053) | `phpmyadmin:latest` | Apache mod_php |
| Webmail | `roundcube/roundcubemail:latest` | Apache mod_php |

`opencli docker-update` pulls the newest of those images, but it cannot install a fix the image's maintainers have not published yet. On October 21, 2026 at 18:12 UTC, hours after PHP's release, `shinsenter/php:8.3-fpm` still had PHP 8.3.35, `shinsenter/php:8.2-fpm` 8.2.34, `phpmyadmin:latest` 8.3.35 and `roundcube/roundcubemail:latest` 8.4.26.

This guide patches PHP inside the images you already run, without waiting. [`Dockerfile`](/scripts/php-rebuild/Dockerfile) recompiles the image's own PHP from the php.net source with the image's own configure flags (the source is checked against php.net's SHA-256 and the PHP release managers' GPG signatures) and installs it over the old one. A patch release keeps the extension API, so the image's extensions (ionCube included) and its configuration stay as they are. [`php-rebuild.sh`](/scripts/php-rebuild/php-rebuild.sh) finds the newest release of the image's PHP branch and builds it. Both were used on an OpenPanel 2.0.15 server on the day of the release; a build takes about 12 minutes on 8 vCPUs.

:::warning
Run everything below as root in a plain SSH session (`tmux` or `nohup` for the builds), not with `systemd-run` or from a systemd service: podman leaves each container's `conmon` monitor in the calling unit, and systemd kills it when the unit ends.
:::

## 1. Check what you run

```bash
podman exec phpmyadmin php -r 'echo PHP_VERSION, PHP_EOL;'
podman exec openadmin_roundcube php -r 'echo PHP_VERSION, PHP_EOL;'

# the PHP a new account or a PHP version switch gets, from the shared image store:
for v in 8.2 8.3 8.4 8.5; do
  id=$(podman --root /var/lib/containers/shared-storage \
    image inspect --format '{{.Id}}' docker.io/shinsenter/php:$v-fpm 2>/dev/null) || continue
  echo "$v: $(podman run --rm --entrypoint php "$id" -r 'echo PHP_VERSION;')"
done

# what the users' running php-fpm containers run:
source /usr/local/opencli/lib/podman.sh
for f in /home/*/docker-compose.yml; do
  u=$(basename "$(dirname "$f")")
  for c in $(podman_user "$u" ps --format '{{.Names}}' | grep '^php-fpm-'); do
    echo "$u/$c: $(podman_user "$u" exec -u 0 "$c" php -r 'echo PHP_VERSION;')"
  done
done
```

Anything below 8.2.35, 8.3.36, 8.4.27 or 8.5.12 on its branch is vulnerable. PHP 8.1 and older get no official fix: move those sites to 8.2 or newer.

Download the two files:

```bash
mkdir -p /root/php-rebuild && cd /root/php-rebuild
curl -fsSLO https://openpanel.com/scripts/php-rebuild/Dockerfile
curl -fsSLO https://openpanel.com/scripts/php-rebuild/php-rebuild.sh
```

## 2. phpMyAdmin

```bash
cd /root/php-rebuild
bash php-rebuild.sh "$(podman inspect -f '{{.Image}}' phpmyadmin)" localhost/php-rebuild/phpmyadmin
podman tag localhost/php-rebuild/phpmyadmin "$(podman inspect -f '{{.ImageName}}' phpmyadmin)"
```

The `openpanel` container depends on `phpmyadmin`, so podman will not remove phpMyAdmin while the panel exists, and the panel is recreated with it (down for about 40 seconds). Remove them one at a time: `podman rm openpanel phpmyadmin` checks both at once and refuses the second.

```bash
cd /root
podman stop phpmyadmin openpanel
podman rm openpanel
podman rm phpmyadmin
podman-compose up -d --no-deps phpmyadmin
podman-compose up -d --no-deps openpanel
podman exec phpmyadmin php -r 'echo PHP_VERSION, PHP_EOL;'
```

Port 2053 can answer 502 for the first few seconds.

## 3. Roundcube

:::danger
Roundcube's SQLite database has no volume: `/var/roundcube/db/sqlite.db` lives inside the container. Recreating the container without copying it out and back deletes every webmail user's contacts, identities and settings.
:::

```bash
cd /root/php-rebuild
bash php-rebuild.sh "$(podman inspect -f '{{.Image}}' openadmin_roundcube)" localhost/php-rebuild/roundcube
podman tag localhost/php-rebuild/roundcube "$(podman inspect -f '{{.ImageName}}' openadmin_roundcube)"

podman stop openadmin_roundcube
podman cp openadmin_roundcube:/var/roundcube/db/sqlite.db /root/php-rebuild/roundcube-sqlite.db
podman rm openadmin_roundcube
cd /usr/local/mail/openmail && podman-compose up -d --no-deps roundcube
```

The new container's first start copies Roundcube into place and only then creates its database. Wait for Apache before putting the old database back:

```bash
until podman logs openadmin_roundcube 2>&1 | grep -q 'resuming normal operations'; do sleep 2; done
podman stop openadmin_roundcube
podman cp /root/php-rebuild/roundcube-sqlite.db openadmin_roundcube:/var/roundcube/db/sqlite.db
podman start openadmin_roundcube
podman exec openadmin_roundcube php -r 'echo PHP_VERSION, PHP_EOL;'
```

Webmail is down for about a minute.

## 4. Users' PHP (shinsenter/php)

Every user's podman reads OpenPanel's shared image store, `/var/lib/containers/shared-storage`, so a rebuilt image placed there serves every account, including new accounts and PHP version switches. Rebuild each PHP version that is still vulnerable (8.3 here) and copy it into the store with `skopeo` (installed by `opencli docker-update` since 2.0.16, otherwise `apt-get install -y skopeo`):

```bash
cd /root/php-rebuild
S=/var/lib/containers/shared-storage
v=8.3
img=docker.io/shinsenter/php:$v-fpm
bash php-rebuild.sh "$(podman --root $S image inspect --format '{{.Id}}' $img)" localhost/php-rebuild/shinsenter-$v
podman save -o shinsenter-$v.tar localhost/php-rebuild/shinsenter-$v
skopeo copy docker-archive:shinsenter-$v.tar "containers-storage:[overlay@$S+/run/containers/shared-storage]$img"
chmod -R o+rX $S
find $S -name '*.lock' -exec chmod o+rw {} \;
```

:::danger
Never pull or load into the shared store with `podman --root /var/lib/containers/shared-storage`. With the default runroot (root's own), a pull or load that creates layers unmounts the layer of every running root container (caddy, the mail server, the panel, MySQL and the rest): they keep running, but the next restart of any of them fails. See [If a container won't restart](#if-a-container-wont-restart). podman refuses a different `--runroot` for that store (`database configuration mismatch`), so writes go through skopeo with the store's own runroot, as above and as `opencli docker-update` does since OpenPanel 2.0.17. Reads (`image inspect`, `images`) are fine with `podman --root`. The `chmod` lines are the ones OpenPanel's own scripts run after writing to the store, so the users' podman can read it and take its locks.
:::

Then recreate the running `php-fpm-<version>` containers on the new image, through each user's own podman socket, and reload the user's web server so it connects to the new container:

```bash
source /usr/local/opencli/lib/podman.sh
for f in /home/*/docker-compose.yml; do
  u=$(basename "$(dirname "$f")")
  podman_user "$u" ps --format '{{.Names}}' | grep -qx "php-fpm-$v" || continue
  (cd "/home/$u" && CONTAINER_HOST=$(podman_user_socket "$u") podman-compose up -d --no-deps --force-recreate "php-fpm-$v")
  until podman_user "$u" logs --since 5m "php-fpm-$v" 2>&1 | grep -q 'ready to handle connections'; do sleep 1; done
  podman_user "$u" exec nginx nginx -s reload 2>/dev/null ||
    podman_user "$u" exec openresty openresty -s reload 2>/dev/null ||
    podman_user "$u" exec apache apachectl -k graceful 2>/dev/null || true
  echo "$u: $(podman_user "$u" exec -u 0 "php-fpm-$v" php -r 'echo PHP_VERSION;')"
done
```

php-fpm listens about 20 seconds after its container starts (the image runs its own start tasks first), and the user's sites answer 502 meanwhile. Check the version printed for each user:

- If a user's web server was created depending on php-fpm, podman-compose cannot remove php-fpm and leaves it on the old image. Recreate that user's web server together with it.
- A user whose own image store holds a copy of the image uses that copy instead of the shared one. Load and tag the rebuild there too: `podman_user <user> load -i shinsenter-$v.tar`, then `podman_user <user> tag localhost/php-rebuild/shinsenter-$v $img`.

OpenLiteSpeed runs its own `lsphp`, not shinsenter/php, and is not covered here.

## 5. When the upstream images catch up

The rebuilt images carry the upstream names, so `opencli docker-update` replaces them with the maintainers' own once those are newer. Check that a newer image really is fixed before you run it, without touching the local tags:

```bash
podman --root /tmp/probe/root --runroot /tmp/probe/run run --rm --entrypoint php docker.io/library/phpmyadmin:latest -r 'echo PHP_VERSION;'
rm -rf /tmp/probe
```

Since OpenPanel 2.0.17, `opencli docker-update` checks the PHP version inside an image and leaves the containers alone while it is still unpatched, and the 2.0.17 update moves Roundcube's database to a volume before it updates Roundcube. Run the checks in step 1 again after any update.

## If a container won't restart

After a pull or load into the shared store with `podman --root`, a root container that is stopped and started fails with `crun: executable file '/init' not found` or `creating temporary passwd file ... no such file or directory`, and `podman exec` can report `unable to find user root`. Find the containers in that state:

```bash
for c in $(podman ps --format '{{.Names}}'); do
  [ -e "$(podman inspect -f '{{.GraphDriver.Data.MergedDir}}' "$c")/etc" ] || echo "$c: layer not mounted"
done
```

The 2.0.17 update runs `podman mount <container>` for each running container in that state, to put its layer back without stopping it. A container that fails to start needs a forced unmount first, or reboot the server:

```bash
podman stop <container> && podman umount --force <container> && podman start <container>
```
