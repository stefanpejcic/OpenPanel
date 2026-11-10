---
sidebar_position: 2
---

# Dashboard

The dashboard page is the hub of the OpenAdmin interface and provides the overview of the current server performance and usage.

For administrator accounts, the dashboard also displays a quick summary bar with counts for **Servers**, **Containers**, **Users**, **Domains**, **Websites**, **Packages**, and **Emails** (each showing current usage against the license's limit, if any). Reseller accounts instead see their own **OpenPanel accounts** and **Disk space used** usage against the limits set by the administrator.

![Summary bar with the number of nodes, containers, users, domains, websites, packages and emails](/img/openadmin-screenshots/001_dashboard-summary.png#gh-light-mode-only)
![Summary bar with the number of nodes, containers, users, domains, websites, packages and emails](/img/openadmin-screenshots/001_dashboard-summary_dark.png#gh-dark-mode-only)

The dashboard page contains widgets:

- **Tasks** widget: Lists background tasks that are running now.
- **Cluster** widget: On Enterprise, shows every server of a [multi-server](/docs/admin/system/multi-server/) cluster with its users, domains, version and unread notifications.
- **Daily summary** tab: Displays yesterday's traffic and issues for all accounts combined, see [Daily Summary](/docs/admin/daily-summary). Shown only when the [Daily Summary module](/docs/admin/settings/modules/#daily-summary) is enabled.
- **User activity** widget: Displays real-time combined activity log of all OpenPanel users.
- **Latest news** widget: Displays blog articles from the OpenPanel blog.
- **System information** widget: Displays information about your server configuration: hostname, IPv4 address, OS, OpenPanel version, server time, kernel, CPU type, uptime, number of running processes and available package updates.
- **Resource usage** widget: Displays a CPU % and RAM % chart for the last hour, with a link to view the full resource usage history.

![OpenAdmin dashboard with the summary bar and the User Activity, Latest News, System Information and Resource Usage widgets](/img/openadmin-screenshots/001_dashboard-window.png#gh-light-mode-only)
![OpenAdmin dashboard with the summary bar and the User Activity, Latest News, System Information and Resource Usage widgets](/img/openadmin-screenshots/001_dashboard-window_dark.png#gh-dark-mode-only)

### Notes on the dashboard

The dashboard flags problems it finds above or inside the related widget:

- **Usage stats are not being collected**: the resource usage chart has no data from the last few minutes, so what it shows is old. Check that the cron service is running and the `opencli sentinel` schedule on [Scheduled actions](/docs/admin/advanced/crons/).
- **Long running tasks**: a task that has been running for over 20 hours is most likely stuck. Check it on the [Tasks](/docs/admin/advanced/tasks/) page, a hung task can block the next run of its cron.
- **Servers with problems**: on a multi-server master, servers that are offline, have a service down or have disk or RAM over 90%. Paused servers are left out.
- **OS support ending**: the System information widget warns when the server's OS stops getting security updates soon, or already has. On a standalone server with few accounts it also points to the [migration guide](/docs/admin/system/migrate/), since everything can be moved to a server with a current OS.
- **OpenPanel update**: versions older than 2.0.22 are asked to update, since 2.0.22 brings a lot of security fixes.
- **Many package updates**: a large number of available package updates means the server hasn't been upgraded in a long time. Back up the server first, then run the upgrade from the [Terminal](/docs/admin/advanced/terminal/).

## Onboarding

After logging in, OpenAdmin opens an onboarding guide (`/onboarding`) until it's dismissed. It walks through getting a new server ready in three steps:

1. **Enable modules & services** – choose which services run on the server. Services already running are checked, and newly checked ones start in the background when you continue. The Community edition offers OpenPanel and the CSF firewall; Enterprise also offers Email + Webmail, FTP, DNS, phpMyAdmin and ImunifyAV.
2. **Server configuration** – recommended settings to review before going live: modules, a custom domain for panel access, SMTP and notifications, and update preferences. Each card opens its settings page.
3. **Users and plans** – the default PHP version and web server, hosting plans and their features, and the first user account (or importing accounts from a backup on Enterprise).

**Skip for now** opens the dashboard and shows the guide again at the next login. **Dismiss completely** hides it for good, and **Skip to dashboard** at the top leaves the guide right away.

## Customize the dashboard

Administrators can arrange the widgets, and the layout is saved on the admin account, so it's the same in every browser and, on a multi-server master, on the dashboard of every server:

- **Move a widget**: drag it by the handle in its top right corner, or focus the handle and use the arrow keys. While you drag, the free slot is shown as a dashed drop zone.
- **Change the width**: hover over the handle and pick **1/3**, **1/2**, **2/3** or **Full** width. By default Resource Usage takes 2/3 of the page and the other widgets 1/3.
- **Change the height**: hover over the handle and pick **1 block high** or **2 blocks high**. A 2 block high widget makes room for two 1 block widgets stacked next to it, like Tasks and Cluster next to Resource Usage. Other widgets move up to fill gaps in the grid.
- **Collapse a widget**: click the arrow in its top right corner to show only its title. Click it again to expand it.

To go back to the default order, widths and heights, click **Reset layout** under the widgets.

## Usage

In the top-right corner of every page in **OpenAdmin**, administrators can monitor real-time resource usage, including **Load**, **Memory**, **CPU**, and **Disk**.

Hovering over each metric provides detailed information:

* **Load** – Average system load over 1, 5, and 15 minutes
* **Memory** – Usage of physical memory and SWAP
* **CPU** – Usage per CPU core
* **Disk** – Usage per disk partition

![Resource usage bar in the top-right corner of OpenAdmin with the load tooltip open](/img/openadmin-screenshots/001_dashboard-sse.png#gh-light-mode-only)
![Resource usage bar in the top-right corner of OpenAdmin with the load tooltip open](/img/openadmin-screenshots/001_dashboard-sse_dark.png#gh-dark-mode-only)

## Daily summary

The daily summary of all accounts is on its own tab, see [Daily Summary](/docs/admin/daily-summary).

## User activity

The **User activity** widget provides a combined list of latest actions from all OpenPanel users (their activity logs).

![User Activity widget with the latest actions of all OpenPanel users](/img/openadmin-screenshots/001_dashboard-activity.png#gh-light-mode-only)
![User Activity widget with the latest actions of all OpenPanel users](/img/openadmin-screenshots/001_dashboard-activity_dark.png#gh-dark-mode-only)

## Latest news

The **Latest news** widget displays news from [OpenPanel Blog](https://openpanel.com/blog).

![Latest News widget with articles from the OpenPanel blog](/img/openadmin-screenshots/001_dashboard-news.png#gh-light-mode-only)
![Latest News widget with articles from the OpenPanel blog](/img/openadmin-screenshots/001_dashboard-news_dark.png#gh-dark-mode-only)

## System information

The **System information** widget displays information about your server:

- Hostname
- IPv4 address
- OS
- OpenPanel version
- Server time
- Kernel version
- CPU model
- Uptime
- Number of running processes
- Available package updates

![System Information widget with hostname, IPv4 address, OS, OpenPanel version, server time, kernel, CPU, uptime, processes and package updates](/img/openadmin-screenshots/001_dashboard-sysinfo.png#gh-light-mode-only)
![System Information widget with hostname, IPv4 address, OS, OpenPanel version, server time, kernel, CPU, uptime, processes and package updates](/img/openadmin-screenshots/001_dashboard-sysinfo_dark.png#gh-dark-mode-only)

## Resource usage

The **Resource usage** widget displays a chart of CPU % and RAM % usage over the last hour. On a multi-server master it draws a line for every server, click a server's name above the chart to show only that one. Click **View history** to open the full Resource Usage history page (**Server > Resource usage**).

![Resource Usage widget with a chart of CPU and RAM usage over the last hour and the View history link](/img/openadmin-screenshots/001_dashboard-usage.png#gh-light-mode-only)
![Resource Usage widget with a chart of CPU and RAM usage over the last hour and the View history link](/img/openadmin-screenshots/001_dashboard-usage_dark.png#gh-dark-mode-only)

## Cluster

On a [multi-server](/docs/admin/system/multi-server/) master, the **Cluster** widget lists every server with three bars for its health: green when it's healthy, orange when its disk or RAM is almost full, red when it's offline or a service is down, and grey when it's paused. Each value opens the page to act on it: users and domains of that server, the updates page, its notifications and the sync settings.

On a single server it shows that the server is standalone, and on a slave which master manages it.

![Cluster widget with a problem notice, the master with its users, domains, version and notifications, a paused server and an offline one](/img/openadmin-screenshots/001_dashboard-cluster.png#gh-light-mode-only)
![Cluster widget with a problem notice, the master with its users, domains, version and notifications, a paused server and an offline one](/img/openadmin-screenshots/001_dashboard-cluster_dark.png#gh-dark-mode-only)

## Tasks

The **Tasks** widget lists background tasks that are currently running on the server. Click **View all** to open the full list of tasks, where you can tick running tasks and **Kill** them all at once from the bulk actions bar.

![Tasks widget with the currently running background tasks and the View all link](/img/openadmin-screenshots/001_dashboard-tasks.png#gh-light-mode-only)
![Tasks widget with the currently running background tasks and the View all link](/img/openadmin-screenshots/001_dashboard-tasks_dark.png#gh-dark-mode-only)

## Found a bug

By default, every page in both the OpenPanel and OpenAdmin UIs includes a **"Found a bug? Let us know"** link at the bottom. This link allows users to report issues directly to our [GitHub Issues](https://github.com/stefanpejcic/OpenPanel/issues) page and includes basic information to help reproduce the problem.

For the OpenPanel UI, administrators can disable this link by going to **Settings > OpenPanel** and toggling off the **"Display link to report bugs"** option.

## Dark mode

To enable dark mode, click your username in the bottom-left corner and turn on the **Dark mode** switch. Turn it off to go back to light mode. Until you pick one, OpenAdmin follows your browser's light or dark setting.

![OpenAdmin dashboard in dark mode](/img/openadmin-screenshots/001_dashboard_dark-window.png#screenshot)

## Menu

OpenAdmin menu lists all available options in the OpenAdmin interface. Simply click on a menu item to open it.

The menu comes in two styles, set by the **Menu style** option on [**Settings > OpenPanel**](/docs/admin/settings/openpanel/#display) (the same option sets the default for OpenPanel users):

- **Classic** (default): items are grouped in expandable sections, like the screenshot below.
- **Modern**: the sidebar has one link per area - Accounts, Hosting Plans, Domains, Emails, Backups, then **Server** (Services, Security, Server, System) and **Settings** (Settings, License & Support) - and the pages of that area are shown as tabs at the top of each page.

These docs use the Modern names, so a path like **System > Server time** means: click **System** in the sidebar, then the **Server time** tab. In the Classic menu the same pages are in the expandable groups, for example **Server > Server time**.

In the Modern menu, a colored dot next to a sidebar item flags something that needs attention - hover it to see what:

- **Services** (red): a monitored service is down, according to an unread notification.
- **Settings** (orange): changes are waiting for an OpenPanel or OpenAdmin restart.

Settings pages (General, OpenPanel, User Defaults, PHP, Notifications, Updates, Custom Code and Service Limits) ask for confirmation before you leave with changes that weren't saved.

![OpenAdmin sidebar menu with the Accounts, Hosting Plans, Domains, Emails, Services, Security, Settings and Advanced sections](/img/openadmin-screenshots/001_dashboard-menu.png#gh-light-mode-only)
![OpenAdmin sidebar menu with the Accounts, Hosting Plans, Domains, Emails, Services, Security, Settings and Advanced sections](/img/openadmin-screenshots/001_dashboard-menu_dark.png#gh-dark-mode-only)

## Search

Search returns:

- OpenPanel users with login link for their OpenPanel
- Website/Domains of users
- Features/pages in the Admin interface, shown as **Area › Page** - pages can be found by their current name, their previous name (for example **General Settings** or **CorazaWAF**) or the menu area they are in

![Search box in the OpenAdmin sidebar with matching pages, users and websites](/img/openadmin-screenshots/001_dashboard-search.png#gh-light-mode-only)
![Search box in the OpenAdmin sidebar with matching pages, users and websites](/img/openadmin-screenshots/001_dashboard-search_dark.png#gh-dark-mode-only)

## Keyboard shortcuts

OpenAdmin UI can be navigated using keyboard shortcuts: [view documentation](/docs/articles/dev-experience/openadmin-keyboard-shortcuts/).

Press **Ctrl + K** (or **Cmd + K** on macOS) on any page to show the list of available shortcuts:

![Keyboard shortcuts dialog opened with Ctrl + K, listing the key combinations for OpenAdmin pages](/img/openadmin-screenshots/001_dashboard-shortcuts.png#gh-light-mode-only)
![Keyboard shortcuts dialog opened with Ctrl + K, listing the key combinations for OpenAdmin pages](/img/openadmin-screenshots/001_dashboard-shortcuts_dark.png#gh-dark-mode-only)

## Logout

To log out of the OpenAdmin account, click your username in the bottom-left corner and select **Sign out**.

![Account menu opened from the username in the bottom-left corner of the sidebar, with the Sign out option](/img/openadmin-screenshots/001_dashboard-logout.png#gh-light-mode-only)
![Account menu opened from the username in the bottom-left corner of the sidebar, with the Sign out option](/img/openadmin-screenshots/001_dashboard-logout_dark.png#gh-dark-mode-only)
