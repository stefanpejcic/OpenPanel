---
sidebar_position: 2
---

# Dashboard

The dashboard page is the hub of the OpenAdmin interface and provides the overview of the current server performance and usage.

For administrator accounts, the dashboard also displays a quick summary bar with counts for **Nodes**, **Containers**, **Users**, **Domains**, **Websites**, **Packages**, and **Emails** (each showing current usage against the license's limit, if any). Reseller accounts instead see their own **OpenPanel Accounts** and **Disk Space Used** usage against the limits set by the administrator.

![Summary bar with the number of nodes, containers, users, domains, websites, packages and emails](/img/openadmin-screenshots/001_dashboard-summary.png)

The dashboard page contains widgets:

- **User Activity** widget: Displays real-time combined activity log of all OpenPanel users.
- **Latest News** widget: Displays blog articles from the OpenPanel blog.
- **System Information** widget: Displays Information about your server configuration: Hostname, IPv4 address, OS, OpenPanel version, Server Time, Kernel, CPU type, Uptime, Number of Running Processes and available Package Updates.
- **Resource Usage** widget: Displays a CPU % and RAM % chart for the last hour, with a link to view the full resource usage history.

![OpenAdmin dashboard with the summary bar and the User Activity, Latest News, System Information and Resource Usage widgets](/img/openadmin-screenshots/001_dashboard-window.png)

## Usage

In the top-right corner of every page in **OpenAdmin**, administrators can monitor real-time resource usage, including **Load**, **Memory**, **CPU**, and **Disk**.

Hovering over each metric provides detailed information:

* **Load** – Average system load over 1, 5, and 15 minutes
* **Memory** – Usage of physical memory and SWAP
* **CPU** – Usage per CPU core
* **Disk** – Usage per disk partition

![Resource usage bar in the top-right corner of OpenAdmin with the load tooltip open](/img/openadmin-screenshots/001_dashboard-sse.png)

## User Activity

The **User Activity** widget provides a combined list of latest actions from all OpenPanel users (their activity logs).

![User Activity widget with the latest actions of all OpenPanel users](/img/openadmin-screenshots/001_dashboard-activity.png)

## Latest News

The **Latest News** widget displays news from [OpenPanel Blog](https://openpanel.com/blog).

![Latest News widget with articles from the OpenPanel blog](/img/openadmin-screenshots/001_dashboard-news.png)

## System Information

The **System Information** widget displays information about your server:

- Hostname
- IPv4 address
- OS
- OpenPanel version
- Server Time
- Kernel version
- CPU model
- Uptime
- Number of running processes
- Available Package updates

![System Information widget with hostname, IPv4 address, OS, OpenPanel version, server time, kernel, CPU, uptime, processes and package updates](/img/openadmin-screenshots/001_dashboard-sysinfo.png)

## Resource Usage

The **Resource Usage** widget displays a chart of CPU % and RAM % usage over the last hour. Click **View history** to open the full Resource Usage history page (**Server > Resource Usage**).

![Resource Usage widget with a chart of CPU and RAM usage over the last hour and the View history link](/img/openadmin-screenshots/001_dashboard-usage.png)

## Found a Bug

By default, every page in both the OpenPanel and OpenAdmin UIs includes a **"Found a bug? Let us know"** link at the bottom. This link allows users to report issues directly to our [GitHub Issues](https://github.com/stefanpejcic/OpenPanel/issues) page and includes basic information to help reproduce the problem.

For the OpenPanel UI, administrators can disable this link by going to **Settings > OpenPanel Settings** and toggling off the **"Display link to report bugs"** option.

## Dark Mode

To enable Dark Mode, click your username in the bottom-left corner and select the Moon icon. To switch back to Light Mode, click the Sun icon.

![OpenAdmin dashboard in dark mode](/img/openadmin-screenshots/001_dashboard_dark-window.png)

## Menu

OpenAdmin menu lists all available options in the OpenAdmin interface. Simply click on a menu item to open it.

![OpenAdmin sidebar menu with the Accounts, Hosting Plans, Domains, Emails, Services, Security, Settings and Advanced sections](/img/openadmin-screenshots/001_dashboard-menu.png)

## Search

Search returns:

- OpenPanel users with login link for their OpenPanel
- Website/Domains of users
- Features/pages in the Admin interface

![Search box in the OpenAdmin sidebar with matching pages, users and websites](/img/openadmin-screenshots/001_dashboard-search.png)

## Keyboard Shortcuts

OpenAdmin UI can be navigated using keyboard shortcuts: [view documentation](/docs/articles/dev-experience/openadmin-keyboard-shortcuts/).

## Logout

To log out of the OpenAdmin account, click your username in the bottom-left corner and select 'Sign out' option.
