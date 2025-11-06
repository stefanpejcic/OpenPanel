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
- **Actions**  
  - Shows whether the service is **Enabled** or **Disabled**.  
  - Clicking the status toggles it.  
  - If the service is running, a **Terminal** link appears — click it to open a web terminal (`docker exec`) for that container.

## Editing Resources

To change CPU or Memory limits for a service:

1. Hover over the **Allocated** value in the table.
2. Click the **pencil** icon.
3. Adjust the value in the input field.
4. Click **Save**.

> OpenPanel will check whether you have available resources to make this change. If valid, the container will restart automatically with the new limits.

## Adding New Services

To add a new Docker service (container), fill in the **Add Service** form with the required details.  

- **Service Name** – Unique name for the container.  
  - Must start with a letter, contain only lowercase letters and digits, and be at least 3 characters long.  
  - Example: `webapp`, `redis1`  

- **Image** – Docker image to use.  
  - Example: `nginx:latest`, `redis:7.2`  

- **Environment Variables** – Optional. Provide variables in `KEY=value` format, one per line.  
  - Example:  
    ```
    REDIS_PASSWORD=secret
    DEBUG=true
    ```

- **CPU Limit** – Maximum CPU allocation for the container. Must be a positive number.  
  - Example: `0.5`, `1`  

- **RAM Limit** – Maximum memory allocation. Must be a number followed by `M` or `G`.  
  - Example: `512M`, `1.5G`  

- **Network** – Docker network to attach the container to.  
  - Example: `backend_network`  

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
- **Environment Variables** – Must be in `KEY=value` format.  
- **Healthcheck** – Must be valid YAML.  

Each service automatically uses an **uppercase prefix** for environment variable keys.  
CPU and RAM values are also stored as environment variables for each service.  

For example, a service named `nginx` will have the following environment keys:

  ```
  NGINX_CPU
  NGINX_RAM
  ``` 

> Once the form is submitted and validated, the service is added to the Docker Compose configuration and environment variables are automatically updated.

