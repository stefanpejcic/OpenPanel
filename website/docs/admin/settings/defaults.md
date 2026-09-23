---
sidebar_position: 3
---

# Defaults

From **OpenAdmin > Settings > User Defaults** Administrators can edit values for the `docker-compose.yml` and `.env` files used for new users.

![Edit Defaults page with the default webserver, database, Varnish cache, PHP version and autostart services for new users](/img/openadmin-screenshots/settings/defaults-page.png#gh-light-mode-only)
![Edit Defaults page with the default webserver, database, Varnish cache, PHP version and autostart services for new users](/img/openadmin-screenshots/settings/defaults-page_dark.png#gh-dark-mode-only)

These files determine services and limits for new users.

> [How to setup Apache, Nginx, OpenResty, OpenLiteSpeed, and Varnish as default webserver](/docs/articles/containers/how-to-set-nginx-apache-varnish-per-user-in-openpanel/)
> [How to setup MySQL, MariaDB or Percona for default database type](/docs/articles/containers/how-to-set-mysql-mariadb-per-user-in-openpanel/)

---

Using the 'Advanced' option you can directly edit the files.

![Service limits on the Edit Defaults page for new user accounts](/img/openadmin-screenshots/settings/defaults-services.png#gh-light-mode-only)
![Service limits on the Edit Defaults page for new user accounts](/img/openadmin-screenshots/settings/defaults-services_dark.png#gh-dark-mode-only)

In these files you can configure additional services (docker containers) and change defaults for existing services.

Keep in mind that this is intended for advanced users and misconfiguration can cause exposed system ports, user hogging resources or exceeding disk limits.

For more information refer to [How to add custom docker images](/docs/articles/containers/how-to-add-custom-docker-image-for-openpanel-user)
