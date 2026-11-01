---
sidebar_label: "Remove Old PHP Versions"
---

# How to Remove Old PHP Versions

By default, every OpenPanel user gets all PHP versions from 5.6 to the latest release. If you don't want users to select end-of-life versions, you can remove them per user and for all new users.

## How OpenPanel lists PHP versions

There is no hardcoded list of PHP versions. OpenPanel reads them from the `php-fpm-X.Y` services in each user's `/home/<USERNAME>/docker-compose.yml` file. When a new user is created, that file is copied from the template at `/etc/openpanel/docker/compose/1.0/docker-compose.yml`.

The same list is used for every PHP selector in OpenPanel: **PHP > Domains**, **PHP > Default Version**, **PHP > Options** and the application installers. Versions that aren't in the list are also rejected when a domain's PHP version is changed through the API. Once you remove a `php-fpm-X.Y` service, users can't select that version anymore.

:::info
On OpenLiteSpeed and LiteSpeed web servers, PHP versions come from the web server's image tags instead, and users can't select a PHP version per domain.
:::

## For new users

Edit the template file `/etc/openpanel/docker/compose/1.0/docker-compose.yml` and delete the service blocks of the versions you want to remove. Each block starts at a `  php-fpm-X.Y:` line and ends right before the next service:

```yaml
  php-fpm-5.6:
    image: shinsenter/php:5.6-fpm
    container_name: php-fpm-5.6
    ...
    command: php-fpm --allow-to-run-as-root
```

Or remove them with a script. Set `VERSIONS` to the versions you want to remove:

```bash
VERSIONS="5.6 7.0 7.1 7.2 7.3 7.4 8.0 8.1"
f=/etc/openpanel/docker/compose/1.0/docker-compose.yml
cp -p "$f" "$f.bak"
for v in $VERSIONS; do
  awk -v svc="php-fpm-$v" '
    $0 ~ "^  " svc ":[ \t]*$" { skip=1; next }
    skip && (/^  [^ #]/ || /^[^ \t#]/) { skip=0 }
    !skip
  ' "$f" > "$f.tmp" && cat "$f.tmp" > "$f" && rm -f "$f.tmp"
done
grep -E '^  php-fpm-' "$f"
```

Then make sure `DEFAULT_PHP_VERSION` in `/etc/openpanel/docker/compose/1.0/.env` is set to a version you are keeping. You can also change it from the [Edit User Defaults](/docs/admin/settings/defaults/) page. The `PHP_FPM_X_Y_RAM` and `PHP_FPM_X_Y_CPU` lines of the removed versions in that `.env` file can stay. They aren't used without the service.

## For existing users

:::warning
Websites that still use a removed PHP version stop working. Move them to a supported version before you remove it.
:::

### 1. Find domains that use old versions

A domain's PHP version is set in its vhost file. To list all domains on the versions you want to remove:

```bash
VERSIONS="5.6 7.0 7.1 7.2 7.3 7.4 8.0 8.1"
for v in $VERSIONS; do
  grep -l "php-fpm-$v:" /home/*/docker-data/volumes/*_webserver_data/_data/*.conf 2>/dev/null |
    sed "s|.*/||; s|\.conf$| (PHP $v)|"
done
```

### 2. Move them to a supported version

For each domain from the list, run [`opencli php-domain`](/docs/articles/opencli/php/#change-version-for-a-domain):

```bash
opencli php-domain example.com --update 8.3
```

Users can also do this themselves from **OpenPanel > PHP > Domains**.

### 3. Change the default version

If a user's default PHP version is one you are removing, change it with [`opencli php-default`](/docs/articles/opencli/php/#change-the-default-version):

```bash
opencli php-default <USERNAME>
opencli php-default <USERNAME> --update 8.3
```

### 4. Stop and remove the old services

Stop the old PHP containers through the user's own podman socket, then remove the service blocks from the user's `docker-compose.yml`. This loops through all users. To do one user only, replace `/home/*/docker-compose.yml` with `/home/<USERNAME>/docker-compose.yml`:

```bash
VERSIONS="5.6 7.0 7.1 7.2 7.3 7.4 8.0 8.1"
source /usr/local/opencli/lib/podman.sh
for f in /home/*/docker-compose.yml; do
  u=$(basename "$(dirname "$f")")
  cp -p "$f" "$f.bak"
  for v in $VERSIONS; do
    grep -q "^  php-fpm-$v:" "$f" || continue
    (cd "/home/$u" && CONTAINER_HOST=$(podman_user_socket "$u") podman-compose down "php-fpm-$v") >/dev/null 2>&1
    awk -v svc="php-fpm-$v" '
      $0 ~ "^  " svc ":[ \t]*$" { skip=1; next }
      skip && (/^  [^ #]/ || /^[^ \t#]/) { skip=0 }
      !skip
    ' "$f" > "$f.tmp" && cat "$f.tmp" > "$f" && rm -f "$f.tmp"
  done
  (cd "/home/$u" && podman-compose config >/dev/null 2>&1) && echo "$u: OK" || { cp -p "$f.bak" "$f"; echo "$u: invalid config, restored backup"; }
done
```

### 5. Restart OpenPanel

OpenPanel caches the PHP version list for up to one hour. To show the change right away, restart the OpenPanel container:

```bash
podman restart openpanel
```

## Don't run `opencli update --php`

[`opencli update --php`](/docs/articles/opencli/update/#php) adds every PHP version from the upstream template that is missing on your server. It adds them to the template and to all users, so it would bring the removed versions back. It only runs when you start it manually. It doesn't run with automatic updates.

The [PHP Settings](/docs/admin/settings/php/) page in OpenAdmin may show a recommendation to install supported PHP versions that are missing. You can ignore it for the versions you removed on purpose.

To add a removed version back later, copy its service block from the [upstream template](https://github.com/stefanpejcic/openpanel-configuration/blob/main/docker/compose/1.0/docker-compose.yml) into the user's `docker-compose.yml`.
