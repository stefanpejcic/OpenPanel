# Managing User Containers from the Terminal

In OpenPanel, each user has their own [Docker context](https://docs.docker.com/engine/manage-resources/contexts/) running in [rootless mode](https://docs.docker.com/engine/security/rootless/). If you are familiar with Docker, you can manage a user’s files and services directly from the terminal.

SSH access as the **root user** on the server is required to switch between Docker contexts.

---

## Compose Setup

Each user’s environment is based on:

* **`docker-compose.yml`** – defines services, networks, and volumes.
* **`.env`** – defines CPU/memory limits, Docker image tags, and custom ports.

> ⚠️ Always edit these files via the OpenAdmin interface. The interface validates your changes. Editing directly from the terminal bypasses validation and can break services. Make a backup before any modifications.

* [Volume Management in OpenPanel](/docs/articles/docker/volume-management-openpanel/)
* [Network Isolation in OpenPanel](/docs/articles/docker/network-isolation-openpanel/)

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
├── docker-data          # Docker storage
│   ├── containerd       # Container runtime data
│   ├── containers       # Individual container data
│   ├── image            # Docker images
│   ├── network          # Docker networks
│   ├── overlay2         # Overlay filesystem storage
│   ├── plugins          # Docker plugins
│   └── volumes          # Persistent data volumes
│       ├── <USERNAME>_html_data   # Website files (/var/www/html/)
│       └── <USERNAME>_mysql_data  # MySQL database files (/var/lib/mysql/)
├── httpd.conf           # Apache configuration
├── my.cnf               # MySQL root credentials for service commands
├── nginx.conf           # NGINX configuration
├── openlitespeed.conf   # OpenLiteSpeed configuration
├── openresty.conf       # OpenResty configuration
├── php.ini/             # PHP configuration files
├── pma.php              # phpMyAdmin entry point
└── sockets              # Sockets (MySQL, Redis, PostgreSQL, etc.)
```

* Website files are stored in the Docker volume `USERNAME_html_data` and physically located at:

```
/home/USERNAME/docker_data/volumes/USERNAME_html_data/_data/
```

* File ownership should match the OpenPanel user ID. Check with:

```bash
id -u USERNAME
```

* Docker runs in **rootless mode**, mapping the user UID to `root` inside containers. This avoids permission issues while giving the user root-level permissions in their services.

* Disk quotas are enforced using the [quota](https://linux.die.net/man/2/quotactl) tool. To check disk and inodes usage for user:

```bash
quota -u USERNAME
```

---

## Managing Services

Services are defined in `/home/USERNAME/docker-compose.yml`. You can extend this file with custom services, but to maintain compatibility with the OpenPanel UI, [follow the guidelines](#).

Use the user’s Docker context to manage their services. Examples:

* **List running services:**

```bash
docker --context=USERNAME ps -a
```

* **Stop a service:**

```bash
docker --context=USERNAME stop mysql
```

* **Restart a service:**

```bash
cd /home/USERNAME && \
docker --context=USERNAME compose down mysql && \
docker --context=USERNAME compose up -d mysql
```

* **Exec into a service:**

```bash
docker --context=USERNAME exec -it mysql bash
```

* **View resource usage:**

```bash
docker --context=USERNAME stats --no-stream
```

---

## Domains

Domain configuration files are stored in:

* **Caddy configuration:**
```
/etc/openpanel/caddy/domains/<DOMAIN>.conf
```

Each domain has its own Caddyfile. Caddy handles SSL certificates and acts as a reverse proxy to the user’s web server.
For more information, see: [How Web Traffic Flows with User Containers](https://openpanel.com/docs/articles/docker/how-traffic-flows-in-openpanel/)


* **BIND9 zone file:**
```
/etc/bind/zones/<DOMAIN>.db
```
Contains the DNS records for the domain.

---

## Backups

OpenPanel uses [offen/docker-volume-backup](https://offen.github.io/docker-volume-backup/) for backups. Configuration is stored in `/home/{CONTEXT}/backup.env`.

* Administrators can schedule automatic backups for users.
* Users can also enable and manage their own backups if permitted.
* Only Docker volumes are backed up, which include the actual user data (website files and databases).

For detailed instructions, see: [Configuring OpenPanel Backups](/docs/articles/backups/configuring-backups)

---

## Crons

OpenPanel uses [mcuadros/ofelia](https://github.com/mcuadros/ofelia) for cron obs. Configuration is stored in `/home/{CONTEXT}/crons.ini`.

---

## Ports

Ports are defined in the user’s `.env` file:

```bash
root@vagabundo:/# grep _PORT home/mysqlbaja/.env 
HTTP_PORT="127.0.0.1:32780:80"
HTTPS_PORT="127.0.0.1:32781:443"
PROXY_HTTP_PORT="127.0.0.1:32782:80" 
MYSQL_PORT="127.0.0.1:32777:3306"
PMA_PORT="32779:80"
POSTGRES_PORT="0:5432"
PGADMIN_PORT="0:80"
```

* Ports are generated during user creation but can be modified by the administrator. After changes, run `compose down && compose up` for the changes to take effect.
* ⚠️ Docker rootless mode prevents the use of ports under **1024**.
* Services should not expose ports externally unless remote access is required (e.g., phpMyAdmin, remote MySQL, or pgAdmin).

---
