---
sidebar_position: 1
---

# Containers

The **Containers** page in OpenPanel allows you to manage Docker services defined via Docker Compose files by your Administrator.

This section provides a clear overview of your containerized services and their resource usage.

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

At the moment, new services can only be added by the Administrator.
