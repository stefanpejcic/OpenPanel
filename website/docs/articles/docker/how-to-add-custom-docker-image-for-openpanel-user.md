# Adding custom docker images for OpenPanel users

Each user has a Docker service running in [Rootless mode](https://docs.docker.com/engine/security/rootless/) and a single `docker-compose.yml` file that has all their services defined.

For each user you can edit their `/home/USER/docker-compose.yml` file to specify custom services.


## Default services

To edit services for all new users that you create, edit the template files:

- `/etc/openpanel/docker/compose/1.0/docker-compose.yml` - services for users, volumes and networks.
- `/etc/openpanel/docker/compose/1.0/.env` - limits for services.

## Guidelines

You can add **any Docker image** by including its Compose configuration. However, there are a few important **rules** to follow so that OpenPanel can:

* Recognize the service as valid
* Allow editing of resource limits from the GUI
* Enable service start/stop control from the GUI
* Monitor usage statistics
* Enforce resource restrictions

Make sure your service definition follows the required structure to ensure full OpenPanel integration.


| **Name**          | **Description**                                                                                                                                                                                                                                                                                                | **Example**                                                                                                                     |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| **service name**  | The service name must exactly match the `container_name` and **cannot contain spaces**.                                                                                                                                                                                                                        | For `uptimekuma`:<br>`container_name: uptimekuma`                                                                               |
| **env variables** | All variables in the `.env` file must start with the service name in **uppercase**, followed by an underscore. Each service must define `_CPU` and `_RAM` to allow OpenPanel to restrict and monitor resource usage.                                                                                           | `UPTIMEKUMA_VERSION="1"`<br>`UPTIMEKUMA_CPU="0.5"`<br>`UPTIMEKUMA_RAM="0.5G"`<br>                                               |
| **image tag**     | The image tag should be a variable, allowing it to be changed via the OpenPanel UI. If not provided, a fallback value is used.                                                                                                                                                                                 | `image: louislam/uptime-kuma:${UPTIMEKUMA_VERSION:-1}`                                                                          |
| **volumes**       | Mounting host OS paths can expose the server. Use relative paths (e.g., `./data`) for app data. To use existing data like `/var/www/html/`, mount the appropriate volume. If Docker socket access is needed, mount `/hostfs/run/user/${USER_ID}/docker.sock` as **read-only** to prevent privilege escalation. | `- ./data:/app/data`<br>`- html_data:/var/www/html/`<br>`- /hostfs/run/user/${USER_ID}/docker.sock:/var/run/docker.sock:ro` |
| **resources**     | Define `cpus`, `memory`, and `pids` under the `deploy.resources.limits` section. Without `pids`, services are vulnerable to fork bombs. Use variables with fallback values for `cpus` and `memory`.                                                                                                            | `cpus: "${BUSYBOX_CPU:-0.1}"`<br>`memory: "${BUSYBOX_RAM:-0.1G}"<br>pids: 100`                                            |
| **networks**      | Only add networks if the service needs access to other containers. Use either `www` or `db` networks. `www` is for app/webserver access; `db` is for database-only access.                                                                                                                                     | `- www`<br>`- db`                                                                                                         |
| **environment**   | Use the `environment` section to pass custom environment variables.                                                                                                                                                                                                                                            | `EULA: "TRUE"`<br>`ENABLE_QUERY: "${MINECRAFT_ENABLE_QUERY:-true}"`                                                       |
| **ports**         | Only define ports if external access is required. For internal-only services (e.g., Redis, Memcached), **do not expose ports**.                                                                                                                                                                                | `- "${MYSQL_PORT}"`<br>`- "${MINECRAFT_PORT:-25565}:25565"`                                                               |
| **labels**        | Labels are optional and ignored by OpenPanel but can be used for external tools or metadata.                                                                                                                                                                                                                   | `- docker-volume-backup.archive-pre=/bin/sh -c '/dump.sh'`                                                                      |
| **healthcheck**   | Optional. If defined, OpenPanel respects the health check and uses it to manage restarts.                                                                                                                                                                                                                      | `test: ['CMD-SHELL', 'mysqladmin ping -h localhost']`<br>`interval: 1s`<br>`timeout: 5s`<br>`retries: 10`                     |
| restart policy                                 | restart policy should be explicitly set to [unless-stopped](https://docs.docker.com/engine/containers/start-containers-automatically/#use-a-restart-policy) so that OpenPanel can auto-restart services in case of failure, except when user account is suspended. | `restart: unless-stopped` |

## Examples

These examples are **drop-in snippets** you can insert into your files to add a new service for an OpenPanel user.

### FileBrowser

To add [FileBrowser](https://github.com/filebrowser/filebrowser) as a service for user in OpenPanel:

add to .env file:

```bash
# FILEBROWSER
FILEBROWSER_VERSION="s6"
FILEBROWSER_CPU="0.25"
FILEBROWSER_RAM="0.35"
```

add to `docker-compose.yml` file in the **services** section:

```bash
  filebrowser:
    image: filebrowser/filebrowser:${FILEBROWSER_VERSION:-s6}
    container_name: filebrowser
    volumes:
      - html_data:/srv
      - ./filebrowser/config/:/config/
      - ./filebrowser/database/:/database/
    environment:
      - PUID=${USER_ID:-0}
      - PGID=${USER_ID:-0}
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: "${FILEBROWSER_CPU:-0.35}"
          memory: "${FILEBROWSER_RAM:-0.35G}"   
          pids: 100
    networks:
      - www
```




### Minecraft

To add [Minecraft](https://github.com/itzg/docker-minecraft-server) as a service for user in OpenPanel:

add to .env file:

```bash
# MINECRAFT
MINECRAFT_VERSION="latest"
MINECRAFT_PORT="25565"
MINECRAFT_CPU="1.0"
MINECRAFT_RAM="1.0G"
MINECRAFT_ENABLE_QUERY="true"
MINECRAFT_MAX_PLAYERS="20"
MINECRAFT_MAX_WORLD_SIZE="10000"
MINECRAFT_ALLOW_NETHER="false"
MINECRAFT_ANNOUNCE_PLAYER_ACHIEVEMENTS="false"
MINECRAFT_ENABLE_COMMAND_BLOCK="false"
```

add to `docker-compose.yml` file in the **volumes** section:

```bash
  mc_data:
    driver: local
    labels:
      description: "This volume holds the minecraft data directory."
      purpose: "storage"
```

add to `docker-compose.yml` file in the **services** section:

```bash
  minecraft:
    image: itzg/minecraft-server:${MINECRAFT_VERSION:-latest}
    container_name: minecraft
    tty: true
    stdin_open: true
    ports:
      - "${MINECRAFT_PORT:-25565}:25565"
    environment:
      EULA: "TRUE"
      ENABLE_QUERY: "${MINECRAFT_ENABLE_QUERY:-true}"
      QUERY_PORT: "${MINECRAFT_PORT:-25565}"
    volumes:
      - mc_data:/data
    deploy:
      resources:
        limits:
          cpus: "${MINECRAFT_CPU:-1.0}"
          memory: "${MINECRAFT_RAM:-1.0G}"
          pids: 100
    healthcheck:
      test: mc-health
      start_period: 1m
      interval: 5s
      retries: 20
    networks:
      - www
```




### MsSQL

To add [MsSQL](https://hub.docker.com/r/microsoft/mssql-server) as a service for user in OpenPanel:

add to .env file:

```bash
# MSSQL
MSSQL_IMAGE="mcr.microsoft.com/mssql/server"
MSSQL_VERSION="latest"
MSSQL_PID="Standard"
MSSQL_PORT="0:1433"
MSSQL_CPU="1.0"
MSSQL_RAM="1.0G"
MSSQL_SA_PASSWORD="rootpassword"
```

add to `docker-compose.yml` file in the **volumes** section:

```bash
  mssql_data:
    driver: local
    labels:
      description: "This volume holds the MSSQL databases."
      purpose: "database"      
```

add to `docker-compose.yml` file in the **services** section:

```bash
  mssql:
    image: ${MSSQL_IMAGE}:${MSSQL_VERSION:-latest}
    container_name: mssql
    restart: unless-stopped
    environment:
      ACCEPT_EULA: "Y"
      MSSQL_SA_PASSWORD: ${MSSQL_SA_PASSWORD:-StrongPassword!}
      MSSQL_PID: ${MSSQL_PID:-Developer}  # Options: Developer, Express, Standard, Enterprise
    ports:
      - "${MSSQL_PORT}"
    volumes:
      - mssql_data:/var/opt/mssql                                      # MSSQL data
      - ./sockets/mssql:/var/opt/mssql/sockets          # MSSQL socket
      - ./mssql.conf:/etc/mssql/mssql.conf:ro           # Custom MSSQL config
    deploy:
      resources:
        limits:
          cpus: "${MSSQL_CPU:-1}"
          memory: "${MSSQL_RAM:-2G}"
          pids: 100
    networks:
      - db
    healthcheck:
      test: ['CMD-SHELL', 'sqlcmd -S localhost -U sa -P "$$MSSQL_SA_PASSWORD" -Q "SELECT 1" || exit 1']
      interval: 10s
      timeout: 5s
      retries: 5

```


### UptimeKuma

To add [UtimeKuma](https://github.com/louislam/uptime-kuma) as a service for user in OpenPanel:

add to .env file:

```bash
# UPTIMEKUMA
UPTIMEKUMA_VERSION="1"
UPTIMEKUMA_CPU="0.5"
UPTIMEKUMA_RAM="0.5G"
```

add to `docker-compose.yml` file in the **services** section:

```bash
  uptimekuma:
    image: louislam/uptime-kuma:${UPTIMEKUMA_VERSION:-1}
    container_name: uptimekuma
    volumes:
      - ./data:/app/data
      - /hostfs/run/user/${USER_ID}/docker.sock:/var/run/docker.sock:ro
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: "${UPTIMEKUMA_CPU:-0.35}"
          memory: "${UPTIMEKUMA_RAM:-0.35G}"   
          pids: 100
    networks:
      - www
```




### BusyBox

This example adds busybox container, its an example on how to add any docker compose service:

add to .env file:

```bash
# BUSYBOX
BUSYBOX_CPU="0.1"
BUSYBOX_RAM="0.1G"
```

add to `docker-compose.yml` file in the **services** section:

```bash
  busybox:
    image: busybox
    container_name: busybox          
    restart: unless-stopped
    working_dir: /var/www/html
    deploy:
      resources:
        limits:
          cpus: "${BUSYBOX_CPU:-0.1}"
          memory: "${BUSYBOX_RAM:-0.1G}"   
          pids: 100
    volumes:
      - html_data:/var/www/html/
```

