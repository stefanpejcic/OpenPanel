---
sidebar_position: 3
---

# Defaults

From **OpenAdmin > Settings > Defaults** Administrators can edit values for the `docker-compose.yml` and `.env` files used for new users.

![defaults basic](https://i.postimg.cc/KFRzLrGY/admin-defaults.png)

These files determine services and limits for new users.

> [How to setup Apache, Nginx, OpenResty, OpenLiteSpeed, and Varnish as default webserver](/docs/articles/docker/how-to-set-nginx-apache-varnish-per-user-in-openpanel/)
> [How to setup MySQL, MariaDB or Percona for default database type](/docs/articles/docker/how-to-set-mysql-mariadb-per-user-in-openpanel/)

---

Using the 'Advanced' option you can directly edit the files.

![defaults advanced](https://i.postimg.cc/74BhfQyc/admin-defaults-advanced.png)

In these files you can configure additional services (docker containers) and change defaults for existing services.

Keep in mind that this is intended for advanced users and misconfiguration can cause exposed system ports, user hogging resources or exceeding disk limits.

For more information refer to [How to add custom docker images](/docs/articles/docker/how-to-add-custom-docker-image-for-openpanel-user)
