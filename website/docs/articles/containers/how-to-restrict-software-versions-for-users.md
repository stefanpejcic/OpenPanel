---
sidebar_label: "Restrict Software Versions"
---

# How to Restrict Software Versions for Users

On the **Containers > [Software Versions](/docs/panel/containers/change/)** page, OpenPanel users can change the image tag, which is the version, of their services. They can pick any tag from Docker Hub or type one in, for example an old MySQL, Nginx or Redis release. If you only want to support the versions you choose, you can turn this off for everyone, for some plans, or for single services.

## How the version is set

Most services in the `docker-compose.yml` template read their tag from a variable in the `.env` file, with a default after `:-`:

```yaml
  redis:
    image: redis:${REDIS_VERSION:-8.6.2-alpine}
```

```bash
REDIS_VERSION="8.6.2-alpine"
```

When a user picks a new tag on the Software Versions page, OpenPanel writes it to that variable in the user's `/home/<USERNAME>/.env` file and recreates the service. MySQL and MariaDB share the `MYSQL_VERSION` variable.

The PHP services and the internal `cron`, `backup` and `docker-proxy` services aren't listed on the page. Their images are managed by OpenPanel. To limit PHP versions, see [How to Remove Old PHP Versions](/docs/articles/user-experience/how-to-remove-old-php-versions/).

## Option 1: Disable the Software Versions page

Disable the [Change Container Image Tag (`change_image`)](/docs/admin/settings/modules/#change-image) module from **OpenAdmin > Settings > Modules**. To disable it for some users only, remove it from the [feature set](/docs/admin/plans/feature-manager) of their plan.

Users keep the versions they have now, and you decide which versions are used:

- **New users** get the tags from `/etc/openpanel/docker/compose/1.0/.env`. You can edit it from **OpenAdmin > Settings > [User Defaults](/docs/admin/settings/defaults/) > Advanced**.
- **Existing users:** change the variable in their `/home/<USERNAME>/.env` file, then recreate the service. For example, to move a user to Redis 8.6:

```bash
USERNAME=dorotea
sed -i 's/^REDIS_VERSION=.*/REDIS_VERSION="8.6.2-alpine"/' /home/$USERNAME/.env
source /usr/local/opencli/lib/podman.sh
cd /home/$USERNAME && CONTAINER_HOST=$(podman_user_socket "$USERNAME") podman-compose up -d --force-recreate redis
```

:::warning
The `docker` module lets users add their own services on the **Containers** page, with any image and version. If you need to make sure users only run the versions you choose, disable the `docker` module too.
:::

## Option 2: Pin the version of one service

To keep the Software Versions page for most services, but fix the version of one of them, replace the variable in the `image:` line with a fixed tag. For example, in `/etc/openpanel/docker/compose/1.0/docker-compose.yml` for new users, or in `/home/<USERNAME>/docker-compose.yml` for an existing user:

```yaml
  redis:
    image: redis:8.6.2-alpine
```

Docker Hub tags are then no longer listed for that service. The service is still shown on the page and a tag can still be typed in, but it's only saved to the `.env` file and isn't used. Tell your users which services are pinned, so they aren't surprised when nothing changes. To fully remove a service instead, see [How to Limit Which Services Users Can Run](/docs/articles/containers/how-to-limit-services-for-users/).

## Find users on old versions

To see which tag every user has set for a service:

```bash
grep -H '^REDIS_VERSION=' /home/*/.env
```

Users without the variable use the default from their `docker-compose.yml`.

## Related

- [Change Image Tag](/docs/panel/containers/change/) (OpenPanel docs for users)
- [How to Limit Which Services Users Can Run](/docs/articles/containers/how-to-limit-services-for-users/)
- [How to Remove Old PHP Versions](/docs/articles/user-experience/how-to-remove-old-php-versions/)
