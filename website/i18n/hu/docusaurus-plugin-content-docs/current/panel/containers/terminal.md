---
sidebar_position: 2
---

# Terminal

The **Terminal** page provides a web-based terminal (`docker exec`) for interacting with your running containers directly through the OpenPanel interface.

## Requirements

To access the Terminal:

- The **Docker** module must be enabled **server-wide** by an Administrator.
- Your account must have the **Docker** feature enabled.

## Accessing the Terminal

1. In the OpenPanel menu, go to **Docker > Terminal**.
2. Click on **Select Service** to display a list of currently running services.
3. Click on the service you want to access.
4. The terminal window will open, allowing you to run commands inside the container.

You can switch the shell type between `sh` and `bash` using the selector in the top-right corner of the terminal.

---

> 💡 This feature uses `docker exec` under the hood, giving you direct access to the container's shell environment in real-time.
