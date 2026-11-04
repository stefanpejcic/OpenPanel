---
sidebar_label: "Limit Services for Users"
---

# How to Limit Which Services Users Can Run

Every OpenPanel user has their own set of services, such as Redis, Memcached, Elasticsearch, OpenSearch, PostgreSQL and MongoDB. If you don't want to offer or support some of them, you can hide them in OpenPanel, or remove them so they can't be started at all.

## How OpenPanel lists services

The services a user can run are the ones defined in their `/home/<USERNAME>/docker-compose.yml` file. When a new user is created, that file is copied from the template at `/etc/openpanel/docker/compose/1.0/docker-compose.yml`. Services are only defined there. They don't run, and don't use any resources, until the user or a feature starts them.

There are two ways to limit a service, and you can combine them:

| | Disable the module | Remove the service |
|---|---|---|
| What it does | Hides the page for that service in OpenPanel | Removes the service from the user's stack |
| Where | **OpenAdmin > Settings > [Modules](/docs/admin/settings/modules/)**, or per plan with [feature sets](/docs/admin/plans/feature-manager) | The `docker-compose.yml` template and users' own files |
| Existing users | Applies right away | Each user's file has to be edited |
| Service can still be started | Yes, from the **Services** or **Containers** page | No |

## Modules per service

| Service | Module | OpenPanel page |
|---|---|---|
| `redis` | `redis` | Cache & Search > Redis |
| `valkey` | `valkey` | Cache & Search > Valkey |
| `memcached` | `memcached` | Cache & Search > Memcached |
| `elasticsearch` | `elasticsearch` | Cache & Search > ElasticSearch |
| `opensearch` | `opensearch` | Cache & Search > OpenSearch |
| `varnish` | `varnish` | Cache & Search > Varnish |
| `postgres`, `pgadmin` | `postgresql` | PostgreSQL pages |
| `mongodb` | `mongodb` | MongoDB pages |

Every service in the user's file is also listed on **Processes & Services > Services** (`services` module) and on **Containers** (`docker` module). To fully hide a service, disable its module and remove it from the stack.

:::warning
The `docker` module lets users add their own services on the **Containers** page, with any image. If you need to make sure users only run what you offer, disable the `docker` module too. The `services` module still lets them start and stop the services you left in place.
:::

## Services you shouldn't remove

- `cron`, `backup` and `docker-proxy` are used by OpenPanel itself.
- `mysql` and `mariadb`: to keep users on one database type, disable the [Switch MySQL Type (`change_db`)](/docs/admin/settings/modules/#change-db) module instead, and set the default type on the [User Defaults](/docs/admin/settings/defaults/) page.
- Webservers and PHP versions can be removed, but they need extra steps. See [How to Disable Webservers for Users](/docs/articles/web-servers/how-to-disable-webservers-for-users/) and [How to Remove Old PHP Versions](/docs/articles/user-experience/how-to-remove-old-php-versions/).

## Remove a service for new users

Remove the service block from the template. You can edit it from **OpenAdmin > Settings > User Defaults > Advanced**, or in the file `/etc/openpanel/docker/compose/1.0/docker-compose.yml`. The block starts at the service name, for example `  elasticsearch:`, and ends right before the next service.

Or remove it with a script. Set `SERVICES` to the services you want to remove:

```bash
SERVICES="elasticsearch opensearch"
f=/etc/openpanel/docker/compose/1.0/docker-compose.yml
cp -p "$f" "$f.bak"
for s in $SERVICES; do
  awk -v svc="$s" '
    $0 ~ "^  " svc ":[ \t]*$" { skip=1; next }
    skip && (/^  [^ #]/ || /^[^ \t#]/) { skip=0 }
    !skip
  ' "$f" > "$f.tmp" && cat "$f.tmp" > "$f" && rm -f "$f.tmp"
done
```

If the service is in the **Autostart services** list on the [User Defaults](/docs/admin/settings/defaults/) page, remove it from there too.

## Remove a service for existing users

### 1. Find users that run it

A service that is running has data or websites that depend on it. Check before you remove it:

```bash
SERVICE=elasticsearch
source /usr/local/opencli/lib/podman.sh
for f in /home/*/docker-compose.yml; do
  u=$(basename "$(dirname "$f")")
  podman_user "$u" ps --format '{{.Names}}' 2>/dev/null | grep -qx "$SERVICE" && echo "$u"
done
```

Ask these users to move off the service, or leave it in place for them.

### 2. Remove it for everyone else

This skips users that run the service, and checks that each file is still valid afterwards:

```bash
SERVICE=elasticsearch
source /usr/local/opencli/lib/podman.sh
for f in /home/*/docker-compose.yml; do
  u=$(basename "$(dirname "$f")")
  grep -q "^  $SERVICE:" "$f" || continue
  if podman_user "$u" ps --format '{{.Names}}' 2>/dev/null | grep -qx "$SERVICE"; then echo "$u: runs $SERVICE, skipped"; continue; fi
  cp -p "$f" "$f.bak"
  awk -v svc="$SERVICE" '
    $0 ~ "^  " svc ":[ \t]*$" { skip=1; next }
    skip && (/^  [^ #]/ || /^[^ \t#]/) { skip=0 }
    !skip
  ' "$f" > "$f.tmp" && cat "$f.tmp" > "$f" && rm -f "$f.tmp"
  (cd "/home/$u" && podman-compose config >/dev/null 2>&1) && echo "$u: OK" || { cp -p "$f.bak" "$f"; echo "$u: invalid config, restored backup"; }
done
```

The service's data volume, for example `es_data`, is not deleted. Remove it yourself when you no longer need the data.

## Don't run `opencli update --compose`

[`opencli update --compose`](/docs/articles/opencli/update/#templates) replaces the whole template with the latest upstream version, which brings back every removed service for new users. Existing users aren't affected. If you run it, remove the services again. The previous template is saved at `/etc/openpanel/docker/compose/1.0/docker-compose.yml.bak`.

To add a removed service back later, copy its block from the [upstream template](https://github.com/stefanpejcic/openpanel-configuration/blob/main/docker/compose/1.0/docker-compose.yml).

## Related

- [How to Restrict Software Versions for Users](/docs/articles/containers/how-to-restrict-software-versions-for-users/)
- [How to Disable Webservers for Users](/docs/articles/web-servers/how-to-disable-webservers-for-users/)
- [How to Remove Old PHP Versions](/docs/articles/user-experience/how-to-remove-old-php-versions/)
