---
sidebar_position: 2
---

# Terminal

The **Terminal** page provides a web-based terminal (`docker exec`) for interacting with your running containers directly through the OpenPanel interface.

## Requirements

To access the Terminal:

- The **Docker** module must be enabled **server-wide** by an Administrator.
- Your account must have the **Docker** feature enabled.
- [`/etc/openpanel/disable_openpanel_terminal_ui` file should not exist](/docs/articles/dev-experience/disable-openpanel-web-terminal/).


## Accessing the Terminal

1. In the OpenPanel menu, go to **Containers > Terminal**.
2. Click on **Select Service** to display a list of currently running services.
3. Click on the service you want to access.
4. The terminal window will open, allowing you to run commands inside the container.

:::tip
Pages that manage a single service, like **MySQL**, **Redis** or **Web Server Config**, also have **Terminal** and **Logs** tabs that open that service's terminal directly, without leaving the page's section.
:::

![Terminal page with the service dropdown opened, listing the running containers to connect to](/img/openpanel-screenshots/containers/terminal_server-select.png#gh-light-mode-only)
![Terminal page with the service dropdown opened, listing the running containers to connect to](/img/openpanel-screenshots/containers/terminal_server-select_dark.png#gh-dark-mode-only)

You can switch the shell type between `sh` and `bash` using the selector in the top-right corner of the terminal.

![Terminal connected to the apache container with a shell prompt and the shell dropdown](/img/openpanel-screenshots/containers/terminal_server-shell.png#gh-light-mode-only)
![Terminal connected to the apache container with a shell prompt and the shell dropdown](/img/openpanel-screenshots/containers/terminal_server-shell_dark.png#gh-dark-mode-only)

---

> 💡 This feature uses `docker exec` under the hood, giving you direct access to the container's shell environment in real-time.
