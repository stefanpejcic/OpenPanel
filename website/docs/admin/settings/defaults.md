---
sidebar_position: 3
---

# Defaults

From **OpenAdmin > Settings > Defaults** Administrators can edit values for the `docker-compose.yml` and `.env` files used for new users.

![defaults basic](https://i.postimg.cc/KFRzLrGY/admin-defaults.png)

These files determine services and limits for new users.

---

Using the 'Advanced' option you can directly edit the files.

![defaults advanced](https://i.postimg.cc/74BhfQyc/admin-defaults-advanced.png)

In these files you can configure additional services (docker containers) and change defaults for existing services.

Keep in mind that this is intended for advanced users and misconfiguration can cause exposed system ports, user hogging resources or exciding disk limits.

When adding new services keep in mind the following:

- container name must be same as service name
- cpu and mempory limits for service must be named in format: `SERVICE_`CPU and `SERVICE_`RAM.
- other variables for service should also be prefixed with `SERVICE_`
- processes inside containers must be run as root (`0`) user in order for container files to be counted against user quota and avoid permission issues.
- configuration files should be mounted in read-only mode.

