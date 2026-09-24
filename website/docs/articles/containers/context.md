# Managing User Containers from the Terminal

In OpenPanel, each user has their own [rootless Podman](https://docs.podman.io/en/latest/markdown/podman.1.html#rootless-mode) instance. If you are familiar with Docker or Podman, you can manage a user’s files and services directly from the terminal.

SSH access as the **root user** on the server is required to manage another user's containers.

---

## Compose Setup

Each user’s environment is based on:

* **`docker-compose.yml`** – defines services, networks, and volumes.
* **`.env`** – defines CPU/memory limits, image tags, and custom ports.

> ⚠️ Always edit these files via the OpenAdmin interface. The interface validates your changes. Editing directly from the terminal bypasses validation and can break services. Make a backup before any modifications.

* [Volume Management in OpenPanel](/docs/articles/containers/volume-management-openpanel/)
* [Network Isolation in OpenPanel](/docs/articles/containers/network-isolation-openpanel/)

---

## File Structure

Each user’s home directory is located at:

```
/home/USERNAME/
```

It contains configuration files and user data. Example structure:

```
├── .env                 # Environment variables: versions, ports, CPU/memory limits
├── backup.env           # Backup configuration
├── crons.ini            # Cron jobs
├── custom.cnf           # MySQL configuration
├── default.vcl          # Varnish cache configuration
├── docker-compose.yml   # Service definitions
├── docker-data          # Podman storage (graphroot)
│   ├── libpod           # Podman state database
│   ├── networks         # Container networks
│   ├── overlay          # Overlay filesystem storage
│   ├── overlay-containers  # Individual container data
│   ├── overlay-images   # User's own images (shared images are reused from /var/lib/containers/shared-storage)
│   └── volumes          # Persistent data volumes
│       ├── <USERNAME>_html_data   # Website files (/var/www/html/)
│       └── <USERNAME>_mysql_data  # MySQL database files (/var/lib/mysql/)
├── .config/containers/  # Podman config for the user (storage.conf, containers.conf)
├── httpd.conf           # Apache configuration
├── my.cnf               # MySQL root credentials for service commands
├── nginx.conf           # NGINX configuration
├── openlitespeed.conf   # OpenLiteSpeed configuration
├── openresty.conf       # OpenResty configuration
├── php.ini/             # PHP configuration files
├── pma.php              # phpMyAdmin entry point
└── sockets              # Sockets (MySQL, Redis, PostgreSQL, etc.)
```

* Website files are stored in the volume `USERNAME_html_data` and physically located at:

```
/home/USERNAME/docker-data/volumes/USERNAME_html_data/_data/
```

* The folder is still called `docker-data` for compatibility with older installations - it holds the user's Podman storage.

* File ownership should match the OpenPanel user ID. Check with:

```bash
id -u USERNAME
```

* Podman runs in **rootless mode**, mapping the user UID to `root` inside containers. This avoids permission issues while giving the user root-level permissions in their services.

* Disk quotas are enforced using the [quota](https://linux.die.net/man/2/quotactl) tool. To check disk and inodes usage for user:

```bash
quota -u USERNAME
```

---

## Managing Services

Services are defined in `/home/USERNAME/docker-compose.yml`. You can extend this file with custom services, but to maintain compatibility with the OpenPanel UI, [follow the guidelines](#).

As root, load the OpenPanel Podman helpers first - they run `podman` and `podman-compose` against the user's own rootless instance:

```bash
source /usr/local/opencli/lib/podman.sh
```

Examples:

* **List running services:**

```bash
podman_user USERNAME ps -a
```

* **Stop a service:**

```bash
podman_user USERNAME stop mysql
```

* **Restart a service:**

```bash
cd /home/USERNAME && \
podman_compose_user USERNAME down mysql && \
podman_compose_user USERNAME up -d mysql
```

* **Open a shell in a service:**

```bash
opencli docker USERNAME mysql
```

Run `opencli docker USERNAME` without a container name to pick one from a list. See [opencli docker](/docs/articles/opencli/docker/).

* **View resource usage:**

```bash
podman_user USERNAME stats --no-stream
```

* **View a service's logs:**

```bash
podman_user USERNAME logs --tail 100 mysql
```

---

## Domains

Domain configuration files are stored in:

* **Caddy configuration:**
```
/etc/openpanel/caddy/domains/<DOMAIN>.conf
```

Each domain has its own Caddyfile. Caddy handles SSL certificates and acts as a reverse proxy to the user’s web server.
For more information, see: [How Web Traffic Flows with User Containers](https://openpanel.com/docs/articles/containers/how-traffic-flows-in-openpanel/)


* **BIND9 zone file:**
```
/etc/bind/zones/<DOMAIN>.zone
```
Contains the DNS records for the domain.

---

## Backups

OpenPanel uses [offen/docker-volume-backup](https://offen.github.io/docker-volume-backup/) for backups. Configuration is stored in `/home/{CONTEXT}/backup.env`.

* Administrators can schedule automatic backups for users.
* Users can also enable and manage their own backups if permitted.
* Only volumes are backed up, which include the actual user data (website files and databases).

For detailed instructions, see: [Configuring OpenPanel Backups](/docs/articles/backups/configuring-backups)

---

## Crons

OpenPanel uses [mcuadros/ofelia](https://github.com/mcuadros/ofelia) for cron jobs. Configuration is stored in `/home/{CONTEXT}/crons.ini`.

---

## Ports

Ports are defined in the user’s `.env` file:

```bash
root@stefan:/# grep _PORT /home/stefan/.env 
HTTP_PORT="127.0.0.1:32780:80"
HTTPS_PORT="127.0.0.1:32781:443"
PROXY_HTTP_PORT="127.0.0.1:32782:80" 
MYSQL_PORT="127.0.0.1:32777:3306"
PMA_PORT="32779:80"
POSTGRES_PORT="0:5432"
PGADMIN_PORT="0:80"
```

* Ports are generated during user creation but can be modified by the administrator. After changes, run `podman_compose_user USERNAME down` and `podman_compose_user USERNAME up -d` from `/home/USERNAME` for the changes to take effect.
* ⚠️ Rootless Podman prevents the use of ports under **1024**.
* Services should not expose ports externally unless remote access is required (e.g., phpMyAdmin, remote MySQL, or pgAdmin).

---
