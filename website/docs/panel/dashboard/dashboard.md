---
sidebar_position: 1
---

# Dashboard

The OpenPanel interface is the central hub of your user account, providing access to all available tools, usage information, and notifications.

The OpenPanel dashboard interface is divided in two sections:

- Features: listing all available options
- Widgets: 2FA, Information, Usage and How-to guides

![OpenPanel dashboard with the sidebar, feature shortcuts grouped by section, and the 2FA, Information and Usage widgets](/img/openpanel-screenshots/dashboard/dashboard-window.png#gh-light-mode-only)
![OpenPanel dashboard with the sidebar, feature shortcuts grouped by section, and the 2FA, Information and Usage widgets](/img/openpanel-screenshots/dashboard/dashboard-window_dark.png#gh-dark-mode-only)

## Two-Factor Authentication

This widget shows whether [Two-Factor Authentication](/docs/panel/account/2fa/) is enabled for your account. If it isn't, click **Click to Enable** to set it up.

![Two-Factor Authentication widget showing 2FA as disabled with the Click to Enable button](/img/openpanel-screenshots/dashboard/dashboard-twofa.png#gh-light-mode-only)
![Two-Factor Authentication widget showing 2FA as disabled with the Click to Enable button](/img/openpanel-screenshots/dashboard/dashboard-twofa_dark.png#gh-dark-mode-only)

## Information

This section displays general server information:

- Username
- Plan name
- Shared or Dedicated IP address
- Last login IP address
- Nameservers (if configured)

![Information widget with the username, plan, IP address and last login IP address](/img/openpanel-screenshots/dashboard/dashboard-information.png#gh-light-mode-only)
![Information widget with the username, plan, IP address and last login IP address](/img/openpanel-screenshots/dashboard/dashboard-information_dark.png#gh-dark-mode-only)

## Usage

This widget displays your current usage against your plan's limits, for each item enabled on your plan:

- Websites
- Domains
- Databases (MySQL and/or PostgreSQL)
- Email accounts
- FTP accounts
- Storage
- Inodes
- CPU Usage
- Memory Usage

![Usage widget with bars for websites, domains, databases, email and FTP accounts, storage, inodes, CPU and memory](/img/openpanel-screenshots/dashboard/dashboard-usage.png#gh-light-mode-only)
![Usage widget with bars for websites, domains, databases, email and FTP accounts, storage, inodes, CPU and memory](/img/openpanel-screenshots/dashboard/dashboard-usage_dark.png#gh-dark-mode-only)

CPU and Memory usage are fetched live from the server. For a more detailed, dedicated view of CPU and RAM usage (and historical charts), see the **Resource Usage** and **Resource Usage History** pages under the Advanced menu, if enabled by your Administrator.

## How-to Guides

If enabled by the Administrator, the How-to guides section will display links to Knowledge Base articles configured by your hosting provider, or to the official OpenPanel documentation by default.

![General How-to widget with links to knowledge base articles](/img/openpanel-screenshots/dashboard/dashboard-howto.png#gh-light-mode-only)
![General How-to widget with links to knowledge base articles](/img/openpanel-screenshots/dashboard/dashboard-howto_dark.png#gh-dark-mode-only)

## Favorites

When [Favorites feature](/docs/admin/settings/modules/#favorites) is enabled by Administrator, users can bookmark up to 10 pages from the OpenPanel interface. These pages will appear in the sidebar menu. To add a page to favorites, simply left-click the star icon in the top-right corner of any page. To remove a page, right-click the same icon.

![Star icon in the top-right corner of the page header, used to add the current page to favorites](/img/openpanel-screenshots/dashboard/dashboard-favorites.png#gh-light-mode-only)
![Star icon in the top-right corner of the page header, used to add the current page to favorites](/img/openpanel-screenshots/dashboard/dashboard-favorites_dark.png#gh-dark-mode-only)

Saved pages are listed under **Favorites** at the top of the sidebar:

![Favorites list at the top of the sidebar with a saved page](/img/openpanel-screenshots/dashboard/dashboard-favorites-sidebar.png#gh-light-mode-only)
![Favorites list at the top of the sidebar with a saved page](/img/openpanel-screenshots/dashboard/dashboard-favorites-sidebar_dark.png#gh-dark-mode-only)

## Services

A service badge with the service name and its status is shown in the top-right corner of the header, but only on pages that manage a container, such as MySQL, Redis or Cron Jobs. Other pages don't show it.

Hover over the service name to see live information about its container:

- Name
- CPU Usage
- RAM Usage
- Memory %
- Network I/O
- Block I/O
- PIDs

If the Services feature is allowed on your plan, a **Manage** link next to the badge opens the service's page. If only the Docker feature is allowed, **Manage** opens the service on the [Containers](/docs/panel/containers/) page instead.

![Service badge in the page header with its tooltip showing the container name and resource usage](/img/openpanel-screenshots/dashboard/dashboard-service.png#gh-light-mode-only)
![Service badge in the page header with its tooltip showing the container name and resource usage](/img/openpanel-screenshots/dashboard/dashboard-service_dark.png#gh-dark-mode-only)

## Search

Click the magnifying glass icon in the top-right corner of the header to open the search box. It searches across pages/features, files and folders, websites, domains, email accounts, FTP accounts, containers, services, cron jobs, and MySQL/PostgreSQL databases and users. Results are limited to what's enabled on your plan — some entity types (databases, domains, emails, FTP, containers, services, websites, cron jobs) additionally require an Enterprise license.

![Search box opened from the magnifying glass icon in the header, with results for mysql](/img/openpanel-screenshots/dashboard/dashboard-search.png#gh-light-mode-only)
![Search box opened from the magnifying glass icon in the header, with results for mysql](/img/openpanel-screenshots/dashboard/dashboard-search_dark.png#gh-dark-mode-only)

## Getting Started Tour

On first login, a short guided tour highlights the main areas of the interface. You can skip it at any time; it won't reappear once dismissed.

The tour can be started from the last step of the [Onboarding](/docs/panel/dashboard/onboarding/) wizard.
