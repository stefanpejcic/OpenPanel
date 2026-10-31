---
sidebar_position: 2
---

# Dashboard

The dashboard page is the hub of the OpenAdmin interface and provides the overview of the current server performance and usage.

For administrator accounts, the dashboard also displays a quick summary bar with counts for **Servers**, **Containers**, **Users**, **Domains**, **Websites**, **Packages**, and **Emails** (each showing current usage against the license's limit, if any). Reseller accounts instead see their own **OpenPanel Accounts** and **Disk Space Used** usage against the limits set by the administrator.

![Summary bar with the number of nodes, containers, users, domains, websites, packages and emails](/img/openadmin-screenshots/001_dashboard-summary.png#gh-light-mode-only)
![Summary bar with the number of nodes, containers, users, domains, websites, packages and emails](/img/openadmin-screenshots/001_dashboard-summary_dark.png#gh-dark-mode-only)

The dashboard page contains widgets:

- **Tasks** widget: Lists background tasks that are running now.
- **Cluster** widget: On Enterprise, shows every server of a [multi-server](/docs/admin/system/multi-server/) cluster with its users, domains, version and unread notifications.
- **Daily Summary** tab: Displays yesterday's traffic and issues for all accounts combined, see [Daily Summary](/docs/admin/daily-summary). Shown only when the [Daily Summary module](/docs/admin/settings/modules/#daily-summary) is enabled.
- **User Activity** widget: Displays real-time combined activity log of all OpenPanel users.
- **Latest News** widget: Displays blog articles from the OpenPanel blog.
- **System Information** widget: Displays Information about your server configuration: Hostname, IPv4 address, OS, OpenPanel version, Server Time, Kernel, CPU type, Uptime, Number of Running Processes and available Package Updates.
- **Resource Usage** widget: Displays a CPU % and RAM % chart for the last hour, with a link to view the full resource usage history.

![OpenAdmin dashboard with the summary bar and the User Activity, Latest News, System Information and Resource Usage widgets](/img/openadmin-screenshots/001_dashboard-window.png#gh-light-mode-only)
![OpenAdmin dashboard with the summary bar and the User Activity, Latest News, System Information and Resource Usage widgets](/img/openadmin-screenshots/001_dashboard-window_dark.png#gh-dark-mode-only)

## Onboarding

After logging in, OpenAdmin opens an onboarding guide (`/onboarding`) until it's dismissed. It walks through getting a new server ready in three steps:

1. **Enable modules & services** – choose which services run on the server. Services already running are checked, and newly checked ones start in the background when you continue. The Community edition offers OpenPanel and the CSF firewall; Enterprise also offers Email + Webmail, FTP, DNS, phpMyAdmin and ImunifyAV.
2. **Server configuration** – recommended settings to review before going live: modules, a custom domain for panel access, SMTP and notifications, and update preferences. Each card opens its settings page.
3. **Users and plans** – the default PHP version and web server, hosting plans and their features, and the first user account (or importing accounts from a backup on Enterprise).

**Skip for now** opens the dashboard and shows the guide again at the next login. **Dismiss completely** hides it for good, and **Skip to dashboard** at the top leaves the guide right away.

## Customize the dashboard

Administrators can arrange the widgets, and the layout is saved on the admin account, so it's the same in every browser and, on a multi-server master, on the dashboard of every server:

- **Move a widget**: hover over it and drag it by the handle in its top right corner.
- **Change the width**: hover over a widget and pick **1/3**, **2/3** or **3/3** of the page. By default Resource Usage takes 2/3 of the page and the other widgets 1/3. Two 1/3 widgets that follow a 2/3 one stack in the column next to it, like Tasks and Cluster next to Resource Usage.
- **Collapse a widget**: click the arrow in its top right corner to show only its title. Click it again to expand it.

To go back to the default layout, click **Reset dashboard layout** under the widgets.

## Usage

In the top-right corner of every page in **OpenAdmin**, administrators can monitor real-time resource usage, including **Load**, **Memory**, **CPU**, and **Disk**.

Hovering over each metric provides detailed information:

* **Load** – Average system load over 1, 5, and 15 minutes
* **Memory** – Usage of physical memory and SWAP
* **CPU** – Usage per CPU core
* **Disk** – Usage per disk partition

![Resource usage bar in the top-right corner of OpenAdmin with the load tooltip open](/img/openadmin-screenshots/001_dashboard-sse.png#gh-light-mode-only)
![Resource usage bar in the top-right corner of OpenAdmin with the load tooltip open](/img/openadmin-screenshots/001_dashboard-sse_dark.png#gh-dark-mode-only)

## Daily Summary

The daily summary of all accounts is on its own tab, see [Daily Summary](/docs/admin/daily-summary).

## User Activity

The **User Activity** widget provides a combined list of latest actions from all OpenPanel users (their activity logs).

![User Activity widget with the latest actions of all OpenPanel users](/img/openadmin-screenshots/001_dashboard-activity.png#gh-light-mode-only)
![User Activity widget with the latest actions of all OpenPanel users](/img/openadmin-screenshots/001_dashboard-activity_dark.png#gh-dark-mode-only)

## Latest News

The **Latest News** widget displays news from [OpenPanel Blog](https://openpanel.com/blog).

![Latest News widget with articles from the OpenPanel blog](/img/openadmin-screenshots/001_dashboard-news.png#gh-light-mode-only)
![Latest News widget with articles from the OpenPanel blog](/img/openadmin-screenshots/001_dashboard-news_dark.png#gh-dark-mode-only)

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

![System Information widget with hostname, IPv4 address, OS, OpenPanel version, server time, kernel, CPU, uptime, processes and package updates](/img/openadmin-screenshots/001_dashboard-sysinfo.png#gh-light-mode-only)
![System Information widget with hostname, IPv4 address, OS, OpenPanel version, server time, kernel, CPU, uptime, processes and package updates](/img/openadmin-screenshots/001_dashboard-sysinfo_dark.png#gh-dark-mode-only)

## Resource Usage

The **Resource Usage** widget displays a chart of CPU % and RAM % usage over the last hour. On a multi-server master it draws a line for every server, click a server's name above the chart to show only that one. Click **View history** to open the full Resource Usage history page (**Server > Resource Usage**).

![Resource Usage widget with a chart of CPU and RAM usage over the last hour and the View history link](/img/openadmin-screenshots/001_dashboard-usage.png#gh-light-mode-only)
![Resource Usage widget with a chart of CPU and RAM usage over the last hour and the View history link](/img/openadmin-screenshots/001_dashboard-usage_dark.png#gh-dark-mode-only)

## Cluster

On a [multi-server](/docs/admin/system/multi-server/) master, the **Cluster** widget lists every server with three bars for its health: green when it's healthy, orange when its disk or RAM is almost full, red when it's offline or a service is down, and grey when it's paused. Each value opens the page to act on it: users and domains of that server, the updates page, its notifications and the sync settings.

On a single server it shows that the server is standalone, and on a slave which master manages it.

![Cluster widget listing the master and two servers with their health bars, users, domains, version and last sync](/img/openadmin-screenshots/001_dashboard-cluster.png#gh-light-mode-only)
![Cluster widget listing the master and two servers with their health bars, users, domains, version and last sync](/img/openadmin-screenshots/001_dashboard-cluster_dark.png#gh-dark-mode-only)

## Tasks

The **Tasks** widget lists background tasks that are currently running on the server. Click **View all** to open the full list of tasks, where you can tick running tasks and **Kill** them all at once from the bulk actions bar.

![Tasks widget with the currently running background tasks and the View all link](/img/openadmin-screenshots/001_dashboard-tasks.png#gh-light-mode-only)
![Tasks widget with the currently running background tasks and the View all link](/img/openadmin-screenshots/001_dashboard-tasks_dark.png#gh-dark-mode-only)

## Found a Bug

By default, every page in both the OpenPanel and OpenAdmin UIs includes a **"Found a bug? Let us know"** link at the bottom. This link allows users to report issues directly to our [GitHub Issues](https://github.com/stefanpejcic/OpenPanel/issues) page and includes basic information to help reproduce the problem.

For the OpenPanel UI, administrators can disable this link by going to **Settings > OpenPanel** and toggling off the **"Display link to report bugs"** option.

## Dark Mode

To enable Dark Mode, click your username in the bottom-left corner and select the Moon icon. To switch back to Light Mode, click the Sun icon.

![OpenAdmin dashboard in dark mode](/img/openadmin-screenshots/001_dashboard_dark-window.png#screenshot)

## Menu

OpenAdmin menu lists all available options in the OpenAdmin interface. Simply click on a menu item to open it.

The menu comes in two styles, set by the **Menu Style** option on [**Settings > OpenPanel**](/docs/admin/settings/openpanel/#display) (the same option sets the default for OpenPanel users):

- **Classic** (default): items are grouped in expandable sections, like the screenshot below.
- **Modern**: the sidebar has one link per area - Accounts, Hosting Plans, Domains, Emails, Backups, then **Server** (Services, Security, Server, System) and **Settings** (Settings, License & Support) - and the pages of that area are shown as tabs at the top of each page.

These docs use the Modern names, so a path like **System > Server Time** means: click **System** in the sidebar, then the **Server Time** tab. In the Classic menu the same pages are in the expandable groups, for example **Server > Server Time**.

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

## Keyboard Shortcuts

OpenAdmin UI can be navigated using keyboard shortcuts: [view documentation](/docs/articles/dev-experience/openadmin-keyboard-shortcuts/).

Press **Ctrl + K** (or **Cmd + K** on macOS) on any page to show the list of available shortcuts:

![Keyboard shortcuts dialog opened with Ctrl + K, listing the key combinations for OpenAdmin pages](/img/openadmin-screenshots/001_dashboard-shortcuts.png#gh-light-mode-only)
![Keyboard shortcuts dialog opened with Ctrl + K, listing the key combinations for OpenAdmin pages](/img/openadmin-screenshots/001_dashboard-shortcuts_dark.png#gh-dark-mode-only)

## Logout

To log out of the OpenAdmin account, click your username in the bottom-left corner and select **Sign out**.

![Account menu opened from the username in the bottom-left corner of the sidebar, with the Sign out option](/img/openadmin-screenshots/001_dashboard-logout.png#gh-light-mode-only)
![Account menu opened from the username in the bottom-left corner of the sidebar, with the Sign out option](/img/openadmin-screenshots/001_dashboard-logout_dark.png#gh-dark-mode-only)
