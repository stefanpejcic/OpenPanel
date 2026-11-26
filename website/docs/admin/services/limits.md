---
sidebar_position: 2
---

# Service limits

The **Service limits** section shows what each system service uses right now and lets Administrators change its CPU and memory limits.

![Service Limits page with CPU and memory limits for each system service](/img/openadmin-screenshots/services/limits-page.png#gh-light-mode-only)
![Service Limits page with CPU and memory limits for each system service](/img/openadmin-screenshots/services/limits-page_dark.png#gh-dark-mode-only)

---

Each row is one service container with:

- a status dot, green when it's running, orange or red when it's close to a limit, gray when it isn't running
- **CPU** and **Memory**: a ring with the current usage against the limit, the usage and the allocated limit
- **Processes**: how many processes the container runs

Click the arrows next to a column name to sort by it.

To change a limit, click the pencil next to **Allocated**, enter the new value and click **Save**. CPU and memory apply to the running container right away. A service that isn't running gets the new limit when it's started from [Service status](/docs/admin/services/status).

`0` means no limit. A running container can't drop a CPU or memory limit, so after saving `0` the page offers to restart the service. Restarting recreates the container, together with the services that depend on it (restarting redis or mysql also restarts openpanel), and it's down for a few seconds.

Click a service name to open its other settings, like the timezone or the version. A `*_VERSION` setting whose image tag comes from `/root/.env` is a select of that image's tags on Docker Hub, cached for a day. A new version or another setting only applies after a restart, the page offers it right after saving.

**Available Services** (by default):

- openpanel
- caddy
- mysql
- clamav
- redis
- bind9
- ftp
- phpmyadmin

This list reflects the entries currently present in `/root/.env`, the page is built from that file, so any additional `*_CPU`/`*_RAM` (or `*_CPUS`/`*_RAM`) pairs added there also appear as editable services. A service's limits are matched to its container through `/root/docker-compose.yml`.

### Recommendations

Under each limit a recommendation is shown when the value is below the recommended minimum or above the recommended maximum, based on the server's CPU and RAM, the enabled modules and the number of users and domains. `0` means unlimited and never shows a recommendation. CPU is in cores, RAM in GB.

| Service | Recommended |
|---|---|
| **openpanel** | at least 2 on servers with over 4 GB RAM, otherwise at least 1 |
| **caddy** | at least 1 per 20 domains so SSL certificates can be issued and renewed, or a third of the server's RAM/CPU on servers under 10 GB RAM. No more than 8 GB RAM and 6 CPU |
| **mysql** | 0.5 per 10 users, no more than a third of the openpanel container's limit |
| **bind9** | at least 0.1, only when the DNS module is enabled |
| **clamav** | 4 (2 is the minimum), no more than 6, only when the malware scanner module is enabled |
| **ftp** | 0.5 plus 0.1 per 10 users, no more than 3 GB RAM and 2 CPU, only when the FTP module is enabled |
| **phpmyadmin** | at least 2 CPU, and RAM of at least 1 GB or more to fit `PHPMYADMIN_MEMORY_LIMIT` and `PHPMYADMIN_UPLOAD_LIMIT`, only when the phpMyAdmin module is enabled |

`CLAMD_STARTUP_DELAY` is recommended between 30 and 600 seconds, since ClamAV downloads its malware signature database on startup.

---

To edit limits via the terminal, modify the `/root/.env` file directly.
