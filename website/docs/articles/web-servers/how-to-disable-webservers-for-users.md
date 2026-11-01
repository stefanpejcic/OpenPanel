---
sidebar_label: "Disable Webservers for Users"
---

# How to Disable Webservers for Users

By default, OpenPanel users can choose between Apache, Nginx, OpenResty and OpenLiteSpeed. If you don't want to support all of them, you can remove a webserver for new users, for existing users, or stop users from changing their webserver at all.

## How OpenPanel lists webservers

There is no hardcoded list of webservers. OpenPanel reads them from the services in each user's `/home/<USERNAME>/docker-compose.yml` file. A service counts as a webserver if it mounts the `webserver_data` volume and publishes ports. When a new user is created, that file is copied from the template at `/etc/openpanel/docker/compose/1.0/docker-compose.yml`.

The webservers in the template are offered on:

- **OpenAdmin > Settings > [User Defaults](/docs/admin/settings/defaults/)**
- the **OpenAdmin > Users > Add User** form
- `opencli user-add --webserver`. A webserver that isn't in the template is ignored, and the template default is used.

The webservers in a user's own `docker-compose.yml` are offered on:

- **OpenPanel > Web Server Config > [Web Server Type](/docs/panel/webserver/webserver_type/)**, and the API endpoint behind it
- the webserver step of the onboarding wizard on the OpenPanel dashboard
- the webserver setting on the user's page in OpenAdmin

A webserver that isn't on the list is rejected, even if someone calls the API directly.

:::info
A user can only change the webserver while the account has no domains. This applies in OpenAdmin too.
:::

## Option 1: Stop users from changing the webserver

If you only want every user to stay on the webserver you assign, disable the [Switch WebServer (`change_ws`)](/docs/admin/settings/modules/#change-ws) module from **OpenAdmin > Settings > Modules**. To disable it for some users only, remove it from the [feature set](/docs/admin/plans/feature-manager) of their plan.

Users then don't have access to the **Web Server Type** page, and the onboarding wizard skips the webserver step. You can still set the webserver when creating the user, or later from the user's page in OpenAdmin.

## Option 2: Remove a webserver for new users

Remove the webserver's service block from the template. You can edit it from **OpenAdmin > Settings > User Defaults > Advanced**, or in the file `/etc/openpanel/docker/compose/1.0/docker-compose.yml`. The block starts at the service name, for example `  openresty:`, and ends right before the next service.

Or remove it with a script. Set `WS` to the webserver you want to remove:

```bash
WS=openresty
f=/etc/openpanel/docker/compose/1.0/docker-compose.yml
cp -p "$f" "$f.bak"
awk -v svc="$WS" '
  $0 ~ "^  " svc ":[ \t]*$" { skip=1; next }
  skip && (/^  [^ #]/ || /^[^ \t#]/) { skip=0 }
  !skip
' "$f" > "$f.tmp" && cat "$f.tmp" > "$f" && rm -f "$f.tmp"
```

Then make sure the default webserver isn't the one you removed. Check the **Webserver** option on the [User Defaults](/docs/admin/settings/defaults/) page, or `WEB_SERVER` in `/etc/openpanel/docker/compose/1.0/.env`.

:::warning
Don't remove a webserver that another service depends on in the template. If Varnish is enabled by default, it proxies to the default webserver.
:::

## Option 3: Remove a webserver for existing users

Existing users keep their own copy of `docker-compose.yml`, so the template change doesn't affect them.

### 1. Find users that use it

```bash
WS=openresty
grep -l "^WEB_SERVER=\"\?$WS\"\?$" /home/*/.env | cut -d/ -f3
```

Leave the service in place for these users. Their websites stop working without it, and their webserver can only be changed after all their domains are removed.

### 2. Remove the service for everyone else

This loops through all users, skips the ones that use the webserver, and checks that each file is still valid afterwards:

```bash
WS=openresty
for f in /home/*/docker-compose.yml; do
  u=$(basename "$(dirname "$f")")
  grep -q "^  $WS:" "$f" || continue
  if grep -q "^WEB_SERVER=\"\?$WS\"\?$" "/home/$u/.env"; then echo "$u: uses $WS, skipped"; continue; fi
  cp -p "$f" "$f.bak"
  awk -v svc="$WS" '
    $0 ~ "^  " svc ":[ \t]*$" { skip=1; next }
    skip && (/^  [^ #]/ || /^[^ \t#]/) { skip=0 }
    !skip
  ' "$f" > "$f.tmp" && cat "$f.tmp" > "$f" && rm -f "$f.tmp"
  (cd "/home/$u" && podman-compose config >/dev/null 2>&1) && echo "$u: OK" || { cp -p "$f.bak" "$f"; echo "$u: invalid config, restored backup"; }
done
```

## Don't run `opencli update --compose`

[`opencli update --compose`](/docs/articles/opencli/update/#templates) replaces the whole `docker-compose.yml` template with the latest upstream version, which brings back every removed webserver. Existing users aren't affected. If you run it, compare the new template with the backup it saves at `/etc/openpanel/docker/compose/1.0/docker-compose.yml.bak`, and remove the webservers again.

To add a removed webserver back later, copy its service block from the [upstream template](https://github.com/stefanpejcic/openpanel-configuration/blob/main/docker/compose/1.0/docker-compose.yml).

## Related

- [How to Remove Old PHP Versions](/docs/articles/user-experience/how-to-remove-old-php-versions/)
- [How to set Apache, Nginx, OpenResty, OpenLiteSpeed or Varnish as the default webserver](/docs/articles/containers/how-to-set-nginx-apache-varnish-per-user-in-openpanel/)
