---
sidebar_position: 2
---

# Dashboard

The dashboard page is the hub of the OpenAdmin interface and provides the overview of the current server performance and usage.

The dashboard page contains 13 widgets:

- **Overview** widget: Displays the total number of nodes (servers), containers, users, domains, websites, hosting packages and email accounts.
- **Resource Usage** widget: Shows the real-time server resource usage for: Load, CPU%, Memory, SWAP, Disk usage and Network I/O.
- **User Activity** widget: Displays real-time combined activity log of all OpenPanel users.
- **Latest News Activity** widget: Displays blog articles from the OpenPanel blog.
- **System Information** widget: Displays Information about your server configuration: Hostname, OS, OpenPanel version, Kernel, CPU type, Uptime, NUmber of Running Processes and available Package Updates.

## SSE Usage

In the top-right corner of every page in **OpenAdmin**, administrators can monitor real-time resource usage, including **Load**, **Memory**, **CPU**, and **Disk**.

Hovering over each metric provides detailed information:

* **Load** – Average system load over 1, 5, and 15 minutes
* **Memory** – Usage of physical memory and SWAP
* **CPU** – Usage per CPU core
* **Disk** – Usage per disk partition

![SSE Widget](https://i.postimg.cc/9Q9DMPH0/openadmin-sse.gif)

## Resource Usage

![openadmin dashboard widget](/img/admin/dashboard/openadmin_dashboard_widget.gif)


The **Resource Usage** widget provides real-time monitoring of server resources, including:

* **Average Load** – Displays the system's current average load.
* **Memory Usage** – Shows current RAM and swap usage, with options to drop cached memory and clear swap space.
* **CPU Usage** – Displays the total CPU usage percentage and usage per individual core.
* **Disk Usage** – Provides details on free, used, and total disk space, device name, filesystem type, and in-depth disk I/O statistics.
* **Network I/O** – Shows real-time input and output data per network interface.

## User Activity

The **User Activity** widget provides a combined list of latest actions from all OpenPanel users (their activity logs).

## Latest News

The **Latest News** widget displays news from [OpenPanel Blog](https://openpanel.com/blog).

## System Information

The **System Informations** widget displays information about your server:

- Hostname
- OS
- OpenPanel version
- Server Time
- Kernel version
- CPU model
- Uptime
- Number of running processes
- Available Package updates

## Try Enterprise

The **Try Enterprise** widget is displayed on Community Editio only and displays features for Enterprise edition and options to upgrade.

## Found a Bug

By default, every page in both the OpenPanel and OpenAdmin UIs includes a **"Found a bug? Let us know"** link at the bottom. This link allows users to report issues directly to our [GitHub Issues](https://github.com/stefanpejcic/OpenPanel/issues) page and includes basic information to help reproduce the problem.

For the OpenPanel UI, administrators can disable this link by going to **OpenAdmin > Settings > OpenPanel** and toggling off the **"Display link to report bugs"** option.

## Dark Mode

To enable Dark Mode, click your username in the bottom-left corner and select the Moon icon. To switch back to Light Mode, click the Sun icon.

![openadmin dark mode](/img/admin/dashboard/openadmin_dark_mode_toggle.gif)

## Menu

OpenAdmin menu lists all available options in the OpenAdmin interface. Simply click on a menu item to open it.

## Search

Search returns:

- OpenPanel users with login link for their OpenPanel
- Website/Domains of users
- Features/pages in the Admin interface



## Logout

To log out of the OpenAdmin account, click your username in the bottom-left corner and select 'Sign out' option.
