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
  - If the service is **running**, **Stop**, **Terminal** and **Logs** buttons are shown. Click **Terminal** to open a web terminal (`docker exec`) for that container, or **Logs** to jump to its log output.  
  - **Edit** and **Delete** links are only available for services you added yourself — core services (webserver, database, mail, etc.) cannot be edited or deleted from this page.

## Editing Resources

To change CPU or Memory limits for a service:

1. Hover over the **Allocated** value in the table.
2. Click the **pencil** icon.
3. Adjust the value in the input field.
4. Click **Save**.

> The change is applied immediately to the running container (no restart required). Setting the value to `0` removes the limit and falls back to the maximum allowed by your hosting plan.

## Adding New Services

To add a new Docker service (container), fill in the **Add Service** form with the required details.  

- **Service Name** – Unique name for the container.  
  - Must start with a letter, contain only lowercase letters and digits, and be at least 3 characters long.  
  - Example: `webapp`, `redis1`  

- **Image** – Docker image to use.  
  - Example: `nginx:latest`, `redis:7.2`  

- **Environment Variables** – Optional. Provide variables in `KEY: value` format, one per line.  
  - Example:  
    ```
    REDIS_PASSWORD: secret
    DEBUG: true
    ```

- **CPU Limit** – Maximum CPU allocation for the container. Must be a positive number.  
  - Example: `0.5`, `1`  

- **RAM Limit** – Maximum memory allocation. Must be a number followed by `M` or `G`.  
  - Example: `512M`, `1.5G`  

- **Volumes (optional)** – Attach storage to the container.  
  - **Mount Docker socket** – Checkbox that mounts the host's Docker/Podman socket read-only into the container.  
  - **Volume rows** – For each volume, select an existing Docker volume, enter the mount path inside the container, and optionally mark it **Read-only**. Click **Add** to add more rows.

- **Network** – Select the Docker network to attach the container to, from the networks already defined in your `docker-compose.yml`.  

- **Healthcheck (optional)** – YAML block defining container health checks.  
  - Example:  
    ```yaml
    test: ["CMD", "curl", "-f", "http://localhost"]
    interval: 30s
    timeout: 10s
    retries: 3
    ```  

**Validation Rules**:

- **Service Name** – Must be unique and follow format rules.  
- **CPU Limit** – Must be a positive number.  
- **RAM Limit** – Must end with `M` or `G`.  
- **Environment Variables** – Must be in `KEY: value` format.  
- **Healthcheck** – Must be valid YAML.  

Each service automatically uses an **uppercase prefix** for environment variable keys.  
CPU and RAM values are also stored as environment variables for each service.  

For example, a service named `nginx` will have the following environment keys:

  ```
  NGINX_CPU
  NGINX_RAM
  ``` 

> Once the form is submitted and validated, the service is added to the Docker Compose configuration and environment variables are automatically updated.

