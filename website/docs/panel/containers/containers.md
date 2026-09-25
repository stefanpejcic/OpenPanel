---
sidebar_position: 1
---

# Containers

The **Containers** page in OpenPanel allows you to manage Docker services defined via Docker Compose files by your Administrator.

This section provides a clear overview of your containerized services and their resource usage.

:::info 
To access this feature:
- The **Docker** module must be enabled **server-wide** by an Administrator.
- Your account must have the **Docker** feature enabled.
:::

## Overview

At the top of the page, you'll see:

- **Running Containers / Total Containers** – Indicates how many services are currently active.
- **Total CPU** – Number of CPU cores assigned to your hosting plan.
- **Total Memory** – Amount of RAM (in GB) assigned to your hosting plan.

You can allocate portions of these total resources to individual services.

## Container Table

Use the **Show Columns** dropdown above the table to toggle optional columns (Block I/O, Net I/O, PIDs) on or off. Name, CPU Usage, Memory Usage, Status and Actions are shown by default.

![Containers page listing services with their image, CPU and memory usage, PIDs and status](/img/openpanel-screenshots/containers/containers-list.png#gh-light-mode-only)
![Containers page listing services with their image, CPU and memory usage, PIDs and status](/img/openpanel-screenshots/containers/containers-list_dark.png#gh-dark-mode-only)

Each row in the table represents a containerized service and displays:

- **Name** – Name of the Docker service, along with the image used and its tag (version).
- **CPU Usage**  
  - **Graph** – Real-time usage as a percentage of the allocated CPU.  
  - **Usage** – How much CPU the container is using from its allocated amount.  
  - **Allocated** – Number of CPU cores allocated to the service.
- **Memory Usage**  
  - **Graph** – Real-time usage as a percentage of the allocated memory.  
  - **Usage** – RAM used by the service from its allocated amount.  
  - **Allocated** – Memory (in GB) allocated to the service.
- **Status** – A badge showing whether the service is currently **Enabled** (running) or **Disabled** (stopped).
- **Actions**  
  - If the service is **stopped**, a **Start** button is shown to start it (with a **Pull & Start** option to pull the latest image first).  
  - If the service is **running**, **Stop**, **Terminal** and **Logs** buttons are shown. Click **Terminal** to open a web terminal (`podman exec`) for that container, or **Logs** to jump to its log output.  
  - **Edit** and **Delete** links are only available for services you added yourself — core services (webserver, database, mail, etc.) cannot be edited or deleted from this page.

## Editing Resources

To change CPU or Memory limits for a service:

1. Hover over the **Allocated** value in the table.
2. Click the **pencil** icon.
3. Adjust the value in the input field.
4. Click **Save**.

> The change is applied immediately to the running container (no restart required). Setting the value to `0` removes the limit and falls back to the maximum allowed by your hosting plan.

All limits of a service (CPU, memory and PIDs), along with its image, environment variables, volumes and networks, can also be changed from its **Edit** page:

![Edit service form with the image, resource limits, networks, storage and environment variables](/img/openpanel-screenshots/containers/edit-form.png#gh-light-mode-only)
![Edit service form with the image, resource limits, networks, storage and environment variables](/img/openpanel-screenshots/containers/edit-form_dark.png#gh-dark-mode-only)

## Adding New Services

Click **New Service** on the Containers page to run any Docker image next to your websites, for example a Redis cache, a queue worker or your own app. The form is split into numbered steps:

![Add a container form with the image and name, resource presets, networks, storage and environment variables](/img/openpanel-screenshots/containers/new-form.png#gh-light-mode-only)
![Add a container form with the image and name, resource presets, networks, storage and environment variables](/img/openpanel-screenshots/containers/new-form_dark.png#gh-dark-mode-only)

1. **Image and name**
   - **Docker image** – An image from Docker Hub such as `redis:7-alpine`, or a full address like `ghcr.io/owner/app:1.0`. Popular images are suggested as you type. Add a tag to pin the version.
   - **Name** – Filled in from the image, and changed automatically if that name is already taken. It must start with a letter, contain only lowercase letters and digits, and be at least 3 characters long. Other services reach the container by this name, for example `redis:6379`.

2. **Resources** – Pick **Small** (0.25 CPU, 0.25 GB, 100 processes), **Medium** (0.5 CPU, 1 GB, 500 processes, the default) or **Large** (1 CPU, 2 GB, 1000 processes), or type your own values:
   - **CPU cores** – Maximum CPU the container can use, for example `0.5` or `1`.
   - **Memory** – Maximum memory, always in GB, for example `0.5` or `2`.
   - **Max processes** – The most processes the container can run at once.

   These limits apply to this container only, it can't use more than your hosting plan allows.

3. **Networks** – Tick one or more networks from your `docker-compose.yml`. The container can talk to every service on the networks you choose. At least one is required.

4. **Storage (optional)**
   - **Add volume** – Pick one of your Docker volumes, enter the path it's mounted at inside the container, and optionally mark it **Read-only**.
   - **Give access to the Docker socket** – Only for tools that manage containers, like Portainer or Watchtower. The container can then control all your other containers. Unchecked by default.

5. **Environment variables (optional)** – Add a key and value per row, or click **Paste as text** to paste them as `KEY: value` lines, one per line.

6. **Health check (optional)** – A docker compose `healthcheck` block. Click **Insert example** to get one that fits the image, for example:
    ```yaml
    test: ["CMD", "redis-cli", "ping"]
    interval: 30s
    timeout: 5s
    retries: 3
    ```
   The Containers page then shows the container as healthy or unhealthy.

Click **Create container** and the progress of each step is shown on the page:

1. **Validating container details** – The name, image, limits and networks are checked.
2. **Saving container to docker-compose.yml** – The service is added to your `docker-compose.yml` and its limits are saved to `.env`.
3. **Downloading image** – The image is pulled, so starting the container later is quick.
4. **Container created** – You're taken back to the Containers page.

If the details are invalid the step turns red with the reason, and **Back to the form** returns you to the form with everything still filled in. If only the image download fails, the container is already saved, and the error from the registry is shown so you can fix the image from its **Edit** page.

:::info
The new container isn't started automatically. Start it from the Containers page when you're ready.
:::

Each service gets an **uppercase prefix** for its environment variable keys, and its CPU, memory and process limits are stored as variables in `.env`. For example, a service named `nginx` gets:

  ```
  NGINX_CPU
  NGINX_RAM
  NGINX_PIDS
  ```

